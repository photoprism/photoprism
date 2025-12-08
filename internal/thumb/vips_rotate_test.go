package thumb

import (
	"os"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestVipsRotate(t *testing.T) {
	if err := os.MkdirAll("testdata/vips/rotate", fs.ModeDir); err != nil {
		t.Fatal(err)
	}
	t.Run("OrientationNormal", func(t *testing.T) {
		src := "testdata/example.jpg"
		dst := "testdata/vips/rotate/0.jpg"

		assert.FileExists(t, src)

		// Load image from file.
		img, err := vips.NewImageFromFile(src)

		if err != nil {
			t.Fatal(err)
		}

		if err = VipsRotate(img, OrientationNormal); err != nil {
			t.Fatal(err)
		}

		params := vips.NewJpegExportParams()
		imageBytes, _, exportErr := img.ExportJpeg(params)

		if exportErr != nil {
			t.Fatal(exportErr)
		}

		// Write thumbnail to file.
		if err = os.WriteFile(dst, imageBytes, fs.ModeFile); err != nil {
			t.Fatal(exportErr)
		}

		assert.FileExists(t, dst)
	})
	t.Run("OrientationRotate90", func(t *testing.T) {
		src := "testdata/example.jpg"
		dst := "testdata/vips/rotate/90.jpg"

		assert.FileExists(t, src)

		// Load image from file.
		img, err := vips.NewImageFromFile(src)

		if err != nil {
			t.Fatal(err)
		}

		if err = VipsRotate(img, OrientationRotate90); err != nil {
			t.Fatal(err)
		}

		params := vips.NewJpegExportParams()
		imageBytes, _, exportErr := img.ExportJpeg(params)

		if exportErr != nil {
			t.Fatal(exportErr)
		}

		// Write thumbnail to file.
		if err = os.WriteFile(dst, imageBytes, fs.ModeFile); err != nil {
			t.Fatal(exportErr)
		}

		assert.FileExists(t, dst)
	})
	t.Run("OrientationRotate180", func(t *testing.T) {
		src := "testdata/example.jpg"
		dst := "testdata/vips/rotate/180.jpg"

		assert.FileExists(t, src)

		// Load image from file.
		img, err := vips.NewImageFromFile(src)

		if err != nil {
			t.Fatal(err)
		}

		if err = VipsRotate(img, OrientationRotate180); err != nil {
			t.Fatal(err)
		}

		params := vips.NewJpegExportParams()
		imageBytes, _, exportErr := img.ExportJpeg(params)

		if exportErr != nil {
			t.Fatal(exportErr)
		}

		// Write thumbnail to file.
		if err = os.WriteFile(dst, imageBytes, fs.ModeFile); err != nil {
			t.Fatal(exportErr)
		}

		assert.FileExists(t, dst)
	})
	t.Run("OrientationRotate270", func(t *testing.T) {
		src := "testdata/example.jpg"
		dst := "testdata/vips/rotate/270.jpg"

		assert.FileExists(t, src)

		// Load image from file.
		img, err := vips.NewImageFromFile(src)

		if err != nil {
			t.Fatal(err)
		}

		if err = VipsRotate(img, OrientationRotate270); err != nil {
			t.Fatal(err)
		}

		params := vips.NewJpegExportParams()
		imageBytes, _, exportErr := img.ExportJpeg(params)

		if exportErr != nil {
			t.Fatal(exportErr)
		}

		// Write thumbnail to file.
		if err = os.WriteFile(dst, imageBytes, fs.ModeFile); err != nil {
			t.Fatal(exportErr)
		}

		assert.FileExists(t, dst)
	})
}

func TestVipsRotate_VariousAndInvalid(t *testing.T) {
	// Ensure vips is initialized for image operations
	VipsInit()

	img, err := vips.LoadImageFromFile("testdata/example.jpg", VipsImportParams())
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	defer img.Close()

	// Valid operations should not return errors
	if c, err := img.Copy(); err == nil {
		defer c.Close()
		assert.NoError(t, VipsRotate(c, OrientationFlipH))
	} else {
		t.Fatal(err)
	}
	if c, err := img.Copy(); err == nil {
		defer c.Close()
		assert.NoError(t, VipsRotate(c, OrientationFlipV))
	} else {
		t.Fatal(err)
	}
	if c, err := img.Copy(); err == nil {
		defer c.Close()
		assert.NoError(t, VipsRotate(c, OrientationRotate90))
	} else {
		t.Fatal(err)
	}
	if c, err := img.Copy(); err == nil {
		defer c.Close()
		assert.NoError(t, VipsRotate(c, OrientationRotate180))
	} else {
		t.Fatal(err)
	}
	if c, err := img.Copy(); err == nil {
		defer c.Close()
		assert.NoError(t, VipsRotate(c, OrientationRotate270))
	} else {
		t.Fatal(err)
	}
	if c, err := img.Copy(); err == nil {
		defer c.Close()
		assert.NoError(t, VipsRotate(c, OrientationTranspose))
	} else {
		t.Fatal(err)
	}
	if c, err := img.Copy(); err == nil {
		defer c.Close()
		assert.NoError(t, VipsRotate(c, OrientationTransverse))
	} else {
		t.Fatal(err)
	}

	// Invalid orientation triggers debug branch but must not error
	if c, err := img.Copy(); err == nil {
		defer c.Close()
		assert.NoError(t, VipsRotate(c, 999))
	} else {
		t.Fatal(err)
	}
}
