package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestLike(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", Like(""))
	})
	t.Run("Special", func(t *testing.T) {
		s := " ' \" \t \n %_''"
		exp := "'' \"\"   %_''''"
		result := Like(s)
		t.Logf("String..: %s", s)
		t.Logf("Expected: %s", exp)
		t.Logf("Result..: %s", result)
		assert.Equal(t, exp, result)
	})
	t.Run("Alnum", func(t *testing.T) {
		assert.Equal(t, "123ABCabc", Like("   123ABCabc%*    "))
	})
}

func TestLikeAny(t *testing.T) {
	t.Run("and_or_search", func(t *testing.T) {
		if w := LikeAny("k.keyword", "table spoon & usa | img json", true, false); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'json%' OR k.keyword LIKE 'usa'", w[1])
		}
	})
	t.Run(" exact and_or_search", func(t *testing.T) {
		if w := LikeAny("k.keyword", "table spoon & usa | img json", true, true); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon' OR k.keyword LIKE 'table'", w[0])
			assert.Equal(t, "k.keyword LIKE 'json' OR k.keyword LIKE 'usa'", w[1])
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
	t.Run("SearchAndOr", func(t *testing.T) {
		if w := LikeAnyWord("k.keyword", "table spoon & usa | img json"); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'img%' OR k.keyword LIKE 'json%' OR k.keyword LIKE 'usa%'", w[1])
		}
	})
	t.Run("SearchAndOrEnglish", func(t *testing.T) {
		if w := LikeAnyWord("k.keyword", "table spoon and usa or img json"); len(w) != 2 {
			t.Fatal("two where conditions expected")
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE 'img%' OR k.keyword LIKE 'json%' OR k.keyword LIKE 'usa%'", w[1])
		}
	})
	t.Run("EscapeSql", func(t *testing.T) {
		if w := LikeAnyWord("k.keyword", "table% | 'spoon' & \"us'a"); len(w) != 2 {
			t.Fatalf("two where conditions expected: %#v", w)
		} else {
			assert.Equal(t, "k.keyword LIKE 'spoon%' OR k.keyword LIKE 'table%'", w[0])
			assert.Equal(t, "k.keyword LIKE '\"\"us''a%'", w[1])
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
	t.Run("string empty", func(t *testing.T) {
		w := LikeAll("k.keyword", "", true, true)
		assert.Empty(t, w)
	})
	t.Run("0 words", func(t *testing.T) {
		w := LikeAll("k.keyword", "ab", true, true)
		assert.Empty(t, w)
	})
}

func TestLikeAllKeywords(t *testing.T) {
	t.Run("keywords", func(t *testing.T) {
		if w := LikeAllKeywords("k.keyword", "Jo Mander 李"); len(w) == 2 {
			assert.Equal(t, "k.keyword LIKE 'mander%'", w[0])
			assert.Equal(t, "k.keyword LIKE '李'", w[1])
		} else {
			t.Fatalf("unexpected result:  %#v", w)
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
			t.Fatalf("unexpected result:  %#v", w)
		}
	})
}

func TestLikeAllNames(t *testing.T) {
	t.Run("MultipleNames", func(t *testing.T) {
		if w := LikeAllNames(Cols{"k.name"}, "j Mander 王"); len(w) == 1 {
			assert.Equal(t, "k.name LIKE 'j Mander 王%'", w[0])
		} else {
			t.Fatalf("unexpected result:  %#v", w)
		}
	})
	t.Run("MultipleColumns", func(t *testing.T) {
		if w := LikeAllNames(Cols{"a.col1", "b.col2"}, "Mo Mander"); len(w) == 1 {
			assert.Equal(t, "a.col1 LIKE 'Mo Mander%' OR b.col2 LIKE 'Mo Mander%'", w[0])
		} else {
			t.Fatalf("unexpected result: %#v", w)
		}
	})
	t.Run("EmptyName", func(t *testing.T) {
		w := LikeAllNames(Cols{"k.name"}, "")
		assert.Empty(t, w)
	})
	t.Run("SingleCharacter", func(t *testing.T) {
		if w := LikeAllNames(Cols{"k.name"}, "a"); len(w) == 1 {
			assert.Equal(t, "k.name LIKE '%a%'", w[0])
		} else {
			t.Fatalf("unexpected result: %#v", w)
		}
	})
	t.Run("FullNames", func(t *testing.T) {
		if w := LikeAllNames(Cols{"j.name", "j.alias"}, "Bill & Melinda Gates"); len(w) == 2 {
			assert.Equal(t, "j.name LIKE '%Bill%' OR j.alias LIKE '%Bill%'", w[0])
			assert.Equal(t, "j.name LIKE 'Melinda Gates%' OR j.alias LIKE 'Melinda Gates%'", w[1])
		} else {
			t.Fatalf("unexpected result: %#v", w)
		}
	})
	t.Run("Plus", func(t *testing.T) {
		if w := LikeAllNames(Cols{"name"}, clean.SearchQuery("Paul + Paula")); len(w) == 2 {
			assert.Equal(t, "name LIKE '%Paul%'", w[0])
			assert.Equal(t, "name LIKE '%Paula%'", w[1])
		} else {
			t.Fatalf("unexpected result:  %#v", w)
		}
	})
	t.Run("And", func(t *testing.T) {
		if w := LikeAllNames(Cols{"name"}, clean.SearchQuery("P and Paula")); len(w) == 2 {
			assert.Equal(t, "name LIKE '%P%'", w[0])
			assert.Equal(t, "name LIKE '%Paula%'", w[1])
		} else {
			t.Fatalf("unexpected result:  %#v", w)
		}
	})
	t.Run("Or", func(t *testing.T) {
		if w := LikeAllNames(Cols{"name"}, clean.SearchQuery("Paul or Paula")); len(w) == 1 {
			assert.Equal(t, "name LIKE '%Paul%' OR name LIKE '%Paula%'", w[0])
		} else {
			t.Fatalf("unexpected result:  %#v", w)
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
		where := AnySlug("custom_slug", "botanical-garden|landscape|bay", txt.Or)
		assert.Equal(t, "custom_slug = 'botanical-garden' OR custom_slug = 'landscape' OR custom_slug = 'bay'", where)
	})

	t.Run("len = 0", func(t *testing.T) {
		where := AnySlug("custom_slug", " ", "")
		assert.Equal(t, "custom_slug = '' OR custom_slug = ''", where)
	})
}

func TestAnyInt(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		where := AnyInt("photos.photo_month", "", txt.Or, entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "", where)
	})

	t.Run("Range", func(t *testing.T) {
		where := AnyInt("photos.photo_month", "-3|0|10|9|11|12|13", txt.Or, entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "photos.photo_month = 10 OR photos.photo_month = 9 OR photos.photo_month = 11 OR photos.photo_month = 12", where)
	})

	t.Run("Chars", func(t *testing.T) {
		where := AnyInt("photos.photo_month", "a|b|c", txt.Or, entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "", where)
	})

	t.Run("CommaSeparated", func(t *testing.T) {
		where := AnyInt("photos.photo_month", "-3,10,9,11,12,13", ",", entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "photos.photo_month = 10 OR photos.photo_month = 9 OR photos.photo_month = 11 OR photos.photo_month = 12", where)
	})

	t.Run("Invalid", func(t *testing.T) {
		where := AnyInt("photos.photo_month", "  , |  ", ",", entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "", where)
	})
}

func TestOrLike(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		where, values := OrLike("k.keyword", "")

		assert.Equal(t, "", where)
		assert.Equal(t, []interface{}{}, values)
	})
	t.Run("OneTerm", func(t *testing.T) {
		where, values := OrLike("k.keyword", "bar")

		assert.Equal(t, "k.keyword LIKE ?", where)
		assert.Equal(t, []interface{}{"bar"}, values)
	})
	t.Run("TwoTerms", func(t *testing.T) {
		where, values := OrLike("k.keyword", "foo*%|bar")

		assert.Equal(t, "k.keyword LIKE ? OR k.keyword LIKE ?", where)
		assert.Equal(t, []interface{}{"foo%", "bar"}, values)
	})
	t.Run("OneFilename", func(t *testing.T) {
		where, values := OrLike("files.file_name", " 2790/07/27900704_070228_D6D51B6C.jpg")

		assert.Equal(t, "files.file_name LIKE ?", where)
		assert.Equal(t, []interface{}{" 2790/07/27900704_070228_D6D51B6C.jpg"}, values)
	})
	t.Run("TwoFilenames", func(t *testing.T) {
		where, values := OrLike("files.file_name", "1990* | 2790/07/27900704_070228_D6D51B6C.jpg")

		assert.Equal(t, "files.file_name LIKE ? OR files.file_name LIKE ?", where)
		assert.Equal(t, []interface{}{"1990%", "2790/07/27900704_070228_D6D51B6C.jpg"}, values)
	})
}

func TestSplit(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		values := Split("", "")

		assert.Equal(t, []string{}, values)
	})
	t.Run("FooBar", func(t *testing.T) {
		values := Split(" foo | Bar ", "&")

		assert.Equal(t, []string{" foo | Bar "}, values)
	})
	t.Run("FooAndBar", func(t *testing.T) {
		values := Split(" foo & Bar ", "o")

		assert.Equal(t, []string{"f", "& Bar"}, values)
	})
	t.Run("FooAndBarAndBaz", func(t *testing.T) {
		values := Split(" foo & Bar&BAZ ", "&")

		assert.Equal(t, []string{"foo", "Bar", "BAZ"}, values)
	})
}
func TestSplitOr(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		values := SplitOr("")

		assert.Equal(t, []string{}, values)
	})
	t.Run("FooBar", func(t *testing.T) {
		values := SplitOr(" foo | Bar ")

		assert.Equal(t, []string{"foo", "Bar"}, values)
	})
	t.Run("FooBarTrim", func(t *testing.T) {
		values := SplitOr(" foo | Bar |")

		assert.Equal(t, []string{"foo", "Bar"}, values)
	})
	t.Run("FooAndBar", func(t *testing.T) {
		values := SplitOr(" foo & Bar ")

		assert.Equal(t, []string{" foo & Bar "}, values)
	})
	t.Run("FooAndBarAndBaz", func(t *testing.T) {
		values := SplitOr(" foo & Bar&BAZ ")

		assert.Equal(t, []string{" foo & Bar&BAZ "}, values)
	})
}

func TestSplitAnd(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		values := SplitAnd("")

		assert.Equal(t, []string{}, values)
	})

	t.Run("FooOrBar", func(t *testing.T) {
		values := SplitAnd(" foo | Bar ")

		assert.Equal(t, []string{" foo | Bar "}, values)
	})

	t.Run("FooAndBar", func(t *testing.T) {
		values := SplitAnd(" foo & Bar ")

		assert.Equal(t, []string{"foo", "Bar"}, values)
	})

	t.Run("FooAndBarAndBaz", func(t *testing.T) {
		values := SplitAnd(" foo & Bar&BAZ ")

		assert.Equal(t, []string{"foo", "Bar", "BAZ"}, values)
	})
}
