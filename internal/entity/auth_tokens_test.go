package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	m := GenerateToken()
	assert.Equal(t, 8, len(m))
}

func TestInvalidDownloadToken(t *testing.T) {
	assert.True(t, InvalidDownloadToken("xxx"))
	assert.True(t, InvalidDownloadToken("1ud3qfpu"))
	DownloadToken.Set("1ud3qfpu", "1ud3qfpu")
	assert.False(t, InvalidDownloadToken("1ud3qfpu"))
}

func TestInvalidPreviewToken(t *testing.T) {
	assert.True(t, InvalidPreviewToken("xxx"))
	assert.True(t, InvalidPreviewToken("2ud3qfpu"))
	PreviewToken.Set("2ud3qfpu", "2ud3qfpu")
	assert.False(t, InvalidPreviewToken("2ud3qfpu"))
}
