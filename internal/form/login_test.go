package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin_Email(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		form := &Login{UserEmail: "", Password: "passwd", ShareToken: ""}
		assert.Equal(t, "", form.Email())
	})
	t.Run("valid", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", ShareToken: "123"}
		assert.Equal(t, "test@test.com", form.Email())
	})
}

func TestLogin_HasToken(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", ShareToken: ""}
		assert.Equal(t, false, form.HasShareToken())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", ShareToken: "123"}
		assert.Equal(t, true, form.HasShareToken())
	})
}

func TestLogin_HasName(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", Password: "passwd", ShareToken: ""}
		assert.Equal(t, false, form.HasUsername())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", ShareToken: "123"}
		assert.Equal(t, true, form.HasUsername())
	})
}

func TestLogin_HasPasscode(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", Passcode: "", ShareToken: ""}
		assert.Equal(t, false, form.HasPasscode())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Passcode: "123456", ShareToken: "123"}
		assert.Equal(t, true, form.HasPasscode())
	})
}

func TestLogin_HasPassword(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", Password: "", ShareToken: ""}
		assert.Equal(t, false, form.HasPassword())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", ShareToken: "123"}
		assert.Equal(t, true, form.HasPassword())
	})
}

func TestLogin_HasCredentials(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", Password: "passwd123", ShareToken: ""}
		assert.Equal(t, false, form.HasCredentials())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", ShareToken: "123"}
		assert.Equal(t, true, form.HasCredentials())
	})
}
