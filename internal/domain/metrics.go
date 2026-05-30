package domain

// SystemMetrics is a snapshot of host resource usage. On non-Linux dev hosts
// the sysstat fallback returns zero values.
type SystemMetrics struct {
	CPU            float64 // 0..1 load fraction
	RAM            float64 // 0..1 used fraction
	Swap           float64 // 0..1 used fraction
	RAMUsedBytes   int64
	RAMTotalBytes  int64
	SwapUsedBytes  int64
	SwapTotalBytes int64
	UptimeSeconds  int64
	DiskSegments   []DiskSegment
}

// DiskSegment is one slice of the disk-usage breakdown shown on the dashboard.
type DiskSegment struct {
	Label      string
	UsedBytes  int64
	TotalBytes int64
}

// TrafficSample is an instantaneous global throughput reading from the core.
type TrafficSample struct {
	UploadBps   int64
	DownloadBps int64
}

// UserTraffic is a cumulative per-client counter reported by a TrafficSource.
type UserTraffic struct {
	Name string
	Up   int64
	Down int64
}

// TrafficDelta is an increment applied to a client's stored counters during a
// batched write from the traffic worker.
type TrafficDelta struct {
	ClientID int64
	Up       int64
	Down     int64
}
