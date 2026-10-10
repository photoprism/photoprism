package thumb

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestOpenInput verifies renderer support, explicit orientation, and decode errors.
func TestOpenInput(t *testing.T) {
	library := Library
	t.Cleanup(func() { Library = library })
	Library = LibVips
	t.Run("ArithmeticJPEG", func(t *testing.T) {
		img, err := openInputTest("../../pkg/fs/testdata/arithmetic.jpg", 1)
		require.NoError(t, err)
		assert.Equal(t, image.Rect(0, 0, 64, 48), img.Bounds())
	})
	t.Run("Orientation", func(t *testing.T) {
		source := image.NewNRGBA(image.Rect(0, 0, 20, 10))
		source.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
		file := filepath.Join(t.TempDir(), "source.png")
		require.NoError(t, Save(source, file))
		img, err := openInputTest(file, 6)
		require.NoError(t, err)
		assert.Equal(t, image.Rect(0, 0, 10, 20), img.Bounds())
		r, _, _, _ := img.At(9, 0).RGBA()
		assert.Equal(t, uint32(65535), r)
	})
	t.Run("MalformedProfile", func(t *testing.T) {
		img, err := openInputTest("testdata/icc_profile_bad_length.jpg", 1)
		require.NoError(t, err)
		require.NotNil(t, img)
	})
	t.Run("Missing", func(t *testing.T) {
		_, err := openInputTest(filepath.Join(t.TempDir(), "missing.jpg"), 1)
		require.Error(t, err)
	})
	t.Run("PixelBudget", func(t *testing.T) {
		limit := fs.MaxImagePixels
		t.Cleanup(func() { fs.MaxImagePixels = limit })
		fs.MaxImagePixels = 4
		_, err := openInputTest("../../pkg/fs/testdata/arithmetic.jpg", 1)
		require.ErrorIs(t, err, fs.ErrImageTooLarge)
	})
	t.Run("GoRenderer", func(t *testing.T) {
		Library = LibAuto
		_, err := openInputTest("../../pkg/fs/testdata/arithmetic.jpg", 1)
		require.Error(t, err)
	})
}

// TestOpenInputColor verifies enabled ICC conversion and unchanged disabled colors.
func TestOpenInputColor(t *testing.T) {
	library, colorMode := Library, Color
	Library = LibVips
	t.Cleanup(func() { Library, Color = library, colorMode })
	source := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	draw.Draw(source, source.Bounds(), &image.Uniform{C: color.NRGBA{R: 180, G: 100, B: 60, A: 255}}, image.Point{}, draw.Src)
	var encoded bytes.Buffer
	require.NoError(t, jpeg.Encode(&encoded, source, &jpeg.Options{Quality: 100}))
	profile, err := os.ReadFile("../../assets/profiles/icc/displayp3-v4.icc")
	require.NoError(t, err)
	payload := append([]byte("ICC_PROFILE\x00\x01\x01"), profile...)
	marker := make([]byte, 4)
	marker[0], marker[1] = 0xff, 0xe2
	require.LessOrEqual(t, len(payload)+2, 65535)
	binary.BigEndian.PutUint16(marker[2:], uint16(len(payload)+2)) //nolint:gosec // The fixture marker length is bounded above.
	data := append([]byte{}, encoded.Bytes()[:2]...)
	data = append(data, marker...)
	data = append(data, payload...)
	data = append(data, encoded.Bytes()[2:]...)
	file := filepath.Join(t.TempDir(), "profile.jpg")
	require.NoError(t, os.WriteFile(file, data, fs.ModeFile)) //nolint:gosec // Write only the isolated profile fixture.
	Color = ColorNone
	raw, err := openInputTest(file, 1)
	require.NoError(t, err)
	Color = colorMode
	if Color == ColorNone {
		Color = ColorAuto
	}
	converted, err := openInputTest(file, 1)
	require.NoError(t, err)
	rawRed, _, _, _ := raw.At(4, 4).RGBA()
	convertedRed, _, _, _ := converted.At(4, 4).RGBA()
	assert.Greater(t, convertedRed, rawRed+1000)
}

// TestInputSourceResample verifies region selection, native dimensions, and release behavior.
func TestInputSourceResample(t *testing.T) {
	library := Library
	t.Cleanup(func() { Library = library })
	for _, backend := range []Lib{LibVips, LibAuto} {
		t.Run(string(backend), func(t *testing.T) {
			Library = backend
			source := image.NewNRGBA(image.Rect(0, 0, 453, 301))
			draw.Draw(source, image.Rect(26, 0, 427, 301), &image.Uniform{C: color.NRGBA{R: 255, A: 255}}, image.Point{}, draw.Src)
			file := filepath.Join(t.TempDir(), "source.png")
			require.NoError(t, Save(source, file))
			input, err := OpenInputSource(file, 1)
			require.NoError(t, err)
			defer input.Close()
			pixels, err := input.Resample(image.Rect(26, 0, 427, 301), 224, 224)
			require.NoError(t, err)
			assert.Equal(t, image.Rect(0, 0, 224, 224), pixels.Bounds())
			for _, x := range []int{0, 112, 223} {
				r, _, _, _ := pixels.At(x, 112).RGBA()
				assert.Equal(t, uint32(65535), r)
			}
			input.Close()
			assert.Empty(t, input.Bounds())
			_, err = input.Resample(image.Rect(0, 0, 1, 1), 224, 224)
			require.Error(t, err)
		})
	}
	t.Run("InvalidRegion", func(t *testing.T) {
		source := NewInputSource(image.NewNRGBA(image.Rect(0, 0, 10, 10)))
		defer source.Close()
		_, err := source.Resample(image.Rect(0, 0, 11, 10), 224, 224)
		require.Error(t, err)
	})
}

// openInputTest materializes a small source for decoder assertions.
func openInputTest(file string, orientation int) (image.Image, error) {
	source, err := OpenInputSource(file, orientation)
	if err != nil {
		return nil, err
	}
	defer source.Close()
	bounds := source.Bounds()
	return source.Resample(bounds, bounds.Dx(), bounds.Dy())
}
