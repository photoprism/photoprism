package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUcFirst(t *testing.T) {
	t.Run("PhotoLover", func(t *testing.T) {
		assert.Equal(t, "Photo-lover", UpperFirst("photo-lover"))
	})
	t.Run("Cat", func(t *testing.T) {
		assert.Equal(t, "Cat", UpperFirst("Cat"))
	})
	t.Run("KwaZuluNatal", func(t *testing.T) {
		assert.Equal(t, "KwaZulu natal", UpperFirst("KwaZulu natal"))
	})
	t.Run("EmptyString", func(t *testing.T) {
		assert.Equal(t, "", UpperFirst(""))
	})
}
