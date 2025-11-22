package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
