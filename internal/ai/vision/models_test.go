package vision

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
)

// TestNewDefaultLabelModel verifies that the default labels model is the registered default classifier.
func TestNewDefaultLabelModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		model := newDefaultLabelModel()
		require.NotNil(t, model)
		assert.Equal(t, ModelTypeLabels, model.Type)
		assert.Equal(t, string(classify.DefaultModelName()), model.Name)
		assert.True(t, model.Default)
		assert.NotNil(t, model.ONNX)
		assert.Empty(t, model.Run)
	})
	t.Run("NotShared", func(t *testing.T) {
		assert.NotSame(t, newDefaultLabelModel(), newDefaultLabelModel())
	})
}
