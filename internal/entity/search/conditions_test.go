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
	t.Run("AndOrSearch", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "table spoon & usa | img json", true, false)
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ?", "k.keyword LIKE ? OR k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"spoon%", "table%"}, {"json%", "usa"}}, v)
	})
	t.Run("ExactAndOrSearch", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "table spoon & usa | img json", true, true)
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ?", "k.keyword LIKE ? OR k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"spoon", "table"}, {"json", "usa"}}, v)
	})
	t.Run("AndOrSearchEn", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "table spoon and usa or img json", true, false)
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ?", "k.keyword LIKE ? OR k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"spoon%", "table%"}, {"json%", "usa"}}, v)
	})
	t.Run("TableSpoonUsaImgJson", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "table spoon usa img json", true, false)
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ? OR k.keyword LIKE ? OR k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"json%", "spoon%", "table%", "usa"}}, v)
	})
	t.Run("CatDog", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "cat dog", true, false)
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"cat", "dog"}}, v)
	})
	t.Run("CatsDogs", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "cats dogs", true, false)
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ? OR k.keyword LIKE ? OR k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"cats%", "cat", "dogs%", "dog"}}, v)
	})
	t.Run("Spoon", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "spoon", true, false)
		assert.Equal(t, []string{"k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"spoon%"}}, v)
	})
	t.Run("Img", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "img", true, false)
		assert.Empty(t, w, "where")
		assert.Empty(t, v, "value")
	})
	t.Run("Empty", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "", true, false)
		assert.Empty(t, w, "where")
		assert.Empty(t, v, "value")
	})
}

func TestLikeAnyKeyword(t *testing.T) {
	t.Run("AndOrSearch", func(t *testing.T) {
		w, v := LikeAnyKeyword("k.keyword", "table spoon & usa | img json")
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ?", "k.keyword LIKE ? OR k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"spoon%", "table%"}, {"json%", "usa"}}, v)
	})
	t.Run("AndOrSearchEn", func(t *testing.T) {
		w, v := LikeAnyKeyword("k.keyword", "table spoon and usa or img json")
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ?", "k.keyword LIKE ? OR k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"spoon%", "table%"}, {"json%", "usa"}}, v)
	})
}

func TestLikeAnyWord(t *testing.T) {
	t.Run("SearchAndOr", func(t *testing.T) {
		w, v := LikeAnyWord("k.keyword", "table spoon & usa | img json")
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ?", "k.keyword LIKE ? OR k.keyword LIKE ? OR k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"spoon%", "table%"}, {"img%", "json%", "usa%"}}, v)
	})
	t.Run("SearchAndOrEnglish", func(t *testing.T) {
		w, v := LikeAnyWord("k.keyword", "table spoon and usa or img json")
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ?", "k.keyword LIKE ? OR k.keyword LIKE ? OR k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"spoon%", "table%"}, {"img%", "json%", "usa%"}}, v)
	})
	t.Run("EscapeSql", func(t *testing.T) {
		w, v := LikeAnyWord("k.keyword", "table% | 'spoon' & \"us'a")
		assert.Equal(t, []string{"k.keyword LIKE ? OR k.keyword LIKE ?", "k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"spoon%", "table%"}, {`"us'a%`}}, v)
	})
}

func TestLikeAll(t *testing.T) {
	t.Run("Keywords", func(t *testing.T) {
		w, v := LikeAll("k.keyword", "Jo Mander 李", true, false)
		assert.Equal(t, []string{"k.keyword LIKE ?", "k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"mander%"}, {"李"}}, v)
	})
	t.Run("Exact", func(t *testing.T) {
		w, v := LikeAll("k.keyword", "Jo Mander 李", true, true)
		assert.Equal(t, []string{"k.keyword LIKE ?", "k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"mander"}, {"李"}}, v)
	})
	t.Run("StringEmpty", func(t *testing.T) {
		w, v := LikeAll("k.keyword", "", true, true)
		assert.Empty(t, w)
		assert.Empty(t, v)
	})
	t.Run("ZeroWords", func(t *testing.T) {
		w, v := LikeAll("k.keyword", "ab", true, true)
		assert.Empty(t, w)
		assert.Empty(t, v)
	})
}

func TestLikeAllKeywords(t *testing.T) {
	t.Run("Keywords", func(t *testing.T) {
		w, v := LikeAllKeywords("k.keyword", "Jo Mander 李")
		assert.Equal(t, []string{"k.keyword LIKE ?", "k.keyword LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"mander%"}, {"李"}}, v)
	})
}

func TestLikeAllWords(t *testing.T) {
	t.Run("Keywords", func(t *testing.T) {
		w, v := LikeAllWords("k.name", "Jo Mander 王")
		assert.Equal(t, []string{"k.name LIKE ?", "k.name LIKE ?", "k.name LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"jo%"}, {"mander%"}, {"王%"}}, v)
	})
}

func TestLikeAllNames(t *testing.T) {
	t.Run("MultipleNames", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"k.name"}, "j Mander 王")
		assert.Equal(t, []string{"k.name LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"j Mander 王%"}}, v)
	})
	t.Run("MultipleColumns", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"a.col1", "b.col2"}, "Mo Mander")
		assert.Equal(t, []string{"a.col1 LIKE ? OR b.col2 LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"Mo Mander%", "Mo Mander%"}}, v)
	})
	t.Run("EmptyName", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"k.name"}, "")
		assert.Empty(t, w)
		assert.Empty(t, v)
	})
	t.Run("SingleCharacter", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"k.name"}, "a")
		assert.Equal(t, []string{"k.name LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"%a%"}}, v)
	})
	t.Run("FullNames", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"j.name", "j.alias"}, "Bill & Melinda Gates")
		assert.Equal(t, []string{"j.name LIKE ? OR j.alias LIKE ?", "j.name LIKE ? OR j.alias LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"%Bill%", "%Bill%"}, {"Melinda Gates%", "Melinda Gates%"}}, v)
	})
	t.Run("Plus", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"name"}, clean.SearchQuery("Paul + Paula"))
		assert.Equal(t, []string{"name LIKE ?", "name LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"%Paul%"}, {"%Paula%"}}, v)
	})
	t.Run("And", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"name"}, clean.SearchQuery("P and Paula"))
		assert.Equal(t, []string{"name LIKE ?", "name LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"%P%"}, {"%Paula%"}}, v)
	})
	t.Run("Or", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"name"}, clean.SearchQuery("Paul or Paula"))
		assert.Equal(t, []string{"name LIKE ? OR name LIKE ?"}, w)
		assert.Equal(t, [][]interface{}{{"%Paul%", "%Paula%"}}, v)
	})
}

func TestAnySlug(t *testing.T) {
	t.Run("Multiple", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "table spoon usa img json", " ")
		assert.Equal(t, "custom_slug = ? OR custom_slug = ? OR custom_slug = ? OR custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []interface{}{"table", "spoon", "usa", "img", "json"}, v)
	})
	t.Run("CatDog", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "cat dog", " ")
		assert.Equal(t, "custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []interface{}{"cat", "dog"}, v)
	})
	t.Run("CatsDogs", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "cats dogs", " ")
		assert.Equal(t, "custom_slug = ? OR custom_slug = ? OR custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []interface{}{"cats", "cat", "dogs", "dog"}, v)
	})
	t.Run("Spoon", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "spoon", " ")
		assert.Equal(t, "custom_slug = ?", w)
		assert.Equal(t, []interface{}{"spoon"}, v)
	})
	t.Run("Img", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "img", " ")
		assert.Equal(t, "custom_slug = ?", w)
		assert.Equal(t, []interface{}{"img"}, v)
	})
	t.Run("Space", func(t *testing.T) {
		w, v := AnySlug("custom_slug", " ", "")
		assert.Equal(t, "custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []interface{}{"", ""}, v)
	})
	t.Run("Empty", func(t *testing.T) {
		where, values := AnySlug("custom_slug", "", " ")
		assert.Equal(t, "", where)
		assert.Empty(t, values)
	})
	t.Run("CommaSeparated", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "botanical-garden,landscape,bay", ",")
		assert.Equal(t, "custom_slug = ? OR custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []interface{}{"botanical-garden", "landscape", "bay"}, v)
	})
	t.Run("PipeSeparated", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "botanical-garden|landscape|bay", txt.Or)
		assert.Equal(t, "custom_slug = ? OR custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []interface{}{"botanical-garden", "landscape", "bay"}, v)
	})
	t.Run("Emoji", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "💐", "|")
		assert.Equal(t, "custom_slug = ?", w)
		assert.Equal(t, []interface{}{"_5cpzfea"}, v)
	})
	t.Run("EmojiSlug", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "_5cpzfea", "|")
		assert.Equal(t, "custom_slug = ?", w)
		assert.Equal(t, []interface{}{"_5cpzfea"}, v)
	})
}

func TestAnyInt(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		where, values := AnyInt("photos.photo_month", "", txt.Or, entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "", where)
		assert.Empty(t, values)
	})
	t.Run("Range", func(t *testing.T) {
		w, v := AnyInt("photos.photo_month", "-3|0|10|9|11|12|13", txt.Or, entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "photos.photo_month = ? OR photos.photo_month = ? OR photos.photo_month = ? OR photos.photo_month = ?", w)
		assert.Equal(t, []interface{}{10, 9, 11, 12}, v)
	})
	t.Run("Chars", func(t *testing.T) {
		where, values := AnyInt("photos.photo_month", "a|b|c", txt.Or, entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "", where)
		assert.Empty(t, values)
	})
	t.Run("CommaSeparated", func(t *testing.T) {
		w, v := AnyInt("photos.photo_month", "-3,10,9,11,12,13", ",", entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "photos.photo_month = ? OR photos.photo_month = ? OR photos.photo_month = ? OR photos.photo_month = ?", w)
		assert.Equal(t, []interface{}{10, 9, 11, 12}, v)
	})
	t.Run("Invalid", func(t *testing.T) {
		where, values := AnyInt("photos.photo_month", "  , |  ", ",", entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "", where)
		assert.Empty(t, values)
	})
}

func TestOrLike(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		where, values := OrLike("k.keyword", "")

		assert.Equal(t, "", where)
		assert.Equal(t, []any{}, values)
	})
	t.Run("OneTerm", func(t *testing.T) {
		where, values := OrLike("k.keyword", "bar")

		assert.Equal(t, "k.keyword LIKE ?", where)
		assert.Equal(t, []any{"bar"}, values)
	})
	t.Run("TwoTerms", func(t *testing.T) {
		where, values := OrLike("k.keyword", "foo*%|bar")

		assert.Equal(t, "k.keyword LIKE ? OR k.keyword LIKE ?", where)
		assert.Equal(t, []any{"foo%", "bar"}, values)
	})
	t.Run("OneFilename", func(t *testing.T) {
		where, values := OrLike("files.file_name", " 2790/07/27900704_070228_D6D51B6C.jpg")

		assert.Equal(t, "files.file_name LIKE ?", where)
		assert.Equal(t, []any{" 2790/07/27900704_070228_D6D51B6C.jpg"}, values)
	})
	t.Run("TwoFilenames", func(t *testing.T) {
		where, values := OrLike("files.file_name", "1990*|2790/07/27900704_070228_D6D51B6C.jpg")

		assert.Equal(t, "files.file_name LIKE ? OR files.file_name LIKE ?", where)
		assert.Equal(t, []any{"1990%", "2790/07/27900704_070228_D6D51B6C.jpg"}, values)
	})
}

func TestOrLikeCols(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		where, values := OrLikeCols([]string{"k.keyword", "p.photo_caption"}, "")

		assert.Equal(t, "", where)
		assert.Equal(t, []any{}, values)
	})
	t.Run("OneTerm", func(t *testing.T) {
		where, values := OrLikeCols([]string{"k.keyword", "p.photo_caption"}, "bar")

		assert.Equal(t, "k.keyword LIKE ? OR p.photo_caption LIKE ?", where)
		assert.Equal(t, []any{"bar", "bar"}, values)
	})
	t.Run("TwoTerms", func(t *testing.T) {
		where, values := OrLikeCols([]string{"k.keyword", "p.photo_caption"}, "foo*%|bar")

		assert.Equal(t, "k.keyword LIKE ? OR k.keyword LIKE ? OR p.photo_caption LIKE ? OR p.photo_caption LIKE ?", where)
		assert.Equal(t, []any{"foo%", "bar", "foo%", "bar"}, values)
	})
	t.Run("OneTermEscaped", func(t *testing.T) {
		where, values := OrLikeCols([]string{"k.keyword", "p.photo_caption"}, "\\|bar")

		assert.Equal(t, "k.keyword LIKE ? OR p.photo_caption LIKE ?", where)
		assert.Equal(t, []any{"|bar", "|bar"}, values)
	})
	t.Run("TwoTermsEscaped", func(t *testing.T) {
		where, values := OrLikeCols([]string{"k.keyword", "p.photo_caption"}, "foo*%|\\|bar")

		assert.Equal(t, "k.keyword LIKE ? OR k.keyword LIKE ? OR p.photo_caption LIKE ? OR p.photo_caption LIKE ?", where)
		assert.Equal(t, []any{"foo%", "|bar", "foo%", "|bar"}, values)
	})
	t.Run("OneFilename", func(t *testing.T) {
		where, values := OrLikeCols([]string{"files.file_name"}, " 2790/07/27900704_070228_D6D51B6C.jpg")

		assert.Equal(t, "files.file_name LIKE ?", where)
		assert.Equal(t, []any{" 2790/07/27900704_070228_D6D51B6C.jpg"}, values)
	})
	t.Run("TwoFilenames", func(t *testing.T) {
		where, values := OrLikeCols([]string{"files.file_name", "photos.photo_name"}, "1990*|2790/07/27900704_070228_D6D51B6C.jpg")

		assert.Equal(t, "files.file_name LIKE ? OR files.file_name LIKE ? OR photos.photo_name LIKE ? OR photos.photo_name LIKE ?", where)
		assert.Equal(t, []any{"1990%", "2790/07/27900704_070228_D6D51B6C.jpg", "1990%", "2790/07/27900704_070228_D6D51B6C.jpg"}, values)
	})
	t.Run("OneFilenameEscaped", func(t *testing.T) {
		where, values := OrLikeCols([]string{"files.file_name"}, " 2790/07/27900704_070228_D6D\\|51B6C.jpg")

		assert.Equal(t, "files.file_name LIKE ?", where)
		assert.Equal(t, []any{" 2790/07/27900704_070228_D6D|51B6C.jpg"}, values)
	})
	t.Run("TwoFilenamesEscaped", func(t *testing.T) {
		where, values := OrLikeCols([]string{"files.file_name", "photos.photo_name"}, "1990*|2790/07/27900704_070228_D6D\\|51B6C.jpg")

		assert.Equal(t, "files.file_name LIKE ? OR files.file_name LIKE ? OR photos.photo_name LIKE ? OR photos.photo_name LIKE ?", where)
		assert.Equal(t, []any{"1990%", "2790/07/27900704_070228_D6D|51B6C.jpg", "1990%", "2790/07/27900704_070228_D6D|51B6C.jpg"}, values)
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
