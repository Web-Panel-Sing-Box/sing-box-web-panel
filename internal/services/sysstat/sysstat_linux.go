//go:build linux

package sysstat

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"sing-box-web-panel/internal/domain"
)

type linuxReader struct {
	mu        sync.Mutex
	prevIdle  uint64
	prevTotal uint64
	havePrev  bool
}

// New returns a Linux /proc-backed metrics reader.
func New() Reader { return &linuxReader{} }

func (r *linuxReader) Read() (domain.SystemMetrics, error) {
	var m domain.SystemMetrics
	m.CPU = r.cpu()
	r.mem(&m)
	m.UptimeSeconds = uptime()
	r.disk(&m)
	return m, nil
}

// cpu computes the CPU busy fraction from the delta of /proc/stat since the
// previous call. The first call returns 0.
func (r *linuxReader) cpu() float64 {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return 0
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	if !sc.Scan() {
		return 0
	}
	fields := strings.Fields(sc.Text())
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0
	}
	var total, idle uint64
	for i, v := range fields[1:] {
		n, _ := strconv.ParseUint(v, 10, 64)
		total += n
		if i == 3 || i == 4 { // idle + iowait
			idle += n
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	defer func() {
		r.prevIdle, r.prevTotal, r.havePrev = idle, total, true
	}()
	if !r.havePrev || total <= r.prevTotal {
		return 0
	}
	totalDelta := float64(total - r.prevTotal)
	idleDelta := float64(idle - r.prevIdle)
	if totalDelta == 0 {
		return 0
	}
	usage := 1 - idleDelta/totalDelta
	return clamp01(usage)
}

func (r *linuxReader) mem(m *domain.SystemMetrics) {
	vals := readMeminfo()
	memTotal := vals["MemTotal"]
	memAvail := vals["MemAvailable"]
	swapTotal := vals["SwapTotal"]
	swapFree := vals["SwapFree"]

	m.RAMTotalBytes = int64(memTotal * 1024)
	if memTotal > 0 {
		used := memTotal - memAvail
		m.RAMUsedBytes = int64(used * 1024)
		m.RAM = clamp01(float64(used) / float64(memTotal))
	}
	m.SwapTotalBytes = int64(swapTotal * 1024)
	if swapTotal > 0 {
		used := swapTotal - swapFree
		m.SwapUsedBytes = int64(used * 1024)
		m.Swap = clamp01(float64(used) / float64(swapTotal))
	}
}

func (r *linuxReader) disk(m *domain.SystemMetrics) {
	var st syscall.Statfs_t
	if err := syscall.Statfs("/", &st); err != nil {
		return
	}
	bs := int64(st.Bsize)
	total := int64(st.Blocks) * bs
	free := int64(st.Bavail) * bs
	used := total - free
	m.DiskSegments = []domain.DiskSegment{
		{Label: "used", UsedBytes: used, TotalBytes: total},
		{Label: "free", UsedBytes: free, TotalBytes: total},
	}
}

func readMeminfo() map[string]uint64 {
	out := make(map[string]uint64)
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return out
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		key, rest, ok := strings.Cut(sc.Text(), ":")
		if !ok {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		n, _ := strconv.ParseUint(fields[0], 10, 64)
		out[key] = n
	}
	return out
}

func uptime() int64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	secs, _ := strconv.ParseFloat(fields[0], 64)
	return int64(secs)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
