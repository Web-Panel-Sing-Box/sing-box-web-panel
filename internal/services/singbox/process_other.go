//go:build !linux

package singbox

import "context"

func externalProcessStatus(_ context.Context, _ ProcessConfig) (externalProcess, bool) {
	return externalProcess{}, false
}
