package sublink_test

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/services/sublink"
)

func TestBuildLinkVLESSReality(t *testing.T) {
	ib := &domain.Inbound{
		ID: 1, Remark: "edge", Protocol: domain.ProtocolVLESS, Port: 44321,
		Transmission: domain.TransmissionTCP, TLS: domain.TLSModeReality,
		SNI: "www.cloudflare.com", Dest: "www.cloudflare.com:443",
		Settings: domain.InboundSettings{
			RealityPublicKey: "PUBKEY", RealityShortID: "abcd1234", Flow: "xtls-rprx-vision",
		},
	}
	c := &domain.Client{Name: "alice", UUID: "11111111-1111-4111-8111-111111111111"}

	link := sublink.BuildLink(ib, c, "203.0.113.10")
	for _, want := range []string{
		"vless://11111111-1111-4111-8111-111111111111@203.0.113.10:44321",
		"security=reality", "pbk=PUBKEY", "sid=abcd1234", "flow=xtls-rprx-vision",
		"sni=www.cloudflare.com", "#alice",
	} {
		if !strings.Contains(link, want) {
			t.Errorf("vless reality link missing %q\n got: %s", want, link)
		}
	}
}

func TestBuildLinkHysteria2(t *testing.T) {
	ib := &domain.Inbound{
		ID: 2, Protocol: domain.ProtocolHysteria2, Port: 51005,
		TLS: domain.TLSModeTLS, SNI: "panel.example",
	}
	c := &domain.Client{Name: "bob", Password: "secret-pass"}

	link := sublink.BuildLink(ib, c, "panel.example")
	for _, want := range []string{"hysteria2://", "secret-pass@panel.example:51005", "sni=panel.example", "insecure=1", "#bob"} {
		if !strings.Contains(link, want) {
			t.Errorf("hysteria2 link missing %q\n got: %s", want, link)
		}
	}
}

func TestBuildLinkHysteria2AllowInsecureFalse(t *testing.T) {
	ib := &domain.Inbound{
		ID: 2, Protocol: domain.ProtocolHysteria2, Port: 51005,
		TLS: domain.TLSModeTLS, SNI: "panel.example",
		Settings: domain.InboundSettings{
			AllowInsecure: boolPtr(false),
		},
	}
	c := &domain.Client{Name: "bob", Password: "secret-pass"}

	link := sublink.BuildLink(ib, c, "panel.example")
	if got := queryValue(t, link, "insecure"); got != "0" {
		t.Fatalf("insecure = %q, want 0\nlink: %s", got, link)
	}
}

func TestBuildLinkVLESSTLSAllowInsecure(t *testing.T) {
	ib := &domain.Inbound{
		ID: 4, Protocol: domain.ProtocolVLESS, Port: 443,
		Transmission: domain.TransmissionTCP, TLS: domain.TLSModeTLS,
		SNI: "panel.example",
	}
	c := &domain.Client{Name: "dave", UUID: "11111111-1111-4111-8111-111111111111"}

	link := sublink.BuildLink(ib, c, "panel.example")
	if got := queryValue(t, link, "allowInsecure"); got != "1" {
		t.Fatalf("allowInsecure = %q, want 1\nlink: %s", got, link)
	}
}

func TestBuildLinkVLESSRealityOmitsAllowInsecure(t *testing.T) {
	ib := &domain.Inbound{
		ID: 5, Protocol: domain.ProtocolVLESS, Port: 443,
		Transmission: domain.TransmissionTCP, TLS: domain.TLSModeReality,
		SNI: "www.cloudflare.com", Dest: "www.cloudflare.com:443",
		Settings: domain.InboundSettings{
			AllowInsecure:     boolPtr(true),
			RealityPublicKey:  "PUBKEY",
			RealityPrivateKey: "PRIVATE",
			RealityShortID:    "abcd1234",
		},
	}
	c := &domain.Client{Name: "erin", UUID: "11111111-1111-4111-8111-111111111111"}

	link := sublink.BuildLink(ib, c, "panel.example")
	if got := queryValue(t, link, "allowInsecure"); got != "" {
		t.Fatalf("allowInsecure = %q, want empty\nlink: %s", got, link)
	}
}

func TestBuildLinkNaive(t *testing.T) {
	ib := &domain.Inbound{ID: 3, Protocol: domain.ProtocolNaive, Port: 38119, TLS: domain.TLSModeTLS}
	c := &domain.Client{Name: "carol", Password: "pw"}

	link := sublink.BuildLink(ib, c, "panel.example")
	if !strings.HasPrefix(link, "naive+https://carol:pw@panel.example:38119") {
		t.Errorf("unexpected naive link: %s", link)
	}
	if got := queryValue(t, link, "allowInsecure"); got != "1" {
		t.Fatalf("allowInsecure = %q, want 1\nlink: %s", got, link)
	}
}

func TestRenderBase64(t *testing.T) {
	ib := &domain.Inbound{ID: 1, Protocol: domain.ProtocolVLESS, Port: 443, TLS: domain.TLSModeNone}
	c := &domain.Client{Name: "x", UUID: "u"}

	res, err := sublink.Render(sublink.FormatBase64, ib, c, "host")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Body) == 0 || strings.Contains(string(res.Body), "vless://") {
		t.Errorf("base64 output should be encoded, got: %s", res.Body)
	}
}

func TestBuildClientConfigAllowInsecure(t *testing.T) {
	ib := &domain.Inbound{
		ID: 2, Protocol: domain.ProtocolHysteria2, Port: 51005,
		TLS: domain.TLSModeTLS, SNI: "panel.example",
	}
	c := &domain.Client{Name: "bob", Password: "secret-pass"}

	body, err := sublink.BuildClientConfig(ib, c, "panel.example")
	if err != nil {
		t.Fatal(err)
	}
	tls := firstOutboundTLS(t, body)
	if got := tls["insecure"]; got != true {
		t.Fatalf("tls.insecure = %#v, want true\nconfig: %s", got, body)
	}
}

func TestBuildClientConfigTrustedTLSOmitsInsecure(t *testing.T) {
	ib := &domain.Inbound{
		ID: 2, Protocol: domain.ProtocolHysteria2, Port: 51005,
		TLS: domain.TLSModeTLS, SNI: "panel.example",
		Settings: domain.InboundSettings{
			ACMEDomain: "panel.example",
		},
	}
	c := &domain.Client{Name: "bob", Password: "secret-pass"}

	body, err := sublink.BuildClientConfig(ib, c, "panel.example")
	if err != nil {
		t.Fatal(err)
	}
	tls := firstOutboundTLS(t, body)
	if _, ok := tls["insecure"]; ok {
		t.Fatalf("tls.insecure should be omitted for trusted TLS\nconfig: %s", body)
	}
}

func queryValue(t *testing.T, rawURL, key string) string {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse link: %v", err)
	}
	return u.Query().Get(key)
}

func firstOutboundTLS(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var cfg struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	if err := json.Unmarshal(body, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if len(cfg.Outbounds) == 0 {
		t.Fatal("expected at least one outbound")
	}
	tls, ok := cfg.Outbounds[0]["tls"].(map[string]any)
	if !ok {
		t.Fatalf("first outbound tls is missing or invalid: %#v", cfg.Outbounds[0]["tls"])
	}
	return tls
}

func boolPtr(v bool) *bool {
	return &v
}
