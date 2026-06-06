package migrator

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"
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
