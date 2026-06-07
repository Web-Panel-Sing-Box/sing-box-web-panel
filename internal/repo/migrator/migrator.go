package migrator

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const migrationTable = "schema_migrations"

type migration struct {
	version int
	name    string
	sql     string
}

func loadMigrations() ([]migration, error) {
	return loadMigrationsFS(migrationsFS, "migrations")
}

func loadMigrationsFS(fsys fs.FS, dir string) ([]migration, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var migrations []migration
	seen := make(map[int]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		name := entry.Name()
		parts := strings.SplitN(name, "_", 2)
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("parse migration version from %q: %w", name, err)
		}

		// Fail fast on duplicate versions. Two migrations sharing a version
		// collide on schema_migrations.version (INTEGER PRIMARY KEY): the
		// second INSERT hits a UNIQUE constraint and its whole transaction —
		// including any schema change — is rolled back, leaving a silently
		// broken schema. Refusing to start is safer than a partial migration.
		if prev, ok := seen[version]; ok {
			return nil, fmt.Errorf("duplicate migration version %d: %q and %q", version, prev, name)
		}
		seen[version] = name

		data, err := fs.ReadFile(fsys, dir+"/"+name)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", name, err)
		}

		migrations = append(migrations, migration{
			version: version,
			name:    name,
			sql:     string(data),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func Migrate(ctx context.Context, db *sql.DB, log *slog.Logger) error {
	if _, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS `+migrationTable+` (
			version    INTEGER PRIMARY KEY,
			name       TEXT    NOT NULL,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
	); err != nil {
		return fmt.Errorf("create %s: %w", migrationTable, err)
	}

	applied := make(map[int]bool)
	rows, err := db.QueryContext(ctx, `SELECT version FROM `+migrationTable)
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return fmt.Errorf("scan migration version: %w", err)
		}
		applied[v] = true
	}
	rows.Close()

	log.Debug("loaded applied migrations from database",
		slog.Int("count", len(applied)),
	)

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	log.Debug("loaded embedded migrations",
		slog.Int("count", len(migrations)),
	)

	if len(migrations) == 0 {
		log.Debug("no migrations found")
		return nil
	}

	pending := 0
	for _, m := range migrations {
		if !applied[m.version] {
			pending++
		}
	}

	if pending == 0 {
		log.Debug("all migrations already applied", slog.Int("total", len(migrations)))
		return nil
	}

	log.Info("applying migrations", slog.Int("pending", pending), slog.Int("total", len(migrations)))

	for _, m := range migrations {
		if applied[m.version] {
			log.Debug("migration already applied, skipping",
				slog.Int("version", m.version),
				slog.String("name", m.name),
			)
			continue
		}

		log.Debug("applying migration",
			slog.Int("version", m.version),
			slog.String("name", m.name),
		)

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx for migration %d: %w", m.version, err)
		}

		if _, err := tx.ExecContext(ctx, strings.TrimSpace(m.sql)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply migration %d (%s): %w", m.version, m.name, err)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO `+migrationTable+` (version, name) VALUES (?, ?)`,
			m.version, m.name,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}

		log.Info("migration applied",
			slog.Int("version", m.version),
			slog.String("name", m.name),
		)
	}

	log.Info("migrations complete", slog.Int("applied", pending))
	return nil
}
