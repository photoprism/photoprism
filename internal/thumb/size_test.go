package thumb

import (
	"image"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/stretchr/testify/assert"
)

func TestSkip(t *testing.T) {
	t.Run("Tile500", func(t *testing.T) {
		bounds := image.Rectangle{Min: image.Point{}, Max: image.Point{X: 1024, Y: 1024}}
		assert.False(t, Skip(SizeTile500, bounds))
	})
	t.Run("Fit1600", func(t *testing.T) {
		bounds := image.Rectangle{Min: image.Point{}, Max: image.Point{X: 2048, Y: 2048}}
		assert.True(t, Skip(SizeFit1600, bounds))
	})
	t.Run("Fit1920", func(t *testing.T) {
		large := image.Rectangle{Min: image.Point{}, Max: image.Point{X: 2048, Y: 2048}}
		small := image.Rectangle{Min: image.Point{}, Max: image.Point{X: 1024, Y: 1024}}
		assert.False(t, Skip(SizeFit1920, large))
		assert.True(t, Skip(SizeFit1920, small))
	})
}

func TestSize_Skip(t *testing.T) {
	// Image Size: 750x500px
	src := "testdata/example.jpg"

	t.Run("Tile500", func(t *testing.T) {
		size := Sizes[Tile500]

		assert.FileExists(t, src)

		img, err := imaging.Open(src, imaging.AutoOrientation(true))

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, size.Skip(img))
	})
	t.Run("Fit720", func(t *testing.T) {
		size := Sizes[Fit720]

		assert.FileExists(t, src)

		img, err := imaging.Open(src, imaging.AutoOrientation(true))

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, size.Skip(img))
	})
	t.Run("Fit1280", func(t *testing.T) {
		size := Sizes[Fit1280]

		assert.FileExists(t, src)

		img, err := imaging.Open(src, imaging.AutoOrientation(true))

		if err != nil {
			t.Fatal(err)
		}

		assert.False(t, size.Skip(img))
	})
	t.Run("Fit2048", func(t *testing.T) {
		size := Sizes[Fit2048]

		assert.FileExists(t, src)

		img, err := imaging.Open(src, imaging.AutoOrientation(true))

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, size.Skip(img))
	})
	t.Run("Fit4096", func(t *testing.T) {
		size := Sizes[Fit4096]

		assert.FileExists(t, src)

		img, err := imaging.Open(src, imaging.AutoOrientation(true))

		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, size.Skip(img))
	})
}

func TestSize_FileName(t *testing.T) {
	size := Sizes[Fit2048]

	r, err := size.FileName("193456789098765432", "testdata/cache")

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, r, "testdata/cache/1/9/3/193456789098765432_2048x2048_fit.jpg")
}
