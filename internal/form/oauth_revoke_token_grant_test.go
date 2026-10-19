package form_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestOAuthRevokeToken_GrantRefreshToken pins that the form recognizes refresh tokens of entity grants.
func TestOAuthRevokeToken_GrantRefreshToken(t *testing.T) {
	f := form.OAuthRevokeToken{Token: rnd.RefreshToken(rnd.GenerateUID(entity.GrantUID))}

	require.NoError(t, f.Validate())
	assert.Equal(t, form.RefreshToken, f.TokenTypeHint)
}
