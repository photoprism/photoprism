package thumb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestVipsCheckPixels(t *testing.T) {
	VipsInit()

	max := fs.MaxImagePixels
	defer func() { fs.MaxImagePixels = max }()

	t.Run("WithinBudget", func(t *testing.T) {
		fs.MaxImagePixels = max
		img, err := vips.LoadImageFromFile("testdata/example.jpg", VipsImportParams())
		assert.NoError(t, err)
		defer img.Close()
		assert.NoError(t, vipsCheckPixels(img, "example.jpg"))
	})
	t.Run("AboveBudget", func(t *testing.T) {
		fs.MaxImagePixels = 4
		img, err := vips.LoadImageFromFile("testdata/example.jpg", VipsImportParams())
		assert.NoError(t, err)
		defer img.Close()
		assert.True(t, errors.Is(vipsCheckPixels(img, "example.jpg"), fs.ErrImageTooLarge))
	})
	t.Run("Disabled", func(t *testing.T) {
		fs.MaxImagePixels = 0
		img, err := vips.LoadImageFromFile("testdata/example.jpg", VipsImportParams())
		assert.NoError(t, err)
		defer img.Close()
		assert.NoError(t, vipsCheckPixels(img, "example.jpg"))
	})
	t.Run("NotLoaded", func(t *testing.T) {
		fs.MaxImagePixels = max
		assert.Error(t, vipsCheckPixels(nil, "example.jpg"))
	})
}

func TestVips_PixelBudget(t *testing.T) {
	max := fs.MaxImagePixels
	defer func() { fs.MaxImagePixels = max }()

	t.Run("AboveBudget", func(t *testing.T) {
		fs.MaxImagePixels = 4
		name, buf, err := Vips("testdata/example.jpg", nil, "193456789012345678901234567890abcdef1234", t.TempDir(), 150, 150, ResampleFit)
		assert.True(t, errors.Is(err, fs.ErrImageTooLarge))
		assert.Empty(t, name)
		assert.Empty(t, buf)
	})
	t.Run("WithinBudget", func(t *testing.T) {
		fs.MaxImagePixels = max
		_, buf, err := Vips("testdata/example.jpg", nil, "193456789012345678901234567890abcdef1234", t.TempDir(), 150, 150, ResampleFit)
		assert.NoError(t, err)
		assert.NotEmpty(t, buf)
	})
}

func TestVerify_PixelBudget(t *testing.T) {
	max := fs.MaxImagePixels
	lib := Library
	defer func() {
		fs.MaxImagePixels = max
		Library = lib
	}()

	Library = LibVips

	t.Run("AboveBudget", func(t *testing.T) {
		fs.MaxImagePixels = 4
		assert.True(t, errors.Is(Verify("testdata/example.jpg"), fs.ErrImageTooLarge))
	})
	t.Run("WithinBudget", func(t *testing.T) {
		fs.MaxImagePixels = max
		assert.NoError(t, Verify("testdata/example.jpg"))
	})
}

func TestVipsLoadedPages(t *testing.T) {
	VipsInit()

	t.Run("SinglePage", func(t *testing.T) {
		img, err := vips.LoadImageFromFile("testdata/example.jpg", VipsImportParams())
		assert.NoError(t, err)
		defer img.Close()
		assert.Equal(t, 1, vipsLoadedPages(img))
	})
	t.Run("AnimatedCountsWhatWasLoaded", func(t *testing.T) {
		// The file declares many pages, but the import parameters ask for one, so counting the
		// declared pages would reject a file whose frames are never decoded.
		name := writeAnimatedGif(t, 60)
		img, err := vips.LoadImageFromFile(name, VipsImportParams())
		assert.NoError(t, err)
		defer img.Close()
		assert.Greater(t, img.Pages(), 1, "the fixture must declare more than one page")
		assert.Equal(t, 1, vipsLoadedPages(img))
	})
}

func TestVipsCheckPixels_Animated(t *testing.T) {
	VipsInit()

	max := fs.MaxImagePixels
	defer func() { fs.MaxImagePixels = max }()

	name := writeAnimatedGif(t, 60)

	t.Run("DeclaredFramesDoNotCount", func(t *testing.T) {
		// A budget that admits one frame but not sixty must admit this file, since only the
		// first frame is loaded.
		fs.MaxImagePixels = animatedGifWidth * animatedGifHeight * 2
		img, err := vips.LoadImageFromFile(name, VipsImportParams())
		assert.NoError(t, err)
		defer img.Close()
		assert.NoError(t, vipsCheckPixels(img, "animated.gif"))
	})
	t.Run("LoadedFrameStillCounts", func(t *testing.T) {
		fs.MaxImagePixels = 4
		img, err := vips.LoadImageFromFile(name, VipsImportParams())
		assert.NoError(t, err)
		defer img.Close()
		assert.ErrorIs(t, vipsCheckPixels(img, "animated.gif"), fs.ErrImageTooLarge)
	})
	t.Run("ThumbnailSucceeds", func(t *testing.T) {
		fs.MaxImagePixels = animatedGifWidth * animatedGifHeight * 2
		_, buf, err := Vips(name, nil, "293456789012345678901234567890abcdef1234", t.TempDir(), 64, 64, ResampleFit)
		assert.NoError(t, err)
		assert.NotEmpty(t, buf)
	})
}

const animatedGifWidth = 200
const animatedGifHeight = 100

// writeAnimatedGif writes an animated GIF with the given number of frames and returns its path.
func writeAnimatedGif(t *testing.T, frames int) string {
	t.Helper()

	g := &gif.GIF{}

	for i := 0; i < frames; i++ {
		p := image.NewPaletted(image.Rect(0, 0, animatedGifWidth, animatedGifHeight), color.Palette{color.Black, color.White})
		p.Set(i%animatedGifWidth, 0, color.White)
		g.Image = append(g.Image, p)
		g.Delay = append(g.Delay, 5)
	}

	name := filepath.Join(t.TempDir(), "animated.gif")

	// #nosec G304 -- the path is a test-owned temporary directory.
	f, err := os.Create(name)
	require.NoError(t, err)
	require.NoError(t, gif.EncodeAll(f, g))
	require.NoError(t, f.Close())

	return name
}

// oversizedJpegHeader returns a JPEG whose SOF0 declares the given geometry while carrying almost
// no scan data, which is the shape a geometry check exists to reject before a decoder sizes its
// buffer from the declaration.
func oversizedJpegHeader(t *testing.T, width, height uint16) []byte {
	t.Helper()

	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, image.NewGray(image.Rect(0, 0, 8, 8)), nil))
	data := buf.Bytes()

	// Patch the SOF0 height and width, which follow the marker, the segment length and the
	// sample precision.
	sof := bytes.Index(data, []byte{0xff, 0xc0})
	require.Greater(t, sof, 0, "the encoder must emit a baseline SOF0 marker")

	binary.BigEndian.PutUint16(data[sof+5:], height)
	binary.BigEndian.PutUint16(data[sof+7:], width)

	return data
}

func TestCheckJpegPixels(t *testing.T) {
	max := fs.MaxImagePixels
	defer func() { fs.MaxImagePixels = max }()

	t.Run("WithinBudget", func(t *testing.T) {
		fs.MaxImagePixels = max
		f, err := os.Open("testdata/example.jpg")
		require.NoError(t, err)
		defer f.Close()
		assert.NoError(t, checkJpegPixels(f, "example.jpg"))
	})
	t.Run("AboveBudget", func(t *testing.T) {
		fs.MaxImagePixels = 4
		f, err := os.Open("testdata/example.jpg")
		require.NoError(t, err)
		defer f.Close()
		assert.ErrorIs(t, checkJpegPixels(f, "example.jpg"), fs.ErrImageTooLarge)
	})
	t.Run("DeclaredGeometryDecidesIt", func(t *testing.T) {
		// The declaration is what a decoder allocates from, so a small file declaring a large
		// frame has to be refused on the declaration rather than on its length.
		fs.MaxImagePixels = 150000000
		data := oversizedJpegHeader(t, 20000, 20000)
		assert.Less(t, len(data), 2048, "the fixture must stay far smaller than what it declares")
		assert.ErrorIs(t, checkJpegPixels(bytes.NewReader(data), "crafted.jpg"), fs.ErrImageTooLarge)
	})
	t.Run("LeavesTheReaderAtTheStart", func(t *testing.T) {
		// decodeImage reads the same reader afterwards, so the guard must not consume it.
		fs.MaxImagePixels = max
		f, err := os.Open("testdata/example.jpg")
		require.NoError(t, err)
		defer f.Close()
		require.NoError(t, checkJpegPixels(f, "example.jpg"))
		pos, err := f.Seek(0, io.SeekCurrent)
		require.NoError(t, err)
		assert.Equal(t, int64(0), pos)
	})
	t.Run("UnreadableHeaderIsLeftToTheDecoder", func(t *testing.T) {
		fs.MaxImagePixels = 4
		assert.NoError(t, checkJpegPixels(bytes.NewReader([]byte("not a jpeg at all")), "junk.jpg"))
	})
	t.Run("Disabled", func(t *testing.T) {
		fs.MaxImagePixels = 0
		data := oversizedJpegHeader(t, 20000, 20000)
		assert.NoError(t, checkJpegPixels(bytes.NewReader(data), "crafted.jpg"))
	})
}

func TestOpenJpeg_PixelBudget(t *testing.T) {
	max := fs.MaxImagePixels
	defer func() { fs.MaxImagePixels = max }()

	t.Run("WithinBudget", func(t *testing.T) {
		fs.MaxImagePixels = max
		img, err := OpenJpeg("testdata/example.jpg", 1)
		assert.NoError(t, err)
		assert.NotNil(t, img)
	})
	t.Run("AboveBudget", func(t *testing.T) {
		// OpenJpeg is reached only when the color path is active, which the current
		// configuration never selects, so this is what keeps the guard from rotting.
		fs.MaxImagePixels = 4
		img, err := OpenJpeg("testdata/example.jpg", 1)
		assert.ErrorIs(t, err, fs.ErrImageTooLarge)
		assert.Nil(t, img)
	})
}
