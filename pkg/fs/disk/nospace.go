package disk

import (
	"errors"
	"strings"
	"syscall"

	"github.com/photoprism/photoprism/pkg/log/status"
)

// IsNoSpace reports whether err indicates the filesystem is out of space.
// It matches both the ENOSPC errno returned by direct file writes and the
// "no space left on device" text that external converters surface via stderr,
// since their errors reach us as plain strings without the original errno.
func IsNoSpace(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, syscall.ENOSPC) {
		return true
	}

	return strings.Contains(strings.ToLower(err.Error()), "no space left on device")
}

// AsInsufficientStorage maps an out-of-space write error to status.ErrInsufficientStorage
// and flushes the cached free-space probe so the next check reflects the full disk;
// nil and unrelated errors are returned unchanged.
func AsInsufficientStorage(err error) error {
	if IsNoSpace(err) {
		FlushFree()
		return status.ErrInsufficientStorage
	}

	return err
}
