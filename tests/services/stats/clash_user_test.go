package stats_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sing-box-web-panel/internal/services/singbox"
	"sing-box-web-panel/internal/services/stats"
)

type connFix struct {
	id           string
	chains       []string
	upload, down int64
}

func clashConnServer(t *testing.T, samples [][]connFix) *httptest.Server {
	t.Helper()
	var i int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/connections" {
			http.NotFound(w, r)
			return
		}
		idx := i
		if idx >= len(samples) {
			idx = len(samples) - 1
		}
		i++
		conns := make([]map[string]any, 0, len(samples[idx]))
		for _, c := range samples[idx] {
			conns = append(conns, map[string]any{
				"id":       c.id,
				"upload":   c.upload,
				"download": c.down,
				"chains":   c.chains,
			})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"connections": conns})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func hostOf(srv *httptest.Server) string {
	return strings.TrimPrefix(srv.URL, "http://")
}

func staticLookup(byID map[int64]string) func(int64) string {
	return func(id int64) string { return byID[id] }
}

func TestClashUserSourceAttributesViaChains(t *testing.T) {
	tagAlice := singbox.ClientOutboundTag(1)
	tagBob := singbox.ClientOutboundTag(2)

	srv := clashConnServer(t, [][]connFix{
		{
			{id: "a", chains: []string{tagAlice, "direct"}, upload: 100, down: 200},
			{id: "b", chains: []string{tagBob, "direct"}, upload: 50, down: 75},
		},
		{
			{id: "a", chains: []string{tagAlice, "direct"}, upload: 300, down: 500},
			{id: "b", chains: []string{tagBob, "direct"}, upload: 50, down: 75},
		},
	})

	src := stats.NewClashUserSource(hostOf(srv), "", staticLookup(map[int64]string{
		1: "alice",
		2: "bob",
	}))

	first, err := src.UserDeltas(context.Background())
	if err != nil {
		t.Fatalf("UserDeltas 1: %v", err)
	}
	byUser := map[string][2]int64{}
	for _, u := range first {
		byUser[u.Name] = [2]int64{u.Up, u.Down}
	}
	if got := byUser["alice"]; got != [2]int64{100, 200} {
		t.Fatalf("alice initial = %v, want [100 200]", got)
	}
	if got := byUser["bob"]; got != [2]int64{50, 75} {
		t.Fatalf("bob initial = %v, want [50 75]", got)
	}

	seen := src.UserSeen()
	if _, ok := seen["alice"]; !ok {
		t.Fatalf("UserSeen missing alice: %v", seen)
	}
	if _, ok := seen["bob"]; !ok {
		t.Fatalf("UserSeen missing bob: %v", seen)
	}

	second, err := src.UserDeltas(context.Background())
	if err != nil {
		t.Fatalf("UserDeltas 2: %v", err)
	}
	if len(second) != 1 || second[0].Name != "alice" {
		t.Fatalf("second deltas = %+v, want alice only", second)
	}
	if second[0].Up != 200 || second[0].Down != 300 {
		t.Fatalf("alice delta = %+v, want {alice 200 300}", second[0])
	}
}

func TestClashUserSourceSkipsUntaggedConns(t *testing.T) {
	srv := clashConnServer(t, [][]connFix{
		{
			{id: "x", chains: []string{"direct"}, upload: 1000, down: 2000},
			{id: "y", chains: nil, upload: 50, down: 100},
		},
	})
	src := stats.NewClashUserSource(hostOf(srv), "", staticLookup(nil))

	got, err := src.UserDeltas(context.Background())
	if err != nil {
		t.Fatalf("UserDeltas: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got = %v, want empty (no user-N chain)", got)
	}
	if len(src.UserSeen()) != 0 {
		t.Fatalf("UserSeen = %v, want empty", src.UserSeen())
	}
}

func TestClashUserSourceSkipsDeletedClient(t *testing.T) {
	tag := singbox.ClientOutboundTag(42)
	srv := clashConnServer(t, [][]connFix{
		{{id: "z", chains: []string{tag}, upload: 1, down: 1}},
	})
	src := stats.NewClashUserSource(hostOf(srv), "", staticLookup(nil)) // 42 → ""

	got, err := src.UserDeltas(context.Background())
	if err != nil {
		t.Fatalf("UserDeltas: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("got = %v, want empty (client deleted)", got)
	}
}

func TestClashUserSourceReusedConnectionID(t *testing.T) {
	tag := singbox.ClientOutboundTag(7)
	srv := clashConnServer(t, [][]connFix{
		{{id: "1", chains: []string{tag}, upload: 1000, down: 2000}},
		{{id: "1", chains: []string{tag}, upload: 50, down: 75}},
	})
	src := stats.NewClashUserSource(hostOf(srv), "", staticLookup(map[int64]string{7: "carol"}))

	if _, err := src.UserDeltas(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := src.UserDeltas(context.Background())
	if err != nil {
		t.Fatalf("UserDeltas: %v", err)
	}
	if len(got) != 1 || got[0].Name != "carol" || got[0].Up != 50 || got[0].Down != 75 {
		t.Fatalf("reused-id delta = %+v, want {carol 50 75}", got)
	}
}

func TestClashUserSourceUserSeenIsReplacedEachPoll(t *testing.T) {
	tagAlice := singbox.ClientOutboundTag(1)
	tagBob := singbox.ClientOutboundTag(2)
	srv := clashConnServer(t, [][]connFix{
		{{id: "a", chains: []string{tagAlice}, upload: 10, down: 20}},
		{{id: "b", chains: []string{tagBob}, upload: 10, down: 20}},
	})
	src := stats.NewClashUserSource(hostOf(srv), "", staticLookup(map[int64]string{
		1: "alice",
		2: "bob",
	}))

	if _, err := src.UserDeltas(context.Background()); err != nil {
		t.Fatal(err)
	}
	if seen := src.UserSeen(); len(seen) != 1 {
		t.Fatalf("after sample 1, UserSeen = %v, want only alice", seen)
	}
	if _, err := src.UserDeltas(context.Background()); err != nil {
		t.Fatal(err)
	}
	seen := src.UserSeen()
	if _, ok := seen["alice"]; ok {
		t.Fatalf("UserSeen still has alice after she dropped: %v", seen)
	}
	if _, ok := seen["bob"]; !ok {
		t.Fatalf("UserSeen missing bob: %v", seen)
	}
}
