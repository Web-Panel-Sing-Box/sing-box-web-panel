package version

import "strings"

var (
	// Version is replaced by release builds through -ldflags.
	Version = "dev"
)

// Panel returns the public Shilka panel version.
func Panel() string {
	v := strings.TrimSpace(Version)
	if v == "" {
		return "dev"
	}
	return v
}
