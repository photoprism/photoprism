package clean

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestSqlLikeCond(t *testing.T) {
	t.Run("Column", func(t *testing.T) {
		assert.Equal(t, "path LIKE ? ESCAPE '!'", SqlLikeCond(false, false, "path"))
		assert.Equal(t, "path ILIKE ? ESCAPE '!'", SqlLikeCond(true, false, "path"))
		assert.Equal(t, "path LIKE ? ESCAPE '!'", SqlLikeCond(false, true, "path"))
		assert.Equal(t, "convert_from(path, 'UTF8') ILIKE ? ESCAPE '!'", SqlLikeCond(true, true, "path"))
		assert.Equal(t, "photos.photo_path LIKE ? ESCAPE '!'", SqlLikeCond(false, false, "photos.photo_path"))
		assert.Equal(t, "photos.photo_path ILIKE ? ESCAPE '!'", SqlLikeCond(true, false, "photos.photo_path"))
		assert.Equal(t, "photos.photo_path LIKE ? ESCAPE '!'", SqlLikeCond(false, true, "photos.photo_path"))
		assert.Equal(t, "convert_from(photos.photo_path, 'UTF8') ILIKE ? ESCAPE '!'", SqlLikeCond(true, true, "photos.photo_path"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond(false, false, ""))
		assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL)", SqlLikeCond(true, false, ""))
		assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond(false, true, ""))
		assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL)", SqlLikeCond(true, true, ""))
		assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond(false, false, "path) OR (1=1"))
		assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond(false, false, "a.b.c"))
		assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond(false, true, "path) OR (1=1"))
		assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond(false, true, "a.b.c"))
	})
}

func TestSqlLikeExpr(t *testing.T) {
	t.Run("Column", func(t *testing.T) {
		assert.Equal(t, "REPLACE(REPLACE(REPLACE(a.path, '!', '!!'), '%', '!%'), '_', '!_')", SqlLikeExpr(false, true, "a.path"))
		assert.Equal(t, "REPLACE(REPLACE(REPLACE(a.path, '!', '!!'), '%', '!%'), '_', '!_')", SqlLikeExpr(true, false, "a.path"))
		assert.Equal(t, "REPLACE(REPLACE(REPLACE(convert_from(a.path, 'UTF8'), '!', '!!'), '%', '!%'), '_', '!_')", SqlLikeExpr(true, true, "a.path"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		assert.Equal(t, "NULL", SqlLikeExpr(false, true, ""))
		assert.Equal(t, "NULL", SqlLikeExpr(false, true, "a.path, '')--"))
		assert.Equal(t, "NULL", SqlLikeExpr(true, true, ""))
		assert.Equal(t, "NULL", SqlLikeExpr(true, true, "a.path, '')--"))
	})
}

func TestSqlPrefixCond(t *testing.T) {
	t.Run("Column", func(t *testing.T) {
		assert.Equal(t, "(photos.photo_path LIKE ? ESCAPE '!' AND SUBSTR(photos.photo_path, 1, LENGTH(?)) = ?)", SqlPrefixCond(false, false, "photos.photo_path"))
		assert.Equal(t, "(photos.photo_path ILIKE ? ESCAPE '!' AND SUBSTR(photos.photo_path, 1, LENGTH(?)) = ?)", SqlPrefixCond(true, false, "photos.photo_path"))
		assert.Equal(t, "(photos.photo_path LIKE ? ESCAPE '!' AND SUBSTR(photos.photo_path, 1, LENGTH(?)) = ?)", SqlPrefixCond(false, true, "photos.photo_path"))
		assert.Equal(t, "(convert_from(photos.photo_path, 'UTF8') ILIKE ? ESCAPE '!' AND SUBSTR(photos.photo_path, 1, LENGTH(?)) = ?)", SqlPrefixCond(true, true, "photos.photo_path"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL AND ? IS NULL)", SqlPrefixCond(false, false, ""))
		assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL AND ? IS NULL)", SqlPrefixCond(false, false, "path) OR (1=1"))
		assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL AND ? IS NULL)", SqlPrefixCond(false, true, ""))
		assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL AND ? IS NULL)", SqlPrefixCond(false, true, "path) OR (1=1"))
		assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL AND CAST(? AS VARCHAR) IS NULL AND CAST(? AS VARCHAR) IS NULL)", SqlPrefixCond(true, false, ""))
		assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL AND CAST(? AS VARCHAR) IS NULL AND CAST(? AS VARCHAR) IS NULL)", SqlPrefixCond(true, false, "path) OR (1=1"))
		assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL AND CAST(? AS VARCHAR) IS NULL AND CAST(? AS VARCHAR) IS NULL)", SqlPrefixCond(true, true, ""))
		assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL AND CAST(? AS VARCHAR) IS NULL AND CAST(? AS VARCHAR) IS NULL)", SqlPrefixCond(true, true, "path) OR (1=1"))
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
		assert.Equal(t, "(user_name LIKE ? ESCAPE '!')", SqlLikeAny(false, false, "user_name"))
		assert.Equal(t, "(user_name LIKE ? ESCAPE '!' OR user_email LIKE ? ESCAPE '!')", SqlLikeAny(false, false, "user_name", "user_email"))
		assert.Equal(t, "(user_name LIKE ? ESCAPE '!')", SqlLikeAny(false, true, "user_name"))
		assert.Equal(t, "(user_name LIKE ? ESCAPE '!' OR user_email LIKE ? ESCAPE '!')", SqlLikeAny(false, true, "user_name", "user_email"))
		assert.Equal(t, "(user_name ILIKE ? ESCAPE '!')", SqlLikeAny(true, false, "user_name"))
		assert.Equal(t, "(user_name ILIKE ? ESCAPE '!' OR user_email ILIKE ? ESCAPE '!')", SqlLikeAny(true, false, "user_name", "user_email"))
		assert.Equal(t, "(convert_from(user_name, 'UTF8') ILIKE ? ESCAPE '!')", SqlLikeAny(true, true, "user_name"))
		assert.Equal(t, "(convert_from(user_name, 'UTF8') ILIKE ? ESCAPE '!' OR convert_from(user_email, 'UTF8') ILIKE ? ESCAPE '!')", SqlLikeAny(true, true, "user_name", "user_email"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		assert.Equal(t, "(1 = 0)", SqlLikeAny(false, false))
		assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL)", SqlLikeAny(false, false, "user_name", "x) OR (1=1"))
		assert.Equal(t, "(1 = 0)", SqlLikeAny(false, true))
		assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL)", SqlLikeAny(false, true, "user_name", "x) OR (1=1"))
		assert.Equal(t, "(1 = 0)", SqlLikeAny(true, false))
		assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL AND CAST(? AS VARCHAR) IS NULL)", SqlLikeAny(true, false, "user_name", "x) OR (1=1"))
		assert.Equal(t, "(1 = 0)", SqlLikeAny(true, true))
		assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL AND CAST(? AS VARCHAR) IS NULL)", SqlLikeAny(true, true, "user_name", "x) OR (1=1"))
	})
}

func TestSqlNoMatch(t *testing.T) {
	assert.Equal(t, "(1 = 0)", sqlNoMatch(false, 0))
	assert.Equal(t, "(1 = 0 AND ? IS NULL)", sqlNoMatch(false, 1))
	assert.Equal(t, 3, strings.Count(sqlNoMatch(false, 3), "?"))
	assert.Equal(t, "(1 = 0 AND CAST(? AS VARCHAR) IS NULL)", sqlNoMatch(true, 1))
	assert.Equal(t, 3, strings.Count(sqlNoMatch(true, 3), "?"))
}
