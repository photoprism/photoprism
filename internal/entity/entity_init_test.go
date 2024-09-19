package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {

	t.Run("UserCounts", func(t *testing.T) {
		m := &User{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(13), count)
	})

	t.Run("PlaceCounts", func(t *testing.T) {
		m := &Place{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(10), count)
	})

	t.Run("Cell-Location-Counts", func(t *testing.T) {
		m := &Cell{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(9), count)
	})

	t.Run("CountryCounts", func(t *testing.T) {
		m := &Country{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(2), count)
	})

	t.Run("CameraCounts", func(t *testing.T) {
		m := &Camera{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(7), count)
	})

	t.Run("LensCounts", func(t *testing.T) {
		m := &Lens{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(3), count)
	})

	t.Run("LabelCounts", func(t *testing.T) {
		m := &Label{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(32), count)
	})

	t.Run("PhotoCounts", func(t *testing.T) {
		m := &Photo{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(55), count)
	})

	t.Run("AlbumCounts", func(t *testing.T) {
		m := &Album{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(30), count)
	})

	t.Run("ServiceCounts", func(t *testing.T) {
		m := &Service{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(2), count)
	})

	t.Run("LinkCounts", func(t *testing.T) {
		m := &Link{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(5), count)
	})

	t.Run("PhotoAlbumCounts", func(t *testing.T) {
		m := &PhotoAlbum{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(28), count)
	})

	t.Run("FolderCounts", func(t *testing.T) {
		m := &Folder{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(3), count)
	})

	t.Run("FileCounts", func(t *testing.T) {
		m := &File{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(65), count)
	})

	t.Run("KeywordCounts", func(t *testing.T) {
		m := &Keyword{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(26), count)
	})

	t.Run("PhotoKeywordCounts", func(t *testing.T) {
		m := &PhotoKeyword{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(31), count)
	})

	t.Run("CategoryCounts", func(t *testing.T) {
		m := &Category{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(1), count)
	})

	t.Run("FileShareCounts", func(t *testing.T) {
		m := &FileShare{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(2), count)
	})

	t.Run("FileSyncCounts", func(t *testing.T) {
		m := &FileSync{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(3), count)
	})

	t.Run("SubjectCounts", func(t *testing.T) {
		m := &Subject{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(6), count)
	})

	t.Run("MarkerCounts", func(t *testing.T) {
		m := &Marker{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(18), count)
	})

	t.Run("FaceCounts", func(t *testing.T) {
		m := &Face{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(7), count)
	})

	t.Run("SessionCounts", func(t *testing.T) {
		m := &Session{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(16), count)
	})

	t.Run("ClientCounts", func(t *testing.T) {
		m := &Client{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(7), count)
	})

	t.Run("ReactionCounts", func(t *testing.T) {
		m := &Reaction{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(3), count)
	})

	t.Run("PasscodeCounts", func(t *testing.T) {
		m := &Passcode{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(3), count)
	})

	t.Run("PasswordCounts", func(t *testing.T) {
		m := &Password{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(10), count)
	})

	t.Run("UserShareCounts", func(t *testing.T) {
		m := &UserShare{}
		stmt := UnscopedDb()
		count := int64(0)

		stmt.Model(m).Count(&count)

		assert.Equal(t, int64(1), count)
	})

}
