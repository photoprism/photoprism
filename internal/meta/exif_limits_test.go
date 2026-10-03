package meta

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestExifFileSize(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		want, err := os.Stat("testdata/photoshop.jpg")
		assert.NoError(t, err)
		size, err := ExifFileSize("testdata/photoshop.jpg")
		assert.NoError(t, err)
		assert.Equal(t, want.Size(), size)
	})
	t.Run("TooLarge", func(t *testing.T) {
		max := ExifMaxFileBytes
		ExifMaxFileBytes = 16
		defer func() { ExifMaxFileBytes = max }()
		size, err := ExifFileSize("testdata/photoshop.jpg")
		assert.True(t, errors.Is(err, ErrExifFileTooLarge))
		assert.Greater(t, size, int64(16))
	})
	t.Run("NotFound", func(t *testing.T) {
		_, err := ExifFileSize("testdata/not-found.jpg")
		assert.Error(t, err)
		assert.False(t, errors.Is(err, ErrExifFileTooLarge))
	})
	t.Run("Directory", func(t *testing.T) {
		_, err := ExifFileSize("testdata")
		assert.Error(t, err)
		assert.False(t, errors.Is(err, ErrExifFileTooLarge))
	})
}

func TestRawExif_FileSizeLimit(t *testing.T) {
	t.Run("WithinLimit", func(t *testing.T) {
		raw, err := RawExif("testdata/photoshop.jpg", fs.ImageJpeg, false)
		assert.NoError(t, err)
		assert.NotEmpty(t, raw)
	})
	t.Run("AboveLimit", func(t *testing.T) {
		max := ExifMaxFileBytes
		ExifMaxFileBytes = 16
		defer func() { ExifMaxFileBytes = max }()
		raw, err := RawExif("testdata/photoshop.jpg", fs.ImageJpeg, false)
		assert.True(t, errors.Is(err, ErrExifFileTooLarge))
		assert.Empty(t, raw)
	})
	t.Run("AboveLimitBruteForce", func(t *testing.T) {
		max := ExifMaxFileBytes
		ExifMaxFileBytes = 16
		defer func() { ExifMaxFileBytes = max }()
		// The brute-force search reads the file as well, so the limit must apply to it too.
		raw, err := RawExif("testdata/photoshop.jpg", fs.ImageRaw, true)
		assert.True(t, errors.Is(err, ErrExifFileTooLarge))
		assert.Empty(t, raw)
	})
	t.Run("TiffAboveLimit", func(t *testing.T) {
		max := ExifMaxFileBytes
		ExifMaxFileBytes = 1024
		defer func() { ExifMaxFileBytes = max }()
		// A TIFF block is the whole file, so the file size is the only bound available for it.
		raw, err := RawExif("testdata/exif-example.tiff", fs.ImageTiff, false)
		assert.True(t, errors.Is(err, ErrExifFileTooLarge))
		assert.Empty(t, raw)
	})
}

func TestData_Exif_TagLimit(t *testing.T) {
	t.Run("WithinLimit", func(t *testing.T) {
		data := Data{}
		assert.NoError(t, data.Exif("testdata/photoshop.jpg", fs.ImageJpeg, false))
		assert.Greater(t, len(data.exif), 8)
	})
	t.Run("AboveLimit", func(t *testing.T) {
		max := ExifMaxTags
		ExifMaxTags = 4
		defer func() { ExifMaxTags = max }()
		data := Data{}
		_ = data.Exif("testdata/photoshop.jpg", fs.ImageJpeg, false)
		assert.LessOrEqual(t, len(data.exif), 4)
	})
	t.Run("FileSizeAboveLimit", func(t *testing.T) {
		max := ExifMaxFileBytes
		ExifMaxFileBytes = 16
		defer func() { ExifMaxFileBytes = max }()
		data := Data{}
		assert.True(t, errors.Is(data.Exif("testdata/photoshop.jpg", fs.ImageJpeg, true), ErrExifFileTooLarge))
		assert.Empty(t, data.exif)
	})
}
