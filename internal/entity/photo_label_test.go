package entity

import (
	"testing"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/stretchr/testify/assert"
)

func TestNewPhotoLabel(t *testing.T) {
	t.Run("name Christmas 2018", func(t *testing.T) {
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
	t.Run("success path 1", func(t *testing.T) {
		model := LabelFixtures.PhotoLabel(1000000, "flower", 38, "image")
		result := FirstOrCreatePhotoLabel(&model)

		if result == nil {
			t.Fatal("result should not be nil")
		}

		assert.NotEqual(t, uint(0x0), model.Label.ID)
		// Validate Preload
		assert.NotNil(t, result.Label)
		if result.Label != nil {
			// Do this way to prevent SIGSEGV
			assert.Equal(t, "Flower", result.Label.LabelName)
		}

		if result.PhotoID != model.PhotoID {
			t.Errorf("PhotoID should be the same: %d %d", result.PhotoID, model.PhotoID)
		}

		if result.LabelID != model.LabelID {
			t.Errorf("LabelID should be the same: %d %d", result.LabelID, model.LabelID)
		}
	})

	t.Run("success path 2", func(t *testing.T) {
		model := LabelFixtures.PhotoLabel(1000000, "flowerz", 38, "image")
		assert.Equal(t, uint(0x0), model.LabelID)
		result := FirstOrCreatePhotoLabel(&model)

		if result == nil {
			t.Fatal("result should not be nil")
		}

		assert.NotEqual(t, uint(0x0), model.LabelID)
		// Validate Preload
		assert.NotNil(t, result.Label)
		if result.Label != nil {
			// Do this way to prevent SIGSEGV
			assert.Equal(t, "Flowerz", result.Label.LabelName)
		}

		if result.PhotoID != model.PhotoID {
			t.Errorf("PhotoID should be the same: %d %d", result.PhotoID, model.PhotoID)
		}

		if result.LabelID != model.LabelID {
			t.Errorf("LabelID should be the same: %d %d", result.LabelID, model.LabelID)
		}
	})

}

func TestPhotoLabel_ClassifyLabel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		pl := LabelFixtures.PhotoLabel(1000000, "flower", 38, "image")
		r := pl.ClassifyLabel()
		assert.Equal(t, "Flower", r.Name)
		assert.Equal(t, 38, r.Uncertainty)
		assert.Equal(t, "image", r.Source)
	})

	t.Run("label = nil", func(t *testing.T) {
		photoLabel := NewPhotoLabel(1, 3, 80, "source")
		result := photoLabel.ClassifyLabel()
		assert.Equal(t, classify.Label{}, result)
	})
}

func TestPhotoLabel_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		newPhoto := &Photo{ID: 567286} // Can't add details if there isn't a photo in the database.
		Db().Create(newPhoto)
		newLabel := &Label{ID: 567383, LabelSlug: "MustBeUnique"}
		Db().Create(newLabel)

		photoLabel := NewPhotoLabel(newPhoto.ID, newLabel.ID, 99, "image")
		err := photoLabel.Save()
		if err != nil {
			t.Fatal(err)
		}
		UnscopedDb().Delete(photoLabel)
		UnscopedDb().Delete(newLabel)
		UnscopedDb().Delete(newPhoto)
	})

	t.Run("photo not nil and label not nil", func(t *testing.T) {
		newLabel := &Label{LabelName: "LabelSaveUnique", LabelSlug: "unique-slug"}
		Db().Create(newLabel) // Foreign keys require the data to be saved.
		newPhoto := &Photo{}
		Db().Create(newPhoto)

		assert.NotEqual(t, 0, newPhoto.ID)
		assert.NotEqual(t, 0, newLabel.ID)

		photoLabel := PhotoLabel{Photo: newPhoto, Label: newLabel}
		err := photoLabel.Save()
		if err != nil {
			t.Fatal(err)
		}
		UnscopedDb().Delete(photoLabel)
		UnscopedDb().Delete(newLabel)
		UnscopedDb().Delete(newPhoto)
	})

	t.Run("photo nil and label not nil", func(t *testing.T) {
		newLabel := &Label{LabelName: "LabelSaveUnique", LabelSlug: "unique-slug"}
		Db().Create(newLabel) // Foreign keys require the data to be saved.

		assert.NotEqual(t, 0, newLabel.ID)

		photoLabel := PhotoLabel{Photo: nil, Label: newLabel}
		err := photoLabel.Save()
		assert.ErrorContains(t, err, "PK value not provided")
		UnscopedDb().Delete(newLabel)
	})
	t.Run("photo zero ID and label not nil", func(t *testing.T) {
		newLabel := &Label{LabelName: "LabelSaveUnique", LabelSlug: "unique-slug"}
		Db().Create(newLabel) // Foreign keys require the data to be saved.
		newPhoto := &Photo{PhotoUID: "Ameaninglessstring"}

		assert.NotEqual(t, 0, newLabel.ID)
		assert.Equal(t, uint(0), newPhoto.ID)

		photoLabel := PhotoLabel{Photo: newPhoto, Label: newLabel}
		err := photoLabel.Save()
		assert.ErrorContains(t, err, "PK value not provided")
		UnscopedDb().Delete(newLabel)
	})

}

func TestPhotoLabel_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photoLabel := PhotoLabel{LabelID: 555, PhotoID: 888}
		assert.Equal(t, uint(0x22b), photoLabel.LabelID)

		err := photoLabel.Update("LabelID", 8)

		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, uint(0x8), photoLabel.LabelID)
	})
}
