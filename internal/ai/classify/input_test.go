package classify

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/onnx"
)

// TestPreparedInputBlob verifies explicit geometry bypass and unchanged normalization.
func TestPreparedInputBlob(t *testing.T) {
	m := NewRegisteredModel(modelsPath, ModelEfficientFormerV2S2, onnx.DefaultProvider, false)
	m.mean = m.meta.Input.Normalization.Mean
	m.scales = m.meta.Input.Normalization.Scales()
	img := image.NewNRGBA(image.Rect(0, 0, 224, 224))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.NRGBA{G: 128, A: 255}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(0, 0, 3, 224), &image.Uniform{C: color.NRGBA{R: 255, A: 255}}, image.Point{}, draw.Src)
	t.Run("Prepared", func(t *testing.T) {
		blob, err := m.buildInputBlob(Input{Image: img, Prepared: true})
		require.NoError(t, err)
		require.Len(t, blob, 3*224*224)
		assert.InDelta(t, (255-imageNetMean[0])/imageNetStdDev[0], blob[0], 1e-6)
		assert.InDelta(t, (128-imageNetMean[1])/imageNetStdDev[1], blob[224*224+112], 1e-6)
		assert.InDelta(t, -imageNetMean[2]/imageNetStdDev[2], blob[2*224*224], 1e-6)
	})
	t.Run("Native224StillCrops", func(t *testing.T) {
		blob, err := m.buildBlob(img)
		require.NoError(t, err)
		assert.InDelta(t, -imageNetMean[0]/imageNetStdDev[0], blob[0], 1e-6)
	})
	t.Run("WrongDimensions", func(t *testing.T) {
		_, err := m.buildInputBlob(Input{Image: image.NewNRGBA(image.Rect(0, 0, 225, 224)), Prepared: true})
		require.ErrorContains(t, err, "expected 224x224")
	})
	t.Run("MissingPixels", func(t *testing.T) {
		_, err := m.buildInputBlob(Input{Prepared: true})
		require.Error(t, err)
	})
	t.Run("OtherModel", func(t *testing.T) {
		other := NewModel(Settings{Name: "custom", Info: m.meta})
		_, err := other.buildInputBlob(Input{Image: img, Prepared: true})
		require.ErrorContains(t, err, "require EfficientFormerV2-S2")
	})
}

// TestPredict verifies disabled and invalid decoded input inference.
func TestPredict(t *testing.T) {
	t.Run("Disabled", func(t *testing.T) {
		m := NewModel(Settings{Disabled: true})
		labels, err := m.Predict(Input{}, 20)
		require.NoError(t, err)
		assert.Empty(t, labels)
	})
	t.Run("Nil", func(t *testing.T) {
		var m *Model
		labels, err := m.Predict(Input{}, 20)
		require.NoError(t, err)
		assert.Empty(t, labels)
	})
	t.Run("InvalidModel", func(t *testing.T) {
		_, err := NewModel(Settings{}).Predict(Input{}, 20)
		require.Error(t, err)
	})
}
