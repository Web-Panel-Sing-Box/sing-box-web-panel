// Package stats collects traffic metrics from the sing-box core and enforces
// per-client quota and expiry.
//
// Two source kinds are defined because the sing-box Clash API does not expose
// per-user traffic attribution:
//   - LiveSource (Clash REST) provides aggregate live metrics for the dashboard.
//   - UserSource (V2Ray gRPC, opt-in) provides cumulative per-user counters for
//     quota enforcement. It is only available with a `with_v2ray_api` binary.
package stats

import (
	"context"
	"sync"
	"time"

	"sing-box-web-panel/internal/domain"
)

// Live is an aggregate snapshot used by the dashboard.
type Live struct {
	UploadBps     int64
	DownloadBps   int64
	Connections   int
	UploadTotal   int64 // cumulative bytes since core start
	DownloadTotal int64
}

// LiveSource yields aggregate live metrics.
type LiveSource interface {
	Sample(ctx context.Context) (Live, error)
}

// UserSource yields per-user traffic deltas since the previous call (the V2Ray
// stats query resets counters), keyed by client name.
type UserSource interface {
	UserDeltas(ctx context.Context) ([]domain.UserTraffic, error)
}

// Point is a single throughput sample for the dashboard traffic chart.
type Point struct {
	T    int64 // unix milliseconds
	Up   int64
	Down int64
}

const historySize = 60

// LiveHolder stores the most recent Live sample plus a short throughput history
// for the dashboard handler. It is written by the worker and read by handlers.
type LiveHolder struct {
	mu      sync.RWMutex
	live    Live
	history []Point
}

func (h *LiveHolder) Set(l Live) {
	now := time.Now().UnixMilli()
	h.mu.Lock()
	h.live = l
	h.history = append(h.history, Point{T: now, Up: l.UploadBps, Down: l.DownloadBps})
	if len(h.history) > historySize {
		h.history = h.history[len(h.history)-historySize:]
	}
	h.mu.Unlock()
}

func (h *LiveHolder) Get() Live {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.live
}

func (h *LiveHolder) History() []Point {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]Point, len(h.history))
	copy(out, h.history)
	return out
}
