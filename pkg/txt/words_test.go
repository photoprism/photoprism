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
	t.Run("Österreich Urlaub", func(t *testing.T) {
		result := Words("Österreich Urlaub")
		assert.Equal(t, []string{"Österreich", "Urlaub"}, result)
	})
	t.Run("Schäferhund", func(t *testing.T) {
		result := Words("Schäferhund")
		assert.Equal(t, []string{"Schäferhund"}, result)
	})
	t.Run("Île de la Réunion", func(t *testing.T) {
		result := Words("Île de la Réunion")
		assert.Equal(t, []string{"Île", "Réunion"}, result)
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
	t.Run("Österreich Urlaub", func(t *testing.T) {
		result := FilenameWords("Österreich Urlaub")
		assert.Equal(t, []string{"Österreich", "Urlaub"}, result)
	})
	t.Run("Schäferhund", func(t *testing.T) {
		result := FilenameWords("Schäferhund")
		assert.Equal(t, []string{"Schäferhund"}, result)
	})
	t.Run("Île de la Réunion", func(t *testing.T) {
		result := FilenameWords("Île de la Réunion")
		assert.Equal(t, []string{"Île", "Réunion"}, result)
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
	t.Run("Österreich Urlaub", func(t *testing.T) {
		result := FilenameKeywords("Österreich Urlaub")
		assert.Equal(t, []string{"österreich", "urlaub"}, result)
	})
	t.Run("Schäferhund", func(t *testing.T) {
		result := FilenameKeywords("Schäferhund")
		assert.Equal(t, []string{"schäferhund"}, result)
	})
	t.Run("Île de la Réunion", func(t *testing.T) {
		result := FilenameKeywords("Île de la Réunion")
		assert.Equal(t, []string{"île", "réunion"}, result)
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
	t.Run("Österreich Urlaub", func(t *testing.T) {
		result := Keywords("Österreich Urlaub")
		assert.Equal(t, []string{"österreich", "urlaub"}, result)
	})
	t.Run("Schäferhund", func(t *testing.T) {
		result := Keywords("Schäferhund")
		assert.Equal(t, []string{"schäferhund"}, result)
	})
	t.Run("Île de la Réunion", func(t *testing.T) {
		result := Keywords("Île de la Réunion")
		assert.Equal(t, []string{"île", "réunion"}, result)
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

func TestRemoveFromWords(t *testing.T) {
	t.Run("brown apple", func(t *testing.T) {
		result := RemoveFromWords([]string{"lazy", "jpg", "Brown", "apple", "brown", "new-york", "JPG"}, "brown apple")
		assert.Equal(t, []string{"jpg", "lazy", "new-york"}, result)
	})
	t.Run("empty", func(t *testing.T) {
		result := RemoveFromWords([]string{"lazy", "jpg", "Brown", "apple"}, "")
		assert.Equal(t, []string{"apple", "brown", "jpg", "lazy"}, result)
	})
}
