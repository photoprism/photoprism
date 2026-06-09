package photoprism

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/log/status"
)

func TestWalkResultLog(t *testing.T) {
	t.Run("NilError", func(t *testing.T) {
		level, msg, emit := walkResultLog("index", nil)
		assert.False(t, emit)
		assert.Equal(t, logrus.Level(0), level)
		assert.Empty(t, msg)
	})
	t.Run("Canceled", func(t *testing.T) {
		level, msg, emit := walkResultLog("index", status.ErrCanceled)
		assert.True(t, emit)
		assert.Equal(t, logrus.InfoLevel, level)
		assert.Equal(t, "index: canceled", msg)
	})
	t.Run("WrappedCanceled", func(t *testing.T) {
		level, _, emit := walkResultLog("index", fmt.Errorf("worker: %w", status.ErrCanceled))
		assert.True(t, emit)
		assert.Equal(t, logrus.InfoLevel, level)
	})
	t.Run("InsufficientStorage", func(t *testing.T) {
		// Suppressed because the storage helper already emitted the actionable line.
		_, _, emit := walkResultLog("import", status.ErrInsufficientStorage)
		assert.False(t, emit)
	})
	t.Run("WrappedInsufficientStorage", func(t *testing.T) {
		_, _, emit := walkResultLog("import", fmt.Errorf("walk: %w", status.ErrInsufficientStorage))
		assert.False(t, emit)
	})
	t.Run("UnknownError", func(t *testing.T) {
		level, msg, emit := walkResultLog("import", errors.New("disk vanished"))
		assert.True(t, emit)
		assert.Equal(t, logrus.ErrorLevel, level)
		assert.Equal(t, "import: disk vanished", msg)
	})
}
