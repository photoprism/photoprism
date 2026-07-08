package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCameraFisheyeFov(t *testing.T) {
	t.Run("InstaX", func(t *testing.T) {
		assert.Equal(t, 204, CameraFisheyeFov("Insta360", "Insta360 X4"))
	})
	t.Run("InstaOne", func(t *testing.T) {
		assert.Equal(t, 200, CameraFisheyeFov("Arashi Vision", "Insta360 ONE"))
	})
	t.Run("Theta", func(t *testing.T) {
		assert.Equal(t, 200, CameraFisheyeFov("RICOH", "RICOH THETA Z1"))
	})
	t.Run("MakerOnlyInsta360", func(t *testing.T) {
		// Bare model name; identified by the maker field (the makeName parameter must be used).
		assert.Equal(t, 200, CameraFisheyeFov("Insta360", "X4"))
	})
	t.Run("MakerOnlyArashiVision", func(t *testing.T) {
		assert.Equal(t, 200, CameraFisheyeFov("Arashi Vision", ""))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, 0, CameraFisheyeFov("Canon", "Canon EOS 6D"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, 0, CameraFisheyeFov("", ""))
	})
}
