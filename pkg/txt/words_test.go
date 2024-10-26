package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknownWord(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.True(t, UnknownWord(""))
	})
	t.Run("qx", func(t *testing.T) {
		assert.True(t, UnknownWord("qx"))
	})
	t.Run("atz", func(t *testing.T) {
		assert.True(t, UnknownWord("atz"))
	})
	t.Run("xqx", func(t *testing.T) {
		assert.True(t, UnknownWord("xqx"))
	})
	t.Run("kuh", func(t *testing.T) {
		assert.False(t, UnknownWord("kuh"))
	})
	t.Run("muh", func(t *testing.T) {
		assert.False(t, UnknownWord("muh"))
	})
	t.Run("桥", func(t *testing.T) {
		assert.False(t, UnknownWord("桥"))
	})
	t.Run("桥船", func(t *testing.T) {
		assert.False(t, UnknownWord("桥船"))
	})
}

func TestWords(t *testing.T) {
	t.Run("桥", func(t *testing.T) {
		result := Words("桥")
		assert.Equal(t, []string{"桥"}, result)
	})
	t.Run("桥船", func(t *testing.T) {
		result := Words("桥船")
		assert.Equal(t, []string{"桥船"}, result)
	})
	t.Run("桥船猫", func(t *testing.T) {
		result := Words("桥船猫")
		assert.Equal(t, []string{"桥船猫"}, result)
	})
	t.Run("谢谢！", func(t *testing.T) {
		result := Words("谢谢！")
		assert.Equal(t, []string{"谢谢"}, result)
	})
	t.Run("I'm a lazy-brown fox!", func(t *testing.T) {
		result := Words("I'm a lazy-BRoWN fox!")
		assert.Equal(t, []string{"I'm", "lazy-BRoWN", "fox"}, result)
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
		assert.Equal(t, []string{"Île", "de", "la", "Réunion"}, result)
	})
	t.Run("empty", func(t *testing.T) {
		result := Words("")
		assert.Empty(t, result)
	})
	t.Run("trim", func(t *testing.T) {
		result := Words(" -foo- -")
		assert.Equal(t, []string{"foo"}, result)
	})
	t.Run("McDonalds", func(t *testing.T) {
		result := Words(" McDonald's FOO'bar-'")
		assert.Equal(t, []string{"McDonald's", "FOO'bar"}, result)
	})
	// McDonald's
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
	t.Run("empty", func(t *testing.T) {
		result := FilenameWords("")
		assert.Empty(t, result)
	})
}

func TestAddToWords(t *testing.T) {
	t.Run("I'm a lazy-BRoWN fox!", func(t *testing.T) {
		result := AddToWords([]string{"foo", "bar", "fox"}, "Yellow banana, apple; pan-pot")
		assert.Equal(t, []string{"apple", "banana", "bar", "foo", "fox", "pan-pot", "yellow"}, result)
	})
}

func TestMergeWords(t *testing.T) {
	t.Run("I'm a lazy-BRoWN fox!", func(t *testing.T) {
		result := MergeWords("I'm a lazy-BRoWN fox!", "Yellow banana, apple; pan-pot")
		assert.Equal(t, "apple, banana, fox, i'm, lazy-brown, pan-pot, yellow", result)
	})
}

func TestFilenameKeywords(t *testing.T) {
	t.Run("桥.jpg", func(t *testing.T) {
		result := FilenameKeywords("桥.jpg")
		assert.Equal(t, []string{"桥"}, result)
	})
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
	t.Run("empty", func(t *testing.T) {
		result := FilenameKeywords("")
		assert.Empty(t, result)
	})
}

func TestKeywords(t *testing.T) {
	t.Run("桥", func(t *testing.T) {
		result := Keywords("桥")
		assert.Equal(t, []string{"桥"}, result)
	})
	t.Run("桥船", func(t *testing.T) {
		result := Keywords("桥船")
		assert.Equal(t, []string{"桥船"}, result)
	})
	t.Run("桥船猫", func(t *testing.T) {
		result := Keywords("桥船猫")
		assert.Equal(t, []string{"桥船猫"}, result)
	})
	t.Run("谢谢！", func(t *testing.T) {
		result := Keywords("谢谢！")
		assert.Equal(t, []string{"谢谢"}, result)
	})
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
	t.Run("empty", func(t *testing.T) {
		result := Keywords("")
		assert.Empty(t, result)
	})
}

func TestUniqueWords(t *testing.T) {
	t.Run("Many", func(t *testing.T) {
		result := UniqueWords([]string{"lazy", "jpg", "Brown", "apple", "brown", "new-york", "JPG"})
		assert.Equal(t, []string{"apple", "brown", "jpg", "lazy", "new-york"}, result)
	})
	t.Run("One", func(t *testing.T) {
		result := UniqueWords([]string{"lazy"})
		assert.Equal(t, []string{"lazy"}, result)
	})
	t.Run("Numerals", func(t *testing.T) {
		result := UniqueWords([]string{"1st", "40.", "52nd", "ma'am", "80s"})
		assert.Equal(t, []string{"1st", "40.", "52nd", "80s", "ma'am"}, result)
	})
}

func TestUniqueKeywords(t *testing.T) {
	t.Run("Many", func(t *testing.T) {
		result := UniqueKeywords("lazy, Brown, apple, new-york, brown, ...")
		assert.Equal(t, []string{"apple", "brown", "lazy", "new-york"}, result)
	})
	t.Run("One", func(t *testing.T) {
		result := UniqueKeywords("")
		assert.Equal(t, []string(nil), result)
	})
	t.Run("Numerals", func(t *testing.T) {
		result := UniqueKeywords("1st, 40., 52nd, ma'am, 80s")
		assert.Equal(t, []string{"1st", "52nd", "80s", "ma'am"}, result)
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

func TestStopwordsOnly(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, StopwordsOnly(""))
	})
	t.Run("FoldersDateienFile", func(t *testing.T) {
		assert.True(t, StopwordsOnly("Folders Dateien File"))
	})
	t.Run("FoldersDateienFile", func(t *testing.T) {
		assert.False(t, StopwordsOnly("Folders Dateien Meme File"))
	})
	t.Run("qx", func(t *testing.T) {
		assert.True(t, StopwordsOnly("qx"))
	})
	t.Run("atz", func(t *testing.T) {
		assert.True(t, StopwordsOnly("atz"))
	})
	t.Run("xqx", func(t *testing.T) {
		assert.True(t, StopwordsOnly("xqx"))
	})
	t.Run("kuh", func(t *testing.T) {
		assert.False(t, StopwordsOnly("kuh"))
	})
	t.Run("muh", func(t *testing.T) {
		assert.False(t, StopwordsOnly("muh"))
	})
	t.Run("桥", func(t *testing.T) {
		assert.False(t, StopwordsOnly("桥"))
	})
	t.Run("桥船", func(t *testing.T) {
		assert.False(t, StopwordsOnly("桥船"))
	})
}
