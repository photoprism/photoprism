package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin_HasToken(t *testing.T) {
	t.Run("false", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Username: "John", Password: "passwd", Token: ""}
		assert.Equal(t, false, form.HasToken())
	})
	t.Run("true", func(t *testing.T) {
		form := &Login{Email: "test@test.com", Username: "John", Password: "passwd", Token: "123"}
		assert.Equal(t, true, form.HasToken())
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
