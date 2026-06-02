package domain

import "time"

// APIToken is a panel-to-panel credential. TokenHash is stored instead of the
// raw secret for locally issued tokens; raw tokens are only shown once.
type APIToken struct {
	ID          int64
	Name        string
	TokenHash   string
	TokenPrefix string
	Scopes      string
	Enabled     bool
	LastUsedAt  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
