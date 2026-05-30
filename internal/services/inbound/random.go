package inbound

import (
	"crypto/rand"
	"encoding/binary"
)

// randomPort returns a pseudo-random port in [10000, 60000) for cloned
// inbounds. The caller persists it and a UNIQUE constraint rejects collisions.
func randomPort() int {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 20000
	}
	return 10000 + int(binary.BigEndian.Uint32(b[:])%50000)
}
