package clean

import (
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
		assert.Equal(t, "", SqlLikeCond(""))
		assert.Equal(t, "", SqlLikeCond("path) OR (1=1"))
		assert.Equal(t, "", SqlLikeCond("a.b.c"))
	})
}

func TestSqlLikeExpr(t *testing.T) {
	t.Run("Column", func(t *testing.T) {
		assert.Equal(t, "REPLACE(REPLACE(REPLACE(a.path, '!', '!!'), '%', '!%'), '_', '!_')", SqlLikeExpr("a.path"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		assert.Equal(t, "", SqlLikeExpr(""))
		assert.Equal(t, "", SqlLikeExpr("a.path, '')--"))
	})
}

func TestSqlPrefixCond(t *testing.T) {
	t.Run("Column", func(t *testing.T) {
		assert.Equal(t, "(photos.photo_path LIKE ? ESCAPE '!' AND SUBSTR(photos.photo_path, 1, LENGTH(?)) = ?)", SqlPrefixCond("photos.photo_path"))
	})
	t.Run("InvalidColumn", func(t *testing.T) {
		assert.Equal(t, "", SqlPrefixCond(""))
		assert.Equal(t, "", SqlPrefixCond("path) OR (1=1"))
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
