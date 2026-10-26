package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpperFirst(t *testing.T) {
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
	t.Run("SingleLetter", func(t *testing.T) {
		assert.Equal(t, "A", UpperFirst("a"))
	})
	t.Run("MultiByte", func(t *testing.T) {
		assert.Equal(t, "Äpfel", UpperFirst("äpfel"))
		assert.Equal(t, "Ärger", UpperFirst("Ärger"))
		assert.Equal(t, "Éclair", UpperFirst("éclair"))
		assert.Equal(t, "Ω", UpperFirst("ω"))
	})
	t.Run("NoUppercase", func(t *testing.T) {
		assert.Equal(t, "123 abc", UpperFirst("123 abc"))
		assert.Equal(t, "日本", UpperFirst("日本"))
		assert.Equal(t, "📷 photo", UpperFirst("📷 photo"))
	})
	t.Run("InvalidUTF8", func(t *testing.T) {
		assert.Equal(t, "\xffabc", UpperFirst("\xffabc"))
		assert.Equal(t, "\xc3", UpperFirst("\xc3"))
	})
}
