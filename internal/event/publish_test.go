package event

import (
	"bytes"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// captureLog redirects the logger to a buffer at info level and restores it afterwards.
func captureLog(t *testing.T) *bytes.Buffer {
	buf := &bytes.Buffer{}
	logger := logrus.New()
	logger.SetOutput(buf)
	logger.SetLevel(logrus.InfoLevel)

	orig := Log
	Log = logger
	t.Cleanup(func() { Log = orig })

	return buf
}

func TestPublishMsg(t *testing.T) {
	t.Run("LogStaysEnglish", func(t *testing.T) {
		// Server logs are read in a terminal and under /library/logs, so they must not follow
		// the instance locale even when the published notification does.
		i18n.SetLocale("he")
		t.Cleanup(func() { i18n.SetLocale(string(i18n.Default)) })

		buf := captureLog(t)
		s := Subscribe("notify.success")
		SuccessMsg(i18n.MsgIndexingCompletedIn, 11)
		msg := <-s.Receiver
		Unsubscribe(s)

		// Guards the case itself: without a loaded catalog the payload would render English
		// and the log assertion below would hold for the wrong reason.
		require.NotEqual(t, "Indexing completed in 11 s", msg.Fields["message"])
		assert.Contains(t, buf.String(), "indexing completed in 11 s")
		assert.Equal(t, "Indexing completed in %d s", msg.Fields["messageId"])
		assert.Equal(t, []any{11}, msg.Fields["messageParams"])
	})
	t.Run("DefaultLocale", func(t *testing.T) {
		buf := captureLog(t)
		s := Subscribe("notify.warning")
		WarnMsg(i18n.ErrBusy)
		<-s.Receiver
		Unsubscribe(s)

		assert.Contains(t, buf.String(), "busy, please try again later")
	})
}

func TestNotifyMsg(t *testing.T) {
	t.Run("WithParams", func(t *testing.T) {
		buf := captureLog(t)
		s := Subscribe("notify.info")
		notifyMsg("notify.info", i18n.MsgIndexingFiles, "/photos")
		msg := <-s.Receiver
		Unsubscribe(s)

		assert.Equal(t, "notify.info", msg.Name)
		assert.Equal(t, "Indexing files in /photos", msg.Fields["message"])
		assert.Equal(t, "Indexing files in %s", msg.Fields["messageId"])
		assert.Equal(t, []any{"/photos"}, msg.Fields["messageParams"])
		assert.Empty(t, buf.String())
	})
	t.Run("WithoutParams", func(t *testing.T) {
		buf := captureLog(t)
		s := Subscribe("notify.error")
		notifyMsg("notify.error", i18n.ErrBusy)
		msg := <-s.Receiver
		Unsubscribe(s)

		assert.Equal(t, "notify.error", msg.Name)
		assert.Equal(t, "Busy, please try again later", msg.Fields["message"])
		assert.Equal(t, "Busy, please try again later", msg.Fields["messageId"])
		assert.Empty(t, msg.Fields["messageParams"])
		assert.Empty(t, buf.String())
	})
}

func TestPublishSuccessMsg(t *testing.T) {
	t.Run("WithParams", func(t *testing.T) {
		buf := captureLog(t)
		s := Subscribe("notify.success")
		PublishSuccessMsg(i18n.MsgIndexingCompletedIn, 11)
		msg := <-s.Receiver
		Unsubscribe(s)

		assert.Equal(t, "notify.success", msg.Name)
		assert.Equal(t, "Indexing completed in 11 s", msg.Fields["message"])
		assert.Equal(t, "Indexing completed in %d s", msg.Fields["messageId"])
		assert.Equal(t, []any{11}, msg.Fields["messageParams"])
		assert.Empty(t, buf.String())
	})
	t.Run("WithoutParams", func(t *testing.T) {
		buf := captureLog(t)
		s := Subscribe("notify.success")
		PublishSuccessMsg(i18n.MsgAlbumCreated)
		msg := <-s.Receiver
		Unsubscribe(s)

		assert.Equal(t, "Album created", msg.Fields["message"])
		assert.Equal(t, "Album created", msg.Fields["messageId"])
		assert.Empty(t, buf.String())
	})
}

func TestSuccessMsg(t *testing.T) {
	t.Run("WithParams", func(t *testing.T) {
		s := Subscribe("notify.success")
		SuccessMsg(i18n.MsgAlbumDeleted, "Holiday")
		msg := <-s.Receiver
		assert.Equal(t, "notify.success", msg.Name)
		assert.Equal(t, "Album Holiday deleted", msg.Fields["message"])
		assert.Equal(t, "Album %s deleted", msg.Fields["messageId"])
		assert.Equal(t, []any{"Holiday"}, msg.Fields["messageParams"])
		Unsubscribe(s)
	})
	t.Run("WithoutParams", func(t *testing.T) {
		s := Subscribe("notify.success")
		SuccessMsg(i18n.MsgAlbumCreated)
		msg := <-s.Receiver
		assert.Equal(t, "notify.success", msg.Name)
		assert.Equal(t, "Album created", msg.Fields["message"])
		assert.Equal(t, "Album created", msg.Fields["messageId"])
		Unsubscribe(s)
	})
}

func TestErrorMsg(t *testing.T) {
	s := Subscribe("notify.error")
	ErrorMsg(i18n.ErrAlreadyExists, "A cat")
	msg := <-s.Receiver
	assert.Equal(t, "notify.error", msg.Name)
	assert.Equal(t, "A cat already exists", msg.Fields["message"])
	assert.Equal(t, "%s already exists", msg.Fields["messageId"])
	assert.Equal(t, []any{"A cat"}, msg.Fields["messageParams"])
	Unsubscribe(s)
}

func TestInfoMsg(t *testing.T) {
	s := Subscribe("notify.info")
	InfoMsg(i18n.MsgIndexingFiles, "/photos")
	msg := <-s.Receiver
	assert.Equal(t, "notify.info", msg.Name)
	assert.Equal(t, "Indexing files in /photos", msg.Fields["message"])
	assert.Equal(t, "Indexing files in %s", msg.Fields["messageId"])
	assert.Equal(t, []any{"/photos"}, msg.Fields["messageParams"])
	Unsubscribe(s)
}

func TestWarnMsg(t *testing.T) {
	s := Subscribe("notify.warning")
	WarnMsg(i18n.ErrBusy)
	msg := <-s.Receiver
	assert.Equal(t, "notify.warning", msg.Name)
	assert.Equal(t, "Busy, please try again later", msg.Fields["message"])
	assert.Equal(t, "Busy, please try again later", msg.Fields["messageId"])
	Unsubscribe(s)
}

func TestPublishCompleted(t *testing.T) {
	// Each case subscribes before publishing, because the hub delivers to current subscribers.
	collect := func(topics []string, want int, fn func()) []Message {
		sub := Subscribe(topics...)
		defer Unsubscribe(sub)

		fn()

		var got []Message

		for range want {
			select {
			case msg := <-sub.Receiver:
				got = append(got, msg)
			case <-time.After(2 * time.Second):
				return got
			}
		}

		return got
	}

	t.Run("Fields", func(t *testing.T) {
		got := collect([]string{"upload.completed"}, 1, func() {
			PublishCompleted([]string{"upload.completed"}, "urqjjgt71ap1w2gr", "", 7)
		})
		require.Len(t, got, 1)
		assert.Equal(t, "urqjjgt71ap1w2gr", got[0].Fields["uid"])
		assert.Equal(t, 7, got[0].Fields["seconds"])
		assert.NotContains(t, got[0].Fields, "action")
		// No path: nothing reads one, and topic routing reaches every permitted session
		// rather than the one that acted.
		assert.NotContains(t, got[0].Fields, "path")
	})
	t.Run("Action", func(t *testing.T) {
		got := collect([]string{"index.completed"}, 1, func() {
			PublishCompleted([]string{"index.completed"}, "urqjjgt71ap1w2gr", "index", 3)
		})
		require.Len(t, got, 1)
		assert.Equal(t, "index", got[0].Fields["action"])
	})
	t.Run("EveryTopicSamePayload", func(t *testing.T) {
		topics := []string{"import.completed", "index.completed", "upload.completed"}
		got := collect(topics, 3, func() {
			PublishCompleted(topics, "urqjjgt71ap1w2gr", "import", 1)
		})
		require.Len(t, got, 3)

		seen := map[string]bool{}

		for _, msg := range got {
			seen[msg.Name] = true
			// Asserted per message, not once: a payload that differs per topic would
			// otherwise pass while only the first carried the fields.
			assert.Equal(t, "urqjjgt71ap1w2gr", msg.Fields["uid"], msg.Name)
			assert.Equal(t, "import", msg.Fields["action"], msg.Name)
			assert.Equal(t, 1, msg.Fields["seconds"], msg.Name)
		}

		assert.Len(t, seen, 3)
	})
}

func TestMessageSepIsQuotedByClean(t *testing.T) {
	// The sanitizer quotes a value holding this character. Format joins fields with it, so the
	// two have to name the same one or a value could read as several fields unquoted.
	assert.Contains(t, MessageSep, string(clean.FieldSep))
}
