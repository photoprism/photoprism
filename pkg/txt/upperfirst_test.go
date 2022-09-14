package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUcFirst(t *testing.T) {
	t.Run("photo-lover", func(t *testing.T) {
		assert.Equal(t, "Photo-lover", UpperFirst("photo-lover"))
	})
	t.Run("cat", func(t *testing.T) {
		assert.Equal(t, "Cat", UpperFirst("Cat"))
	})
	t.Run("KwaZulu natal", func(t *testing.T) {
		assert.Equal(t, "KwaZulu natal", UpperFirst("KwaZulu natal"))
	})
	t.Run("empty string", func(t *testing.T) {
		assert.Equal(t, "", UpperFirst(""))
	})
}
