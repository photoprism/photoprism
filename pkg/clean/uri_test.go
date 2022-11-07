package clean

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUri(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		result := Uri("https://docs.photoprism.app/getting-started/config-options/#file-converters")
		assert.Equal(t, "https://docs.photoprism.app/getting-started/config-options/#file-converters", result)
	})
	t.Run("Invalid", func(t *testing.T) {
		result := Uri("https://..docs.photoprism.app/gettin\\g-started/config-options/\tfile-converters")
		assert.Equal(t, "", result)
	})
}
