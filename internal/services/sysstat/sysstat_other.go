//go:build !linux

package sysstat

import "sing-box-web-panel/internal/domain"

// stubReader is used on non-Linux dev hosts; host metrics are unavailable there.
type stubReader struct{}

// New returns a no-op reader for non-Linux platforms.
func New() Reader { return &stubReader{} }

func (stubReader) Read() (domain.SystemMetrics, error) {
	return domain.SystemMetrics{}, nil
}
