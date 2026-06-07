package migrator

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMigrationsHaveUniqueVersions(t *testing.T) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	seen := make(map[int]string)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		name := entry.Name()
		parts := strings.SplitN(name, "_", 2)
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			t.Errorf("parse migration version from %q: %v", name, err)
			continue
		}

		if prev, ok := seen[version]; ok {
			t.Errorf("duplicate migration version %d: %q and %q", version, prev, name)
			continue
		}
		seen[version] = name
	}
}

// TestLoadMigrationsRealFSLoads exercises the production loader over the real
// embedded migrations: it must succeed and return versions in ascending order.
func TestLoadMigrationsRealFSLoads(t *testing.T) {
	migrations, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("expected at least one embedded migration")
	}
	for i := 1; i < len(migrations); i++ {
		if migrations[i-1].version >= migrations[i].version {
			t.Fatalf("migrations not sorted ascending: %d before %d",
				migrations[i-1].version, migrations[i].version)
		}
	}
}

// TestLoadMigrationsFSRejectsDuplicateVersion is the runtime fail-fast guard:
// two files sharing a version must make the loader error before any migration
// is applied, instead of letting the second one collide on
// schema_migrations.version and silently roll back its schema change.
func TestLoadMigrationsFSRejectsDuplicateVersion(t *testing.T) {
	fsys := fstest.MapFS{
		"m/000010_add_client_last_used_at.sql":  {Data: []byte("ALTER TABLE clients ADD COLUMN last_used_at DATETIME;")},
		"m/000010_add_node_skip_tls_verify.sql": {Data: []byte("ALTER TABLE nodes ADD COLUMN skip_tls_verify BOOLEAN NOT NULL DEFAULT 0;")},
	}

	_, err := loadMigrationsFS(fsys, "m")
	if err == nil {
		t.Fatal("expected error for duplicate migration version, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate migration version 10") {
		t.Fatalf("expected duplicate-version error, got: %v", err)
	}
}

// TestLoadMigrationsFSSortsByVersion confirms unique versions load and sort
// regardless of directory (filename) order.
func TestLoadMigrationsFSSortsByVersion(t *testing.T) {
	fsys := fstest.MapFS{
		"m/000002_b.sql": {Data: []byte("SELECT 2;")},
		"m/000001_a.sql": {Data: []byte("SELECT 1;")},
		"m/000010_c.sql": {Data: []byte("SELECT 10;")},
	}

	migrations, err := loadMigrationsFS(fsys, "m")
	if err != nil {
		t.Fatalf("loadMigrationsFS: %v", err)
	}
	got := []int{}
	for _, m := range migrations {
		got = append(got, m.version)
	}
	want := []int{1, 2, 10}
	if len(got) != len(want) {
		t.Fatalf("got %v versions, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
