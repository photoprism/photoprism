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
