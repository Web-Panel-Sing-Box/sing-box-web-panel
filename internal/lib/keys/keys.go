// Package keys generates the cryptographic material used by inbounds and
// clients: Reality x25519 key pairs, short IDs, UUIDs, passwords and
// subscription tokens. All randomness comes from crypto/rand.
package keys

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// RealityKeyPair holds base64.RawURLEncoding-encoded x25519 keys, matching the
// output of `sing-box generate reality-keypair`.
type RealityKeyPair struct {
	PrivateKey string
	PublicKey  string
}

// GenerateRealityKeyPair creates a clamped x25519 key pair for VLESS Reality.
func GenerateRealityKeyPair() (RealityKeyPair, error) {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return RealityKeyPair{}, fmt.Errorf("read random: %w", err)
	}
	// X25519 / WireGuard scalar clamping.
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64

	pub, err := curve25519.X25519(priv[:], curve25519.Basepoint)
	if err != nil {
		return RealityKeyPair{}, fmt.Errorf("derive public key: %w", err)
	}
	return RealityKeyPair{
		PrivateKey: base64.RawURLEncoding.EncodeToString(priv[:]),
		PublicKey:  base64.RawURLEncoding.EncodeToString(pub),
	}, nil
}

// GenerateShortID returns an 8-character hex Reality short_id.
func GenerateShortID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// GenerateToken returns a URL-safe random token of nBytes entropy.
func GenerateToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateSubToken returns a subscription token (24 bytes of entropy).
func GenerateSubToken() (string, error) { return GenerateToken(24) }

// GeneratePassword returns a credential for Naive / Hysteria2 clients.
func GeneratePassword() (string, error) { return GenerateToken(18) }

// GenerateUUID returns a random RFC 4122 version-4 UUID string.
func GenerateUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("read random: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
