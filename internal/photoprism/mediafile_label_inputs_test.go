package photoprism

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// TestPrepareLabelInputsGeometry verifies bounded crops using independently placed edge markers.
func TestPrepareLabelInputsGeometry(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		w, h, left, top, cw, ch int
	}{
		{"FourThirds", 400, 300, 0, 0, 400, 300},
		{"ThreeTwo", 450, 300, 25, 0, 400, 300},
		{"Wide", 1600, 900, 200, 0, 1200, 900},
		{"Panorama", 1200, 300, 400, 0, 400, 300},
		{"Portrait", 300, 400, 0, 0, 300, 400},
		{"Tall", 300, 1200, 0, 400, 300, 400},
		{"Rounding", 453, 301, 26, 0, 401, 301},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := image.NewNRGBA(image.Rect(0, 0, tc.w, tc.h))
			draw.Draw(source, source.Bounds(), &image.Uniform{C: color.NRGBA{B: 255, A: 255}}, image.Point{}, draw.Src)
			region := image.Rect(tc.left, tc.top, tc.left+tc.cw, tc.top+tc.ch)
			draw.Draw(source, region, &image.Uniform{C: color.NRGBA{R: 255, A: 255}}, image.Point{}, draw.Src)
			for y := region.Min.Y; y < region.Max.Y; y++ {
				for x := region.Min.X; x < region.Max.X; x++ {
					source.SetNRGBA(x, y, color.NRGBA{R: uint8((x*17 + y*13) % 256), G: uint8((x*7 + y*19) % 256), B: uint8((x + y) % 256), A: 255}) //nolint:gosec // Synthetic channels are bounded to 0-255.
				}
			}
			expectedSource := image.NewNRGBA(image.Rect(0, 0, tc.cw, tc.ch))
			draw.Draw(expectedSource, expectedSource.Bounds(), source, region.Min, draw.Src)
			expected := thumb.ResampleWithFilter(expectedSource, 224, 224, thumb.ResampleResize, thumb.ResampleCubic)
			inputs, geometry, err := prepareLabelInputs(source.Bounds(), func() (*thumb.InputSource, error) { return thumb.NewInputSource(source), nil }, func() (*thumb.InputSource, error) { return thumb.NewInputSource(source), nil })
			require.NoError(t, err)
			require.Len(t, inputs, 2)
			for _, input := range inputs {
				assert.Equal(t, image.Rect(0, 0, 224, 224), input.Image.Bounds())
				assert.True(t, input.Prepared)
			}
			assert.Equal(t, expected, inputs[1].Image)
			require.Len(t, geometry, 2)
			assert.Equal(t, source.Bounds(), geometry[1].SourceBounds)
			assert.Equal(t, region, geometry[1].CropBounds)

		})
	}
	t.Run("EdgeAnimal", func(t *testing.T) {
		source := image.NewNRGBA(image.Rect(0, 0, 600, 400))
		draw.Draw(source, image.Rect(60, 100, 100, 300), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
		inputs, _, err := prepareLabelInputs(source.Bounds(), func() (*thumb.InputSource, error) { return thumb.NewInputSource(source), nil }, func() (*thumb.InputSource, error) { return thumb.NewInputSource(source), nil })
		require.NoError(t, err)
		r, _, _, _ := inputs[1].Image.At(20, 112).RGBA()
		assert.Greater(t, r, uint32(60000))
		r, _, _, _ = inputs[0].Image.At(20, 112).RGBA()
		assert.Zero(t, r)
	})
	t.Run("LoadError", func(t *testing.T) {
		_, _, err := prepareLabelInputs(image.Rect(0, 0, 300, 300), func() (*thumb.InputSource, error) { return nil, errors.New("read failed") }, nil)
		require.ErrorContains(t, err, "read failed")
	})
	t.Run("Invalid", func(t *testing.T) {
		_, _, err := prepareLabelInputs(image.Rectangle{}, nil, nil)
		require.Error(t, err)
	})
}

// TestMediaFilePreparedLabels verifies production selection and exact input counts.
func TestMediaFilePreparedLabels(t *testing.T) {
	previous, previousVision := Config(), vision.Config
	cfg := config.NewMinimalTestConfig(t.TempDir())
	require.NoError(t, cfg.CreateDirectories())
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(previous)
		vision.Config = previousVision
		vision.SetPreparedLabelsFunc(nil)
		vision.SetLabelsFunc(nil)
	})
	for _, tc := range []struct {
		name        string
		w, h, count int
	}{
		{"Square", 300, 300, 1}, {"NearSquare", 301, 300, 2}, {"Landscape", 600, 400, 2},
		{"Small", 180, 120, 1}, {"SmallTall", 120, 600, 1}, {"Boundary", 336, 224, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "source.png")
			require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, tc.w, tc.h)), file))
			m, err := NewMediaFile(file)
			require.NoError(t, err)
			vision.Config = &vision.ConfigValues{Models: vision.Models{vision.NewLabelModel(classify.ModelEfficientFormerV2S2)}}
			calls := 0
			vision.SetPreparedLabelsFunc(func(inputs []classify.Input, src entity.Src) (classify.Labels, error) {
				calls++
				assert.Len(t, inputs, tc.count)
				for _, input := range inputs {
					assert.True(t, input.Prepared)
					assert.Equal(t, image.Rect(0, 0, 224, 224), input.Image.Bounds())
				}
				return classify.Labels{{Name: "test", Source: src}}, nil
			})
			vision.SetLabelsFunc(func(vision.Files, media.Src, entity.Src) (classify.Labels, error) {
				t.Fatal("unexpected native path")
				return nil, nil
			})
			labels := m.GenerateLabels(entity.SrcManual)
			require.Len(t, labels, 1)
			assert.Equal(t, 1, calls)
			assert.Equal(t, entity.SrcManual, labels[0].Source)
		})
	}
}

// TestLabelWholeSource verifies that the smallest adequate cached rendition is used, the largest cached one
// otherwise, and the original only when nothing is cached.
func TestLabelWholeSource(t *testing.T) {
	cached, demand := thumb.SizeCached, thumb.SizeOnDemand
	thumb.SizeCached, thumb.SizeOnDemand = 4096, 4096
	t.Cleanup(func() { thumb.SizeCached, thumb.SizeOnDemand = cached, demand })
	previous := Config()
	cfg := config.NewMinimalTestConfig(t.TempDir())
	require.NoError(t, cfg.CreateDirectories())
	SetConfig(cfg)
	t.Cleanup(func() { SetConfig(previous) })
	file := filepath.Join(t.TempDir(), "source.png")
	require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 6000, 600)), file))
	m, err := NewMediaFile(file)
	require.NoError(t, err)
	for _, size := range []thumb.Size{thumb.SizeFit720, thumb.SizeFit1920} {
		_, err = m.Thumbnail(cfg.ThumbCachePath(), size.Name)
		require.NoError(t, err)
	}
	// No cached rendition has a 224 px short side, so the largest cached one is used, not the original.
	img, err := m.labelWholeSource()
	require.NoError(t, err)
	assert.Equal(t, 1920, img.Bounds().Dx())
	img.Close()
	_, err = m.Thumbnail(cfg.ThumbCachePath(), thumb.Fit4096)
	require.NoError(t, err)
	img, err = m.labelWholeSource()
	require.NoError(t, err)
	assert.Equal(t, 4096, img.Bounds().Dx())
	assert.Less(t, img.Bounds().Dy(), 600)
	img.Close()

	// A 3:2 source: fit_720 is adequate and preferred over larger renditions.
	file = filepath.Join(t.TempDir(), "landscape.png")
	require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 3000, 2000)), file))
	m, err = NewMediaFile(file)
	require.NoError(t, err)
	img, err = m.labelWholeSource()
	require.NoError(t, err)
	assert.Equal(t, image.Rect(0, 0, 3000, 2000), img.Bounds(), "original when nothing is cached")
	img.Close()
	for _, size := range []thumb.Size{thumb.SizeFit720, thumb.SizeFit1920} {
		_, err = m.Thumbnail(cfg.ThumbCachePath(), size.Name)
		require.NoError(t, err)
	}
	img, err = m.labelWholeSource()
	require.NoError(t, err)
	assert.Equal(t, 720, img.Bounds().Dx())
	img.Close()
}

// TestMediaFileNativeLabelModels verifies unchanged custom-size and remote thumbnail branches.
func TestMediaFileNativeLabelModels(t *testing.T) {
	previous, previousVision := Config(), vision.Config
	cfg := config.NewMinimalTestConfig(t.TempDir())
	require.NoError(t, cfg.CreateDirectories())
	SetConfig(cfg)
	t.Cleanup(func() {
		SetConfig(previous)
		vision.Config = previousVision
		vision.SetPreparedLabelsFunc(nil)
		vision.SetLabelsFunc(nil)
	})
	file := filepath.Join(t.TempDir(), "source.png")
	require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 600, 400)), file))
	m, err := NewMediaFile(file)
	require.NoError(t, err)
	for _, tc := range []struct {
		name  string
		model *vision.Model
		count int
	}{
		{"OtherONNX", vision.NewLabelModel(classify.ModelRepViTM10), 3},
		{"CustomSize", &vision.Model{Type: vision.ModelTypeLabels, Name: string(classify.ModelEfficientFormerV2S2), Resolution: 384}, 1},
		{"Remote", &vision.Model{Type: vision.ModelTypeLabels, Name: string(classify.ModelEfficientFormerV2S2), Service: vision.Service{Uri: "http://localhost:5000"}}, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vision.Config = &vision.ConfigValues{Models: vision.Models{tc.model}}
			vision.SetPreparedLabelsFunc(func([]classify.Input, entity.Src) (classify.Labels, error) {
				t.Fatal("unexpected prepared path")
				return nil, nil
			})
			calls := 0
			vision.SetLabelsFunc(func(files vision.Files, source media.Src, labelSource entity.Src) (classify.Labels, error) {
				calls++
				assert.Len(t, files, tc.count)
				assert.Equal(t, media.SrcLocal, source)
				return classify.Labels{{Name: "native", Source: labelSource}}, nil
			})
			require.Len(t, m.GenerateLabels(entity.SrcImage), 1)
			assert.Equal(t, 1, calls)
		})
	}
}

// TestSmallLabelInputGeometry verifies the single small input preserves proportions.
func TestSmallLabelInputGeometry(t *testing.T) {
	source := image.NewNRGBA(image.Rect(0, 0, 180, 120))
	draw.Draw(source, image.Rect(30, 0, 150, 120), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	calls := 0
	inputs, _, err := prepareLabelInputs(source.Bounds(), func() (*thumb.InputSource, error) { calls++; return thumb.NewInputSource(source), nil }, func() (*thumb.InputSource, error) {
		t.Fatal("small image must not request a whole-photo input")
		return nil, nil
	})
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	assert.Equal(t, 1, calls)
	for _, x := range []int{0, 112, 223} {
		r, _, _, _ := inputs[0].Image.At(x, 112).RGBA()
		assert.Equal(t, uint32(65535), r)
	}
}

// TestPrepareLabelInputsDecodeFallback verifies renderer and center-cache recovery.
func TestPrepareLabelInputsDecodeFallback(t *testing.T) {
	previous, library := Config(), thumb.Library
	cfg := config.NewMinimalTestConfig(t.TempDir())
	require.NoError(t, cfg.CreateDirectories())
	SetConfig(cfg)
	thumb.Library = thumb.LibVips
	t.Cleanup(func() { SetConfig(previous); thumb.Library = library })
	t.Run("ArithmeticJPEG", func(t *testing.T) {
		m, err := NewMediaFile("../../pkg/fs/testdata/arithmetic.jpg")
		require.NoError(t, err)
		inputs, err := m.PrepareLabelInputs()
		require.NoError(t, err)
		require.Len(t, inputs, 1)
	})
	t.Run("CorruptCenter", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "source.png")
		require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 600, 400)), file))
		m, err := NewMediaFile(file)
		require.NoError(t, err)
		cached, err := m.Thumbnail(cfg.ThumbCachePath(), thumb.Tile224)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(cached, []byte("invalid image"), fs.ModeFile))
		inputs, err := m.PrepareLabelInputs()
		require.NoError(t, err)
		require.Len(t, inputs, 2)
		data, err := os.ReadFile(cached) //nolint:gosec // Read the isolated thumbnail fixture.
		require.NoError(t, err)
		assert.Equal(t, "invalid image", string(data))
	})

}

// TestPrepareSmallSquareSource verifies direct original pixels without an intermediate tile.
func TestPrepareSmallSquareSource(t *testing.T) {
	previous := Config()
	cfg := config.NewMinimalTestConfig(t.TempDir())
	require.NoError(t, cfg.CreateDirectories())
	SetConfig(cfg)
	t.Cleanup(func() { SetConfig(previous) })
	source := image.NewNRGBA(image.Rect(0, 0, 224, 224))
	draw.Draw(source, source.Bounds(), &image.Uniform{C: color.NRGBA{R: 255, A: 255}}, image.Point{}, draw.Src)
	file := filepath.Join(t.TempDir(), "square.png")
	require.NoError(t, thumb.Save(source, file))
	m, err := NewMediaFile(file)
	require.NoError(t, err)
	cached, err := thumb.SizeTile224.FileName(m.Hash(), cfg.ThumbCachePath())
	require.NoError(t, err)
	require.NoError(t, thumb.Save(image.NewNRGBA(image.Rect(0, 0, 224, 224)), cached))
	inputs, err := m.PrepareLabelInputs()
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	r, _, _, _ := inputs[0].Image.At(112, 112).RGBA()
	assert.Equal(t, uint32(65535), r)
}
