package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestOrderExpr(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "", OrderExpr("", false))
		assert.Equal(t, "photos.edited_at", OrderExpr("photos.edited_at", false))
		assert.Equal(t, "photos.edited_at ASC", OrderExpr("photos.edited_at ASC", false))
		assert.Equal(t, "photos.edited_at DESC, files.media_id", OrderExpr("photos.edited_at DESC, files.media_id", false))
		assert.Equal(t, "photos.edited_at DESC, files.media_id ASC", OrderExpr("photos.edited_at DESC, files.media_id ASC", false))
	})
	t.Run("Reverse", func(t *testing.T) {
		assert.Equal(t, "", OrderExpr("", true))
		assert.Equal(t, "photos.edited_at", OrderExpr("photos.edited_at", true))
		assert.Equal(t, "photos.edited_at DESC", OrderExpr("photos.edited_at ASC", true))
		assert.Equal(t, "photos.edited_at ASC, files.media_id", OrderExpr("photos.edited_at DESC, files.media_id", true))
		assert.Equal(t, "photos.edited_at ASC, files.media_id DESC", OrderExpr("photos.edited_at DESC, files.media_id ASC", true))
	})
}

func TestOrderExprNulls(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		if entity.DbDialect() == dsn.DialectPostgreSQL {
			assert.Equal(t, "login_at DESC NULLS LAST, id ASC NULLS FIRST", OrderExprNulls("login_at DESC, id ASC", false))
		} else {
			assert.Equal(t, "login_at DESC, id ASC", OrderExprNulls("login_at DESC, id ASC", false))
		}
	})
	t.Run("Reverse", func(t *testing.T) {
		if entity.DbDialect() == dsn.DialectPostgreSQL {
			assert.Equal(t, "login_at ASC NULLS FIRST, id DESC NULLS LAST", OrderExprNulls("login_at DESC, id ASC", true))
		} else {
			assert.Equal(t, "login_at ASC, id DESC", OrderExprNulls("login_at DESC, id ASC", true))
		}
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, "", OrderExprNulls("", false))
	})
}
