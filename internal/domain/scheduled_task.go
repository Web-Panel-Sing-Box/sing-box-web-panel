package domain

import "time"

const (
	ActionResetTrafficAll    = "reset_traffic_all"
	ActionDeleteExpired      = "delete_expired_clients"
	ActionBackupDB           = "backup_database"
	ActionRotateRealityKeys  = "rotate_reality_keys"
)

type ScheduledTask struct {
	ID         int64
	Name       string
	CronExpr   string
	Action     string
	ParamsJSON string
	Enabled    bool
	LastRunAt  *time.Time
	NextRunAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
