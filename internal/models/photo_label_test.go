package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPhotoLabel(t *testing.T) {
	t.Run("name Christmas 2018", func(t *testing.T) {
		photoLabel := NewPhotoLabel(1, 3, 80, "source")
		assert.Equal(t, uint(0x1), photoLabel.PhotoID)
		assert.Equal(t, uint(0x3), photoLabel.LabelID)
		assert.Equal(t, 80, photoLabel.LabelUncertainty)
		assert.Equal(t, "source", photoLabel.LabelSource)
	})
}
