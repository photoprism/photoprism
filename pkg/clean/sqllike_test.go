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
		assert.Equal(t, "path LIKE ? ESCAPE '!'", SqlLikeCond("path"))
		assert.Equal(t, "photos.photo_path LIKE ? ESCAPE '!'", SqlLikeCond("photos.photo_path"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond(""))
		assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond("path) OR (1=1"))
		assert.Equal(t, "(1 = 0 AND ? IS NULL)", SqlLikeCond("a.b.c"))
	})
}

func TestSqlLikeExpr(t *testing.T) {
	t.Run("Column", func(t *testing.T) {
		assert.Equal(t, "REPLACE(REPLACE(REPLACE(a.path, '!', '!!'), '%', '!%'), '_', '!_')", SqlLikeExpr("a.path"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		assert.Equal(t, "NULL", SqlLikeExpr(""))
		assert.Equal(t, "NULL", SqlLikeExpr("a.path, '')--"))
	})
}

func TestSqlPrefixCond(t *testing.T) {
	t.Run("Column", func(t *testing.T) {
		assert.Equal(t, "(photos.photo_path LIKE ? ESCAPE '!' AND SUBSTR(photos.photo_path, 1, LENGTH(?)) = ?)", SqlPrefixCond("photos.photo_path"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL AND ? IS NULL)", SqlPrefixCond(""))
		assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL AND ? IS NULL)", SqlPrefixCond("path) OR (1=1"))
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
		assert.Equal(t, "(user_name LIKE ? ESCAPE '!')", SqlLikeAny("user_name"))
		assert.Equal(t, "(user_name LIKE ? ESCAPE '!' OR user_email LIKE ? ESCAPE '!')", SqlLikeAny("user_name", "user_email"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		assert.Equal(t, "(1 = 0)", SqlLikeAny())
		assert.Equal(t, "(1 = 0 AND ? IS NULL AND ? IS NULL)", SqlLikeAny("user_name", "x) OR (1=1"))
	})
}

func TestSqlNoMatch(t *testing.T) {
	assert.Equal(t, "(1 = 0)", sqlNoMatch(0))
	assert.Equal(t, "(1 = 0 AND ? IS NULL)", sqlNoMatch(1))
	assert.Equal(t, 3, strings.Count(sqlNoMatch(3), "?"))
}
