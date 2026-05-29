package main

//	@title			SingGrok API
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
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"sing-box-web-panel/docs"
	"sing-box-web-panel/internal/config"
	libauth "sing-box-web-panel/internal/lib/auth"
	"sing-box-web-panel/internal/lib/sl"
	sqliterepo "sing-box-web-panel/internal/repo/sqlite"
	"sing-box-web-panel/internal/services/auth"
	"sing-box-web-panel/internal/transport/handler"
	"sing-box-web-panel/internal/transport/middleware"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

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
	totpMgr := libauth.NewTOTPManager("SingGrok")

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

	mux := http.NewServeMux()

	docs.SwaggerInfo.BasePath = "/api"
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	authHandler := handler.NewAuthHandler(authSvc, log)
	authHandler.Register(mux)

	healthHandler := handler.NewHealthHandler()
	healthHandler.Register(mux)

	corsOrigins := []string{"http://localhost:3000", "http://127.0.0.1:3000"}

	stack := middleware.Auth(jwtMgr)(mux)
	stack = middleware.CORS(corsOrigins)(stack)
	stack = middleware.Logger(log)(stack)

	server := &http.Server{
		Addr:           cfg.HTTP.Address,
		Handler:        stack,
		ReadTimeout:    cfg.HTTP.ReadTimeout,
		WriteTimeout:   cfg.HTTP.WriteTimeout,
		IdleTimeout:    cfg.HTTP.IdleTimeout,
		MaxHeaderBytes: cfg.HTTP.MaxHeaderBytes,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("http server listening", slog.String("addr", cfg.HTTP.Address))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server error", sl.Error(err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown error", sl.Error(err))
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

func setupLogger(env string) *slog.Logger {
	switch env {
	case "dev", "local":
		return sl.SetupPrettySlog()
	default:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
