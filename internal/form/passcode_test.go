package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasscode_HasPassword(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		form := &Passcode{Type: "totp", Password: "passwd1234", Code: "123456"}
		assert.Equal(t, true, form.HasPassword())
	})
	t.Run("False", func(t *testing.T) {
		form := &Passcode{Type: "totp", Password: "", Code: "123456"}
		assert.Equal(t, false, form.HasPassword())
	})
}

func TestPasscode_HasPasscode(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		form := &Passcode{Type: "totp", Password: "passwd1234", Code: "123456"}
		assert.Equal(t, true, form.HasPasscode())
	})
	t.Run("False", func(t *testing.T) {
		form := &Passcode{Type: "totp", Password: "passwd1234", Code: ""}
		assert.Equal(t, false, form.HasPasscode())
	})
}

func TestPasscode_Passcode(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		form := &Passcode{Type: "totp", Password: "passwd1234", Code: "123456"}
		assert.Equal(t, "123456", form.Passcode())
	})
	t.Run("Recovery", func(t *testing.T) {
		form := &Passcode{Type: "totp", Password: "passwd1234", Code: " A23 456 H7l pwf"}
		assert.Equal(t, "a23456h7lpwf", form.Passcode())
	})
	t.Run("Empty", func(t *testing.T) {
		form := &Passcode{Type: "totp", Password: "passwd1234", Code: ""}
		assert.Equal(t, "", form.Passcode())
	})
}
