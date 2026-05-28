package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sing-box-web-panel/internal/config"
	"sing-box-web-panel/internal/lib/sl"
	"sing-box-web-panel/internal/repo/sqlite"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("starting server", slog.String("env", cfg.Env))

	storage, err := sqlite.New(cfg.Database, log)
	if err != nil {
		log.Error("failed to connect to database", sl.Error(err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	_ = storage

	<-ctx.Done()

	log.Info("shutting down")
	shutdown(storage, log, cfg)
	log.Info("stopped")
}

func shutdown(storage interface{ Close() error }, log *slog.Logger, cfg *config.Config) {
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
