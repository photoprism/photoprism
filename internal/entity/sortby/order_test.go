package sortby

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestOrderAsc(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, DirAsc, OrderAsc(false))
	})
	t.Run("Reverse", func(t *testing.T) {
		assert.Equal(t, DirDesc, OrderAsc(true))
	})
}

func TestOrderDesc(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, DirDesc, OrderDesc(false))
	})
	t.Run("Reverse", func(t *testing.T) {
		assert.Equal(t, DirAsc, OrderDesc(true))
	})
}
