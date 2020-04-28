package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWords(t *testing.T) {
	t.Run("I'm a lazy-brown fox!", func(t *testing.T) {
		result := Words("I'm a lazy-BRoWN fox!")
		assert.Equal(t, []string{"lazy-BRoWN", "fox"}, result)
	})
	t.Run("no result", func(t *testing.T) {
		result := Words("x")
		assert.Equal(t, []string(nil), result)
	})
}

func TestReplaceSpaces(t *testing.T) {
	t.Run("I love Cats", func(t *testing.T) {
		result := ReplaceSpaces("I love Cats", "dog")
		assert.Equal(t, "IdoglovedogCats", result)
	})
}

func TestFilenameWords(t *testing.T) {
	t.Run("I'm a lazy-brown fox!", func(t *testing.T) {
		result := FilenameWords("I'm a lazy-BRoWN fox!")
		assert.Equal(t, []string{"lazy", "BRoWN", "fox"}, result)
	})
	t.Run("no result", func(t *testing.T) {
		result := FilenameWords("x")
		assert.Equal(t, []string(nil), result)
	})
}

func TestFilenameKeywords(t *testing.T) {
	t.Run("I'm a lazy-brown var fox.jpg!", func(t *testing.T) {
		result := FilenameKeywords("I'm a lazy-brown var fox.jpg!")
		assert.Equal(t, []string{"lazy", "brown", "fox"}, result)
	})
	t.Run("no result", func(t *testing.T) {
		result := FilenameKeywords("x")
		assert.Equal(t, []string(nil), result)
	})
}

func TestKeywords(t *testing.T) {
	t.Run("I'm a lazy brown fox!", func(t *testing.T) {
		result := Keywords("I'm a lazy BRoWN img!")
		assert.Equal(t, []string{"lazy", "brown"}, result)
	})
	t.Run("no result", func(t *testing.T) {
		result := Keywords("was")
		assert.Equal(t, []string(nil), result)
	})
}

func TestUniqueWords(t *testing.T) {
	t.Run("many", func(t *testing.T) {
		result := UniqueWords([]string{"lazy", "jpg", "Brown", "apple", "brown", "new-york", "JPG"})
		assert.Equal(t, []string{"apple", "brown", "jpg", "lazy", "new-york"}, result)
	})
	t.Run("one", func(t *testing.T) {
		result := UniqueWords([]string{"lazy"})
		assert.Equal(t, []string{"lazy"}, result)
	})
}

func TestUniqueKeywords(t *testing.T) {
	t.Run("many", func(t *testing.T) {
		result := UniqueKeywords("lazy, Brown, apple, new-york, brown, ...")
		assert.Equal(t, []string{"apple", "brown", "lazy", "new-york"}, result)
	})
	t.Run("one", func(t *testing.T) {
		result := UniqueKeywords("")
		assert.Equal(t, []string(nil), result)
	})
}
