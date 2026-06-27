package event

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/i18n"
)

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
