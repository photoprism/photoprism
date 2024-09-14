package photoprism

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestMediaFile_RelatedFiles(t *testing.T) {
	c := config.TestConfig()

	t.Run("example.tif", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/example.tif")

		if err != nil {
			t.Fatal(err)
		}

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, related.Files, 6)
		assert.True(t, related.HasPreview())

		for _, result := range related.Files {
			t.Logf("FileName: %s", result.FileName())

			filename := result.FileName()

			if len(filename) < 2 {
				t.Fatalf("filename not be longer: %s", filename)
			}

			extension := result.Extension()

			if len(extension) < 2 {
				t.Fatalf("extension should be longer: %s", extension)
			}

			relativePath := result.RelPath(c.ExamplesPath())

			if len(relativePath) > 0 {
				t.Fatalf("relative path should be empty: %s", relativePath)
			}
		}
	})

	t.Run("canon_eos_6d.dng", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/canon_eos_6d.dng")

		if err != nil {
			t.Fatal(err)
		}

		expectedBaseFilename := c.ExamplesPath() + "/canon_eos_6d"

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, related.Files, 3)
		assert.False(t, related.HasPreview())

		for _, result := range related.Files {
			t.Logf("FileName: %s", result.FileName())

			filename := result.FileName()

			extension := result.Extension()

			baseFilename := filename[0 : len(filename)-len(extension)]

			assert.Equal(t, expectedBaseFilename, baseFilename)
		}
	})

	t.Run("iphone_7.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/iphone_7.heic")

		if err != nil {
			t.Fatal(err)
		}

		expectedBaseFilename := c.ExamplesPath() + "/iphone_7"

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(related.Files), 3)

		for _, result := range related.Files {
			t.Logf("FileName: %s", result.FileName())

			filename := result.FileName()
			extension := result.Extension()
			baseFilename := filename[0 : len(filename)-len(extension)]

			if result.IsJpeg() {
				assert.Contains(t, expectedBaseFilename, "examples/iphone_7")
			} else {
				assert.Equal(t, expectedBaseFilename, baseFilename)
			}
		}
	})

	t.Run("iphone_15_pro.heic", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/iphone_15_pro.heic")

		if err != nil {
			t.Fatal(err)
		}

		expectedBaseFilename := c.ExamplesPath() + "/iphone_15_pro"

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		assert.GreaterOrEqual(t, len(related.Files), 2)

		for _, result := range related.Files {
			t.Logf("FileName: %s", result.FileName())

			filename := result.FileName()
			extension := result.Extension()
			baseFilename := filename[0 : len(filename)-len(extension)]

			if result.IsJpeg() {
				assert.Contains(t, expectedBaseFilename, "examples/iphone_15_pro")
			} else {
				assert.Equal(t, expectedBaseFilename, baseFilename)
			}
		}
	})

	t.Run("2015-02-04.jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile("testdata/2015-02-04.jpg")

		if err != nil {
			t.Fatal(err)
		}

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		if related.Main == nil {
			t.Fatal("main media file must not be nil")
		}

		if len(related.Files) != 4 {
			t.Fatalf("length is %d, should be 4", len(related.Files))
		}

		t.Logf("FILE: %s, %s", related.Main.FileType(), related.Main.MimeType())

		assert.Equal(t, "2015-02-04.jpg", related.Main.BaseName())

		assert.Equal(t, "2015-02-04.jpg", related.Files[0].BaseName())
		assert.Equal(t, "2015-02-04(1).jpg", related.Files[1].BaseName())
		assert.Equal(t, "2015-02-04.jpg.json", related.Files[2].BaseName())
		assert.Equal(t, "2015-02-04.jpg(1).json", related.Files[3].BaseName())
	})

	t.Run("2015-02-04(1).jpg", func(t *testing.T) {
		mediaFile, err := NewMediaFile("testdata/2015-02-04(1).jpg")

		if err != nil {
			t.Fatal(err)
		}

		related, err := mediaFile.RelatedFiles(false)

		if err != nil {
			t.Fatal(err)
		}

		if related.Main == nil {
			t.Fatal("main media file must not be nil")
		}

		if len(related.Files) != 1 {
			t.Fatalf("length is %d, should be 1", len(related.Files))
		}

		assert.Equal(t, "2015-02-04(1).jpg", related.Main.BaseName())

		assert.Equal(t, "2015-02-04(1).jpg", related.Files[0].BaseName())
	})

	t.Run("2015-02-04(1).jpg stacked", func(t *testing.T) {
		mediaFile, err := NewMediaFile("testdata/2015-02-04(1).jpg")

		if err != nil {
			t.Fatal(err)
		}

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		if related.Main == nil {
			t.Fatal("main media file must not be nil")
		}

		if len(related.Files) != 4 {
			t.Fatalf("length is %d, should be 4", len(related.Files))
		}

		assert.Equal(t, "2015-02-04.jpg", related.Main.BaseName())

		assert.Equal(t, "2015-02-04.jpg", related.Files[0].BaseName())
		assert.Equal(t, "2015-02-04(1).jpg", related.Files[1].BaseName())
		assert.Equal(t, "2015-02-04.jpg.json", related.Files[2].BaseName())
		assert.Equal(t, "2015-02-04.jpg(1).json", related.Files[3].BaseName())
	})

	t.Run("Ordering", func(t *testing.T) {
		mediaFile, err := NewMediaFile(c.ExamplesPath() + "/IMG_4120.JPG")

		if err != nil {
			t.Fatal(err)
		}

		related, err := mediaFile.RelatedFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, related.Files, 5)

		assert.Equal(t, c.ExamplesPath()+"/IMG_4120.AAE", related.Files[0].FileName())
		assert.Equal(t, c.ExamplesPath()+"/IMG_4120.JPG", related.Files[1].FileName())

		for _, result := range related.Files {
			filename := result.FileName()
			t.Logf("FileName: %s", filename)
		}
	})
}

func TestMediaFile_RelatedSidecarFiles(t *testing.T) {
	t.Run("FindEdited", func(t *testing.T) {
		file, err := NewMediaFile("testdata/related/IMG_1234 (2).JPEG")

		if err != nil {
			t.Fatal(err)
		}

		files, err := file.RelatedSidecarFiles(false)

		if err != nil {
			t.Fatal(err)
		}

		expected := []string{"testdata/related/IMG_E1234 (2).JPEG"}

		assert.Len(t, files, len(expected))
		assert.Equal(t, expected, files)
	})
	t.Run("StripSequence", func(t *testing.T) {
		file, err := NewMediaFile("testdata/related/IMG_1234 (2).JPEG")

		if err != nil {
			t.Fatal(err)
		}

		files, err := file.RelatedSidecarFiles(true)

		if err != nil {
			t.Fatal(err)
		}

		expected := []string{"testdata/related/IMG_E1234 (2).JPEG", "testdata/related/IMG_1234_HEVC.JPEG"}

		assert.Len(t, files, len(expected))
		assert.Equal(t, expected, files)
	})
}
