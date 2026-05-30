package singbox

// Minimal sing-box configuration schema. Only the fields the panel emits are
// modelled; JSON tags match the sing-box option structs. Targets the widely
// deployed 1.10/1.11 schema (direct/block outbounds, classic route rules).

type sbConfig struct {
	Log          *sbLog          `json:"log,omitempty"`
	Inbounds     []any           `json:"inbounds,omitempty"`
	Outbounds    []sbOutbound    `json:"outbounds,omitempty"`
	Route        *sbRoute        `json:"route,omitempty"`
	Experimental *sbExperimental `json:"experimental,omitempty"`
}

type sbLog struct {
	Level     string `json:"level,omitempty"`
	Timestamp bool   `json:"timestamp,omitempty"`
}

type sbOutbound struct {
	Type string `json:"type"`
	Tag  string `json:"tag"`
}

type sbRoute struct {
	Rules []sbRouteRule `json:"rules,omitempty"`
	Final string        `json:"final,omitempty"`
}

// sbRouteRule uses rule actions (sing-box 1.11+) rather than a block outbound,
// keeping the config valid through sing-box 1.14 where block/dns outbounds and
// the legacy DNS server format were removed.
type sbRouteRule struct {
	Protocol string `json:"protocol,omitempty"`
	Action   string `json:"action,omitempty"`
}

type sbExperimental struct {
	CacheFile *sbCacheFile `json:"cache_file,omitempty"`
	ClashAPI  *sbClashAPI  `json:"clash_api,omitempty"`
	V2RayAPI  *sbV2RayAPI  `json:"v2ray_api,omitempty"`
}

type sbCacheFile struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path,omitempty"`
}

type sbClashAPI struct {
	ExternalController string `json:"external_controller,omitempty"`
	Secret             string `json:"secret,omitempty"`
}

type sbV2RayAPI struct {
	Listen string        `json:"listen,omitempty"`
	Stats  *sbV2RayStats `json:"stats,omitempty"`
}

type sbV2RayStats struct {
	Enabled   bool     `json:"enabled"`
	Inbounds  []string `json:"inbounds,omitempty"`
	Outbounds []string `json:"outbounds,omitempty"`
	Users     []string `json:"users,omitempty"`
}

// --- inbounds ---

type sbVLESSInbound struct {
	Type       string        `json:"type"`
	Tag        string        `json:"tag"`
	Listen     string        `json:"listen"`
	ListenPort int           `json:"listen_port"`
	Users      []sbVLESSUser `json:"users"`
	TLS        *sbInboundTLS `json:"tls,omitempty"`
	Transport  *sbTransport  `json:"transport,omitempty"`
}

type sbVLESSUser struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
	Flow string `json:"flow,omitempty"`
}

type sbHysteria2Inbound struct {
	Type       string             `json:"type"`
	Tag        string             `json:"tag"`
	Listen     string             `json:"listen"`
	ListenPort int                `json:"listen_port"`
	Users      []sbHysteria2User  `json:"users"`
	TLS        *sbInboundTLS      `json:"tls,omitempty"`
}

type sbHysteria2User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type sbNaiveInbound struct {
	Type       string         `json:"type"`
	Tag        string         `json:"tag"`
	Listen     string         `json:"listen"`
	ListenPort int            `json:"listen_port"`
	Users      []sbNaiveUser  `json:"users"`
	TLS        *sbInboundTLS  `json:"tls"`
}

type sbNaiveUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type sbInboundTLS struct {
	Enabled    bool       `json:"enabled"`
	ServerName string     `json:"server_name,omitempty"`
	ALPN       []string   `json:"alpn,omitempty"`
	CertPath   string     `json:"certificate_path,omitempty"`
	KeyPath    string     `json:"key_path,omitempty"`
	ACME       *sbACME    `json:"acme,omitempty"`
	Reality    *sbReality `json:"reality,omitempty"`
}

type sbACME struct {
	Domain []string `json:"domain,omitempty"`
	Email  string   `json:"email,omitempty"`
}

type sbReality struct {
	Enabled    bool        `json:"enabled"`
	Handshake  sbHandshake `json:"handshake"`
	PrivateKey string      `json:"private_key"`
	ShortID    []string    `json:"short_id"`
}

type sbHandshake struct {
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
}

type sbTransport struct {
	Type        string `json:"type"`
	Path        string `json:"path,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
}
