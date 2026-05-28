package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"sing-box-web-panel/internal/config"
	"sing-box-web-panel/internal/repo/migrator"

	_ "modernc.org/sqlite"
)

func New(cfg config.DBConfig, log *slog.Logger) (*sql.DB, error) {
	log.Debug("opening database", slog.String("path", cfg.Path))

	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := applyPragmas(db, cfg); err != nil {
		db.Close()
		return nil, err
	}

	if err := migrator.Migrate(context.Background(), db, log); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

func applyPragmas(db *sql.DB, cfg config.DBConfig) error {
	pragmas := []struct {
		name  string
		value string
	}{
		{"journal_mode", cfg.JournalMode},
		{"synchronous", cfg.Synchronous},
		{"busy_timeout", fmt.Sprintf("%d", cfg.BusyTimeoutMS)},
		{"temp_store", cfg.TempStore},
	}

	if cfg.ForeignKeys {
		pragmas = append(pragmas, struct {
			name  string
			value string
		}{"foreign_keys", "ON"})
	}

	if cfg.MmapSizeMB > 0 {
		pragmas = append(pragmas, struct {
			name  string
			value string
		}{"mmap_size", fmt.Sprintf("%d", cfg.MmapSizeMB*1024*1024)})
	}

	cacheKb := -cfg.CacheSizeKB
	if cacheKb > 0 {
		cacheKb = -cacheKb
	}
	pragmas = append(pragmas, struct {
		name  string
		value string
	}{"cache_size", fmt.Sprintf("%d", cacheKb)})

	for _, p := range pragmas {
		query := fmt.Sprintf("PRAGMA %s = %s", p.name, p.value)
		if _, err := db.Exec(query); err != nil && !strings.Contains(err.Error(), "not available") {
			return fmt.Errorf("pragma %s=%s: %w", p.name, p.value, err)
		}
	}

	return nil
}
