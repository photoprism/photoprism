package media

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase64(t *testing.T) {
	t.Run("Gopher", func(t *testing.T) {
		data, err := DecodeBase64(gopher)
		assert.NoError(t, err)
		assert.Equal(t, gopher, EncodeBase64(data))
	})
}
