package node_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"sing-box-web-panel/internal/domain"
	"sing-box-web-panel/internal/services/node"
)

func TestHTTPClientBlocksPrivateAddressByDefault(t *testing.T) {
	client := node.NewHTTPClient()
	n := &domain.Node{
		Scheme:         "http",
		Address:        "127.0.0.1",
		Port:           8080,
		APITokenSecret: "secret",
	}

	_, _, err := client.Status(context.Background(), n)
	if !errors.Is(err, node.ErrUnsafeAddress) {
		t.Fatalf("expected unsafe address error, got %v", err)
	}
}

func TestHTTPClientAllowsPrivateAddressWhenExplicit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("missing bearer token")
		}
		_ = json.NewEncoder(w).Encode(node.RemoteStatus{PanelVersion: "shilka-test"})
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	host, portRaw, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portRaw)
	if err != nil {
		t.Fatal(err)
	}

	client := node.NewHTTPClient()
	n := &domain.Node{
		Scheme:              u.Scheme,
		Address:             host,
		Port:                port,
		APITokenSecret:      "secret",
		AllowPrivateAddress: true,
	}

	status, _, err := client.Status(context.Background(), n)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.PanelVersion != "shilka-test" {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestHTTPClientRequiresTrustedTLSByDefault(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(node.RemoteStatus{PanelVersion: "shilka-test"})
	}))
	defer srv.Close()

	n := nodeFromServerURL(t, srv.URL)
	client := node.NewHTTPClient()
	if _, _, err := client.Status(context.Background(), n); err == nil {
		t.Fatal("expected TLS verification error")
	}
}

func TestHTTPClientSkipsTLSVerificationWhenExplicit(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("missing bearer token")
		}
		_ = json.NewEncoder(w).Encode(node.RemoteStatus{PanelVersion: "shilka-test"})
	}))
	defer srv.Close()

	n := nodeFromServerURL(t, srv.URL)
	n.SkipTLSVerify = true

	client := node.NewHTTPClient()
	status, _, err := client.Status(context.Background(), n)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.PanelVersion != "shilka-test" {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func nodeFromServerURL(t *testing.T, rawURL string) *domain.Node {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	host, portRaw, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portRaw)
	if err != nil {
		t.Fatal(err)
	}
	return &domain.Node{
		Scheme:              u.Scheme,
		Address:             host,
		Port:                port,
		APITokenSecret:      "secret",
		AllowPrivateAddress: true,
	}
}
