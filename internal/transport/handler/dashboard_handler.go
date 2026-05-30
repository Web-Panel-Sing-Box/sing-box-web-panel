package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/services/singbox"
	"sing-box-web-panel/internal/services/stats"
)

// Dashboard dependencies (consumer-side interfaces).

type sysReader interface {
	Read() (domain.SystemMetrics, error)
}

type clientStats interface {
	Count(ctx context.Context) (int, error)
}

type inboundStats interface {
	ListEnabled(ctx context.Context) ([]domain.Inbound, error)
}

type rollupStats interface {
	Day(ctx context.Context, day string) (up, down int64, err error)
	SumSince(ctx context.Context, since string) (up, down int64, err error)
}

type procStatus interface {
	Status(ctx context.Context) (singbox.Status, error)
}

type DashboardHandler struct {
	sys      sysReader
	live     *stats.LiveHolder
	clients  clientStats
	inbounds inboundStats
	rollup   rollupStats
	proc     procStatus
	log      *slog.Logger
}

func NewDashboardHandler(sys sysReader, live *stats.LiveHolder, clients clientStats, inbounds inboundStats, rollup rollupStats, proc procStatus, log *slog.Logger) *DashboardHandler {
	return &DashboardHandler{sys: sys, live: live, clients: clients, inbounds: inbounds, rollup: rollup, proc: proc, log: log}
}

func (h *DashboardHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/dashboard/metrics", h.Metrics)
	mux.HandleFunc("GET /api/dashboard/traffic", h.Traffic)
}

type diskSegmentDTO struct {
	Label      string `json:"label"`
	UsedBytes  int64  `json:"usedBytes"`
	TotalBytes int64  `json:"totalBytes"`
	Color      string `json:"color,omitempty"`
}

type metricsDTO struct {
	CPU            float64          `json:"cpu"`
	RAM            float64          `json:"ram"`
	Swap           float64          `json:"swap"`
	RAMUsedBytes   int64            `json:"ramUsedBytes"`
	RAMTotalBytes  int64            `json:"ramTotalBytes"`
	SwapUsedBytes  int64            `json:"swapUsedBytes"`
	SwapTotalBytes int64            `json:"swapTotalBytes"`
	UptimeSeconds  int64            `json:"uptimeSeconds"`
	UploadBps      int64            `json:"uploadBps"`
	DownloadBps    int64            `json:"downloadBps"`
	TodayBytes     int64            `json:"todayBytes"`
	MonthBytes     int64            `json:"monthBytes"`
	TotalSent      int64            `json:"totalSent"`
	TotalReceived  int64            `json:"totalReceived"`
	DiskSegments   []diskSegmentDTO `json:"diskSegments"`
	InboundsActive int              `json:"inboundsActive"`
	TotalUsers     int              `json:"totalUsers"`
	OnlineNow      int              `json:"onlineNow"`
	CoreRunning    bool             `json:"coreRunning"`
	CoreVersion    string           `json:"coreVersion"`
}

// Metrics godoc
//
//	@Summary	Dashboard metrics snapshot
//	@Tags		dashboard
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	metricsDTO
//	@Router		/dashboard/metrics [get]
func (h *DashboardHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sys, _ := h.sys.Read()
	live := h.live.Get()

	totalUsers, _ := h.clients.Count(ctx)
	enabled, _ := h.inbounds.ListEnabled(ctx)

	now := time.Now().UTC()
	todayUp, todayDown, _ := h.rollup.Day(ctx, now.Format("2006-01-02"))
	monthUp, monthDown, _ := h.rollup.SumSince(ctx, now.Format("2006-01")+"-01")

	st, _ := h.proc.Status(ctx)

	dto := metricsDTO{
		CPU:            sys.CPU,
		RAM:            sys.RAM,
		Swap:           sys.Swap,
		RAMUsedBytes:   sys.RAMUsedBytes,
		RAMTotalBytes:  sys.RAMTotalBytes,
		SwapUsedBytes:  sys.SwapUsedBytes,
		SwapTotalBytes: sys.SwapTotalBytes,
		UptimeSeconds:  sys.UptimeSeconds,
		UploadBps:      live.UploadBps,
		DownloadBps:    live.DownloadBps,
		TodayBytes:     todayUp + todayDown,
		MonthBytes:     monthUp + monthDown,
		TotalSent:      live.UploadTotal,
		TotalReceived:  live.DownloadTotal,
		DiskSegments:   diskSegments(sys.DiskSegments),
		InboundsActive: len(enabled),
		TotalUsers:     totalUsers,
		OnlineNow:      live.Connections,
		CoreRunning:    st.Running,
		CoreVersion:    st.Version,
	}
	writeJSON(w, http.StatusOK, dto)
}

func diskSegments(segs []domain.DiskSegment) []diskSegmentDTO {
	out := make([]diskSegmentDTO, 0, len(segs))
	for _, s := range segs {
		color := "#ffffff14"
		if s.Label == "used" {
			color = "#10a37f"
		}
		out = append(out, diskSegmentDTO{Label: s.Label, UsedBytes: s.UsedBytes, TotalBytes: s.TotalBytes, Color: color})
	}
	return out
}

type trafficPointDTO struct {
	T    int64 `json:"t"`
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

// Traffic godoc
//
//	@Summary	Recent throughput history
//	@Tags		dashboard
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{array}	trafficPointDTO
//	@Router		/dashboard/traffic [get]
func (h *DashboardHandler) Traffic(w http.ResponseWriter, r *http.Request) {
	points := h.live.History()
	out := make([]trafficPointDTO, 0, len(points))
	for _, p := range points {
		out = append(out, trafficPointDTO{T: p.T, Up: p.Up, Down: p.Down})
	}
	writeJSON(w, http.StatusOK, out)
}
