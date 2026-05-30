package singbox_test

import (
	"context"
	"encoding/json"
	"testing"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/services/singbox"
)

type fakeInbounds struct{ list []domain.Inbound }

func (f fakeInbounds) ListEnabled(context.Context) ([]domain.Inbound, error) { return f.list, nil }

type fakeClients struct{ list []domain.Client }

func (f fakeClients) ListEnabled(context.Context) ([]domain.Client, error) { return f.list, nil }

func realityInbound() domain.Inbound {
	return domain.Inbound{
		ID: 1, Remark: "edge", Protocol: domain.ProtocolVLESS, Port: 44321,
		Transmission: domain.TransmissionTCP, TLS: domain.TLSModeReality,
		SNI: "www.cloudflare.com", Dest: "www.cloudflare.com:443", Enabled: true,
		Settings: domain.InboundSettings{
			RealityPrivateKey: "PRIV", RealityPublicKey: "PUB", RealityShortID: "abcd1234",
			Flow: "xtls-rprx-vision",
		},
	}
}

func render(t *testing.T, source string) map[string]any {
	t.Helper()
	gen := singbox.NewGenerator(
		fakeInbounds{list: []domain.Inbound{realityInbound()}},
		fakeClients{list: []domain.Client{{ID: 1, InboundID: 1, Name: "alice", UUID: "uuid-1", Status: domain.ClientStatusActive}}},
		singbox.GeneratorConfig{
			ClashAPIAddress: "127.0.0.1:9090", ClashAPISecret: "secret",
			StatsSource: source, V2RayAPIListen: "127.0.0.1:8088",
		},
	)
	data, err := gen.Render(context.Background())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return cfg
}

func TestGeneratorClashSource(t *testing.T) {
	cfg := render(t, "clash")

	inbounds := cfg["inbounds"].([]any)
	if len(inbounds) != 1 {
		t.Fatalf("want 1 inbound, got %d", len(inbounds))
	}
	ib := inbounds[0].(map[string]any)
	if ib["type"] != "vless" {
		t.Errorf("want vless, got %v", ib["type"])
	}
	tls := ib["tls"].(map[string]any)
	reality := tls["reality"].(map[string]any)
	if reality["private_key"] != "PRIV" {
		t.Errorf("reality private_key missing")
	}
	users := ib["users"].([]any)
	if len(users) != 1 || users[0].(map[string]any)["uuid"] != "uuid-1" {
		t.Errorf("unexpected users: %v", users)
	}

	exp := cfg["experimental"].(map[string]any)
	if _, ok := exp["clash_api"]; !ok {
		t.Errorf("clash_api must always be present")
	}
	if _, ok := exp["v2ray_api"]; ok {
		t.Errorf("v2ray_api must NOT be present for clash source")
	}
}

func TestGeneratorV2RaySource(t *testing.T) {
	cfg := render(t, "v2ray")
	exp := cfg["experimental"].(map[string]any)
	v2 := exp["v2ray_api"].(map[string]any)
	stats := v2["stats"].(map[string]any)
	users := stats["users"].([]any)
	if len(users) != 1 || users[0] != "alice" {
		t.Errorf("v2ray stats.users should list active client names, got %v", users)
	}
}

func TestGeneratorSkipsInactiveClients(t *testing.T) {
	gen := singbox.NewGenerator(
		fakeInbounds{list: []domain.Inbound{realityInbound()}},
		fakeClients{list: []domain.Client{{ID: 1, InboundID: 1, Name: "expired", Status: domain.ClientStatusExpired}}},
		singbox.GeneratorConfig{ClashAPIAddress: "127.0.0.1:9090"},
	)
	data, _ := gen.Render(context.Background())
	var cfg map[string]any
	json.Unmarshal(data, &cfg)
	ib := cfg["inbounds"].([]any)[0].(map[string]any)
	if users := ib["users"].([]any); len(users) != 0 {
		t.Errorf("inactive clients must be excluded, got %v", users)
	}
}
