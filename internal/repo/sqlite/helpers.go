package sqlite

import "strings"

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
