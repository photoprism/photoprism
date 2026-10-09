package search

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/txt"
)

func TestPathLike(t *testing.T) {
	t.Run("MySQLCaseInsensitiveCollation", func(t *testing.T) {
		// MySQL compares VARBINARY byte-exact, so the path is converted to a case-insensitive collation.
		assert.Equal(t, "CONVERT(albums.album_path USING utf8mb4) COLLATE utf8mb4_general_ci LIKE ? ESCAPE '!'", PathLike(dsn.DialectMySQL, "albums.album_path"))
	})
	t.Run("SQLitePlainLike", func(t *testing.T) {
		// SQLite LIKE is already ASCII case-insensitive, so a plain LIKE suffices.
		assert.Equal(t, "albums.album_path LIKE ? ESCAPE '!'", PathLike(dsn.DialectSQLite, "albums.album_path"))
	})
	t.Run("PostgreSQLDecodedILike", func(t *testing.T) {
		// PostgreSQL stores the path as bytes, so it is decoded and compared with ILIKE.
		assert.Equal(t, "convert_from(albums.album_path, 'UTF8') ILIKE ? ESCAPE '!'", PathLike(dsn.DialectPostgreSQL, "albums.album_path"))
	})
	t.Run("UnknownDialectFallsBackToPlainLike", func(t *testing.T) {
		// A future or unknown dialect must not error; it falls back to a plain LIKE.
		assert.Equal(t, "albums.album_path LIKE ? ESCAPE '!'", PathLike("", "albums.album_path"))
	})
}

func TestLikeExpr(t *testing.T) {
	t.Run("MySQLAndSQLite", func(t *testing.T) {
		for _, d := range []string{dsn.DialectMySQL, dsn.DialectSQLite} {
			assert.Equal(t, "k.keyword LIKE ? ESCAPE '!'", likeExpr(d, "k.keyword", false))
			assert.Equal(t, "photos.photo_path LIKE ? ESCAPE '!'", likeExpr(d, "photos.photo_path", true))
		}
	})
	t.Run("PostgreSQL", func(t *testing.T) {
		assert.Equal(t, "k.keyword ILIKE ? ESCAPE '!'", likeExpr(dsn.DialectPostgreSQL, "k.keyword", false))
		assert.Equal(t, "convert_from(photos.photo_path, 'UTF8') LIKE ? ESCAPE '!'", likeExpr(dsn.DialectPostgreSQL, "photos.photo_path", true))
	})
}

func TestSqlParam(t *testing.T) {
	t.Run("Literal", func(t *testing.T) {
		assert.Equal(t, "a!_b%", SqlParam(" a_b%", "", "%"))
		assert.Equal(t, "%a!!b%", SqlParam("a!b", "%", "%"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", SqlParam("", "", ""))
	})
	t.Run("Wildcards", func(t *testing.T) {
		assert.Equal(t, "%foo%", SqlParam("foo", "%", "%"))
	})
	t.Run("StripsTrimChars", func(t *testing.T) {
		// Leading/trailing operators and wildcards are stripped before the value is bound.
		assert.Equal(t, "spoon", SqlParam(" |&*%spoon%*&| ", "", ""))
	})
	t.Run("KeepsQuotes", func(t *testing.T) {
		// Quotes are preserved unchanged; the parameter binder is responsible for escaping.
		assert.Equal(t, "O'Reilly", SqlParam("O'Reilly", "", ""))
	})
	t.Run("StripsControl", func(t *testing.T) {
		assert.Equal(t, "ab", SqlParam("a\tb\n", "", ""))
	})
	t.Run("PrePostWrappers", func(t *testing.T) {
		assert.Equal(t, "<table%", SqlParam("table", "<", "%"))
	})
}

func TestLikeAny(t *testing.T) {
	t.Run("AndOrSearch", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "table spoon & usa | img json", true, false, false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"spoon%", "table%"}, {"json%", "usa"}}, v)
	})
	t.Run("ExactAndOrSearch", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "table spoon & usa | img json", true, true, false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"spoon", "table"}, {"json", "usa"}}, v)
	})
	t.Run("AndOrSearchEn", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "table spoon and usa or img json", true, false, false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"spoon%", "table%"}, {"json%", "usa"}}, v)
	})
	t.Run("TableSpoonUsaImgJson", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "table spoon usa img json", true, false, false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"json%", "spoon%", "table%", "usa"}}, v)
	})
	t.Run("CatDog", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "cat dog", true, false, false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"cat", "dog"}}, v)
	})
	t.Run("CatsDogs", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "cats dogs", true, false, false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"cats%", "cat", "dogs%", "dog"}}, v)
	})
	t.Run("Spoon", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "spoon", true, false, false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"spoon%"}}, v)
	})
	t.Run("Img", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "img", true, false, false)
		assert.Empty(t, w)
		assert.Empty(t, v)
	})
	t.Run("Empty", func(t *testing.T) {
		w, v := LikeAny("k.keyword", "", true, false, false)
		assert.Empty(t, w)
		assert.Empty(t, v)
	})
	t.Run("LengthAlignment", func(t *testing.T) {
		// Callers rely on wheres[i] and values[i] being aligned; assert it explicitly.
		w, v := LikeAny("k.keyword", "table spoon & usa | img json", true, false, false)
		assert.Equal(t, len(w), len(v))
	})
}

func TestLikeAnyKeyword(t *testing.T) {
	t.Run("AndOrSearch", func(t *testing.T) {
		w, v := LikeAnyKeyword("k.keyword", "table spoon & usa | img json", false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"spoon%", "table%"}, {"json%", "usa"}}, v)
	})
	t.Run("AndOrSearchEn", func(t *testing.T) {
		w, v := LikeAnyKeyword("k.keyword", "table spoon and usa or img json", false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"spoon%", "table%"}, {"json%", "usa"}}, v)
	})
}

func TestLikeAnyWord(t *testing.T) {
	t.Run("SearchAndOr", func(t *testing.T) {
		w, v := LikeAnyWord("k.keyword", "table spoon & usa | img json", false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"spoon%", "table%"}, {"img%", "json%", "usa%"}}, v)
	})
	t.Run("SearchAndOrEnglish", func(t *testing.T) {
		w, v := LikeAnyWord("k.keyword", "table spoon and usa or img json", false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"spoon%", "table%"}, {"img%", "json%", "usa%"}}, v)
	})
	t.Run("EscapeSql", func(t *testing.T) {
		// Quote characters survive in the bound value — the parameter binder, not the
		// query builder, is responsible for escaping them.
		w, v := LikeAnyWord("k.keyword", "table% | 'spoon' & \"us'a", false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"spoon%", "table%"}, {"\"us'a%"}}, v)
	})
}

func TestLikeAll(t *testing.T) {
	t.Run("Keywords", func(t *testing.T) {
		w, v := LikeAll("k.keyword", "Jo Mander 李", true, false, false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"mander%"}, {"李"}}, v)
	})
	t.Run("Exact", func(t *testing.T) {
		w, v := LikeAll("k.keyword", "Jo Mander 李", true, true, false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"mander"}, {"李"}}, v)
	})
	t.Run("StringEmpty", func(t *testing.T) {
		w, v := LikeAll("k.keyword", "", true, true, false)
		assert.Empty(t, w)
		assert.Empty(t, v)
	})
	t.Run("ZeroWords", func(t *testing.T) {
		w, v := LikeAll("k.keyword", "ab", true, true, false)
		assert.Empty(t, w)
		assert.Empty(t, v)
	})
	t.Run("LengthAlignment", func(t *testing.T) {
		w, v := LikeAll("k.keyword", "Jo Mander 李", true, false, false)
		assert.Equal(t, len(w), len(v))
	})
}

func TestLikeAllKeywords(t *testing.T) {
	t.Run("Keywords", func(t *testing.T) {
		w, v := LikeAllKeywords("k.keyword", "Jo Mander 李", false)
		assert.Equal(t, []string{"k.keyword LIKE ? ESCAPE '!'", "k.keyword LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"mander%"}, {"李"}}, v)
	})
}

func TestLikeAllWords(t *testing.T) {
	t.Run("Keywords", func(t *testing.T) {
		w, v := LikeAllWords("k.name", "Jo Mander 王", false)
		assert.Equal(t, []string{"k.name LIKE ? ESCAPE '!'", "k.name LIKE ? ESCAPE '!'", "k.name LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"jo%"}, {"mander%"}, {"王%"}}, v)
	})
}

func TestLikeAllNames(t *testing.T) {
	t.Run("MultipleNames", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"k.name"}, "j Mander 王", false)
		assert.Equal(t, []string{"k.name LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"j Mander 王%"}}, v)
	})
	t.Run("MultipleColumns", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"a.col1", "b.col2"}, "Mo Mander", false)
		assert.Equal(t, []string{"a.col1 LIKE ? ESCAPE '!' OR b.col2 LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"Mo Mander%", "Mo Mander%"}}, v)
	})
	t.Run("EmptyName", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"k.name"}, "", false)
		assert.Empty(t, w)
		assert.Empty(t, v)
	})
	t.Run("EmptyCols", func(t *testing.T) {
		w, v := LikeAllNames(Cols{}, "anything", false)
		assert.Empty(t, w)
		assert.Empty(t, v)
	})
	t.Run("SingleCharacter", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"k.name"}, "a", false)
		assert.Equal(t, []string{"k.name LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"%a%"}}, v)
	})
	t.Run("FullNames", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"j.name", "j.alias"}, "Bill & Melinda Gates", false)
		assert.Equal(t, []string{"j.name LIKE ? ESCAPE '!' OR j.alias LIKE ? ESCAPE '!'", "j.name LIKE ? ESCAPE '!' OR j.alias LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"%Bill%", "%Bill%"}, {"Melinda Gates%", "Melinda Gates%"}}, v)
	})
	t.Run("Plus", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"name"}, clean.SearchQuery("Paul + Paula"), false)
		assert.Equal(t, []string{"name LIKE ? ESCAPE '!'", "name LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"%Paul%"}, {"%Paula%"}}, v)
	})
	t.Run("And", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"name"}, clean.SearchQuery("P and Paula"), false)
		assert.Equal(t, []string{"name LIKE ? ESCAPE '!'", "name LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"%P%"}, {"%Paula%"}}, v)
	})
	t.Run("Or", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"name"}, clean.SearchQuery("Paul or Paula"), false)
		assert.Equal(t, []string{"name LIKE ? ESCAPE '!' OR name LIKE ? ESCAPE '!'"}, w)
		assert.Equal(t, [][]any{{"%Paul%", "%Paula%"}}, v)
	})
	t.Run("LengthAlignment", func(t *testing.T) {
		w, v := LikeAllNames(Cols{"j.name", "j.alias"}, "Bill & Melinda Gates", false)
		assert.Equal(t, len(w), len(v))
	})
}

func TestAnySlug(t *testing.T) {
	t.Run("Multiple", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "table spoon usa img json", " ")
		assert.Equal(t, "custom_slug = ? OR custom_slug = ? OR custom_slug = ? OR custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []any{"table", "spoon", "usa", "img", "json"}, v)
	})
	t.Run("CatDog", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "cat dog", " ")
		assert.Equal(t, "custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []any{"cat", "dog"}, v)
	})
	t.Run("CatsDogs", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "cats dogs", " ")
		assert.Equal(t, "custom_slug = ? OR custom_slug = ? OR custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []any{"cats", "cat", "dogs", "dog"}, v)
	})
	t.Run("Spoon", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "spoon", " ")
		assert.Equal(t, "custom_slug = ?", w)
		assert.Equal(t, []any{"spoon"}, v)
	})
	t.Run("Img", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "img", " ")
		assert.Equal(t, "custom_slug = ?", w)
		assert.Equal(t, []any{"img"}, v)
	})
	t.Run("Space", func(t *testing.T) {
		w, v := AnySlug("custom_slug", " ", "")
		assert.Equal(t, "custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []any{"", ""}, v)
	})
	t.Run("Empty", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "", " ")
		assert.Equal(t, "", w)
		assert.Empty(t, v)
	})
	t.Run("CommaSeparated", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "botanical-garden,landscape,bay", ",")
		assert.Equal(t, "custom_slug = ? OR custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []any{"botanical-garden", "landscape", "bay"}, v)
	})
	t.Run("PipeSeparated", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "botanical-garden|landscape|bay", txt.Or)
		assert.Equal(t, "custom_slug = ? OR custom_slug = ? OR custom_slug = ?", w)
		assert.Equal(t, []any{"botanical-garden", "landscape", "bay"}, v)
	})
	t.Run("Emoji", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "💐", "|")
		assert.Equal(t, "custom_slug = ?", w)
		assert.Equal(t, []any{"_5cpzfea"}, v)
	})
	t.Run("EmojiSlug", func(t *testing.T) {
		w, v := AnySlug("custom_slug", "_5cpzfea", "|")
		assert.Equal(t, "custom_slug = ?", w)
		assert.Equal(t, []any{"_5cpzfea"}, v)
	})
	t.Run("SqlInjectionAttempt", func(t *testing.T) {
		// Single quotes survive in the bound value because the parameter binder
		// escapes them; nothing leaks into the SQL fragment itself.
		w, v := AnySlug("custom_slug", "foo'; DROP TABLE photos--", "|")
		assert.Equal(t, "custom_slug = ?", w)
		assert.Equal(t, []any{"foo-drop-table-photos"}, v)
	})
}

func TestAnyInt(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		w, v := AnyInt("photos.photo_month", "", txt.Or, entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "", w)
		assert.Empty(t, v)
	})
	t.Run("Range", func(t *testing.T) {
		w, v := AnyInt("photos.photo_month", "-3|0|10|9|11|12|13", txt.Or, entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "photos.photo_month = ? OR photos.photo_month = ? OR photos.photo_month = ? OR photos.photo_month = ?", w)
		assert.Equal(t, []any{10, 9, 11, 12}, v)
	})
	t.Run("Chars", func(t *testing.T) {
		w, v := AnyInt("photos.photo_month", "a|b|c", txt.Or, entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "", w)
		assert.Empty(t, v)
	})
	t.Run("CommaSeparated", func(t *testing.T) {
		w, v := AnyInt("photos.photo_month", "-3,10,9,11,12,13", ",", entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "photos.photo_month = ? OR photos.photo_month = ? OR photos.photo_month = ? OR photos.photo_month = ?", w)
		assert.Equal(t, []any{10, 9, 11, 12}, v)
	})
	t.Run("Invalid", func(t *testing.T) {
		w, v := AnyInt("photos.photo_month", "  , |  ", ",", entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "", w)
		assert.Empty(t, v)
	})
	t.Run("SqlInjectionAttempt", func(t *testing.T) {
		// Any token that isn't a clean integer is discarded entirely — the SQL
		// fragment never sees the injected payload.
		w, v := AnyInt("photos.photo_month", "1; DROP TABLE photos|2", txt.Or, entity.UnknownMonth, txt.MonthMax)
		assert.Equal(t, "photos.photo_month = ?", w)
		assert.Equal(t, []any{2}, v)
	})
}

func TestOrLike(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		where, values := OrLike("k.keyword", "", false)

		assert.Equal(t, "", where)
		assert.Equal(t, []any{}, values)
	})
	t.Run("OneTerm", func(t *testing.T) {
		where, values := OrLike("k.keyword", "bar", false)

		assert.Equal(t, "k.keyword LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{"bar"}, values)
	})
	t.Run("TwoTerms", func(t *testing.T) {
		where, values := OrLike("k.keyword", "foo*%|bar", false)

		assert.Equal(t, "k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{"foo%", "bar"}, values)
	})
	t.Run("OneFilename", func(t *testing.T) {
		where, values := OrLike("files.file_name", " 2790/07/27900704_070228_D6D51B6C.jpg", true)

		assert.Equal(t, "files.file_name LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{" 2790/07/27900704!_070228!_D6D51B6C.jpg"}, values)
	})
	t.Run("TwoFilenames", func(t *testing.T) {
		where, values := OrLike("files.file_name", "1990*|2790/07/27900704_070228_D6D51B6C.jpg", true)

		assert.Equal(t, "files.file_name LIKE ? ESCAPE '!' OR files.file_name LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{"1990%", "2790/07/27900704!_070228!_D6D51B6C.jpg"}, values)
	})
}

func TestOrLikeCols(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		where, values := OrLikeCols([]string{"k.keyword", "p.photo_caption"}, "", false)

		assert.Equal(t, "", where)
		assert.Equal(t, []any{}, values)
	})
	t.Run("OneTerm", func(t *testing.T) {
		where, values := OrLikeCols([]string{"k.keyword", "p.photo_caption"}, "bar", false)

		assert.Equal(t, "k.keyword LIKE ? ESCAPE '!' OR p.photo_caption LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{"bar", "bar"}, values)
	})
	t.Run("TwoTerms", func(t *testing.T) {
		where, values := OrLikeCols([]string{"k.keyword", "p.photo_caption"}, "foo*%|bar", false)

		assert.Equal(t, "k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!' OR p.photo_caption LIKE ? ESCAPE '!' OR p.photo_caption LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{"foo%", "bar", "foo%", "bar"}, values)
	})
	t.Run("OneTermEscaped", func(t *testing.T) {
		where, values := OrLikeCols([]string{"k.keyword", "p.photo_caption"}, "\\|bar", false)

		assert.Equal(t, "k.keyword LIKE ? ESCAPE '!' OR p.photo_caption LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{"|bar", "|bar"}, values)
	})
	t.Run("TwoTermsEscaped", func(t *testing.T) {
		where, values := OrLikeCols([]string{"k.keyword", "p.photo_caption"}, "foo*%|\\|bar", false)

		assert.Equal(t, "k.keyword LIKE ? ESCAPE '!' OR k.keyword LIKE ? ESCAPE '!' OR p.photo_caption LIKE ? ESCAPE '!' OR p.photo_caption LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{"foo%", "|bar", "foo%", "|bar"}, values)
	})
	t.Run("OneFilename", func(t *testing.T) {
		where, values := OrLikeCols([]string{"files.file_name"}, " 2790/07/27900704_070228_D6D51B6C.jpg", true)

		assert.Equal(t, "files.file_name LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{" 2790/07/27900704!_070228!_D6D51B6C.jpg"}, values)
	})
	t.Run("TwoFilenames", func(t *testing.T) {
		where, values := OrLikeCols([]string{"files.file_name", "photos.photo_name"}, "1990*|2790/07/27900704_070228_D6D51B6C.jpg", true)

		assert.Equal(t, "files.file_name LIKE ? ESCAPE '!' OR files.file_name LIKE ? ESCAPE '!' OR photos.photo_name LIKE ? ESCAPE '!' OR photos.photo_name LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{"1990%", "2790/07/27900704!_070228!_D6D51B6C.jpg", "1990%", "2790/07/27900704!_070228!_D6D51B6C.jpg"}, values)
	})
	t.Run("OneFilenameEscaped", func(t *testing.T) {
		where, values := OrLikeCols([]string{"files.file_name"}, " 2790/07/27900704_070228_D6D\\|51B6C.jpg", true)

		assert.Equal(t, "files.file_name LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{" 2790/07/27900704!_070228!_D6D|51B6C.jpg"}, values)
	})
	t.Run("TwoFilenamesEscaped", func(t *testing.T) {
		where, values := OrLikeCols([]string{"files.file_name", "photos.photo_name"}, "1990*|2790/07/27900704_070228_D6D\\|51B6C.jpg", true)

		assert.Equal(t, "files.file_name LIKE ? ESCAPE '!' OR files.file_name LIKE ? ESCAPE '!' OR photos.photo_name LIKE ? ESCAPE '!' OR photos.photo_name LIKE ? ESCAPE '!'", where)
		assert.Equal(t, []any{"1990%", "2790/07/27900704!_070228!_D6D|51B6C.jpg", "1990%", "2790/07/27900704!_070228!_D6D|51B6C.jpg"}, values)
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
	t.Run("BoundsGroupCount", func(t *testing.T) {
		// Each group adds a condition of its own, so the count is bounded as well as
		// the input length.
		groups := SplitAnd(strings.TrimSuffix(strings.Repeat("aa&", 50000), "&"))

		assert.Len(t, groups, MaxSearchGroups)
	})
}

func TestClipSearchTerms(t *testing.T) {
	t.Run("WithinLimit", func(t *testing.T) {
		s := " " + strings.Repeat("日", 1300) + " "
		assert.Equal(t, s, ClipSearchTerms(s))
	})
	t.Run("AboveLimit", func(t *testing.T) {
		s := ClipSearchTerms(strings.Repeat("a", clean.LengthLimit*4))
		assert.Len(t, s, clean.LengthLimit)
		assert.True(t, utf8.ValidString(s))
	})
}

// TestConditionsBoundExpansion covers that a long search value cannot size the statement:
// every term expands into its own predicate and bind parameter, so input past the limit
// produces exactly what the clipped input does.
func TestConditionsBoundExpansion(t *testing.T) {
	// Distinct terms, so deduplication alone would not bound the result.
	terms := func(n int) string {
		out := make([]string, n)
		for i := range out {
			out[i] = fmt.Sprintf("q%d", i)
		}
		return strings.Join(out, "|")
	}

	long := terms(50000)
	clipped := ClipSearchTerms(long)

	assert.Greater(t, len(long), clean.LengthLimit*10)
	assert.Len(t, clipped, clean.LengthLimit)

	joined := func(wheres []string) string { return strings.Join(wheres, " AND ") }

	t.Run("LikeAllNames", func(t *testing.T) {
		a, av := LikeAllNames(Cols{"subj_name", "subj_alias"}, long, false)
		b, bv := LikeAllNames(Cols{"subj_name", "subj_alias"}, clipped, false)
		assert.Equal(t, joined(b), joined(a))
		assert.Equal(t, bv, av)
	})
	t.Run("LikeAllNamesDeduplicates", func(t *testing.T) {
		// Repeats collapse, as they do in the sibling builders.
		wheres, values := LikeAllNames(Cols{"subj_name"}, "jane|jane|jane", false)
		if assert.Len(t, wheres, 1) {
			assert.Equal(t, "subj_name LIKE ? ESCAPE '!'", wheres[0])
			assert.Len(t, values[0], 1)
		}
	})
	t.Run("LikeAllNamesDeduplicatesPerGroup", func(t *testing.T) {
		// Each AND group is deduplicated on its own, so a term shared between groups
		// survives in both and the groups keep their meaning.
		wheres, values := LikeAllNames(Cols{"subj_name"}, "a|b&b|c", false)
		if assert.Len(t, wheres, 2) {
			assert.Equal(t, []any{"%a%", "%b%"}, values[0])
			assert.Equal(t, []any{"%b%", "%c%"}, values[1])
		}
	})
	t.Run("LikeAllNamesKeepsDistinctTerms", func(t *testing.T) {
		wheres, values := LikeAllNames(Cols{"subj_name"}, "jane|john", false)
		if assert.Len(t, wheres, 1) {
			assert.Equal(t, "subj_name LIKE ? ESCAPE '!' OR subj_name LIKE ? ESCAPE '!'", wheres[0])
			assert.Equal(t, []any{"%jane%", "%john%"}, values[0])
		}
	})
	t.Run("LikeAnyKeyword", func(t *testing.T) {
		a, _ := LikeAnyKeyword("photos.photo_title", long, false)
		b, _ := LikeAnyKeyword("photos.photo_title", clipped, false)
		assert.Equal(t, joined(b), joined(a))
	})
	t.Run("LikeAllKeywords", func(t *testing.T) {
		a, _ := LikeAllKeywords("photos.photo_title", long, false)
		b, _ := LikeAllKeywords("photos.photo_title", clipped, false)
		assert.Equal(t, joined(b), joined(a))
	})
	t.Run("AnySlug", func(t *testing.T) {
		a, av := AnySlug("labels.label_slug", long, "|")
		b, bv := AnySlug("labels.label_slug", clipped, "|")
		assert.Equal(t, b, a)
		assert.Equal(t, bv, av)
	})
	t.Run("OrLike", func(t *testing.T) {
		a, av := OrLike("photos.photo_title", long, false)
		b, bv := OrLike("photos.photo_title", clipped, false)
		assert.Equal(t, b, a)
		assert.Equal(t, bv, av)
	})
	t.Run("OrLikeCols", func(t *testing.T) {
		a, av := OrLikeCols([]string{"a", "b"}, long, false)
		b, bv := OrLikeCols([]string{"a", "b"}, clipped, false)
		assert.Equal(t, b, a)
		assert.Equal(t, bv, av)
	})
}

func TestLikeCond(t *testing.T) {
	assert.Equal(t, "photos.photo_name LIKE ? ESCAPE '!'", likeCond("photos.photo_name", true))
}

func TestLikePattern(t *testing.T) {
	t.Run("Wildcards", func(t *testing.T) {
		assert.Equal(t, "IMG%", likePattern("IMG*"))
		assert.Equal(t, "IMG%", likePattern("IMG%"))
		assert.Equal(t, "IMG%", likePattern("IMG**"))
		assert.Equal(t, "a%b", likePattern("a***%%*b"))
		assert.Equal(t, "a%b", likePattern("a*b"))
	})
	t.Run("Literal", func(t *testing.T) {
		assert.Equal(t, "IMG!_1234", likePattern("IMG_1234"))
		assert.Equal(t, "Hi!!", likePattern("Hi!"))
	})
}

func TestSqlValue(t *testing.T) {
	assert.Equal(t, "cat", sqlValue(" *cat%| "))
	assert.Equal(t, "a_b", sqlValue("a_b"))
	assert.Equal(t, "", sqlValue("*%"))
}
