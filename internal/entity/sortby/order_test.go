package sortby

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestOrderExpr(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, "", OrderExpr("", false, dsn.DialectMySQL))
		assert.Equal(t, "photos.edited_at", OrderExpr("photos.edited_at", false, dsn.DialectMySQL))
		assert.Equal(t, "photos.edited_at ASC", OrderExpr("photos.edited_at ASC", false, dsn.DialectMySQL))
		assert.Equal(t, "photos.edited_at DESC, files.media_id", OrderExpr("photos.edited_at DESC, files.media_id", false, dsn.DialectMySQL))
		assert.Equal(t, "photos.edited_at DESC, files.media_id ASC", OrderExpr("photos.edited_at DESC, files.media_id ASC", false, dsn.DialectMySQL))
		assert.Equal(t, "photo_count DESC NULLS LAST, albums.album_title, albums.album_uid DESC", OrderExpr("photo_count DESC NULLS LAST, albums.album_title, albums.album_uid DESC", false, dsn.DialectMySQL))
	})
	t.Run("Reverse", func(t *testing.T) {
		assert.Equal(t, "", OrderExpr("", true, dsn.DialectMySQL))
		assert.Equal(t, "photos.edited_at", OrderExpr("photos.edited_at", true, dsn.DialectMySQL))
		assert.Equal(t, "photos.edited_at DESC", OrderExpr("photos.edited_at ASC", true, dsn.DialectMySQL))
		assert.Equal(t, "photos.edited_at ASC, files.media_id", OrderExpr("photos.edited_at DESC, files.media_id", true, dsn.DialectMySQL))
		assert.Equal(t, "photos.edited_at ASC, files.media_id DESC", OrderExpr("photos.edited_at DESC, files.media_id ASC", true, dsn.DialectMySQL))
		assert.Equal(t, "photo_count ASC NULLS FIRST, albums.album_title, albums.album_uid ASC", OrderExpr("photo_count DESC NULLS LAST, albums.album_title, albums.album_uid DESC", true, dsn.DialectMySQL))
	})
	t.Run("DefaultPostgreSQL", func(t *testing.T) {
		assert.Equal(t, "", OrderExpr("", false, dsn.DialectPostgreSQL))
		assert.Equal(t, "photos.photo_title COLLATE \"caseinsensitive\"", OrderExpr("photos.photo_title", false, dsn.DialectPostgreSQL))
		assert.Equal(t, "photos.photo_title COLLATE \"caseinsensitive\" ASC", OrderExpr("photos.photo_title ASC", false, dsn.DialectPostgreSQL))
		assert.Equal(t, "photos.photo_title COLLATE \"caseinsensitive\" DESC, albums.album_category COLLATE \"caseinsensitive\"", OrderExpr("photos.photo_title DESC, albums.album_category", false, dsn.DialectPostgreSQL))
		assert.Equal(t, "photos.photo_title COLLATE \"caseinsensitive\" DESC, albums.album_category COLLATE \"caseinsensitive\" ASC", OrderExpr("photos.photo_title DESC, albums.album_category ASC", false, dsn.DialectPostgreSQL))
		assert.Equal(t, "photo_count DESC NULLS LAST, albums.album_title COLLATE \"caseinsensitive\", albums.album_uid DESC", OrderExpr("photo_count DESC NULLS LAST, albums.album_title, albums.album_uid DESC", false, dsn.DialectPostgreSQL))
	})
	t.Run("ReversePostgreSQL", func(t *testing.T) {
		assert.Equal(t, "", OrderExpr("", true, dsn.DialectPostgreSQL))
		assert.Equal(t, "photos.photo_title COLLATE \"caseinsensitive\"", OrderExpr("photos.photo_title", true, dsn.DialectPostgreSQL))
		assert.Equal(t, "photos.photo_title COLLATE \"caseinsensitive\" DESC", OrderExpr("photos.photo_title ASC", true, dsn.DialectPostgreSQL))
		assert.Equal(t, "photos.photo_title COLLATE \"caseinsensitive\" ASC, albums.album_category COLLATE \"caseinsensitive\"", OrderExpr("photos.photo_title DESC, albums.album_category", true, dsn.DialectPostgreSQL))
		assert.Equal(t, "photos.photo_title COLLATE \"caseinsensitive\" ASC, albums.album_category COLLATE \"caseinsensitive\" DESC", OrderExpr("photos.photo_title DESC, albums.album_category ASC", true, dsn.DialectPostgreSQL))
		assert.Equal(t, "photo_count ASC NULLS FIRST, albums.album_title COLLATE \"caseinsensitive\", albums.album_uid ASC", OrderExpr("photo_count DESC NULLS LAST, albums.album_title, albums.album_uid DESC", true, dsn.DialectPostgreSQL))
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
