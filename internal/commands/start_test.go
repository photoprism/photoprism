package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
)

func TestStartCommand(t *testing.T) {
	t.Run("AuditRecordedThroughHub", func(t *testing.T) {
		requireTestDb(t)
		prev := event.SetAuditRecorder(func(event.Data) {})
		t.Cleanup(func() { event.SetAuditRecorder(prev) })

		output, err := RunWithTestContext(StartCommand, []string{"start", "--config"})
		require.NoError(t, err)
		assert.Contains(t, output, "http-port")
		assert.False(t, event.AuditRecordedSync())
	})
}
