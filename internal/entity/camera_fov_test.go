package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCameraFisheyeFov(t *testing.T) {
	ValidateFixtures(t)
	t.Run("InstaX", func(t *testing.T) {
		assert.Equal(t, 190, CameraFisheyeFov("Insta360", "Insta360 X4"))
	})
	t.Run("InstaOne", func(t *testing.T) {
		assert.Equal(t, 190, CameraFisheyeFov("Arashi Vision", "Insta360 ONE"))
	})
	t.Run("InstaOneRS", func(t *testing.T) {
		assert.Equal(t, 190, CameraFisheyeFov("Arashi Vision", "Insta360 OneRS"))
	})
	t.Run("Theta", func(t *testing.T) {
		assert.Equal(t, 190, CameraFisheyeFov("RICOH", "RICOH THETA Z1"))
	})
	t.Run("MakerOnlyInsta360", func(t *testing.T) {
		// Bare model name; identified by the maker field (the makeName parameter must be used).
		assert.Equal(t, 190, CameraFisheyeFov("Insta360", "X4"))
	})
	t.Run("MakerOnlyArashiVision", func(t *testing.T) {
		assert.Equal(t, 190, CameraFisheyeFov("Arashi Vision", ""))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, 0, CameraFisheyeFov("Canon", "Canon EOS 6D"))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Equal(t, 0, CameraFisheyeFov("", ""))
	})
	t.Run("IPhoneXNotFisheye", func(t *testing.T) {
		// "one x" is a substring of "iphone x", so bare model names must not match on their own.
		assert.Equal(t, 0, CameraFisheyeFov("Apple", "iPhone X"))
	})
	t.Run("BareModelWithoutVendor", func(t *testing.T) {
		assert.Equal(t, 0, CameraFisheyeFov("", "ONE RS"))
	})
	t.Run("MeasuredValue", func(t *testing.T) {
		assert.Equal(t, MeasuredFisheyeFov, CameraFisheyeFov("Insta360", "Insta360 OneRS"))
	})
}

// TestCameraFisheyeRoll verifies that only the validated OneRS profile receives a correction.
func TestCameraFisheyeRoll(t *testing.T) {
	ValidateFixtures(t)
	t.Run("OneRS", func(t *testing.T) {
		assert.Equal(t, 180, CameraFisheyeRoll("Insta360", "Insta360 OneRS"))
	})
	t.Run("OneRSLegacyMaker", func(t *testing.T) {
		assert.Equal(t, 180, CameraFisheyeRoll("Arashi Vision", "Insta360 OneRS"))
	})
	t.Run("OneRSVideoTrailer", func(t *testing.T) {
		assert.Equal(t, 180, CameraFisheyeRoll("", "Insta360 OneRS"))
	})
	t.Run("OtherInsta360", func(t *testing.T) {
		assert.Equal(t, 0, CameraFisheyeRoll("Insta360", "Insta360 X4"))
	})
	t.Run("ConflictingMaker", func(t *testing.T) {
		assert.Equal(t, 0, CameraFisheyeRoll("Canon", "Insta360 OneRS"))
	})
	t.Run("Unknown", func(t *testing.T) {
		assert.Equal(t, 0, CameraFisheyeRoll("", ""))
	})
}
