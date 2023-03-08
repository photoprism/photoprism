package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin_HasToken(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", AuthToken: ""}
		assert.Equal(t, false, form.HasToken())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", AuthToken: "123"}
		assert.Equal(t, true, form.HasToken())
	})
}

func TestLogin_HasName(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", Password: "passwd", AuthToken: ""}
		assert.Equal(t, false, form.HasUsername())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", AuthToken: "123"}
		assert.Equal(t, true, form.HasUsername())
	})
}

func TestLogin_HasPassword(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", Password: "", AuthToken: ""}
		assert.Equal(t, false, form.HasPassword())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", AuthToken: "123"}
		assert.Equal(t, true, form.HasPassword())
	})
}

func TestLogin_HasCredentials(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", Password: "passwd123", AuthToken: ""}
		assert.Equal(t, false, form.HasCredentials())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{UserEmail: "test@test.com", UserName: "John", Password: "passwd", AuthToken: "123"}
		assert.Equal(t, true, form.HasCredentials())
	})
}
