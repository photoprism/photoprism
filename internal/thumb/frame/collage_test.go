package frame

import (
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// collagePalette holds opaque colors that differ from the background and from the polaroid frame,
// so a test can tell which of several images a layout composited and where it put each one.
var collagePalette = []color.NRGBA{
	{R: 200, G: 30, B: 30, A: 255},
	{R: 30, G: 200, B: 30, A: 255},
	{R: 30, G: 30, B: 200, A: 255},
}

// solidImages returns n distinct opaque images taken from collagePalette.
func solidImages(n int) ([]image.Image, []color.NRGBA) {
	images := make([]image.Image, 0, n)

	for i := range n {
		images = append(images, newCanvas(500, 500, collagePalette[i]))
	}

	return images, collagePalette[:n]
}

// leftmostColumn returns the smallest x at which an image holds a color, or -1 when it holds it
// nowhere.
func leftmostColumn(img image.Image, c color.NRGBA) int {
	b := img.Bounds()

	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			if color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA) == c {
				return x
			}
		}
	}

	return -1
}

// requireComposited fails the test unless every color is visible in the rendered collage.
func requireComposited(t *testing.T, img image.Image, colors []color.NRGBA) {
	t.Helper()

	for i, c := range colors {
		assert.NotEqual(t, -1, leftmostColumn(img, c), "image %d must be visible", i)
	}
}

// embeddedPixels counts the pixels of a collage that differ from its background, so a test can tell
// a rendered collage from an empty canvas.
func embeddedPixels(img image.Image) int {
	b := img.Bounds()
	n := 0

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA) != CollageBackground {
				n++
			}
		}
	}

	return n
}

func TestCollage(t *testing.T) {
	t.Run("Polaroid", func(t *testing.T) {
		var images []image.Image

		img, _, err := fs.DecodeImageFile("testdata/500x500.jpg")
		assert.NoError(t, err)

		for i := 0; i <= 5; i++ {
			images = append(images, img)
		}

		saveName := "testdata/test-polaroid-collage.jpg"
		preview, err := Collage(Polaroid, images)

		assert.NoError(t, err)

		err = thumb.Save(preview, saveName)

		assert.NoError(t, err)
		mimeType, _ := fs.DetectMimeType(saveName)
		assert.Equal(t, header.ContentTypeJpeg, mimeType)

		_ = os.Remove(saveName)
	})
	t.Run("Two", func(t *testing.T) {
		var images []image.Image

		img, _, err := fs.DecodeImageFile("testdata/500x500.jpg")
		assert.NoError(t, err)

		for i := 0; i <= 1; i++ {
			images = append(images, img)
		}

		saveName := "testdata/test-polaroid-collage-two.jpg"
		preview, err := Collage(Polaroid, images)

		assert.NoError(t, err)

		err = thumb.Save(preview, saveName)

		assert.NoError(t, err)
		mimeType, _ := fs.DetectMimeType(saveName)
		assert.Equal(t, header.ContentTypeJpeg, mimeType)

		_ = os.Remove(saveName)
	})
	t.Run("One", func(t *testing.T) {
		img, _, err := fs.DecodeImageFile("testdata/500x500.jpg")
		require.NoError(t, err)

		preview, err := Collage(Polaroid, []image.Image{img})

		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.Equal(t, 1600, preview.Bounds().Dx())
		assert.Equal(t, 900, preview.Bounds().Dy())
		assert.NotZero(t, embeddedPixels(preview), "the image must be embedded")

		saveName := "testdata/test-polaroid-collage-one.jpg"
		require.NoError(t, thumb.Save(preview, saveName))
		mimeType, _ := fs.DetectMimeType(saveName)
		assert.Equal(t, header.ContentTypeJpeg, mimeType)

		_ = os.Remove(saveName)
	})
	t.Run("NoImages", func(t *testing.T) {
		var images []image.Image

		saveName := "testdata/test-no-images-collage.jpg"
		preview, err := Collage(Polaroid, images)

		assert.NoError(t, err)

		err = thumb.Save(preview, saveName)

		assert.NoError(t, err)
		mimeType, _ := fs.DetectMimeType(saveName)
		assert.Equal(t, header.ContentTypeJpeg, mimeType)

		_ = os.Remove(saveName)
	})
	t.Run("UnknownCollageType", func(t *testing.T) {
		var images []image.Image

		img, _, err := fs.DecodeImageFile("testdata/500x500.jpg")
		assert.NoError(t, err)

		for i := 0; i <= 5; i++ {
			images = append(images, img)
		}

		saveName := "testdata/test-unknown-type-collage.jpg"

		preview, err := Collage("Unknown", images)

		assert.Error(t, err)
		assert.Equal(t, "unknown collage type Unknown", err.Error())

		err = thumb.Save(preview, saveName)

		assert.NoError(t, err)

		mimeType, _ := fs.DetectMimeType(saveName)
		assert.Equal(t, header.ContentTypeJpeg, mimeType)

		_ = os.Remove(saveName)

	})
}

func TestPolaroidCollage(t *testing.T) {
	// canvas is what Collage hands the layout, so a case starts from the same background.
	canvas := func(t *testing.T) image.Image {
		t.Helper()
		empty, err := Collage(Polaroid, nil)
		require.NoError(t, err)

		return empty
	}

	t.Run("None", func(t *testing.T) {
		out, err := polaroidCollage(canvas(t), nil)

		require.NoError(t, err)
		assert.Zero(t, embeddedPixels(out))
	})
	t.Run("One", func(t *testing.T) {
		images, colors := solidImages(1)
		out, err := polaroidCollage(canvas(t), images)

		require.NoError(t, err)
		assert.Equal(t, image.Rect(0, 0, 1600, 900), out.Bounds())
		requireComposited(t, out, colors)

		// The single image is placed left of center rather than at the origin.
		assert.Greater(t, leftmostColumn(out, colors[0]), 300)
		assert.Less(t, leftmostColumn(out, colors[0]), 800)
	})
	t.Run("Two", func(t *testing.T) {
		images, colors := solidImages(2)
		out, err := polaroidCollage(canvas(t), images)

		require.NoError(t, err)
		assert.Equal(t, image.Rect(0, 0, 1600, 900), out.Bounds())
		requireComposited(t, out, colors)

		// The pair has a layout of its own, which puts the second image left of where the fan
		// layout would start it.
		assert.Less(t, leftmostColumn(out, colors[1]), 810)
	})
	t.Run("Three", func(t *testing.T) {
		images, colors := solidImages(3)
		out, err := polaroidCollage(canvas(t), images)

		require.NoError(t, err)
		assert.Equal(t, image.Rect(0, 0, 1600, 900), out.Bounds())

		// The fan layout places all but the first image at a random point, and one of those can end
		// up behind another, so only the image it draws last is asserted to be visible.
		requireComposited(t, out, colors[:1])
	})
}
