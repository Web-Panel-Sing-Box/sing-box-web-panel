package domain

// Setting is a single key/value panel setting row. Values are stored as plain
// strings (JSON-encoded when structured).
type Setting struct {
	Key   string
	Value string
}

// Panel setting keys.
const (
	SettingPanelName    = "panel_name"
	SettingBinaryPath   = "binary_path"
	SettingLogLevel     = "log_level"
	SettingSubPublicURL = "sub_public_url"
	SettingInboundHost  = "inbound_host" // host/IP advertised in generated client links
	SettingTokenTTL     = "token_ttl"
)
