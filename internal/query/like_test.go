package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLikeAny(t *testing.T) {
	t.Run("and_or_search", func(t *testing.T) {
		if w := LikeAny("k.keyword", "table spoon & usa | img json", true, false); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'json%' OR k.keyword LIKE 'usa'", w[1])
		}
	})
	t.Run("and_or_search_en", func(t *testing.T) {
		if w := LikeAny("k.keyword", "table spoon and usa or img json", true, false); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'json%' OR k.keyword LIKE 'usa'", w[1])
		}
	})
	t.Run("table spoon usa img json", func(t *testing.T) {
		if w := LikeAny("k.keyword", "table spoon usa img json", true, false); len(w) != 1 {
			t.Fatal("one where condition expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'json%' OR k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%' OR k.keyword LIKE 'usa'", w[0])
		}
	})

	t.Run("cat dog", func(t *testing.T) {
		if w := LikeAny("k.keyword", "cat dog", true, false); len(w) != 1 {
			t.Fatal("one where condition expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'cat' OR k.keyword LIKE 'dog'", w[0])
		}
	})

	t.Run("cats dogs", func(t *testing.T) {
		if w := LikeAny("k.keyword", "cats dogs", true, false); len(w) != 1 {
			t.Fatal("one where condition expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'cats%' OR k.keyword LIKE 'cat' OR k.keyword LIKE 'dogs%' OR k.keyword LIKE 'dog'", w[0])
		}
	})

	t.Run("spoon", func(t *testing.T) {
		if w := LikeAny("k.keyword", "spoon", true, false); len(w) != 1 {
			t.Fatal("one where condition expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%'", w[0])
		}
	})

	t.Run("img", func(t *testing.T) {
		if w := LikeAny("k.keyword", "img", true, false); len(w) > 0 {
			t.Fatal("no where condition expected")
		}
	})

	t.Run("empty", func(t *testing.T) {
		if w := LikeAny("k.keyword", "", true, false); len(w) > 0 {
			t.Fatal("no where condition expected")
		}
	})
}

func TestLikeAnyKeyword(t *testing.T) {
	t.Run("and_or_search", func(t *testing.T) {
		if w := LikeAnyKeyword("k.keyword", "table spoon & usa | img json"); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'json%' OR k.keyword LIKE 'usa'", w[1])
		}
	})
	t.Run("and_or_search_en", func(t *testing.T) {
		if w := LikeAnyKeyword("k.keyword", "table spoon and usa or img json"); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'json%' OR k.keyword LIKE 'usa'", w[1])
		}
	})
}

func TestLikeAnyWord(t *testing.T) {
	t.Run("and_or_search", func(t *testing.T) {
		if w := LikeAnyWord("k.keyword", "table spoon & usa | img json"); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'img%' OR k.keyword LIKE 'json%' OR k.keyword LIKE 'usa%'", w[1])
		}
	})
	t.Run("and_or_search_en", func(t *testing.T) {
		if w := LikeAnyWord("k.keyword", "table spoon and usa or img json"); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'img%' OR k.keyword LIKE 'json%' OR k.keyword LIKE 'usa%'", w[1])
		}
	})
}

func TestLikeAll(t *testing.T) {
	t.Run("keywords", func(t *testing.T) {
		if w := LikeAll("k.keyword", "Jo Mander 李", true, false); len(w) == 2 {
			assert.Equal(t, "k.keyword LIKE 'mander%'", w[0])
			assert.Equal(t, "k.keyword LIKE '李'", w[1])
		} else {
			t.Logf("wheres: %#v", w)
			t.Fatal("two where conditions expected")
		}
	})
	t.Run("exact", func(t *testing.T) {
		if w := LikeAll("k.keyword", "Jo Mander 李", true, true); len(w) == 2 {
			assert.Equal(t, "k.keyword LIKE 'mander'", w[0])
			assert.Equal(t, "k.keyword LIKE '李'", w[1])
		} else {
			t.Logf("wheres: %#v", w)
			t.Fatal("two where conditions expected")
		}
	})
}

func TestLikeAllKeywords(t *testing.T) {
	t.Run("keywords", func(t *testing.T) {
		if w := LikeAllKeywords("k.keyword", "Jo Mander 李"); len(w) == 2 {
			assert.Equal(t, "k.keyword LIKE 'mander%'", w[0])
			assert.Equal(t, "k.keyword LIKE '李'", w[1])
		} else {
			t.Logf("wheres: %#v", w)
			t.Fatal("two where conditions expected")
		}
	})
}

func TestLikeAllWords(t *testing.T) {
	t.Run("keywords", func(t *testing.T) {
		if w := LikeAllWords("k.name", "Jo Mander 王"); len(w) == 3 {
			assert.Equal(t, "k.name LIKE 'jo%'", w[0])
			assert.Equal(t, "k.name LIKE 'mander%'", w[1])
			assert.Equal(t, "k.name LIKE '王%'", w[2])
		} else {
			t.Logf("wheres: %#v", w)
			t.Fatal("two where conditions expected")
		}
	})
}

func TestLikeAllNames(t *testing.T) {
	t.Run("keywords", func(t *testing.T) {
		if w := LikeAllNames("k.name", "j Mander 王"); len(w) == 4 {
			assert.Equal(t, "k.name LIKE 'mander'", w[0])
			assert.Equal(t, "k.name LIKE 'mander %'", w[1])
			assert.Equal(t, "k.name LIKE '王'", w[2])
			assert.Equal(t, "k.name LIKE '王 %'", w[3])
		} else {
			t.Logf("wheres: %#v", w)
			t.Fatal("4 where conditions expected")
		}
	})
}

func TestAnySlug(t *testing.T) {
	t.Run("table spoon usa img json", func(t *testing.T) {
		where := AnySlug("custom_slug", "table spoon usa img json", " ")
		assert.Equal(t, "custom_slug = 'table' OR custom_slug = 'spoon' OR custom_slug = 'usa' OR custom_slug = 'img' OR custom_slug = 'json'", where)
	})

	t.Run("cat dog", func(t *testing.T) {
		where := AnySlug("custom_slug", "cat dog", " ")
		assert.Equal(t, "custom_slug = 'cat' OR custom_slug = 'dog'", where)
	})

	t.Run("cats dogs", func(t *testing.T) {
		where := AnySlug("custom_slug", "cats dogs", " ")
		assert.Equal(t, "custom_slug = 'cats' OR custom_slug = 'cat' OR custom_slug = 'dogs' OR custom_slug = 'dog'", where)
	})

	t.Run("spoon", func(t *testing.T) {
		where := AnySlug("custom_slug", "spoon", " ")
		assert.Equal(t, "custom_slug = 'spoon'", where)
	})

	t.Run("img", func(t *testing.T) {
		where := AnySlug("custom_slug", "img", " ")
		assert.Equal(t, "custom_slug = 'img'", where)
	})

	t.Run("empty", func(t *testing.T) {
		where := AnySlug("custom_slug", "", " ")
		assert.Equal(t, "", where)
	})

	t.Run("comma separated", func(t *testing.T) {
		where := AnySlug("custom_slug", "botanical-garden|landscape|bay", Or)
		assert.Equal(t, "custom_slug = 'botanical-garden' OR custom_slug = 'landscape' OR custom_slug = 'bay'", where)
	})

	t.Run("len = 0", func(t *testing.T) {
		where := AnySlug("custom_slug", " ", "")
		assert.Equal(t, "custom_slug = '' OR custom_slug = ''", where)
	})
}
