package form

import (
	"testing"

	"github.com/photoprism/photoprism/pkg/authn"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		client := NewClient()
		assert.Equal(t, authn.MethodOAuth2, client.Method())
		assert.Equal(t, "", client.Scope())
		assert.Equal(t, "", client.Name())
	})
}
