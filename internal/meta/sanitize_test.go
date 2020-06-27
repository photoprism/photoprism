package meta

import "testing"

func TestSanitizeTitle(t *testing.T) {
	t.Run("IMG_0599", func(t *testing.T) {
		result := SanitizeTitle("IMG_0599")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})

	t.Run("IMG_0599.JPG", func(t *testing.T) {
		result := SanitizeTitle("IMG_0599.JPG")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})

	t.Run("IMG_0599 ABC", func(t *testing.T) {
		result := SanitizeTitle("IMG_0599 ABC")

		if result != "IMG_0599 ABC" {
			t.Fatal("result should be IMG_0599 ABC")
		}
	})

	t.Run("DSC10599", func(t *testing.T) {
		result := SanitizeTitle("DSC10599")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})
}

func TestSanitizeDescription(t *testing.T) {
	t.Run("IMG_0599", func(t *testing.T) {
		result := SanitizeDescription("IMG_0599")

		if result == "" {
			t.Fatal("result should not be empty")
		}
	})

	t.Run("OLYMPUS DIGITAL CAMERA", func(t *testing.T) {
		result := SanitizeDescription("OLYMPUS DIGITAL CAMERA")

		if result != "" {
			t.Fatal("result should be empty")
		}
	})
}
