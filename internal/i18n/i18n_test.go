package i18n

import (
	"os"
	"testing"

	"github.com/leonelquinteros/gotext"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gotext.Configure(localeDir, string(locale), "default")

	code := m.Run()

	os.Exit(code)
}

func TestMsg(t *testing.T) {
	t.Run("already exists", func(t *testing.T) {
		msg := Msg(ErrAlreadyExists, "A cat")
		assert.Equal(t, "A cat already exists", msg)
	})

	t.Run("unexpected error", func(t *testing.T) {
		msg := Msg(ErrUnexpected, "A cat")
		assert.Equal(t, "Something went wrong, try again", msg)
	})

	t.Run("already exists german", func(t *testing.T) {
		SetLocale("de")
		msgTrans := Msg(ErrAlreadyExists, "Eine Katze")
		assert.Equal(t, "Eine Katze existiert bereits", msgTrans)
		SetLocale("")
		msgDefault := Msg(ErrAlreadyExists, "A cat")
		assert.Equal(t, "A cat already exists", msgDefault)
	})

	t.Run("already exists polish", func(t *testing.T) {
		SetLocale("pl")
		msgTrans := Msg(ErrAlreadyExists, "Kot")
		assert.Equal(t, "Kot już istnieje", msgTrans)
		SetLocale("")
		msgDefault := Msg(ErrAlreadyExists, "A cat")
		assert.Equal(t, "A cat already exists", msgDefault)
	})

	t.Run("Brazilian Portuguese", func(t *testing.T) {
		SetLocale("pt_BR")
		msgTrans := Msg(ErrAlreadyExists, "Gata")
		assert.Equal(t, "Gata já existe", msgTrans)
		SetLocale("")
		msgDefault := Msg(ErrAlreadyExists, "A cat")
		assert.Equal(t, "A cat already exists", msgDefault)
	})
}

func TestError(t *testing.T) {
	t.Run("already exists", func(t *testing.T) {
		err := Error(ErrAlreadyExists, "A cat")
		assert.EqualError(t, err, "A cat already exists")
	})

	t.Run("unexpected error", func(t *testing.T) {
		err := Error(ErrUnexpected, "A cat")
		assert.EqualError(t, err, "Something went wrong, try again")
	})

	t.Run("already exists german", func(t *testing.T) {
		SetLocale("de")
		errGerman := Error(ErrAlreadyExists, "Eine Katze")
		assert.EqualError(t, errGerman, "Eine Katze existiert bereits")
		SetLocale("")
		errDefault := Error(ErrAlreadyExists, "A cat")
		assert.EqualError(t, errDefault, "A cat already exists")
	})
}

func TestLower(t *testing.T) {
	t.Run("ErrAlreadyExists", func(t *testing.T) {
		msg := Lower(ErrAlreadyExists, "Eine Katze")
		assert.Equal(t, "eine katze already exists", msg)
		SetLocale("de")
		errGerman := Lower(ErrAlreadyExists, "Eine Katze")
		assert.Equal(t, errGerman, "eine katze already exists")
		SetLocale("")
		errDefault := Lower(ErrAlreadyExists, "Eine Katze")
		assert.Equal(t, errDefault, "eine katze already exists")
	})

	t.Run("ErrForbidden", func(t *testing.T) {
		msg := Lower(ErrForbidden, "A cat")
		assert.Equal(t, "permission denied", msg)
	})
}
