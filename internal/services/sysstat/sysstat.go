// Package sysstat reads host resource usage (CPU, memory, swap, disk, uptime).
// The Linux implementation reads /proc and statfs directly (no dependencies);
// other platforms (dev) return a zero-valued snapshot so builds work everywhere.
package sysstat

import "sing-box-web-panel/internal/domain"

// Reader returns a snapshot of host metrics.
type Reader interface {
	Read() (domain.SystemMetrics, error)
}
