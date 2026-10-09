package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestSqlLike(t *testing.T) {
	t.Run("Wildcards", func(t *testing.T) {
		assert.Equal(t, "a!_b", SqlLike("a_b"))
		assert.Equal(t, "50!%", SqlLike("50%"))
		assert.Equal(t, "!%!_!%", SqlLike("%_%"))
	})
	t.Run("EscapeCharacter", func(t *testing.T) {
		assert.Equal(t, "Hi!!", SqlLike("Hi!"))
		assert.Equal(t, "!!!_", SqlLike("!_"))
	})
	t.Run("Unchanged", func(t *testing.T) {
		assert.Equal(t, "", SqlLike(""))
		assert.Equal(t, "2024/Holiday*\\x", SqlLike("2024/Holiday*\\x"))
		assert.Equal(t, "Café 🎉", SqlLike("Café 🎉"))
	})
}

// sqlDialects are the dialects whose conditions are the same, as MySQL and SQLite share one syntax.
var sqlDialects = []string{dsn.DialectMySQL, dsn.DialectSQLite}

func TestSqlLikeCond(t *testing.T) {
	t.Run("Column", func(t *testing.T) {
		for _, d := range sqlDialects {
			for _, binary := range []bool{false, true} {
				assert.Equal(t, "path LIKE ? ESCAPE '!'", SqlLikeCond(d, binary, "path"))
				assert.Equal(t, "photos.photo_path LIKE ? ESCAPE '!'", SqlLikeCond(d, binary, "photos.photo_path"))
			}
		}
	})
	t.Run("PostgreSQL", func(t *testing.T) {
		assert.Equal(t, "user_name ILIKE ? ESCAPE '!'", SqlLikeCond(dsn.DialectPostgreSQL, false, "user_name"))
		assert.Equal(t, "convert_from(photos.photo_path, 'UTF8') LIKE ? ESCAPE '!'", SqlLikeCond(dsn.DialectPostgreSQL, true, "photos.photo_path"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		for _, d := range sqlDialects {
			assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond(d, false, ""))
			assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond(d, true, "path) OR (1=1"))
			assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond(d, false, "a.b.c"))
		}

		assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL)", SqlLikeCond(dsn.DialectPostgreSQL, true, "path) OR (1=1"))
	})
}

func TestSqlLikeExpr(t *testing.T) {
	t.Run("Column", func(t *testing.T) {
		for _, d := range sqlDialects {
			assert.Equal(t, "REPLACE(REPLACE(REPLACE(a.path, '!', '!!'), '%', '!%'), '_', '!_')", SqlLikeExpr(d, true, "a.path"))
		}
	})
	t.Run("PostgreSQL", func(t *testing.T) {
		assert.Equal(t, "REPLACE(REPLACE(REPLACE(convert_from(a.path, 'UTF8'), '!', '!!'), '%', '!%'), '_', '!_')", SqlLikeExpr(dsn.DialectPostgreSQL, true, "a.path"))
		assert.Equal(t, "REPLACE(REPLACE(REPLACE(a.title, '!', '!!'), '%', '!%'), '_', '!_')", SqlLikeExpr(dsn.DialectPostgreSQL, false, "a.title"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		for _, d := range append(sqlDialects, dsn.DialectPostgreSQL) {
			assert.Equal(t, "NULL", SqlLikeExpr(d, true, ""))
			assert.Equal(t, "NULL", SqlLikeExpr(d, true, "a.path, '')--"))
		}
	})
}

func TestSqlPrefixCond(t *testing.T) {
	t.Run("Column", func(t *testing.T) {
		for _, d := range sqlDialects {
			assert.Equal(t, "(photos.photo_path LIKE ? ESCAPE '!' AND SUBSTR(photos.photo_path, 1, LENGTH(?)) = ?)", SqlPrefixCond(d, true, "photos.photo_path"))
		}
	})
	t.Run("PostgreSQL", func(t *testing.T) {
		// The prefix is compared case-sensitively, so a binary column is decoded and matched with LIKE.
		assert.Equal(t, "(convert_from(photos.photo_path, 'UTF8') LIKE ? ESCAPE '!' AND SUBSTR(convert_from(photos.photo_path, 'UTF8'), 1, LENGTH(?)) = ?)",
			SqlPrefixCond(dsn.DialectPostgreSQL, true, "photos.photo_path"))
		assert.Equal(t, "(user_name LIKE ? ESCAPE '!' AND SUBSTR(user_name, 1, LENGTH(?)) = ?)", SqlPrefixCond(dsn.DialectPostgreSQL, false, "user_name"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		for _, d := range sqlDialects {
			assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL AND ? IS NULL)", SqlPrefixCond(d, true, ""))
			assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL AND ? IS NULL)", SqlPrefixCond(d, true, "path) OR (1=1"))
		}
	})
}

func TestSqlPrefixArgs(t *testing.T) {
	t.Run("Escaped", func(t *testing.T) {
		assert.Equal(t, []any{"a!_b!!!%/%", "a_b!%/", "a_b!%/"}, SqlPrefixArgs("a_b!%/"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, []any{"%", "", ""}, SqlPrefixArgs(""))
	})
}

func TestSqlLikeAny(t *testing.T) {
	t.Run("Columns", func(t *testing.T) {
		for _, d := range sqlDialects {
			assert.Equal(t, "(user_name LIKE ? ESCAPE '!')", SqlLikeAny(d, false, "user_name"))
			assert.Equal(t, "(user_name LIKE ? ESCAPE '!' OR user_email LIKE ? ESCAPE '!')", SqlLikeAny(d, false, "user_name", "user_email"))
		}
	})
	t.Run("PostgreSQL", func(t *testing.T) {
		assert.Equal(t, "(user_name ILIKE ? ESCAPE '!' OR user_email ILIKE ? ESCAPE '!')", SqlLikeAny(dsn.DialectPostgreSQL, false, "user_name", "user_email"))
		assert.Equal(t, "(convert_from(auth_id, 'UTF8') LIKE ? ESCAPE '!')", SqlLikeAny(dsn.DialectPostgreSQL, true, "auth_id"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		for _, d := range sqlDialects {
			assert.Equal(t, "(1 = 0)", SqlLikeAny(d, false))
			assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL)", SqlLikeAny(d, false, "user_name", "x) OR (1=1"))
		}
	})
}

func TestSqlNoMatch(t *testing.T) {
	for _, d := range sqlDialects {
		assert.Equal(t, "(1 = 0)", sqlNoMatch(d, 0))
		assert.Equal(t, "(1 = 0 AND ? IS NULL)", sqlNoMatch(d, 1))
		assert.Equal(t, 3, strings.Count(sqlNoMatch(d, 3), "?"))
	}

	assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL)", sqlNoMatch(dsn.DialectPostgreSQL, 1))
	assert.Equal(t, 3, strings.Count(sqlNoMatch(dsn.DialectPostgreSQL, 3), "?"))
}

func TestSqlText(t *testing.T) {
	assert.Equal(t, "convert_from(path, 'UTF8')", sqlText(dsn.DialectPostgreSQL, true, "path"))
	assert.Equal(t, "title", sqlText(dsn.DialectPostgreSQL, false, "title"))
	assert.Equal(t, "path", sqlText(dsn.DialectMySQL, true, "path"))
	assert.Equal(t, "path", sqlText(dsn.DialectSQLite, true, "path"))
}
