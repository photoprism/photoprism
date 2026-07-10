package pwa

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewScreenshots(t *testing.T) {
	t.Run("Standard", func(t *testing.T) {
		c := Config{StaticUri: "https://demo-cdn.photoprism.app/static"}
		result := NewScreenshots(c)
		assert.Len(t, result, 2)
		assert.Equal(t, "https://demo-cdn.photoprism.app/static/img/screenshots/wide.jpg", result[0].Src)
		assert.Equal(t, "1280x900", result[0].Sizes)
		assert.Equal(t, "image/jpeg", result[0].Type)
		assert.Equal(t, "wide", result[0].FormFactor)
		assert.Equal(t, "https://demo-cdn.photoprism.app/static/img/screenshots/narrow.jpg", result[1].Src)
		assert.Equal(t, "375x667", result[1].Sizes)
		assert.Equal(t, "narrow", result[1].FormFactor)
	})
}
