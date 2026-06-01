package sqlite

import (
	"database/sql"
	"strings"
	"time"
)

// rowScanner is satisfied by both *sql.Row and *sql.Rows, letting scan helpers
// serve single-row and multi-row queries alike.
type rowScanner interface {
	Scan(dest ...any) error
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed") ||
		strings.Contains(err.Error(), "duplicate")
}

func nullableInt64(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func ptrFromNullInt64(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func ptrFromNullTime(v sql.NullTime) *time.Time {
	if !v.Valid {
		return nil
	}
	return &v.Time
}
