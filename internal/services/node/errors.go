package node

import (
	"context"
	"errors"
	"net"
	"syscall"
)

// ErrNodeUnreachable marks a failure to reach the remote node (network/timeout),
// as opposed to an internal panel error or a remote node that answered with an
// HTTP error status (see ErrRemote / RemoteHTTPError).
var ErrNodeUnreachable = errors.New("node unreachable")

// UnreachableError wraps a transport-level failure with a human-friendly reason.
// Timeout selects 504 vs 502 at the HTTP layer. The underlying error is kept for
// errors.Is/As traversal and debugging only; it never carries the bearer token,
// which lives in a request header rather than in the URL or net error.
type UnreachableError struct {
	Detail     string // "timeout" | "connection refused" | "dns lookup failed" | ...
	Timeout    bool   // true → 504 Gateway Timeout, false → 502 Bad Gateway
	underlying error
}

func (e *UnreachableError) Error() string { return "node unreachable: " + e.Detail }

func (e *UnreachableError) Unwrap() error { return e.underlying }

func (e *UnreachableError) Is(target error) bool { return target == ErrNodeUnreachable }

// classifyTransportError converts a raw client.Do error into an UnreachableError
// when it is a network/timeout failure; otherwise it returns the error unchanged.
func classifyTransportError(err error) error {
	if err == nil {
		return nil
	}
	// Config/safety block — not a transport failure. Keep as-is so existing
	// errors.Is(err, ErrUnsafeAddress) callers and tests keep working.
	if errors.Is(err, ErrUnsafeAddress) {
		return err
	}
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return &UnreachableError{Detail: "timeout", Timeout: true, underlying: err}
	}
	var dnsErr *net.DNSError
	switch {
	case errors.As(err, &dnsErr):
		return &UnreachableError{Detail: "dns lookup failed", underlying: err}
	case errors.Is(err, syscall.ECONNREFUSED):
		return &UnreachableError{Detail: "connection refused", underlying: err}
	case errors.Is(err, syscall.ECONNRESET):
		return &UnreachableError{Detail: "connection reset", underlying: err}
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return &UnreachableError{Detail: "connection failed", underlying: err}
	}
	return err // not a transport error → unchanged (falls through to 500)
}
