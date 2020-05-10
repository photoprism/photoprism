package entity

import (
	"testing"

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

func TestPhotoLabel_FirstOrCreate(t *testing.T) {
	pl := LabelFixtures.PhotoLabel(1000000, "flower", 38, "image")
	r := pl.FirstOrCreate()
	assert.Equal(t, uint(1000000), r.PhotoID)
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
		assert.Panics(t, func() {

			photoLabel.ClassifyLabel()

		}, "photo label: label is nil")
	})
}

func TestPhotoLabel_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		photoLabel := NewPhotoLabel(13, 1000, 99, "image")
		err := photoLabel.Save()
		if err != nil {
			t.Fatal(err)
		}
	})
}
