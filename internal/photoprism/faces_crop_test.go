package photoprism

import (
	"image"
	"image/color"
	"image/draw"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestEmbedCropWidth pins that the width a rendition is rendered for covers both paths an
// embedding can take, since which one a face takes is only known once its landmarks were fitted.
func TestEmbedCropWidth(t *testing.T) {
	t.Run("CoversEveryRegisteredModel", func(t *testing.T) {
		for name := range face.EmbeddingModels {
			width := embedCropWidth(name)

			assert.GreaterOrEqual(t, width, face.CropSize.Width, name)

			if input, _ := face.FindEmbeddingModel(name).InputSize(); input > 0 {
				assert.GreaterOrEqual(t, width, input, name)
			}
		}
	})
	t.Run("UnknownModel", func(t *testing.T) {
		assert.Equal(t, face.CropSize.Width, embedCropWidth("nonesuch"))
	})
}

func TestCropSourceWidth(t *testing.T) {
	t.Run("RoundsUp", func(t *testing.T) {
		// 160/0.3 is 533.3, and a rendition one pixel short still upscales the crop.
		assert.Equal(t, 534, cropSourceWidth(0.3, 160))
	})
	t.Run("Exact", func(t *testing.T) {
		assert.Equal(t, 1600, cropSourceWidth(0.1, 160))
	})
	t.Run("WithoutGeometry", func(t *testing.T) {
		assert.Zero(t, cropSourceWidth(0, 160))
		assert.Zero(t, cropSourceWidth(-0.5, 160))
		assert.Zero(t, cropSourceWidth(0.1, 0))
	})
}

func TestMarkersCropSourceWidth(t *testing.T) {
	t.Run("SmallestFaceDecides", func(t *testing.T) {
		markers := entity.Markers{{W: 0.5}, {W: 0.05}, {W: 0.25}}

		assert.Equal(t, 3200, markersCropSourceWidth(markers, 160))
	})
	t.Run("WithoutGeometry", func(t *testing.T) {
		assert.Zero(t, markersCropSourceWidth(entity.Markers{{W: 0}}, 160))
		assert.Zero(t, markersCropSourceWidth(nil, 160))
	})
}

func TestFacesCropSourceWidth(t *testing.T) {
	t.Run("SmallestFaceDecides", func(t *testing.T) {
		faces := face.Faces{testCropFace(0.5), testCropFace(0.05)}

		assert.Equal(t, 3200, facesCropSourceWidth(faces, 160))
	})
	t.Run("WithoutFaces", func(t *testing.T) {
		assert.Zero(t, facesCropSourceWidth(nil, 160))
	})
}

// TestCropThumbSize covers the rendition a crop of a given source width is taken from. The name of
// a size bounds it from above but not below, so the pixels it delivers are what decide.
func TestCropThumbSize(t *testing.T) {
	limit := thumb.Sizes[thumb.Fit4096]

	t.Run("MeasuresDeliveredPixelsNotTheBox", func(t *testing.T) {
		// Fit1920 is a 1920x1200 box, so this 4:3 picture is 1600 px wide in it: a requirement
		// between the two is met by the next rung up, not by the one whose name allows it.
		assert.Equal(t, thumb.Fit1920, cropThumbSize(3648, 2736, 1600, limit))
		assert.Equal(t, thumb.Fit2560, cropThumbSize(3648, 2736, 1601, limit))
		assert.Equal(t, thumb.Fit2560, cropThumbSize(3648, 2736, 1920, limit))
	})
	t.Run("BoundedByTheLimit", func(t *testing.T) {
		// A lower ceiling is what the caller asked to render up to, and nothing above it is
		// named however much the crop would gain from it.
		assert.Equal(t, thumb.Fit1920, cropThumbSize(3648, 2736, 99000, thumb.Sizes[thumb.Fit1920]))
		assert.Equal(t, thumb.Fit720, cropThumbSize(3648, 2736, 99000, thumb.Sizes[thumb.Fit720]))
	})
	t.Run("SmallOriginal", func(t *testing.T) {
		// The widest rendition this picture is given, which is also what indexing would write.
		assert.Equal(t, thumb.Fit720, cropThumbSize(600, 400, 99000, limit))
	})
}

// TestFaceCropSourceLimit covers the ceiling on demand rendering obeys, which is the largest
// rendition a crop can be taken from within the configured one.
func TestFaceCropSourceLimit(t *testing.T) {
	cached, onDemand, faceSize := thumb.SizeCached, thumb.SizeOnDemand, thumb.SizeFace
	t.Cleanup(func() { thumb.SizeCached, thumb.SizeOnDemand, thumb.SizeFace = cached, onDemand, faceSize })

	thumb.SizeCached, thumb.SizeOnDemand, thumb.SizeFace = 720, 720, 15360

	t.Run("LargestWithinTheCeiling", func(t *testing.T) {
		assert.Equal(t, thumb.Fit4096, faceCropSourceLimit(4096).Name)
		assert.Equal(t, thumb.Fit2560, faceCropSourceLimit(4095).Name)
		assert.Equal(t, thumb.Fit720, faceCropSourceLimit(720).Name)
	})
	t.Run("Disabled", func(t *testing.T) {
		// Nothing a lower ceiling allows is a rendition a crop is taken from.
		assert.Zero(t, faceCropSourceLimit(0).Width)
		assert.Zero(t, faceCropSourceLimit(719).Width)
	})
	t.Run("NeverAboveWhatThisProcessRenders", func(t *testing.T) {
		thumb.SizeFace = 2560
		defer func() { thumb.SizeFace = 15360 }()

		assert.Equal(t, thumb.Fit2560, faceCropSourceLimit(15360).Name)
	})
}

// TestCacheFaceCropSource covers the rendition indexing renders for a face crop when the cache
// holds none wide enough. Without it an embedding is drawn from upscaled pixels and nothing in the
// vector says so afterwards, so the whole library has to be migrated to find out.
func TestCacheFaceCropSource(t *testing.T) {
	c := newMigrateTestConfig(t, "facecropsource")

	cached, onDemand, faceSize := thumb.SizeCached, thumb.SizeOnDemand, thumb.SizeFace
	t.Cleanup(func() { thumb.SizeCached, thumb.SizeOnDemand, thumb.SizeFace = cached, onDemand, faceSize })

	// What indexing pre-generates and what a request may render stay at the detection size, so
	// only the face ceiling can produce the rendition below.
	thumb.SizeCached, thumb.SizeOnDemand, thumb.SizeFace = 720, 720, 4096
	c.Options().ThumbSizeFace = 4096

	m := newCropSourceTestFile(t, c, "source.jpg", 2000, 1500)

	// 0.09 of a 2000 px picture asks for more than the 1600 px Fit1920 delivers for it, so a
	// check that reads the box width rather than the pixels renders a rendition that upscales.
	faces := face.Faces{testCropFace(0.09)}
	required := cropSourceWidth(0.09, embedCropWidth(face.EmbeddingModelName()))
	thumbPath := c.ThumbCachePath()

	t.Run("Renders", func(t *testing.T) {
		rendered, err := cacheFaceCropSource(c, m, faces)
		require.NoError(t, err)
		require.True(t, rendered)

		// Against the pixels on disk rather than the size that was named, since the name is the
		// box the picture was fitted into and not what it delivers.
		assert.GreaterOrEqual(t, cachedCropWidth(m.Hash(), m.Width(), m.Height(), thumbPath), required)
		assert.False(t, crop.CachedSizeExists(thumb.Sizes[thumb.Fit1920], m.Hash(), thumbPath),
			"the 1920 box supplies 1600 px of this picture, which is short of the crop")

		// The point of writing it: the selection a crop goes through has to choose the rendition.
		area := crop.Area{Name: "face", X: 0.1, Y: 0.1, W: 0.09, H: 0.09}
		size := crop.Size{Width: face.CropSize.Width, Height: face.CropSize.Height, Options: crop.DefaultOptions}

		selected, err := crop.ThumbFileName(m.Hash(), area, size, thumbPath)
		require.NoError(t, err)
		assert.Contains(t, selected, "_2560x1600_fit.jpg")
	})
	t.Run("RendersNothingTwice", func(t *testing.T) {
		rendered, err := cacheFaceCropSource(c, m, faces)
		require.NoError(t, err)
		assert.False(t, rendered, "the cache already covers these faces")
	})
	t.Run("Disabled", func(t *testing.T) {
		// Zero restores what indexing did before: the crop is taken from what the cache holds.
		other := newCropSourceTestFile(t, c, "disabled.jpg", 2000, 1500)

		c.Options().ThumbSizeFace = 0
		defer func() { c.Options().ThumbSizeFace = 4096 }()

		rendered, err := cacheFaceCropSource(c, other, faces)
		require.NoError(t, err)
		assert.False(t, rendered)
		assert.Equal(t, 720, cachedCropWidth(other.Hash(), other.Width(), other.Height(), thumbPath))
	})
	t.Run("LargeFace", func(t *testing.T) {
		// A face the cache already covers renders nothing, so this costs an indexed library
		// with ordinary portraits no decode at all.
		other := newCropSourceTestFile(t, c, "large.jpg", 2000, 1500)

		rendered, err := cacheFaceCropSource(c, other, face.Faces{testCropFace(0.35)})
		require.NoError(t, err)
		assert.False(t, rendered)
	})
	t.Run("WithoutFaces", func(t *testing.T) {
		rendered, err := cacheFaceCropSource(c, m, nil)
		require.NoError(t, err)
		assert.False(t, rendered)
	})
	t.Run("NilInput", func(t *testing.T) {
		rendered, err := cacheFaceCropSource(nil, m, faces)
		require.NoError(t, err)
		assert.False(t, rendered)

		rendered, err = cacheFaceCropSource(c, nil, faces)
		require.NoError(t, err)
		assert.False(t, rendered)
	})
}

// testCropFace returns a face covering the given fraction of the image it was detected in, which
// is the frame a marker's relative geometry is measured against.
func testCropFace(relWidth float32) face.Face {
	const cols = 10000

	return face.Face{
		Rows: cols,
		Cols: cols,
		Area: face.Area{Name: "face", Row: cols / 2, Col: cols / 2, Scale: int(relWidth * cols)},
	}
}

// newCropSourceTestFile writes a picture of the given dimensions and its detection thumbnail, so a
// test states what the cache holds rather than what the configured limit would generate.
func newCropSourceTestFile(t *testing.T, c *config.Config, name string, width, height int) *MediaFile {
	t.Helper()

	fileName := filepath.Join(c.OriginalsPath(), name)
	require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))

	// Shaded by name, or two pictures of the same dimensions would hash alike and share a cache.
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	shade := color.NRGBA{R: uint8(len(name)%16*7 + 1), G: 32, B: 32, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: shade}, image.Point{}, draw.Src)

	require.NoError(t, thumb.Save(img, fileName))

	m, err := NewMediaFile(fileName)
	require.NoError(t, err)
	require.Equal(t, width, m.Width())

	_, err = m.Thumbnail(c.ThumbCachePath(), thumb.Fit720)
	require.NoError(t, err)

	return m
}
