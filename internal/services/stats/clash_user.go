package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/services/singbox"
)

// ClashUserSource attributes per-connection traffic from sing-box's Clash API
// to clients by reading `chains[0]` on each /connections entry. The generator
// pins every client to a unique `user-{id}` direct outbound (see
// singbox.ClientOutboundTag), so the tag the Clash API reports is a stable,
// log-free client identifier.
//
// The same instance implements UserHeartbeat: any client whose tag appears in
// the most recent snapshot is "online now", and the worker mirrors that into
// last_used_at.
type ClashUserSource struct {
	baseURL   string
	secret    string
	client    *http.Client
	nameByID  func(int64) string

	mu        sync.Mutex
	prevConn  map[string]connBytes
	lastSeen  map[string]time.Time
}

type connBytes struct {
	up, down int64
}

// NewClashUserSource builds a source against the same external_controller
// address used by ClashSource. nameByID resolves a client ID to its display
// name so UserTraffic / UserSeen stay keyed by name (matching the worker's
// existing byName lookup).
func NewClashUserSource(apiAddress, secret string, nameByID func(int64) string) *ClashUserSource {
	return &ClashUserSource{
		baseURL:  "http://" + apiAddress,
		secret:   secret,
		client:   &http.Client{Timeout: 4 * time.Second},
		nameByID: nameByID,
	}
}

type clashUserConn struct {
	ID       string   `json:"id"`
	Upload   int64    `json:"upload"`
	Download int64    `json:"download"`
	Chains   []string `json:"chains"`
}

type clashUserPayload struct {
	Connections []clashUserConn `json:"connections"`
}

// UserDeltas polls /connections, computes per-conn byte deltas, and aggregates
// them per client. Connections whose chain doesn't match the panel's per-user
// tag (e.g. panel UI egress, bittorrent reject) are skipped. Disappeared
// connections lose at most one polling interval of tail bytes.
func (s *ClashUserSource) UserDeltas(ctx context.Context) ([]domain.UserTraffic, error) {
	var data clashUserPayload
	if err := s.fetch(ctx, "/connections", &data); err != nil {
		return nil, err
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.prevConn == nil {
		s.prevConn = make(map[string]connBytes)
	}

	next := make(map[string]connBytes, len(data.Connections))
	byID := make(map[int64]*domain.UserTraffic)
	seen := make(map[string]time.Time, len(data.Connections))

	for _, c := range data.Connections {
		if len(c.Chains) == 0 {
			continue
		}
		id, ok := singbox.ParseClientOutboundTag(c.Chains[0])
		if !ok {
			continue
		}
		name := s.nameByID(id)
		if name == "" {
			// Client was deleted; skip silently.
			continue
		}

		next[c.ID] = connBytes{up: c.Upload, down: c.Download}
		seen[name] = now

		prev, hadPrev := s.prevConn[c.ID]
		dUp, dDown := c.Upload, c.Download
		if hadPrev {
			// Backwards counter = id reused for a fresh conn; charge the full
			// current value.
			if c.Upload >= prev.up {
				dUp = c.Upload - prev.up
			}
			if c.Download >= prev.down {
				dDown = c.Download - prev.down
			}
		}
		if dUp == 0 && dDown == 0 {
			continue
		}
		ut := byID[id]
		if ut == nil {
			ut = &domain.UserTraffic{Name: name}
			byID[id] = ut
		}
		ut.Up += dUp
		ut.Down += dDown
	}
	s.prevConn = next
	s.lastSeen = seen

	out := make([]domain.UserTraffic, 0, len(byID))
	for _, ut := range byID {
		out = append(out, *ut)
	}
	return out, nil
}

// UserSeen returns the set of clients whose per-user outbound was present in
// the most recent /connections snapshot, each mapped to the snapshot time.
// Worker mirrors that into last_used_at so the heartbeat freezes the moment a
// client disappears from /connections.
func (s *ClashUserSource) UserSeen() map[string]time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.lastSeen) == 0 {
		return nil
	}
	out := make(map[string]time.Time, len(s.lastSeen))
	for name, at := range s.lastSeen {
		out[name] = at
	}
	return out
}

func (s *ClashUserSource) fetch(ctx context.Context, path string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+path, nil)
	if err != nil {
		return err
	}
	if s.secret != "" {
		req.Header.Set("Authorization", "Bearer "+s.secret)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("clash api %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("clash api %s: status %d", path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode clash %s: %w", path, err)
	}
	return nil
}
