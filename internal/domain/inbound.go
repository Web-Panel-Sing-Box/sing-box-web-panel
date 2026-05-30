package domain

import "time"

// Protocol is the inbound proxy protocol. v1 supports the set modelled by the
// frontend: VLESS, Naive and Hysteria2.
type Protocol string

const (
	ProtocolVLESS     Protocol = "vless"
	ProtocolNaive     Protocol = "naive"
	ProtocolHysteria2 Protocol = "hysteria2"
)

// Transmission is the stream transport for VLESS. Naive and Hysteria2 always
// use their own transport and ignore this field.
type Transmission string

const (
	TransmissionTCP  Transmission = "tcp"
	TransmissionWS   Transmission = "ws"
	TransmissionGRPC Transmission = "grpc"
)

// TLSMode selects the security layer applied to an inbound.
type TLSMode string

const (
	TLSModeNone    TLSMode = "none"
	TLSModeTLS     TLSMode = "tls"
	TLSModeReality TLSMode = "reality"
)

// Inbound is a configured sing-box listener. Protocol-specific and rarely
// changing details live in Settings (persisted as JSON).
type Inbound struct {
	ID           int64
	Remark       string
	Protocol     Protocol
	Port         int
	Transmission Transmission
	TLS          TLSMode
	SNI          string
	Dest         string
	Enabled      bool
	Settings     InboundSettings
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// InboundSettings holds protocol/transport/TLS material that is generated or
// configured once and reused on every config render. The Reality private key
// is a secret and must never be logged or returned by the public API.
type InboundSettings struct {
	// Reality (VLESS).
	RealityPrivateKey string `json:"realityPrivateKey,omitempty"`
	RealityPublicKey  string `json:"realityPublicKey,omitempty"`
	RealityShortID    string `json:"realityShortId,omitempty"`
	// Flow for VLESS (e.g. "xtls-rprx-vision" with Reality).
	Flow string `json:"flow,omitempty"`
	// Transport options.
	WSPath          string `json:"wsPath,omitempty"`
	GRPCServiceName string `json:"grpcServiceName,omitempty"`
	// Standard TLS certificate material (mode = tls, no ACME).
	CertPath string `json:"certPath,omitempty"`
	KeyPath  string `json:"keyPath,omitempty"`
	// ACME (Let's Encrypt) for the inbound, handled by the sing-box core.
	ACMEDomain string `json:"acmeDomain,omitempty"`
	ACMEEmail  string `json:"acmeEmail,omitempty"`
}
