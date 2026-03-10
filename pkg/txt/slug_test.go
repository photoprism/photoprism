package txt

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestSlug(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Slug(""))
	})
	t.Run("Gates", func(t *testing.T) {
		assert.Equal(t, "william-henry-gates-iii", Slug("William  Henry Gates III"))
	})
	t.Run("Quotes", func(t *testing.T) {
		assert.Equal(t, "william-henry-gates", Slug("william \"HenRy\" gates' "))
	})
	t.Run("Chinese", func(t *testing.T) {
		assert.Equal(t, "chen-zhao", Slug(" 陈  赵"))
	})
	t.Run("Emoji", func(t *testing.T) {
		assert.Equal(t, "_5cpzfdq", Slug("💎"))
		assert.Equal(t, "_5cpzfea", Slug("💐"))
		assert.Equal(t, "_5cpzfea", Slug("   💐   "))
		assert.Equal(t, "_5cpzfdxqt5jja", Slug("💎💐"))
		assert.Equal(t, "photoprism-u1f48e", Slug("PhotoPrism 💎"))
		assert.Equal(t, "ins-u1f377", Slug("ins/🍷"))
		assert.Equal(t, "work-u1f618", Slug("Work 😘"))
		assert.Equal(t, "_3kmib24yr3", Slug("_3kmib24yr3"))
		assert.Equal(t, "-", Slug("-"))
		assert.Equal(t, "_", Slug("_"))
		assert.Equal(t, "_a", Slug("_a"))
		assert.Equal(t, "_5cpzfea", Slug("_5cpzfea"))
		assert.Equal(t, "_5cpzfdxqt5jja", Slug("_5cpzfdxqt5jja"))
	})
	t.Run("LongInputUsesHashSuffix", func(t *testing.T) {
		base := strings.Repeat("Very Long Prefix 💎 ", 8)
		slugA := Slug(base + "Alpha")
		slugB := Slug(base + "Beta")

		assert.NotEqual(t, slugA, slugB)
		assert.Equal(t, ClipSlug, utf8.RuneCountInString(slugA))
		assert.Equal(t, ClipSlug, utf8.RuneCountInString(slugB))
	})
}

func TestSlugToTitle(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SlugToTitle(""))
	})
	t.Run("Kitten", func(t *testing.T) {
		assert.Equal(t, "Cute-Kitten", SlugToTitle("cute-kitten"))
	})
	t.Run("Emoji", func(t *testing.T) {
		assert.Equal(t, "💎", SlugToTitle("_5cpzfdq"))
		assert.Equal(t, "💐", SlugToTitle("_5cpzfea"))
		assert.Equal(t, "💎💐", SlugToTitle("_5cpzfdxqt5jja"))
		assert.Equal(t, "PhotoPrism", SlugToTitle("photoprism"))
	})
}
