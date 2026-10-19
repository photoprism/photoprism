package form

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuthConsent(t *testing.T) {
	t.Run("Unmarshal", func(t *testing.T) {
		var f OAuthConsent
		assert.NoError(t, json.Unmarshal([]byte(`{"request":"abc","scopes":["mcp"]}`), &f))
		assert.Equal(t, "abc", f.Request)
		assert.Equal(t, []string{"mcp"}, f.Scopes)
	})
}
