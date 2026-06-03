package domain

import "time"

type NodeStatus string

const (
	NodeStatusUnknown NodeStatus = "unknown"
	NodeStatusOnline  NodeStatus = "online"
	NodeStatusOffline NodeStatus = "offline"
)

// Node is another Shilka panel managed by this panel.
type Node struct {
	ID                  int64
	Name                string
	Remark              string
	Scheme              string
	Address             string
	Port                int
	BasePath            string
	APITokenSecret      string
	Enabled             bool
	AllowPrivateAddress bool
	SkipTLSVerify       bool
	Status              NodeStatus
	LastHeartbeatAt     *time.Time
	LatencyMS           int64
	PanelVersion        string
	CoreVersion         string
	CPUPct              float64
	RAMPct              float64
	UptimeSeconds       int64
	LastError           string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
