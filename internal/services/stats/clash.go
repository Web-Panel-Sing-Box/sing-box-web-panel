package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ClashSource reads aggregate metrics from the sing-box Clash API (REST). It
// derives throughput from the delta of cumulative totals between samples, which
// avoids the streaming /traffic websocket.
type ClashSource struct {
	baseURL string
	secret  string
	client  *http.Client

	mu        sync.Mutex
	prevUp    int64
	prevDown  int64
	prevAt    time.Time
	havePrev  bool
}

// NewClashSource builds a source for a Clash external_controller address such as
// "127.0.0.1:9090".
func NewClashSource(apiAddress, secret string) *ClashSource {
	return &ClashSource{
		baseURL: "http://" + apiAddress,
		secret:  secret,
		client:  &http.Client{Timeout: 4 * time.Second},
	}
}

type clashConnections struct {
	DownloadTotal int64 `json:"downloadTotal"`
	UploadTotal   int64 `json:"uploadTotal"`
	Connections   []struct {
		ID string `json:"id"`
	} `json:"connections"`
}

// Sample fetches /connections and computes throughput from the totals delta.
func (s *ClashSource) Sample(ctx context.Context) (Live, error) {
	var data clashConnections
	if err := s.get(ctx, "/connections", &data); err != nil {
		return Live{}, err
	}

	now := time.Now()
	live := Live{
		Connections:   len(data.Connections),
		UploadTotal:   data.UploadTotal,
		DownloadTotal: data.DownloadTotal,
	}

	s.mu.Lock()
	if s.havePrev {
		if dt := now.Sub(s.prevAt).Seconds(); dt > 0 {
			if up := data.UploadTotal - s.prevUp; up > 0 {
				live.UploadBps = int64(float64(up) / dt)
			}
			if down := data.DownloadTotal - s.prevDown; down > 0 {
				live.DownloadBps = int64(float64(down) / dt)
			}
		}
	}
	s.prevUp = data.UploadTotal
	s.prevDown = data.DownloadTotal
	s.prevAt = now
	s.havePrev = true
	s.mu.Unlock()

	return live, nil
}

func (s *ClashSource) get(ctx context.Context, path string, dst any) error {
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
