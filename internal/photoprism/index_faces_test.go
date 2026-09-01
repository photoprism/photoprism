package photoprism

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestDetectFacesCropSource is the end-to-end case THUMB_SIZE_FACE exists for: a face small enough
// in its picture that the pre-generated renditions cannot supply the crop the embedder asks for.
//
// Both runs index the same picture at the same thumbnail limit and differ only in the option, so
// what the second one records is what every index produced before it: a vector interpolated up to
// the template, and nothing in the row that says so.
func TestDetectFacesCropSource(t *testing.T) {
	c := Config()

	cached, onDemand, faceSize := thumb.SizeCached, thumb.SizeOnDemand, thumb.SizeFace
	previous := c.Options().ThumbSizeFace

	t.Cleanup(func() {
		thumb.SizeCached, thumb.SizeOnDemand, thumb.SizeFace = cached, onDemand, faceSize
		c.Options().ThumbSizeFace = previous
	})

	// Nothing above the detection thumbnail is pre-generated or may be rendered for delivery, so
	// only the face ceiling can produce a wider source.
	thumb.SizeCached, thumb.SizeOnDemand, thumb.SizeFace = 720, 720, 4096

	detect := func(t *testing.T, name string, shade uint8) face.Face {
		t.Helper()

		m := newSmallFaceTestFile(t, c, name, shade)

		faces, err := DetectFaces(m, 0)
		require.NoError(t, err)

		// The largest, not the first: the canvas is flat enough that the detector also returns a
		// smaller candidate, and its order is not a contract this test should depend on.
		largest := face.Face{}

		for i := range faces {
			if faces[i].Size() > largest.Size() && !faces[i].Embeddings.Empty() {
				largest = faces[i]
			}
		}

		if largest.Embeddings.Empty() {
			t.Skip("faces: skipping, the detector or the embedder is not available here")
		}

		return largest
	}

	t.Run("RenderedSource", func(t *testing.T) {
		c.Options().ThumbSizeFace = 4096

		f := detect(t, "crop-source-rendered.jpg", 210)

		assert.Greater(t, f.ThumbSize, f.Size(), "the crop was taken from a wider rendition than detection ran on")
		assert.Equal(t, 100, f.EmbedUpscaled, "so the source supplied the whole crop")
	})
	t.Run("CacheOnly", func(t *testing.T) {
		// The control, and what the option turns off: the same face embedded from the detection
		// thumbnail, because the selection only ever stats what is already on disk.
		c.Options().ThumbSizeFace = 0

		f := detect(t, "crop-source-cached.jpg", 90)

		assert.Equal(t, f.Size(), f.ThumbSize, "the detection thumbnail is all the cache holds")
		assert.Less(t, f.EmbedUpscaled, 100, "so the crop was interpolated up to the template")
		assert.Positive(t, f.EmbedUpscaled)
	})
}

// newSmallFaceTestFile writes a picture holding one face too small for the detection thumbnail to
// supply its crop, shaded so that two of them do not share a file hash and therefore a cache.
func newSmallFaceTestFile(t *testing.T, c *config.Config, name string, shade uint8) *MediaFile {
	t.Helper()

	src, _, err := fs.DecodeImageFile(filepath.Join("..", "ai", "face", "testdata", "1.jpg"))
	require.NoError(t, err)

	// A portrait scaled into a frame three times its width: the face then covers less of the
	// picture than the 112 px template needs from a 720 px thumbnail.
	scaled := thumb.Resample(src, 700, 700, thumb.ResampleFit, thumb.ResampleDefault)
	canvas := image.NewNRGBA(image.Rect(0, 0, 2048, 2048))

	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: color.NRGBA{R: shade, G: shade, B: shade, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(canvas, scaled.Bounds().Add(image.Pt(600, 600)), scaled, scaled.Bounds().Min, draw.Src)

	fileName := filepath.Join(c.OriginalsPath(), name)
	require.NoError(t, fs.MkdirAll(filepath.Dir(fileName)))
	require.NoError(t, thumb.Save(canvas, fileName))
	t.Cleanup(func() { _ = os.Remove(fileName) })

	m, err := NewMediaFile(fileName)
	require.NoError(t, err)

	return m
}
