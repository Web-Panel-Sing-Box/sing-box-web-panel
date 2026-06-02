package main

//	@title			Shilka API
//	@version		0.1.0
//	@description	Local-first web panel for managing a sing-box process.
//	@host			localhost:8080
//	@BasePath		/api

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter "Bearer " followed by the JWT token.

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"

	"sing-box-web-panel/docs"
	"sing-box-web-panel/internal/config"
	libauth "sing-box-web-panel/internal/lib/auth"
	"sing-box-web-panel/internal/lib/sl"
	sqliterepo "sing-box-web-panel/internal/repo/sqlite"
	"sing-box-web-panel/internal/services/auth"
	svcclient "sing-box-web-panel/internal/services/client"
	svcinbound "sing-box-web-panel/internal/services/inbound"
	"sing-box-web-panel/internal/services/logbuf"
	"sing-box-web-panel/internal/services/singbox"
	"sing-box-web-panel/internal/services/stats"
	"sing-box-web-panel/internal/services/sysstat"
	"sing-box-web-panel/internal/services/tlsmgr"
	"sing-box-web-panel/internal/transport/handler"
	"sing-box-web-panel/internal/transport/middleware"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	cfg := config.MustLoad()
	logBuf := logbuf.New(cfg.Logging.MaxMemoryLines)
	log := setupLogger(cfg.Env, logBuf)

	log.Info("starting server", slog.String("env", cfg.Env))

	storage, err := sqliterepo.New(cfg.Database, log)
	if err != nil {
		log.Error("failed to connect to database", sl.Error(err))
		os.Exit(1)
	}

	hasher := libauth.NewArgon2Hasher(
		cfg.Auth.Argon2MemoryKB,
		cfg.Auth.Argon2Iterations,
		cfg.Auth.Argon2Parallelism,
	)
	jwtMgr := libauth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry)
	totpMgr := libauth.NewTOTPManager("Shilka")

	adminRepo := sqliterepo.NewAdminRepo(storage)
	recoveryRepo := sqliterepo.NewRecoveryRepo(storage)

	totpAdapter := auth.NewTOTPAdapter(totpMgr)

	authSvc := auth.NewService(
		adminRepo,
		recoveryRepo,
		hasher,
		jwtMgr,
		totpAdapter,
		libauth.GenerateRecoveryCode,
	)

	if err := authSvc.SeedAdmin(context.Background(), cfg.Auth.AdminUser, cfg.Auth.AdminPassword); err != nil {
		log.Error("bootstrap admin", sl.Error(err))
		os.Exit(1)
	}

	debug.SetMemoryLimit(mustBytes(cfg.Runtime.GoMemLimit))
	debug.SetGCPercent(cfg.Runtime.GoGC)

	mux := http.NewServeMux()

	if cfg.Env == "dev" || cfg.Env == "local" {
		docs.SwaggerInfo.BasePath = "/api"
		mux.Handle("GET /swagger/", httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
		))
	}

	authHandler := handler.NewAuthHandler(authSvc, log)
	authHandler.Register(mux)

	healthHandler := handler.NewHealthHandler()
	healthHandler.Register(mux)

	// Background worker context: cancelled during shutdown to stop the apply loop.
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()

	inboundRepo := sqliterepo.NewInboundRepo(storage)
	clientRepo := sqliterepo.NewClientRepo(storage)
	configRevRepo := sqliterepo.NewConfigRevisionRepo(storage)

	// Resolve sing-box paths to absolute so the subprocess working dir does not
	// double-apply to a relative config path.
	absConfigPath := absPath(cfg.SingBox.ConfigPath)
	absWorkingDir := absPath(cfg.SingBox.WorkingDir)

	absCoreLogPath := ""
	if cfg.SingBox.CoreLogPath != "" {
		absCoreLogPath = absPath(cfg.SingBox.CoreLogPath)
	}

	// sing-box config generation + lifecycle.
	generator := singbox.NewGenerator(inboundRepo, clientRepo, singbox.GeneratorConfig{
		LogLevel:        "info",
		InboundListen:   "::",
		ClashAPIAddress: cfg.SingBox.APIAddress,
		ClashAPISecret:  cfg.SingBox.APISecret,
		CacheFilePath:   filepath.Join(absWorkingDir, "cache.db"),
		StatsSource:     cfg.Stats.Source,
		V2RayAPIListen:  cfg.Stats.V2RayAPIAddress,
		CoreLogPath:     absCoreLogPath,
	})
	checker := singbox.NewChecker(cfg.SingBox.BinaryPath, cfg.SingBox.CheckTimeout)
	processMgr := singbox.NewProcessManager(singbox.ProcessConfig{
		Mode:         cfg.SingBox.ProcessMode,
		Binary:       cfg.SingBox.BinaryPath,
		ConfigPath:   absConfigPath,
		WorkingDir:   absWorkingDir,
		ServiceName:  cfg.SingBox.ServiceName,
		RestartDelay: cfg.SingBox.RestartDelay,
	}, logBuf.Writer(), log)
	applier := singbox.NewApplier(generator, checker, processMgr, configRevRepo, absConfigPath, log)
	go applier.Run(rootCtx)

	log.Debug("starting sing-box core")
	// Bootstrap the initial config and start the core automatically.
	if err := bootCore(context.Background(), applier, processMgr, log); err != nil {
		log.Warn("boot core", sl.Error(err))
	}

	// Inbound and client management; the applier is the (debounced) ConfigTrigger.
	inboundSvc := svcinbound.NewService(inboundRepo, clientRepo, applier)
	clientSvc := svcclient.NewService(clientRepo, inboundRepo, applier)

	settingRepo := sqliterepo.NewSettingRepo(storage)
	trafficRepo := sqliterepo.NewTrafficRepo(storage)

	// Traffic stats + quota worker. Clash REST is the live dashboard source; the
	// per-user source (V2Ray gRPC) stays nil unless explicitly enabled.
	liveHolder := &stats.LiveHolder{}
	clashSource := stats.NewClashSource(cfg.SingBox.APIAddress, cfg.SingBox.APISecret)
	// Per-user accounting source: V2Ray gRPC stats, only when explicitly enabled
	// (requires a with_v2ray_api binary). Otherwise quota is enforced by expiry.
	var userSource stats.UserSource
	if cfg.Stats.Source == "v2ray" {
		userSource = stats.NewV2RaySource(cfg.Stats.V2RayAPIAddress)
		log.Info("per-user stats via v2ray api", slog.String("addr", cfg.Stats.V2RayAPIAddress))
	}
	statsWorker := stats.NewWorker(clashSource, userSource, clientRepo, trafficRepo, applier, liveHolder, stats.WorkerConfig{
		SampleInterval: cfg.Metrics.TrafficInterval,
		FlushInterval:  cfg.Metrics.BatchFlushInterval,
	}, log)
	statsWorker.Run(rootCtx)

	handler.NewInboundHandler(inboundSvc, log).Register(mux)
	handler.NewClientHandler(clientSvc, cfg.Sub.PublicURL, log).Register(mux)
	handler.NewCoreHandler(processMgr, applier, log, absCoreLogPath).Register(mux)
	handler.NewSubscriptionHandler(clientRepo, inboundRepo, settingRepo, cfg.Sub.PublicURL, "", log).Register(mux)
	handler.NewDashboardHandler(sysstat.New(), liveHolder, clientRepo, inboundRepo, trafficRepo, processMgr, log).Register(mux)
	handler.NewLogsHandler(logBuf).Register(mux)

	frontendHandler := handler.NewFrontendHandler(cfg.Frontend.ServeMode, cfg.Frontend.DiskPath, cfg.Frontend.CacheTTL, frontendDist)
	mux.Handle("/", frontendHandler)

	corsOrigins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}

	stack := middleware.Auth(jwtMgr, log)(mux)
	stack = middleware.CORS(corsOrigins)(stack)
	// Stricter brute-force limit on login, then a general per-IP API limit.
	stack = middleware.RateLimit(cfg.Auth.LoginRateLimit, middleware.LoginPathMatcher, log)(stack)
	stack = middleware.RateLimit(cfg.Auth.APIRateLimit, middleware.APIPathMatcher, log)(stack)
	stack = middleware.Logger(log)(stack)

	server := &http.Server{
		Addr:           cfg.HTTP.Address,
		Handler:        stack,
		ReadTimeout:    cfg.HTTP.ReadTimeout,
		WriteTimeout:   cfg.HTTP.WriteTimeout,
		IdleTimeout:    cfg.HTTP.IdleTimeout,
		MaxHeaderBytes: cfg.HTTP.MaxHeaderBytes,
	}

	tlsMgr := tlsmgr.New(tlsmgr.Config{
		Mode:            cfg.TLS.Mode,
		CertFile:        cfg.TLS.CertFile,
		KeyFile:         cfg.TLS.KeyFile,
		ACMEEmail:       cfg.TLS.ACMEEmail,
		ACMEDomains:     cfg.TLS.ACMEDomains,
		ACMECacheDir:    cfg.TLS.ACMECacheDir,
		SelfSignedHosts: cfg.TLS.SelfSignedHosts,
		SelfSignedDir:   cfg.TLS.SelfSignedDir,
	})
	tlsConf, err := tlsMgr.TLSConfig()
	if err != nil {
		log.Error("tls setup", sl.Error(err))
		os.Exit(1)
	}
	server.TLSConfig = tlsConf

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		scheme := "http"
		if tlsMgr.Enabled() {
			scheme = "https"
		}
		log.Info("server listening", slog.String("addr", cfg.HTTP.Address), slog.String("scheme", scheme))

		var serveErr error
		if tlsMgr.Enabled() {
			serveErr = server.ListenAndServeTLS("", "") // certs come from TLSConfig
		} else {
			serveErr = server.ListenAndServe()
		}
		if serveErr != nil && serveErr != http.ErrServerClosed {
			log.Error("http server error", sl.Error(serveErr))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	rootCancel() // stop background workers (apply loop, etc.)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown error", sl.Error(err))
	}

	if err := processMgr.Stop(context.Background()); err != nil {
		log.Warn("stop core", sl.Error(err))
	}

	shutdown(storage, log, cfg)
	log.Info("stopped")
}

func shutdown(storage interface{ Close() error }, log *slog.Logger, _ *config.Config) {
	if err := storage.Close(); err != nil {
		log.Error("failed to close database", sl.Error(err))
	} else {
		log.Info("database connection closed")
	}
}

func bootCore(ctx context.Context, applier *singbox.Applier, pm singbox.ProcessManager, log *slog.Logger) error {
	if err := applier.ApplyIfMissing(ctx); err != nil {
		return fmt.Errorf("apply initial config: %w", err)
	}
	if err := pm.Start(ctx); err != nil {
		return fmt.Errorf("start core: %w", err)
	}
	log.Info("core started")
	return nil
}

// absPath resolves p to an absolute path, falling back to p on error.
func absPath(p string) string {
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

func setupLogger(env string, buf *logbuf.Buffer) *slog.Logger {
	var base slog.Handler
	switch env {
	case "dev", "local":
		base = sl.SetupPrettySlog().Handler()
	default:
		base = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	return slog.New(logbuf.NewTeeHandler(base, buf))
}

func mustBytes(s string) int64 {
	if s == "" {
		return 0
	}
	var n float64
	unit := "B"
	if _, err := fmt.Sscanf(s, "%f%s", &n, &unit); err != nil {
		return 0
	}
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "B":
		return int64(n)
	case "KB", "KIB":
		return int64(n * 1024)
	case "MB", "MIB":
		return int64(n * 1024 * 1024)
	case "GB", "GIB":
		return int64(n * 1024 * 1024 * 1024)
	default:
		return int64(n)
	}
}
