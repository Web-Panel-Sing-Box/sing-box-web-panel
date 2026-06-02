package node

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"sing-box-web-panel/internal/domain"
)

type Repo interface {
	Create(ctx context.Context, n *domain.Node) error
	GetByID(ctx context.Context, id int64) (*domain.Node, error)
	List(ctx context.Context) ([]domain.Node, error)
	ListEnabled(ctx context.Context) ([]domain.Node, error)
	Update(ctx context.Context, n *domain.Node) error
	Delete(ctx context.Context, id int64) error
	SetEnabled(ctx context.Context, id int64, enabled bool) error
	SetStatus(ctx context.Context, id int64, status domain.NodeStatus, heartbeatAt *time.Time, latencyMS int64, panelVersion, coreVersion string, cpuPct, ramPct float64, uptimeSeconds int64, lastErr string) error
}

type InboundCache interface {
	UpsertRemote(ctx context.Context, nodeID int64, remoteID string, ib *domain.Inbound) (int64, error)
}

type ClientCache interface {
	UpsertRemote(ctx context.Context, nodeID int64, remoteID string, inboundID int64, c *domain.Client) error
}

type Service struct {
	repo     Repo
	inbounds InboundCache
	clients  ClientCache
	remote   RemoteClienter
	now      func() time.Time
}

func NewService(repo Repo, inbounds InboundCache, clients ClientCache, remote RemoteClienter) *Service {
	return &Service{repo: repo, inbounds: inbounds, clients: clients, remote: remote, now: time.Now}
}

func (s *Service) List(ctx context.Context) ([]domain.Node, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id int64) (*domain.Node, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, in Input) (*domain.Node, error) {
	n, err := nodeFromInput(in)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Service) Update(ctx context.Context, id int64, in Input) (*domain.Node, error) {
	n, err := nodeFromInput(in)
	if err != nil {
		return nil, err
	}
	n.ID = id
	if err := s.repo.Update(ctx, n); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Toggle(ctx context.Context, id int64) (*domain.Node, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	n.Enabled = !n.Enabled
	if err := s.repo.SetEnabled(ctx, id, n.Enabled); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Service) Probe(ctx context.Context, id int64) (*domain.Node, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	status, latency, err := s.remote.Status(ctx, n)
	now := s.now().UTC()
	if err != nil {
		_ = s.repo.SetStatus(ctx, id, domain.NodeStatusOffline, nil, 0, "", "", 0, 0, 0, err.Error())
		n.Status = domain.NodeStatusOffline
		n.LastError = err.Error()
		return n, nil
	}
	n.Status = domain.NodeStatusOnline
	n.LastHeartbeatAt = &now
	n.LatencyMS = latency.Milliseconds()
	n.PanelVersion = status.PanelVersion
	n.CoreVersion = status.CoreVersion
	n.CPUPct = status.CPUPct
	n.RAMPct = status.RAMPct
	n.UptimeSeconds = status.UptimeSeconds
	n.LastError = ""
	if err := s.repo.SetStatus(ctx, id, n.Status, n.LastHeartbeatAt, n.LatencyMS, n.PanelVersion, n.CoreVersion, n.CPUPct, n.RAMPct, n.UptimeSeconds, ""); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Service) Sync(ctx context.Context, id int64) (*SyncResult, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	snap, err := s.remote.Snapshot(ctx, n)
	if err != nil {
		now := s.now().UTC()
		_ = s.repo.SetStatus(ctx, id, domain.NodeStatusOffline, &now, 0, "", "", 0, 0, 0, err.Error())
		return nil, err
	}

	inboundIDs := make(map[string]int64, len(snap.Inbounds))
	for _, rib := range snap.Inbounds {
		ib := &domain.Inbound{
			NodeID:       &id,
			RemoteID:     rib.ID,
			Remark:       rib.Remark,
			Protocol:     rib.Protocol,
			Port:         rib.Port,
			Transmission: rib.Transmission,
			TLS:          rib.TLS,
			SNI:          rib.SNI,
			Dest:         rib.Dest,
			Enabled:      rib.Enabled,
			Settings:     rib.Settings,
		}
		localID, err := s.inbounds.UpsertRemote(ctx, id, rib.ID, ib)
		if err != nil {
			return nil, err
		}
		inboundIDs[rib.ID] = localID
	}
	for _, rc := range snap.Clients {
		inboundID, ok := inboundIDs[rc.InboundID]
		if !ok {
			continue
		}
		expiry, err := parseTimePtr(rc.Expiry)
		if err != nil {
			return nil, err
		}
		firstUsedAt, err := parseTimePtr(rc.FirstUsedAt)
		if err != nil {
			return nil, err
		}
		c := &domain.Client{
			NodeID:             &id,
			RemoteID:           rc.ID,
			InboundID:          inboundID,
			Name:               rc.Name,
			UUID:               rc.UUID,
			Password:           rc.Password,
			UsedUp:             rc.UsedUp,
			UsedDown:           rc.UsedDown,
			TotalQuota:         rc.TotalQuota,
			Expiry:             expiry,
			Status:             rc.Status,
			SubToken:           rc.SubToken,
			StartAfterFirstUse: rc.StartAfterFirstUse,
			Enabled:            rc.Enabled,
			FirstUsedAt:        firstUsedAt,
		}
		if err := s.clients.UpsertRemote(ctx, id, rc.ID, inboundID, c); err != nil {
			return nil, err
		}
	}
	now := s.now().UTC()
	_ = s.repo.SetStatus(ctx, id, domain.NodeStatusOnline, &now, 0, snap.Status.PanelVersion, snap.Status.CoreVersion, snap.Status.CPUPct, snap.Status.RAMPct, snap.Status.UptimeSeconds, "")
	return &SyncResult{NodeID: id, InboundCount: len(snap.Inbounds), ClientCount: len(snap.Clients), SyncedAt: now}, nil
}

func (s *Service) Run(ctx context.Context, heartbeatInterval, syncInterval time.Duration) {
	if heartbeatInterval <= 0 {
		heartbeatInterval = 10 * time.Second
	}
	if syncInterval <= 0 {
		syncInterval = 15 * time.Second
	}
	heartbeat := time.NewTicker(heartbeatInterval)
	syncTick := time.NewTicker(syncInterval)
	defer heartbeat.Stop()
	defer syncTick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			s.probeEnabled(ctx)
		case <-syncTick.C:
			s.syncEnabled(ctx)
		}
	}
}

func (s *Service) probeEnabled(ctx context.Context) {
	nodes, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 16)
	for i := range nodes {
		id := nodes[i].ID
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
			_, _ = s.Probe(ctx, id)
		}()
	}
	wg.Wait()
}

func (s *Service) syncEnabled(ctx context.Context) {
	nodes, err := s.repo.ListEnabled(ctx)
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i := range nodes {
		id := nodes[i].ID
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
			_, _ = s.Sync(ctx, id)
		}()
	}
	wg.Wait()
}

func nodeFromInput(in Input) (*domain.Node, error) {
	if strings.TrimSpace(in.Name) == "" {
		return nil, fmt.Errorf("name is required")
	}
	scheme := strings.ToLower(strings.TrimSpace(in.Scheme))
	if scheme == "" {
		scheme = "https"
	}
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("scheme must be http or https")
	}
	if strings.TrimSpace(in.Address) == "" {
		return nil, fmt.Errorf("address is required")
	}
	if in.Port < 1 || in.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1 and 65535")
	}
	return &domain.Node{
		Name:                strings.TrimSpace(in.Name),
		Remark:              strings.TrimSpace(in.Remark),
		Scheme:              scheme,
		Address:             strings.TrimSpace(in.Address),
		Port:                in.Port,
		BasePath:            strings.TrimSpace(in.BasePath),
		APITokenSecret:      strings.TrimSpace(in.APITokenSecret),
		Enabled:             in.Enabled,
		AllowPrivateAddress: in.AllowPrivateAddress,
		Status:              domain.NodeStatusUnknown,
	}, nil
}

func parseTimePtr(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func ParseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
