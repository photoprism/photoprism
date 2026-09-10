package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	ValidateFixtures(t)
	m := GenerateToken()
	assert.Equal(t, 8, len(m))
}

func TestInvalidPreviewToken(t *testing.T) {
	ValidateFixtures(t)
	assert.True(t, InvalidPreviewToken("xxx"))
	assert.True(t, InvalidPreviewToken("2ud3qfpu"))
	PreviewToken.Set("sesspvtoken2ud3qfpu", "2ud3qfpu")
	t.Cleanup(func() { PreviewToken.Unset("sesspvtoken2ud3qfpu") })
	assert.False(t, InvalidPreviewToken("2ud3qfpu"))

	// The instance-wide preview token is registered under the reserved TokenConfig key
	// (Config.Propagate does PreviewToken.Set(TokenConfig, token)); a config token used
	// for shared albums and session-less thumbnails must still resolve as valid.
	assert.True(t, InvalidPreviewToken("sharedtok"))
	PreviewToken.Set(TokenConfig, "sharedtok")
	t.Cleanup(func() { PreviewToken.Unset(TokenConfig) })
	assert.False(t, InvalidPreviewToken("sharedtok"))
}
