package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientToken_Empty(t *testing.T) {
	t.Run("AuthTokenAndTypeHintEmpty", func(t *testing.T) {
		m := ClientToken{
			AuthToken: "",
			TypeHint:  "",
		}
		assert.True(t, m.Empty())
	})
	t.Run("AuthTokenNotEmpty", func(t *testing.T) {
		m := ClientToken{
			AuthToken: "abc",
			TypeHint:  "",
		}
		assert.False(t, m.Empty())
	})
	t.Run("TypeHintNotEmpty", func(t *testing.T) {
		m := ClientToken{
			AuthToken: "",
			TypeHint:  "test",
		}
		assert.False(t, m.Empty())
	})
}

func TestClientToken_Validate(t *testing.T) {
	t.Run("AuthTokenEmpty", func(t *testing.T) {
		m := ClientToken{
			AuthToken: "",
			TypeHint:  "test",
		}
		assert.Error(t, m.Validate())
	})
	t.Run("AuthTokenInvalid", func(t *testing.T) {
		m := ClientToken{
			AuthToken: "abc   234",
			TypeHint:  "test",
		}
		assert.Error(t, m.Validate())
	})
	t.Run("UnsupportedToken", func(t *testing.T) {
		m := ClientToken{
			AuthToken: "abc234",
			TypeHint:  "test",
		}
		assert.Error(t, m.Validate())
	})
	t.Run("Valid", func(t *testing.T) {
		m := ClientToken{
			AuthToken: "abc234",
			TypeHint:  "access_token",
		}
		assert.NoError(t, m.Validate())
	})
}
