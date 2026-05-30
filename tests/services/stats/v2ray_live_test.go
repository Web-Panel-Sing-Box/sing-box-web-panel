package stats_test

import (
	"context"
	"os"
	"testing"
	"time"

	"sing-box-web-panel/internal/services/stats"
)

// TestV2RayUserDeltasLive exercises the hand-rolled gRPC StatsService codec
// against a real sing-box V2Ray API. It is skipped unless SING_GROK_LIVE_V2RAY
// is set to the api address (e.g. 127.0.0.1:8088).
func TestV2RayUserDeltasLive(t *testing.T) {
	addr := os.Getenv("SING_GROK_LIVE_V2RAY")
	if addr == "" {
		t.Skip("set SING_GROK_LIVE_V2RAY=host:port to run")
	}

	src := stats.NewV2RaySource(addr)
	defer src.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	deltas, err := src.UserDeltas(ctx)
	if err != nil {
		t.Fatalf("UserDeltas: %v", err)
	}
	t.Logf("decoded %d user(s) from v2ray stats", len(deltas))
	for _, d := range deltas {
		t.Logf("user=%q up=%d down=%d", d.Name, d.Up, d.Down)
	}
}
