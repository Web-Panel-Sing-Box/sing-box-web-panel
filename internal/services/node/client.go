package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"sing-box-web-panel/internal/domain"
)

var ErrUnsafeAddress = errors.New("node address is private or loopback")

type RemoteClienter interface {
	Status(ctx context.Context, n *domain.Node) (*RemoteStatus, time.Duration, error)
	Snapshot(ctx context.Context, n *domain.Node) (*RemoteSnapshot, error)
}

type HTTPClient struct {
	statusTimeout time.Duration
	writeTimeout  time.Duration
}

func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		statusTimeout: 4 * time.Second,
		writeTimeout:  10 * time.Second,
	}
}

func (c *HTTPClient) Status(ctx context.Context, n *domain.Node) (*RemoteStatus, time.Duration, error) {
	start := time.Now()
	var out RemoteStatus
	if err := c.get(ctx, n, "/api/node/v1/status", c.statusTimeout, &out); err != nil {
		return nil, 0, err
	}
	return &out, time.Since(start), nil
}

func (c *HTTPClient) Snapshot(ctx context.Context, n *domain.Node) (*RemoteSnapshot, error) {
	var out RemoteSnapshot
	if err := c.get(ctx, n, "/api/node/v1/snapshot", c.writeTimeout, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *HTTPClient) get(ctx context.Context, n *domain.Node, suffix string, timeout time.Duration, out any) error {
	base, err := baseURL(n)
	if err != nil {
		return err
	}
	base.Path = path.Join(base.Path, suffix)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.APITokenSecret)

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy:       http.ProxyFromEnvironment,
			DialContext: safeDialer(n.AllowPrivateAddress).DialContext,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("node returned status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode node response: %w", err)
	}
	return nil
}

func baseURL(n *domain.Node) (*url.URL, error) {
	scheme := strings.ToLower(strings.TrimSpace(n.Scheme))
	if scheme == "" {
		scheme = "https"
	}
	if scheme != "http" && scheme != "https" {
		return nil, fmt.Errorf("unsupported node scheme %q", n.Scheme)
	}
	if n.Address == "" {
		return nil, fmt.Errorf("node address is required")
	}
	if n.Port < 1 || n.Port > 65535 {
		return nil, fmt.Errorf("node port must be between 1 and 65535")
	}
	u := &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(strings.TrimSpace(n.Address), strconv.Itoa(n.Port)),
		Path:   strings.TrimSpace(n.BasePath),
	}
	return u, nil
}

func safeDialer(allowPrivate bool) *net.Dialer {
	return &net.Dialer{
		Timeout: 4 * time.Second,
		Control: func(network, address string, _ syscall.RawConn) error {
			if allowPrivate {
				return nil
			}
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ips, err := net.LookupIP(host)
			if err != nil {
				return err
			}
			for _, ip := range ips {
				if isUnsafeIP(ip) {
					return ErrUnsafeAddress
				}
			}
			return nil
		},
	}
}

func isUnsafeIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	return false
}
