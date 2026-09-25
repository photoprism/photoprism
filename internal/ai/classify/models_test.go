package classify

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestModels verifies every registry entry has a complete ImageNet contract.
func TestModels(t *testing.T) {
	assert.Len(t, AutoModelPreference, len(Models))

	for name, model := range Models {
		t.Run(string(name), func(t *testing.T) {
			require.NotNil(t, model)
			require.NotNil(t, model.ONNX)
			require.NotNil(t, model.ONNX.Input)
			require.NotNil(t, model.ONNX.Output)
			assert.Equal(t, name, model.Name)
			assert.NotEmpty(t, model.DisplayName)
			assert.Equal(t, "Apache-2.0", model.ONNX.License)
			assert.Contains(t, model.ONNX.Source, "/resolve/")
			assert.True(t, strings.HasSuffix(model.ONNX.Source, "/model.safetensors"))
			assert.Len(t, model.ONNX.SHA256, 64)
			assert.Equal(t, "input", model.ONNX.Input.Name)
			assert.Equal(t, 224, model.ONNX.Input.Width)
			assert.Equal(t, 224, model.ONNX.Input.Height)
			assert.Equal(t, onnx.LayoutNCHW, model.ONNX.Input.Layout)
			assert.Equal(t, onnx.RGB, model.ONNX.Input.ColorOrder)
			assert.Equal(t, onnx.InterpolationBicubic, model.ONNX.Input.Resize.Interpolation)
			assert.Positive(t, model.ONNX.Input.Resize.ShortEdge)
			assert.Positive(t, model.ONNX.Input.Resize.CropRatio)
			assert.Equal(t, ImageNetClasses, model.ONNX.Output.Width)
			assert.Equal(t, "logits", model.ONNX.Output.Name)
			assert.Equal(t, 1, model.ONNX.Output.Count)
			assert.True(t, model.ONNX.Output.OutputsLogits())
			assert.Empty(t, model.LabelFile)
			assert.True(t, model.CanonicalOrder)
		})
	}
}

// TestModelArtifacts verifies installed candidates and the shared vocabulary match the registry.
func TestModelArtifacts(t *testing.T) {
	labels := imageNetLabelNames
	require.Len(t, labels, ImageNetClasses)
	assert.Equal(t, "tench fish", labels[0])
	assert.Equal(t, "goldfish", labels[1])
	assert.Equal(t, "toilet tissue", labels[len(labels)-1])
	digest := sha256.Sum256([]byte(imageNetLabels))
	assert.Equal(t, "7137cf8c32076dfd6369ce6b08f2e26207164a385141a61aff0ae7c4a8fcdc13", hex.EncodeToString(digest[:]))
	assert.NotEqual(t, "background", strings.ToLower(strings.TrimSpace(labels[0])))

	for name, model := range Models {
		t.Run(string(name), func(t *testing.T) {
			modelPath := model.ONNX.FilePath(filepath.Join(modelsPath, string(name)))
			if _, statErr := os.Stat(modelPath); os.IsNotExist(statErr) {
				t.Skipf("%s is not installed", name)
			}

			assert.Equal(t, model.ONNX.SHA256, fs.Sha256(modelPath))
		})
	}
}

// TestModelDescriptionInstalled verifies registered artifact discovery.
func TestModelDescriptionInstalled(t *testing.T) {
	description := DefaultModel()
	require.NotNil(t, description)
	modelsDir := t.TempDir()
	modelDir := filepath.Join(modelsDir, string(description.Name))
	modelPath := description.ONNX.FilePath(modelDir)

	assert.False(t, description.Installed(modelsDir))
	require.NoError(t, os.MkdirAll(modelDir, fs.ModeDir))
	require.NoError(t, os.WriteFile(modelPath, nil, fs.ModeFile))
	assert.False(t, description.Installed(modelsDir))
	require.NoError(t, os.WriteFile(modelPath, []byte("model"), fs.ModeFile))
	assert.True(t, description.Installed(modelsDir))
}

// TestModelDownloadRegistry verifies every selectable candidate has a checksum-pinned installer entry.
func TestModelDownloadRegistry(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "..", "scripts", "dist", "download-models.sh")
	data, err := os.ReadFile(scriptPath) //nolint:gosec // repository test fixture
	require.NoError(t, err)
	script := string(data)

	for name, model := range Models {
		entry := string(name) + "|"
		assert.Contains(t, script, entry)
		assert.Contains(t, script, model.ONNX.SHA256)
		assert.Contains(t, script, model.ONNX.File)
	}
}

// TestModelNames verifies selection parsing and help text.
func TestModelNames(t *testing.T) {
	assert.Equal(t, ModelAuto, ParseModelName(""))
	assert.Equal(t, ModelNone, ParseModelName("NONE"))
	assert.Equal(t, ModelRepViTM10, ParseModelName("RepViT M1.0"))
	assert.NotNil(t, DefaultModel())
	assert.Contains(t, ModelUsageString(), string(DefaultModelName()))
}

// TestDefaultModelName pins the default and automatic label-model preference.
func TestDefaultModelName(t *testing.T) {
	assert.Equal(t, ModelEfficientFormerV2S2, DefaultModelName())
	require.NotEmpty(t, AutoModelPreference)
	assert.Equal(t, DefaultModelName(), AutoModelPreference[0])
}
