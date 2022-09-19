package thumb

import (
	"testing"
)

func TestOpenJpeg(t *testing.T) {
	t.Run("testdata/example.jpg", func(t *testing.T) {
		img, err := OpenJpeg("testdata/example.jpg", 0)

		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("testdata/broken.jpg", func(t *testing.T) {
		img, err := OpenJpeg("testdata/broken.jpg", 0)

		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("testdata/fixed.jpg", func(t *testing.T) {
		img, err := OpenJpeg("testdata/fixed.jpg", 0)

		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
}
