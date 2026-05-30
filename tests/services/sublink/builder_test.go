package sublink_test

import (
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
	for _, want := range []string{"hysteria2://", "secret-pass@panel.example:51005", "sni=panel.example", "#bob"} {
		if !strings.Contains(link, want) {
			t.Errorf("hysteria2 link missing %q\n got: %s", want, link)
		}
	}
}

func TestBuildLinkNaive(t *testing.T) {
	ib := &domain.Inbound{ID: 3, Protocol: domain.ProtocolNaive, Port: 38119, TLS: domain.TLSModeTLS}
	c := &domain.Client{Name: "carol", Password: "pw"}

	link := sublink.BuildLink(ib, c, "panel.example")
	if !strings.HasPrefix(link, "naive+https://carol:pw@panel.example:38119") {
		t.Errorf("unexpected naive link: %s", link)
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
