package api

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
)

func TestWebsocketTopicsExcludeSystem(t *testing.T) {
	// Operator-only detail, including server paths, is reported on the system channel
	// because no client subscribes to it. Two properties keep that true.
	t.Run("NoSubscription", func(t *testing.T) {
		for _, topic := range WebsocketTopics {
			require.False(t, strings.HasPrefix(topic, string(acl.ChannelSystem)+"."),
				"websocket topic %q would forward the system channel to clients", topic)
		}
	})
	t.Run("NoHooks", func(t *testing.T) {
		// A hook on the system logger would republish its entries as "log.*", which is
		// subscribed. The ordinary logger carries one by design.
		systemLog, ok := event.SystemLog.(*logrus.Logger)
		require.True(t, ok)
		require.Empty(t, systemLog.Hooks)
	})
}

func TestAppendWebsocketTopics(t *testing.T) {
	original := append([]string(nil), WebsocketTopics...)

	t.Cleanup(func() {
		WebsocketTopics = original
	})

	AppendWebsocketTopics("audit.log.*", "custom.topic")

	require.Len(t, WebsocketTopics, len(original)+2)
	require.Contains(t, WebsocketTopics, "audit.log.*")
	require.Contains(t, WebsocketTopics, "custom.topic")
}
