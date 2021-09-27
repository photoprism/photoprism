package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlug(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Slug(""))
	})
	t.Run("BillGates", func(t *testing.T) {
		assert.Equal(t, "william-henry-gates-iii", Slug("William  Henry Gates III"))
	})
	t.Run("Quotes", func(t *testing.T) {
		assert.Equal(t, "william-henry-gates", Slug("william \"HenRy\" gates' "))
	})
	t.Run("Chinese", func(t *testing.T) {
		assert.Equal(t, "chen-zhao", Slug(" 陈  赵"))
	})
}

func TestSlugToTitle(t *testing.T) {
	t.Run("cute_Kitten", func(t *testing.T) {
		assert.Equal(t, "Cute-Kitten", SlugToTitle("cute-kitten"))
	})
	t.Run("empty", func(t *testing.T) {
		assert.Equal(t, "", SlugToTitle(""))
	})
}
