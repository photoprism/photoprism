package api

import (
	"errors"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/clean"
)

// TestLogVisionErr checks that vision request errors reach the system log with URI userinfo and queries redacted.
func TestLogVisionErr(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		hook := captureLog(t)
		systemHook := captureSystemLog(t)

		logVisionErr("labels", errors.New("Get \"https://user:pass@example.com/cat.jpg?w=224&sig=abc123\": 403 Forbidden\u0007"))

		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
		assert.Equal(t, "vision: labels request failed (details in system log)", hook.LastEntry().Message)

		require.Len(t, systemHook.AllEntries(), 1)
		assert.Equal(t, "vision: labels request failed › Get 'https://user:***@example.com/cat.jpg?***': 403 Forbidden", systemHook.LastEntry().Message)
	})
	t.Run("NoError", func(t *testing.T) {
		hook := captureLog(t)
		systemHook := captureSystemLog(t)

		logVisionErr("labels", nil)

		assert.Empty(t, hook.AllEntries())
		assert.Empty(t, systemHook.AllEntries())
	})
}

// TestPostVisionErrorLog checks the logs of vision requests that fail on a caller-supplied image URL.
func TestPostVisionErrorLog(t *testing.T) {
	app, router, _ := NewApiTest()
	PostVisionCaption(router)
	PostVisionLabels(router)
	PostVisionNsfw(router)
	PostVisionFace(router)

	body := `{"id":"3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8","images":["https://exa mple.com/cat.jpg?w=224"]}`

	for _, tc := range []struct {
		path   string
		action string
	}{
		{"caption", "caption"},
		{"labels", "labels"},
		{"nsfw", "nsfw"},
		{"face", "face image"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			hook := captureLog(t)
			systemHook := captureSystemLog(t)

			r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/"+tc.path, body)
			assert.Equal(t, http.StatusBadRequest, r.Code)

			require.Len(t, hook.AllEntries(), 1)
			assert.Equal(t, "vision: "+tc.action+" request failed (details in system log)", hook.LastEntry().Message)

			require.Len(t, systemHook.AllEntries(), 1)
			assert.Contains(t, systemHook.LastEntry().Message, "vision: "+tc.action+" request failed")
			assert.Contains(t, systemHook.LastEntry().Message, "invalid url")
			assert.Contains(t, systemHook.LastEntry().Message, "cat.jpg?"+clean.UriRedactedValue)
			assert.NotContains(t, systemHook.LastEntry().Message, "w=224")
		})
	}
}
