package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLikeAny(t *testing.T) {
	t.Run("and_or_search", func(t *testing.T) {
		if w := LikeAny("k.keyword", "table spoon & usa | img json"); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'json%' OR k.keyword = 'usa'", w[1])
		}
	})
	t.Run("and_or_search_en", func(t *testing.T) {
		if w := LikeAny("k.keyword", "table spoon and usa or img json"); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'json%' OR k.keyword = 'usa'", w[1])
		}
	})
	t.Run("table spoon usa img json", func(t *testing.T) {
		if w := LikeAny("k.keyword", "table spoon usa img json"); len(w) != 1 {
			t.Fatal("one where condition expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'json%' OR k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%' OR k.keyword = 'usa'", w[0])
		}
	})

	t.Run("cat dog", func(t *testing.T) {
		if w := LikeAny("k.keyword", "cat dog"); len(w) != 1 {
			t.Fatal("one where condition expected")
		} else {
			assert.Equal(t, "k.keyword = 'cat' OR k.keyword = 'dog'", w[0])
		}
	})

	t.Run("cats dogs", func(t *testing.T) {
		if w := LikeAny("k.keyword", "cats dogs"); len(w) != 1 {
			t.Fatal("one where condition expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'cats%' OR k.keyword = 'cat' OR k.keyword LIKE 'dogs%' OR k.keyword = 'dog'", w[0])
		}
	})

	t.Run("spoon", func(t *testing.T) {
		if w := LikeAny("k.keyword", "spoon"); len(w) != 1 {
			t.Fatal("one where condition expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%'", w[0])
		}
	})

	t.Run("img", func(t *testing.T) {
		if w := LikeAny("k.keyword", "img"); len(w) > 0 {
			t.Fatal("no where condition expected")
		}
	})

	t.Run("empty", func(t *testing.T) {
		if w := LikeAny("k.keyword", ""); len(w) > 0 {
			t.Fatal("no where condition expected")
		}
	})
}

func TestLikeAll(t *testing.T) {
	t.Run("keywords", func(t *testing.T) {
		if w := LikeAll("k.keyword", "Jo Mander 李"); len(w) == 2 {
			assert.Equal(t, "k.keyword LIKE 'mander%'", w[0])
			assert.Equal(t, "k.keyword = '李'", w[1])
		} else {
			t.Logf("wheres: %#v", w)
			t.Fatal("two where conditions expected")
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
