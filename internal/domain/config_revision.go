package domain

import "time"

// ConfigRevision records one attempt to render and apply a sing-box config,
// for audit and rollback. A failed revision (OK == false) carries the
// `sing-box check` error and means the live config was left untouched.
type ConfigRevision struct {
	ID        int64
	SHA256    string
	OK        bool
	Error     string
	AppliedAt time.Time
}
