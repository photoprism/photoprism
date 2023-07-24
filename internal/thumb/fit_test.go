package thumb

import (
	"testing"

	"github.com/disintegration/imaging"

	"github.com/stretchr/testify/assert"
)

func TestFit(t *testing.T) {
	assert.Equal(t, Sizes[Fit720], Fit(54, 453))
	assert.Equal(t, Sizes[Fit1280], Fit(1000, 1000))
	assert.Equal(t, Sizes[Fit1280], Fit(1250, 1000))
	assert.Equal(t, Sizes[Fit2048], Fit(1300, 1300))
	assert.Equal(t, Sizes[Fit2048], Fit(1600, 1600))
	assert.Equal(t, Sizes[Fit4096], Fit(1000, 3000))
	assert.Equal(t, Sizes[Fit3840], Fit(2300, 2000))
	assert.Equal(t, Sizes[Fit7680], Fit(5000, 5000))
}

func TestFitBounds(t *testing.T) {
	t.Run("example.jpg", func(t *testing.T) {
		src := "testdata/example.jpg"

		assert.FileExists(t, src)

		img, err := imaging.Open(src, imaging.AutoOrientation(true))

		if err != nil {
			t.Fatal(err)
		}

		bounds := img.Bounds()

		assert.Equal(t, 750, bounds.Max.X)
		assert.Equal(t, 500, bounds.Max.Y)

		size := FitBounds(img.Bounds())

		assert.Equal(t, "fit_1280", size.Name.String())
	})
	t.Run("example.bmp", func(t *testing.T) {
		src := "testdata/example.bmp"

		assert.FileExists(t, src)

		img, err := imaging.Open(src, imaging.AutoOrientation(true))

		if err != nil {
			t.Fatal(err)
		}

		bounds := img.Bounds()

		assert.Equal(t, 100, bounds.Max.X)
		assert.Equal(t, 67, bounds.Max.Y)

		size := FitBounds(img.Bounds())

		assert.Equal(t, "fit_720", size.Name.String())
	})
	t.Run("animated-earth.jpg", func(t *testing.T) {
		src := "testdata/animated-earth.jpg"

		assert.FileExists(t, src)

		img, err := imaging.Open(src, imaging.AutoOrientation(true))

		if err != nil {
			t.Fatal(err)
		}

		bounds := img.Bounds()

		assert.Equal(t, 300, bounds.Max.X)
		assert.Equal(t, 300, bounds.Max.Y)

		size := FitBounds(img.Bounds())

		assert.Equal(t, "fit_720", size.Name.String())
	})
}
