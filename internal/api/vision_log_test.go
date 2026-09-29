package api

import (
	"errors"
	"net/http"
	"strings"
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
	// shrinking returns text of about n bytes that redaction shortens to less than a tenth of that.
	shrinking := func(n int) string {
		item := "https://example.com/cat.jpg?q=" + strings.Repeat("1", 200) + " "
		return strings.Repeat(item, n/len(item))
	}
	t.Run("PartialUriBeforeRedaction", func(t *testing.T) {
		captureLog(t)
		systemHook := captureSystemLog(t)

		// The last URI begins just before 8 KiB.
		prefix := shrinking(2*clean.LengthLimit - 200)
		prefix += strings.Repeat("x ", (2*clean.LengthLimit-len(prefix)-len("https://user:pa"))/2)
		logVisionErr("labels", errors.New(prefix+"https://user:pass@example.com/cat.jpg"))

		require.Len(t, systemHook.AllEntries(), 1)
		msg := systemHook.LastEntry().Message
		assert.Contains(t, msg, "https://user:***@example.com/cat.jpg")
		assert.NotContains(t, msg, "user:pa")
	})
	t.Run("PartialUriAtFirstCut", func(t *testing.T) {
		captureLog(t)
		systemHook := captureSystemLog(t)

		// The first cut falls within the password of the last URI, which redaction moves within the second.
		prefix := shrinking(4*clean.LengthLimit - 200)
		prefix += strings.Repeat("x ", (4*clean.LengthLimit-len(prefix)-len("https://user:pa"))/2)
		logVisionErr("labels", errors.New(prefix+"https://user:pass@example.com/cat.jpg"))

		require.Len(t, systemHook.AllEntries(), 1)
		msg := systemHook.LastEntry().Message
		assert.NotContains(t, msg, "user:")
		assert.LessOrEqual(t, len(msg), clean.LengthLimit+64)
	})
	t.Run("PartialUriWithApostrophe", func(t *testing.T) {
		captureLog(t)
		systemHook := captureSystemLog(t)

		// The first cut falls after an apostrophe in the password of the last URI.
		prefix := shrinking(4*clean.LengthLimit - 200)
		prefix += strings.Repeat("x ", (4*clean.LengthLimit-len(prefix)-len(`Get "https://user:pa'ss`))/2)
		logVisionErr("labels", errors.New(prefix+`Get "https://user:pa'ssw0rd@example.com/cat.jpg"`))

		require.Len(t, systemHook.AllEntries(), 1)
		assert.NotContains(t, systemHook.LastEntry().Message, "user:")
	})
	t.Run("PartialUriWithNonAsciiSpace", func(t *testing.T) {
		captureLog(t)
		systemHook := captureSystemLog(t)

		// The first cut falls after a no-break space in the password of the last URI.
		prefix := shrinking(4*clean.LengthLimit - 200)
		prefix += strings.Repeat("x ", (4*clean.LengthLimit-len(prefix)-len("Get \"https://user:pa\u00a0ss"))/2)
		logVisionErr("labels", errors.New(prefix+"Get \"https://user:pa\u00a0ssw0rd@example.com/cat.jpg\""))

		require.Len(t, systemHook.AllEntries(), 1)
		assert.NotContains(t, systemHook.LastEntry().Message, "user:")
	})
	t.Run("CutInRedactedUri", func(t *testing.T) {
		captureLog(t)
		systemHook := captureSystemLog(t)

		// The final cut falls within the redacted userinfo of the last URI.
		filler := strings.Repeat("x ", (clean.LengthLimit-len("https://user:*"))/2)
		logVisionErr("labels", errors.New(filler+"https://user:pass@example.com/cat.jpg?sig=abc123"))

		require.Len(t, systemHook.AllEntries(), 1)
		msg := systemHook.LastEntry().Message
		assert.NotContains(t, msg, "https:")
		assert.NotContains(t, msg, "pass")
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

func TestClipTokens(t *testing.T) {
	t.Run("Short", func(t *testing.T) {
		assert.Equal(t, "a b", clipTokens("a b", 3))
	})
	t.Run("PartialToken", func(t *testing.T) {
		assert.Equal(t, "get", clipTokens("get https://example.com/", 12))
		assert.Equal(t, "get ", clipTokens(`get "https://example.com/"`, 12))
		assert.Equal(t, "get", clipTokens("get https://user:pa'ss@example.com/", 20))
	})
	t.Run("Boundary", func(t *testing.T) {
		assert.Equal(t, "get url", clipTokens("get url more", 8))
		assert.Equal(t, "get url", clipTokens("get url more", 7))
	})
	t.Run("NonAsciiSpace", func(t *testing.T) {
		assert.Equal(t, "get", clipTokens("get https://user:pa\u00a0ss@example.com/", 20))
		assert.Equal(t, "get", clipTokens("get https://user:pa\vss@example.com/", 20))
	})
	t.Run("SingleToken", func(t *testing.T) {
		assert.Equal(t, "", clipTokens("https://user:pass@example.com/", 12))
	})
}
