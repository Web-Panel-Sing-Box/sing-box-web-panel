package node

import (
	"time"

	"sing-box-web-panel/internal/domain"
)

type Input struct {
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
}

type RemoteStatus struct {
	PanelVersion  string  `json:"panelVersion"`
	CoreVersion   string  `json:"coreVersion"`
	CoreStatus    string  `json:"coreStatus"`
	CPUPct        float64 `json:"cpuPct"`
	RAMPct        float64 `json:"ramPct"`
	UptimeSeconds int64   `json:"uptimeSeconds"`
	InboundCount  int     `json:"inboundCount"`
	ClientCount   int     `json:"clientCount"`
	OnlineCount   int     `json:"onlineCount"`
	DepletedCount int     `json:"depletedCount"`
}

type RemoteInbound struct {
	ID           string                 `json:"id"`
	Remark       string                 `json:"remark"`
	Protocol     domain.Protocol        `json:"protocol"`
	Port         int                    `json:"port"`
	Transmission domain.Transmission    `json:"transmission"`
	TLS          domain.TLSMode         `json:"tls"`
	SNI          string                 `json:"sni,omitempty"`
	Dest         string                 `json:"dest,omitempty"`
	Enabled      bool                   `json:"enabled"`
	Settings     domain.InboundSettings `json:"settings,omitempty"`
	UpdatedAt    string                 `json:"updatedAt,omitempty"`
}

type RemoteClient struct {
	ID                 string              `json:"id"`
	InboundID          string              `json:"inboundId"`
	Name               string              `json:"name"`
	UUID               string              `json:"uuid"`
	Password           string              `json:"password"`
	UsedUp             int64               `json:"usedUp"`
	UsedDown           int64               `json:"usedDown"`
	TotalQuota         int64               `json:"totalQuota"`
	Expiry             string              `json:"expiry,omitempty"`
	Status             domain.ClientStatus `json:"status"`
	SubToken           string              `json:"subToken"`
	StartAfterFirstUse bool                `json:"startAfterFirstUse"`
	Enabled            bool                `json:"enabled"`
	FirstUsedAt        string              `json:"firstUsedAt,omitempty"`
}

type RemoteSnapshot struct {
	Status   RemoteStatus    `json:"status"`
	Inbounds []RemoteInbound `json:"inbounds"`
	Clients  []RemoteClient  `json:"clients"`
}

type SyncResult struct {
	NodeID       int64
	InboundCount int
	ClientCount  int
	SyncedAt     time.Time
}
