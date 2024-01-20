package form

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientCredentials_Validate(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		m := ClientCredentials{
			ClientID:     "cs5gfen1bgxz7s9i",
			ClientSecret: "abc",
			AuthScope:    "*",
		}

		assert.NoError(t, m.Validate())
		assert.Equal(t, "*", m.Scope())
	})
	t.Run("NoClientID", func(t *testing.T) {
		m := ClientCredentials{
			ClientID:     "",
			ClientSecret: "Alice123!",
			AuthScope:    "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("InvalidClientID", func(t *testing.T) {
		m := ClientCredentials{
			ClientID:     "s5gfen1bgxz7s9i",
			ClientSecret: "Alice123!",
			AuthScope:    "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("NoSecret", func(t *testing.T) {
		m := ClientCredentials{
			ClientID:     "cs5gfen1bgxz7s9i",
			ClientSecret: "",
			AuthScope:    "*",
		}

		assert.Error(t, m.Validate())
	})
	t.Run("InvalidSecret", func(t *testing.T) {
		m := ClientCredentials{
			ClientID:     "cs5gfen1bgxz7s9i",
			ClientSecret: "abc  123",
			AuthScope:    "*",
		}

		assert.Error(t, m.Validate())
	})
}
