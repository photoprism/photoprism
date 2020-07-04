package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMsg(t *testing.T) {
	t.Run("already exists", func(t *testing.T) {
		msg := Msg(ErrAlreadyExists, "A cat")
		assert.Equal(t, "A cat already exists", msg)
	})

	t.Run("unexpected error", func(t *testing.T) {
		msg := Msg(ErrUnexpected, "A cat")
		assert.Equal(t, "Unexpected error, please try again", msg)
	})

	t.Run("already exists german", func(t *testing.T) {
		SetLang("de")
		msgGerman := Msg(ErrAlreadyExists, "Eine Katze")
		assert.Equal(t, "Eine Katze existiert bereits", msgGerman)
		SetLang("")
		msgDefault := Msg(ErrAlreadyExists, "A cat")
		assert.Equal(t, "A cat already exists", msgDefault)
	})
}

func TestLangMsg(t *testing.T) {
	t.Run("already exists", func(t *testing.T) {
		msgDefault := LangMsg(ErrAlreadyExists, Default, "A cat")
		assert.Equal(t, "A cat already exists", msgDefault)
		msgEnglish := LangMsg(ErrAlreadyExists, English, "A cat")
		assert.Equal(t, msgEnglish, msgDefault)
	})

	t.Run("unexpected error", func(t *testing.T) {
		msgDefault := LangMsg(ErrUnexpected, Default, "A cat")
		assert.Equal(t, "Unexpected error, please try again", msgDefault)
		msgEnglish := LangMsg(ErrUnexpected, English, "A cat")
		assert.Equal(t, msgEnglish, msgDefault)
	})

	t.Run("already exists german", func(t *testing.T) {
		msg := LangMsg(ErrAlreadyExists, German, "Eine Katze")
		assert.Equal(t, "Eine Katze existiert bereits", msg)
	})

	t.Run("unexpected error german", func(t *testing.T) {
		msg := LangMsg(ErrUnexpected, German, "Eine Katze")
		assert.Equal(t, "Unerwarteter Fehler, bitte erneut versuchen", msg)
	})
}

func TestDefaultMsg(t *testing.T) {
	t.Run("already exists", func(t *testing.T) {
		msg := DefaultMsg(ErrAlreadyExists, "A cat")
		assert.Equal(t, "A cat already exists", msg)
	})

	t.Run("unexpected error", func(t *testing.T) {
		msg := DefaultMsg(ErrUnexpected, "A cat")
		assert.Equal(t, "Unexpected error, please try again", msg)
	})
}

func TestError(t *testing.T) {
	t.Run("already exists", func(t *testing.T) {
		err := Error(ErrAlreadyExists, "A cat")
		assert.EqualError(t, err, "A cat already exists")
	})

	t.Run("unexpected error", func(t *testing.T) {
		err := Error(ErrUnexpected, "A cat")
		assert.EqualError(t, err, "Unexpected error, please try again")
	})

	t.Run("already exists german", func(t *testing.T) {
		SetLang("de")
		errGerman := Error(ErrAlreadyExists, "Eine Katze")
		assert.EqualError(t, errGerman, "Eine Katze existiert bereits")
		SetLang("")
		errDefault := Error(ErrAlreadyExists, "A cat")
		assert.EqualError(t, errDefault, "A cat already exists")
	})
}

func TestLangError(t *testing.T) {
	t.Run("already exists", func(t *testing.T) {
		err := LangError(ErrAlreadyExists, English, "A cat")
		assert.EqualError(t, err, "A cat already exists")
	})

	t.Run("unexpected error", func(t *testing.T) {
		err := LangError(ErrUnexpected, English, "A cat")
		assert.EqualError(t, err, "Unexpected error, please try again")
	})

	t.Run("already exists german", func(t *testing.T) {
		err := LangError(ErrAlreadyExists, German, "Eine Katze")
		assert.EqualError(t, err, "Eine Katze existiert bereits")
	})

	t.Run("unexpected error german", func(t *testing.T) {
		err := LangError(ErrUnexpected, German, "Eine Katze")
		assert.EqualError(t, err, "Unerwarteter Fehler, bitte erneut versuchen")
	})
}

func TestDefaultError(t *testing.T) {
	t.Run("already exists", func(t *testing.T) {
		err := DefaultError(ErrAlreadyExists, "A cat")
		assert.EqualError(t, err, "A cat already exists")
	})

	t.Run("unexpected error", func(t *testing.T) {
		err := DefaultError(ErrUnexpected, "A cat")
		assert.EqualError(t, err, "Unexpected error, please try again")
	})
}
