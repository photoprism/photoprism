package thumb

import (
	"testing"
)

func TestOpen(t *testing.T) {
	t.Run("JPEG", func(t *testing.T) {
		img, err := Open("testdata/example.jpg", 0)
		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("BMP", func(t *testing.T) {
		img, err := Open("testdata/example.bmp", 0)
		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("GIF", func(t *testing.T) {
		img, err := Open("testdata/example.gif", 0)
		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("PNG", func(t *testing.T) {
		img, err := Open("testdata/example.png", 0)
		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
	t.Run("TIFF", func(t *testing.T) {
		img, err := Open("testdata/example.tif", 0)
		if err != nil {
			t.Fatal(err)
		}

		if img == nil {
			t.Error("img must not be nil")
		}
	})
}
