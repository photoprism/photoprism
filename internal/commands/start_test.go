package commands

import (
	"testing"

	"github.com/dustin/go-humanize/english"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
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

func TestWarnNonASCIIUsernames(t *testing.T) {
	requireTestDb(t)

	orig := event.SystemLog
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.SystemLog = logger
	t.Cleanup(func() { event.SystemLog = orig })

	base := entity.CountNonASCIIUsernames()

	if base == 0 {
		assert.Equal(t, 0, warnNonASCIIUsernames())
		assert.Nil(t, hook.LastEntry())
	}

	newExactTestUser(t, "warn-josé", false)

	assert.Equal(t, base+1, warnNonASCIIUsernames())
	require.NotNil(t, hook.LastEntry())
	assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
	assert.Equal(t, "users: "+english.Plural(base+1, "account", "accounts")+" with characters other than ASCII in the username › rename with photoprism users mod --username", hook.LastEntry().Message)
}
