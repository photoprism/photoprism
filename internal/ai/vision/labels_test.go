package vision

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/media"
)

func TestGenerateLabels(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		result, err := GenerateLabels(Files{examplesPath + "/chameleon_lime.jpg"}, media.SrcLocal, entity.SrcAuto)

		assert.NoError(t, err)
		assert.IsType(t, classify.Labels{}, result)
		assert.Equal(t, 1, len(result))

		t.Log(result)

		assert.Equal(t, "chameleon", result[0].Name)
		assert.Equal(t, 7, result[0].Uncertainty)
	})
	t.Run("Cat224", func(t *testing.T) {
		result, err := GenerateLabels(Files{examplesPath + "/cat_224.jpeg"}, media.SrcLocal, entity.SrcAuto)

		assert.NoError(t, err)
		assert.IsType(t, classify.Labels{}, result)
		assert.Equal(t, 1, len(result))

		t.Log(result)

		assert.Equal(t, "cat", result[0].Name)
		assert.InDelta(t, 59, result[0].Uncertainty, 10)
		assert.InDelta(t, float32(0.41), result[0].Confidence(), 0.1)
	})
	t.Run("Cat720", func(t *testing.T) {
		result, err := GenerateLabels(Files{examplesPath + "/cat_720.jpeg"}, media.SrcLocal, entity.SrcAuto)

		assert.NoError(t, err)
		assert.IsType(t, classify.Labels{}, result)
		assert.Equal(t, 1, len(result))

		t.Log(result)

		assert.Equal(t, "cat", result[0].Name)
		assert.InDelta(t, 60, result[0].Uncertainty, 10)
		assert.InDelta(t, float32(0.4), result[0].Confidence(), 0.1)
	})
	t.Run("InvalidFile", func(t *testing.T) {
		_, err := GenerateLabels(Files{examplesPath + "/notexisting.jpg"}, media.SrcLocal, entity.SrcAuto)
		assert.Error(t, err)
	})
}
