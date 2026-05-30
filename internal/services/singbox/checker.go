package singbox

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Checker validates a rendered config with `sing-box check` before it is
// applied. A failing check must abort the apply so the live network is never
// brought down by an invalid config.
type Checker struct {
	binaryPath string
	timeout    time.Duration
}

func NewChecker(binaryPath string, timeout time.Duration) *Checker {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	return &Checker{binaryPath: binaryPath, timeout: timeout}
}

// Check runs `sing-box check -c <path>`. It returns nil when the config is
// valid, otherwise an error carrying the core's combined output.
func (c *Checker) Check(ctx context.Context, configPath string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.binaryPath, "check", "-c", configPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("sing-box check failed: %s", msg)
	}
	return nil
}
