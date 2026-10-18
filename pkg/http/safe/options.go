package safe

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Options controls Download and Fetch behavior.
type Options struct {
	Timeout      time.Duration
	MaxSizeBytes int64
	AllowPrivate bool
	Accept       string
	NoRedirects  bool
}

var (
	// Defaults are tuned for general downloads (not just avatars).
	defaultTimeout = 30 * time.Second
	defaultMaxSize = int64(200 * 1024 * 1024) // 200 MiB

	// ErrSchemeNotAllowed is returned when a URL scheme is not permitted.
	ErrSchemeNotAllowed = errors.New("invalid scheme (only http/https allowed)")
	// ErrSizeExceeded is returned when a response exceeds the configured limit.
	ErrSizeExceeded = errors.New("response exceeds maximum allowed size")
	// ErrRedirect is returned when a response redirects and redirects are not allowed.
	ErrRedirect = errors.New("redirect not allowed")
	// ErrPrivateIP is returned when the target resolves to a private or loopback address.
	ErrPrivateIP = errors.New("connection to private or loopback address not allowed")
)

// StatusError is returned by Fetch when a response has a status other than 200 OK.
type StatusError struct {
	Code int
}

// Error returns the error message.
func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected status %d", e.Code)
}

// envInt64 returns an int64 from env or -1 if unset/invalid.
func envInt64(key string) int64 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return -1
}

// envDuration returns a duration from env seconds or 0 if unset/invalid.
func envDuration(key string) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Duration(n) * time.Second
		}
	}
	return 0
}
