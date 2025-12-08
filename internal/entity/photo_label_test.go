package entity

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
)

func TestNewPhotoLabel(t *testing.T) {
	t.Run("NameChristmasNum2018", func(t *testing.T) {
		photoLabel := NewPhotoLabel(1, 3, 80, "source")
		assert.Equal(t, uint(0x1), photoLabel.PhotoID)
		assert.Equal(t, uint(0x3), photoLabel.LabelID)
		assert.Equal(t, 80, photoLabel.Uncertainty)
		assert.Equal(t, "source", photoLabel.LabelSrc)
	})
}

func TestPhotoLabel_TableName(t *testing.T) {
	photoLabel := &PhotoLabel{}
	tableName := photoLabel.TableName()

	assert.Equal(t, "photos_labels", tableName)
}

func TestFirstOrCreatePhotoLabel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		model := LabelFixtures.PhotoLabel(1000000, "flower", 38, "image")
		result := FirstOrCreatePhotoLabel(&model)

		if result == nil {
			t.Fatal("result must not be nil")
		}

		if result.PhotoID != model.PhotoID {
			t.Errorf("PhotoID should be the same: %d %d", result.PhotoID, model.PhotoID)
		}

		if result.LabelID != model.LabelID {
			t.Errorf("LabelID should be the same: %d %d", result.LabelID, model.LabelID)
		}
	})
	t.Run("Invalid", func(t *testing.T) {
		assert.Nil(t, FirstOrCreatePhotoLabel(NewPhotoLabel(0, 1, 38, "image")))
		assert.Nil(t, FirstOrCreatePhotoLabel(NewPhotoLabel(1, 0, 38, "image")))
	})
}

func TestPhotoLabel_ClassifyLabel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pl := LabelFixtures.PhotoLabel(1000000, "flower", 38, "image")
		r := pl.ClassifyLabel()
		assert.Equal(t, "Flower", r.Name)
		assert.Equal(t, 38, r.Uncertainty)
		assert.Equal(t, "image", r.Source)
	})
	t.Run("Invalid", func(t *testing.T) {
		photoLabel := NewPhotoLabel(1, 3, 80, "source")
		result := photoLabel.ClassifyLabel()
		assert.Equal(t, classify.Label{}, result)
	})
}

func TestPhotoLabel_Save(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		photoLabel := NewPhotoLabel(13, 1000, 99, "image")
		err := photoLabel.Save()
		if err != nil {
			t.Fatal(err)
		}
	})
	//TODO fails on mariadb
	t.Run("PhotoNotNilAndLabelNotNil", func(t *testing.T) {
		label := &Label{LabelName: "LabelSaveUnique", LabelSlug: "unique-slug"}
		photo := &Photo{}

		photoLabel := PhotoLabel{Photo: photo, Label: label}
		err := photoLabel.Save()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestPhotoLabel_Update(t *testing.T) {
	t.Run("FlushesCache", func(t *testing.T) {
		FlushPhotoLabelCache()
		relation := createTestPhotoLabel(t)

		photoLabelCache.SetDefault(relation.CacheKey(), *relation)

		require.NoError(t, relation.Update("uncertainty", relation.Uncertainty+1))

		_, found := photoLabelCache.Get(relation.CacheKey())
		assert.False(t, found)
	})

	t.Run("NilPhotoLabel", func(t *testing.T) {
		var label *PhotoLabel
		err := label.Update("uncertainty", 0)
		assert.EqualError(t, err, "photo label must not be nil - you may have found a bug")
	})

	t.Run("MissingID", func(t *testing.T) {
		label := &PhotoLabel{PhotoID: 0, LabelID: 1}
		err := label.Update("uncertainty", 0)
		assert.EqualError(t, err, "photo label ID must not be empty - you may have found a bug")
	})
}

func TestPhotoLabel_Updates(t *testing.T) {
	t.Run("FlushesCache", func(t *testing.T) {
		FlushPhotoLabelCache()
		relation := createTestPhotoLabel(t)

		photoLabelCache.SetDefault(relation.CacheKey(), *relation)

		require.NoError(t, relation.Updates(&PhotoLabel{Uncertainty: relation.Uncertainty + 1}))

		_, found := photoLabelCache.Get(relation.CacheKey())
		assert.False(t, found)
	})

	t.Run("NilPhotoLabel", func(t *testing.T) {
		var label *PhotoLabel
		err := label.Updates(&PhotoLabel{Uncertainty: 0})
		assert.EqualError(t, err, "photo label must not be nil - you may have found a bug")
	})

	t.Run("MissingID", func(t *testing.T) {
		label := &PhotoLabel{PhotoID: 0, LabelID: 1}
		err := label.Updates(&PhotoLabel{Uncertainty: 0})
		assert.EqualError(t, err, "photo label ID must not be empty - you may have found a bug")
	})
}

func TestPhotoLabel_Delete(t *testing.T) {
	FlushPhotoLabelCache()
	relation := createTestPhotoLabel(t)
	photoLabelCache.SetDefault(relation.CacheKey(), *relation)

	require.NoError(t, relation.Delete())

	_, found := photoLabelCache.Get(relation.CacheKey())
	assert.False(t, found)
}

func TestPhotoLabel_HasID(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var label *PhotoLabel
		assert.False(t, label.HasID())
	})

	t.Run("Missing", func(t *testing.T) {
		label := &PhotoLabel{PhotoID: 1}
		assert.False(t, label.HasID())
	})

	t.Run("Complete", func(t *testing.T) {
		label := &PhotoLabel{PhotoID: 1, LabelID: 2}
		assert.True(t, label.HasID())
	})
}

func TestPhotoLabel_CacheKey(t *testing.T) {
	label := &PhotoLabel{PhotoID: 1, LabelID: 2}
	assert.Equal(t, "1-2", label.CacheKey())
}

func createTestPhotoLabel(t *testing.T) *PhotoLabel {
	t.Helper()
	photo := &Photo{}
	require.NoError(t, Db().First(photo).Error)
	label := NewLabel(fmt.Sprintf("photo-label-test-%d", time.Now().UnixNano()), 0)
	require.NoError(t, label.Save())

	relation := NewPhotoLabel(photo.ID, label.ID, 0, SrcManual)
	require.NoError(t, relation.Create())

	t.Cleanup(func() {
		_ = Db().Where("photo_id = ? AND label_id = ?", relation.PhotoID, relation.LabelID).Delete(&PhotoLabel{}).Error
		_ = Db().Delete(label).Error
	})

	return relation
}
