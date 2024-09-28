package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {

	t.Run("UserCounts", func(t *testing.T) {
		m := &entity.User{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(13), count)
	})

	t.Run("PlaceCounts", func(t *testing.T) {
		m := &entity.Place{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(10), count)
	})

	t.Run("Cell-Location-Counts", func(t *testing.T) {
		m := &entity.Cell{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(9), count)
	})

	t.Run("CountryCounts", func(t *testing.T) {
		m := &entity.Country{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(2), count)
	})

	t.Run("CameraCounts", func(t *testing.T) {
		m := &entity.Camera{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(7), count)
	})

	t.Run("LensCounts", func(t *testing.T) {
		m := &entity.Lens{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(3), count)
	})

	t.Run("LabelCounts", func(t *testing.T) {
		m := &entity.Label{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(32), count)
	})

	t.Run("PhotoLabelCounts", func(t *testing.T) {
		m := &entity.PhotoLabel{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(62), count)
	})

	t.Run("PhotoCounts", func(t *testing.T) {
		m := &entity.Photo{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(55), count)
	})

	t.Run("AlbumCounts", func(t *testing.T) {
		m := &entity.Album{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(31), count)
	})

	t.Run("ServiceCounts", func(t *testing.T) {
		m := &entity.Service{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(2), count)
	})

	t.Run("LinkCounts", func(t *testing.T) {
		m := &entity.Link{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(5), count)
	})

	t.Run("PhotoAlbumCounts", func(t *testing.T) {
		m := &entity.PhotoAlbum{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(29), count)
	})

	t.Run("FolderCounts", func(t *testing.T) {
		m := &entity.Folder{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(3), count)
	})

	t.Run("FileCounts", func(t *testing.T) {
		m := &entity.File{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		if entity.DbDialect() == "sqlite" {
			// sqlite allows "" as a FK value, which equates to NULL
			// See file_fixture 1000031
			assert.Equal(t, int64(67), count)
		} else {
			assert.Equal(t, int64(66), count)
		}

	})

	t.Run("KeywordCounts", func(t *testing.T) {
		m := &entity.Keyword{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(26), count)
	})

	t.Run("PhotoKeywordCounts", func(t *testing.T) {
		m := &entity.PhotoKeyword{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(35), count)
	})

	t.Run("CategoryCounts", func(t *testing.T) {
		m := &entity.Category{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(1), count)
	})

	t.Run("FileShareCounts", func(t *testing.T) {
		m := &entity.FileShare{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(2), count)
	})

	t.Run("FileSyncCounts", func(t *testing.T) {
		m := &entity.FileSync{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(3), count)
	})

	t.Run("SubjectCounts", func(t *testing.T) {
		m := &entity.Subject{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(6), count)
	})

	t.Run("MarkerCounts", func(t *testing.T) {
		m := &entity.Marker{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(18), count)
	})

	t.Run("FaceCounts", func(t *testing.T) {
		m := &entity.Face{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(7), count)
	})

	t.Run("SessionCounts", func(t *testing.T) {
		m := &entity.Session{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(16), count)
	})

	t.Run("ClientCounts", func(t *testing.T) {
		m := &entity.Client{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(7), count)
	})

	t.Run("ReactionCounts", func(t *testing.T) {
		m := &entity.Reaction{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(3), count)
	})

	t.Run("PasscodeCounts", func(t *testing.T) {
		m := &entity.Passcode{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(3), count)
	})

	t.Run("PasswordCounts", func(t *testing.T) {
		m := &entity.Password{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(10), count)
	})

	t.Run("UserShareCounts", func(t *testing.T) {
		m := &entity.UserShare{}
		stmt := entity.UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(1), count)
	})

}
