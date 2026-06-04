// Package tlsmgr configures TLS for the panel's own HTTP server: from cert
// files, a generated self-signed cert (usable on a bare IP), or Let's Encrypt
// via autocert (TLS-ALPN-01).
package tlsmgr

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// Modes.
const (
	ModeOff        = "off"
	ModeFile       = "file"
	ModeSelfSigned = "self_signed"
	ModeACME       = "acme"
)

// Config mirrors config.TLSConfig.
type Config struct {
	Mode            string
	CertFile        string
	KeyFile         string
	ACMEEmail       string
	ACMEDomains     []string
	ACMECacheDir    string
	SelfSignedHosts []string
	SelfSignedDir   string
}

type Manager struct {
	cfg Config
}

func New(cfg Config) *Manager { return &Manager{cfg: cfg} }

// Enabled reports whether the panel should serve HTTPS.
func (m *Manager) Enabled() bool {
	return m.cfg.Mode != "" && m.cfg.Mode != ModeOff
}

// TLSConfig builds the *tls.Config for the server, or nil when TLS is off.
func (m *Manager) TLSConfig() (*tls.Config, error) {
	switch m.cfg.Mode {
	case "", ModeOff:
		return nil, nil

	case ModeFile:
		if m.cfg.CertFile == "" || m.cfg.KeyFile == "" {
			return nil, fmt.Errorf("tls mode=file requires cert_file and key_file")
		}
		cert, err := tls.LoadX509KeyPair(m.cfg.CertFile, m.cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load keypair: %w", err)
		}
		return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}, nil

	case ModeSelfSigned:
		cert, err := m.ensureSelfSigned()
		if err != nil {
			return nil, err
		}
		return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}, nil

	case ModeACME:
		if len(m.cfg.ACMEDomains) == 0 {
			return nil, fmt.Errorf("tls mode=acme requires acme_domains")
		}
		cacheDir := m.cfg.ACMECacheDir
		if cacheDir == "" {
			cacheDir = "./storage/acme"
		}
		if err := os.MkdirAll(cacheDir, 0o700); err != nil {
			return nil, fmt.Errorf("create acme cache dir: %w", err)
		}
		am := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache(cacheDir),
			HostPolicy: autocert.HostWhitelist(m.cfg.ACMEDomains...),
			Email:      m.cfg.ACMEEmail,
		}
		tc := am.TLSConfig()
		// TLS-ALPN-01 needs the acme NextProto, which am.TLSConfig already sets.
		tc.MinVersion = tls.VersionTLS12
		if len(tc.NextProtos) == 0 {
			tc.NextProtos = []string{"h2", "http/1.1", acme.ALPNProto}
		}
		return tc, nil

	default:
		return nil, fmt.Errorf("unknown tls mode %q", m.cfg.Mode)
	}
}

// ensureSelfSigned loads a cached self-signed cert, generating one (with the
// configured hosts plus loopback) if absent.
func (m *Manager) ensureSelfSigned() (tls.Certificate, error) {
	certPath, keyPath, err := EnsureSelfSigned(m.cfg.SelfSignedDir, m.cfg.SelfSignedHosts)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.LoadX509KeyPair(certPath, keyPath)
}

// EnsureSelfSigned makes sure a self-signed keypair exists in dir, generating one
// (covering loopback plus extraHosts) when absent, and returns the cert and key
// file paths. Reused both for the panel's own HTTPS cert and as the default
// certificate source for TLS inbounds that have no ACME/cert configured.
func EnsureSelfSigned(dir string, extraHosts []string) (certPath, keyPath string, err error) {
	if dir == "" {
		dir = "./storage/tls"
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", "", fmt.Errorf("create tls dir: %w", err)
	}
	certPath = filepath.Join(dir, "cert.pem")
	keyPath = filepath.Join(dir, "key.pem")

	_, certErr := os.Stat(certPath)
	_, keyErr := os.Stat(keyPath)
	if certErr == nil && keyErr == nil {
		return certPath, keyPath, nil
	}

	hosts := append([]string{"127.0.0.1", "::1", "localhost"}, extraHosts...)
	certPEM, keyPEM, err := generateSelfSigned(hosts)
	if err != nil {
		return "", "", err
	}
	if err := os.WriteFile(certPath, certPEM, 0o600); err != nil {
		return "", "", fmt.Errorf("write cert: %w", err)
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return "", "", fmt.Errorf("write key: %w", err)
	}
	return certPath, keyPath, nil
}

func generateSelfSigned(hosts []string) (certPEM, keyPEM []byte, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate key: %w", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("serial: %w", err)
	}

	cn := "shilka"
	if len(hosts) > 0 {
		cn = hosts[0]
	}
	tmpl := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: cn, Organization: []string{"shilka"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else {
			tmpl.DNSNames = append(tmpl.DNSNames, h)
		}
	}

	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("create certificate: %w", err)
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal key: %w", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM, nil
}
