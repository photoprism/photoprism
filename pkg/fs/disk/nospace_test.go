package disk

import (
	"errors"
	"fmt"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/log/status"
)

func TestIsNoSpace(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.False(t, IsNoSpace(nil))
	})
	t.Run("Errno", func(t *testing.T) {
		assert.True(t, IsNoSpace(syscall.ENOSPC))
	})
	t.Run("WrappedErrno", func(t *testing.T) {
		assert.True(t, IsNoSpace(fmt.Errorf("convert: failed (%w)", syscall.ENOSPC)))
	})
	t.Run("StringMatch", func(t *testing.T) {
		assert.True(t, IsNoSpace(errors.New("vips2png: unable to write to target: No space left on device")))
	})
	t.Run("Unrelated", func(t *testing.T) {
		assert.False(t, IsNoSpace(errors.New("permission denied")))
	})
}

func TestAsInsufficientStorage(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Nil(t, AsInsufficientStorage(nil))
	})
	t.Run("NoSpaceMapsToSentinel", func(t *testing.T) {
		// Seed the cache so we can confirm the flush.
		SetFree("/tmp/disk-nospace-test", 1<<30, 1<<40)
		err := AsInsufficientStorage(syscall.ENOSPC)
		assert.ErrorIs(t, err, status.ErrInsufficientStorage)

		freeMu.RLock()
		_, cached := freeCache["/tmp/disk-nospace-test"]
		freeMu.RUnlock()
		assert.False(t, cached, "free-space cache should be flushed after an out-of-space error")
	})
	t.Run("Unrelated", func(t *testing.T) {
		orig := errors.New("permission denied")
		assert.Equal(t, orig, AsInsufficientStorage(orig))
	})
}
