package event

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHook(t *testing.T) {
	hub := NewHub()
	hook := NewHook(hub)

	assert.IsType(t, &Hook{hub: hub}, hook)
}

func TestHook_Fire(t *testing.T) {
	const hash = "2cad9168fa6acc5c5c2965ddf6ec465ca42fd818"

	// fire passes an entry to a hook on its own hub and returns the published message, if any.
	fire := func(t *testing.T, msg string) (published string, ok bool) {
		t.Helper()

		h := NewHub()
		s := h.Subscribe(10, "log.*")
		defer h.Unsubscribe(s)

		entry := &logrus.Entry{Level: logrus.ErrorLevel, Message: msg, Time: time.Now()}

		require.NoError(t, NewHook(h).Fire(entry))

		// The entry is not changed, so the console output keeps the full text.
		assert.Equal(t, msg, entry.Message)

		select {
		case m := <-s.Receiver:
			return m.Fields["message"].(string), true
		case <-time.After(time.Second):
			return "", false
		}
	}

	t.Run("MasksHashes", func(t *testing.T) {
		published, ok := fire(t, "thumb: "+hash+" not found in /cache/2/c/a/"+hash+"_720x720_fit.jpg")

		require.True(t, ok)
		assert.Equal(t, "thumb: 2ca*** not found in /cache/2/c/a/2ca***_720x720_fit.jpg", published)
	})
	t.Run("Unchanged", func(t *testing.T) {
		published, ok := fire(t, "media: failed to create tile_224 thumbnail for 2015/11/reunion.jpg")

		require.True(t, ok)
		assert.Equal(t, "media: failed to create tile_224 thumbnail for 2015/11/reunion.jpg", published)
	})
	t.Run("EmptyMessage", func(t *testing.T) {
		assert.Error(t, NewHook(NewHub()).Fire(&logrus.Entry{}))
	})
	t.Run("NoEntry", func(t *testing.T) {
		assert.Error(t, NewHook(NewHub()).Fire(nil))
	})
}
