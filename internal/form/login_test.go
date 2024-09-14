package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin_CleanEmail(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &Login{Email: "", Password: "passwd", Token: ""}
		assert.Equal(t, "", form.CleanEmail())
	})
	t.Run("valid", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Username: "John", Password: "passwd", Token: "123"}
		assert.Equal(t, "test@test.com", form.CleanEmail())
	})
}

func TestLogin_HasToken(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Username: "John", Password: "passwd", Token: ""}
		assert.Equal(t, false, form.HasShareToken())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Username: "John", Password: "passwd", Token: "123"}
		assert.Equal(t, true, form.HasShareToken())
	})
}

func TestLogin_HasUsername(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Password: "passwd", Token: ""}
		assert.Equal(t, false, form.HasUsername())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Username: "John", Password: "passwd", Token: "123"}
		assert.Equal(t, true, form.HasUsername())
	})
}

func TestLogin_CleanUsername(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Password: "passwd", Token: ""}
		assert.Equal(t, "", form.CleanUsername())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Username: " John", Password: "passwd", Token: "123"}
		assert.Equal(t, "john", form.CleanUsername())
	})
}

func TestLogin_HasPasscode(t *testing.T) {
	t.Run("False", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Code: "", Token: ""}
		assert.Equal(t, false, form.HasPasscode())
	})
	t.Run("True", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Username: "John", Code: "123456", Token: "123"}
		assert.Equal(t, true, form.HasPasscode())
	})
}

func TestLogin_Passcode(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &Login{Username: "admin", Password: "passwd1234", Code: "", Token: ""}
		assert.Equal(t, "", form.Passcode())
	})
	t.Run("Recovery", func(t *testing.T) {
		form := &Login{Username: "admin", Password: "passwd1234", Code: " A23 456 H7l pwf"}
		assert.Equal(t, "a23456h7lpwf", form.Passcode())
	})
	t.Run("Valid", func(t *testing.T) {
		form := &Login{Username: "admin", Password: "passwd1234", Code: "123456", Token: "123"}
		assert.Equal(t, "123456", form.Passcode())
	})
}

func TestLogin_HasPassword(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Password: "", Token: ""}
		assert.Equal(t, false, form.HasPassword())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Username: "John", Password: "passwd", Token: "123"}
		assert.Equal(t, true, form.HasPassword())
	})
}

func TestLogin_HasCredentials(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Password: "passwd123", Token: ""}
		assert.Equal(t, false, form.HasCredentials())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Username: "John", Password: "passwd", Token: "123"}
		assert.Equal(t, true, form.HasCredentials())
	})
}
