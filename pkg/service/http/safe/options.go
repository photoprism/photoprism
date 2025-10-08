package safe

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"
)

// Options controls Download behavior.
type Options struct {
	Timeout      time.Duration
	MaxSizeBytes int64
	AllowPrivate bool
	Accept       string
}

var (
	// Defaults are tuned for general downloads (not just avatars).
	defaultTimeout = 30 * time.Second
	defaultMaxSize = int64(200 * 1024 * 1024) // 200 MiB

	ErrSchemeNotAllowed = errors.New("invalid scheme (only http/https allowed)")
	ErrSizeExceeded     = errors.New("response exceeds maximum allowed size")
	ErrPrivateIP        = errors.New("connection to private or loopback address not allowed")
)

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
