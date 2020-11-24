package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugToTitle(t *testing.T) {
	t.Run("cute_Kitten", func(t *testing.T) {
		assert.Equal(t, "Cute-Kitten", SlugToTitle("cute-kitten"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, "", SlugToTitle(""))
	})
}
