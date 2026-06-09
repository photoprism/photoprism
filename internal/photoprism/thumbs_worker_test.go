package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/mutex"
)

func TestCancelThumbsInsufficientStorage(t *testing.T) {
	t.Run("CancelsRunningWorker", func(t *testing.T) {
		require.NoError(t, mutex.IndexWorker.Start())
		defer mutex.IndexWorker.Stop()
		require.False(t, mutex.IndexWorker.Canceled())

		cancelThumbsInsufficientStorage()
		assert.True(t, mutex.IndexWorker.Canceled())
	})
	t.Run("NoopWhenNotRunning", func(t *testing.T) {
		// Cancel only takes effect while the worker is running, so an idle worker stays uncanceled.
		require.False(t, mutex.IndexWorker.Canceled())

		cancelThumbsInsufficientStorage()
		assert.False(t, mutex.IndexWorker.Canceled())
	})
}
