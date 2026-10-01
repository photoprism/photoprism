package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

func TestVisionEndpoint(t *testing.T) {
	cases := []struct {
		name   string
		uri    string
		method string
		want   string
	}{
		{
			name:   "Plain",
			uri:    "http://ollama:11434/api/generate",
			method: "POST",
			want:   "POST http://ollama:11434/api/generate",
		},
		{ //nolint:gosec // example URL, the password in it is exactly what this case redacts
			name:   "BasicAuth",
			uri:    "https://vision:secret@vision.example.com/api/generate",
			method: "POST",
			want:   "POST https://vision:***@vision.example.com/api/generate",
		},
		{
			// Nothing distinguishes a name in this position from an access token, which several
			// services carry there, so the name goes and the endpoint is still identified by its
			// scheme, host and path.
			name:   "NameWithoutPassword",
			uri:    "https://vision@vision.example.com/api/generate",
			method: "POST",
			want:   "POST https://***@vision.example.com/api/generate",
		},
		{
			name:   "QueryIsKept",
			uri:    "https://api.example.com/v1/responses?tier=flex",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?tier=flex",
		},
		{
			name:   "MissingUri",
			uri:    "",
			method: "POST",
			want:   "",
		},
		{
			name:   "MissingMethod",
			uri:    "https://api.example.com/v1/responses",
			method: "",
			want:   "",
		},
		{
			name:   "Unparsable",
			uri:    "://nope",
			method: "POST",
			want:   "POST ?",
		},
		{
			name:   "ApiKeyQuery",
			uri:    "https://api.example.com/v1/responses?api_key=notreal",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?api_key=***",
		},
		{
			name:   "AccessTokenQuery",
			uri:    "https://api.example.com/v1/responses?access_token=notreal",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?access_token=***",
		},
		{
			name:   "MixedCaseKeyQuery",
			uri:    "https://api.example.com/v1/responses?X-Api-Key=notreal",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?X-Api-Key=***",
		},
		{
			name:   "CredentialQueryBesideAKeptOne",
			uri:    "https://api.example.com/v1/responses?tier=flex&token=notreal",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?tier=flex&token=***",
		},
		{
			name:   "RepeatedCredentialQuery",
			uri:    "https://api.example.com/v1/responses?secret=one&secret=two",
			method: "POST",
			want:   "POST https://api.example.com/v1/responses?secret=***&secret=***",
		},
		{ //nolint:gosec // example URL, the credentials in it are exactly what this case redacts
			name:   "UserinfoAndQueryTogether",
			uri:    "https://vision:notreal@api.example.com/v1?signature=notreal",
			method: "POST",
			want:   "POST https://vision:***@api.example.com/v1?signature=***",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, visionEndpoint(tc.uri, tc.method))
		})
	}
}

// TestVisionInstalled verifies listing inspects artifacts without loading models.
func TestVisionInstalled(t *testing.T) {
	modelsPath := t.TempDir()
	model := vision.NewLabelModel(classify.DefaultModelName())
	assert.Equal(t, report.No, visionInstalled(model, modelsPath))
	description := classify.DefaultModel()
	filename := description.ONNX.FilePath(filepath.Join(modelsPath, string(description.Name)))
	require.NoError(t, os.MkdirAll(filepath.Dir(filename), fs.ModeDir))
	require.NoError(t, os.WriteFile(filename, []byte("artifact"), fs.ModeFile))
	model.DisabledByMode = true
	assert.Equal(t, report.Yes, visionInstalled(model, modelsPath))
	detector := vision.NewNsfwModel(nsfw.ModelYahoo)
	assert.Equal(t, report.No, visionInstalled(detector, modelsPath))
	custom := &vision.Model{Type: vision.ModelTypeNsfw, Name: "custom", Path: "custom/model.onnx"}
	assert.Equal(t, report.No, visionInstalled(custom, modelsPath))
	require.NoError(t, os.MkdirAll(filepath.Join(modelsPath, "custom"), fs.ModeDir))
	require.NoError(t, os.WriteFile(filepath.Join(modelsPath, "custom/model.onnx"), []byte("artifact"), fs.ModeFile))
	assert.Equal(t, report.Yes, visionInstalled(custom, modelsPath))
	custom.Service = vision.Service{Uri: "https://example.com", Method: "POST"}
	assert.Equal(t, report.NotAssigned, visionInstalled(custom, modelsPath))
	assert.Equal(t, report.NotAssigned, visionInstalled(&vision.Model{Type: vision.ModelTypeCaption}, modelsPath))
}

// TestVisionListCommand verifies installation status is rendered in table and JSON output.
func TestVisionListCommand(t *testing.T) {
	t.Run("Table", func(t *testing.T) {
		output, err := RunWithTestContext(VisionListCommand, []string{"ls"})
		require.NoError(t, err)
		assert.Contains(t, strings.ToLower(output), "installed")
	})
	t.Run("JSON", func(t *testing.T) {
		output, err := RunWithTestContext(VisionListCommand, []string{"ls", "--json"})
		require.NoError(t, err)
		var rows []map[string]any
		require.NoError(t, json.Unmarshal([]byte(output), &rows))
		require.NotEmpty(t, rows)
		for _, row := range rows {
			assert.Contains(t, row, "installed")
			assert.Contains(t, row, "provider")
			assert.Contains(t, row, "status")
		}
	})
}

// visionListValue returns the value of the named column in a "vision ls" row.
func visionListValue(t *testing.T, row []string, col string) string {
	t.Helper()

	for i, name := range visionListCols {
		if name == col {
			return row[i]
		}
	}

	t.Fatalf("unknown column %s", col)

	return ""
}

// TestVisionListRow verifies that rows report what each model runs with, including the face
// embedding model FACE_MODEL selects and the execution provider of local ONNX models.
func TestVisionListRow(t *testing.T) {
	settings := visionListSettings{
		ModelsPath: t.TempDir(),
		Provider:   onnx.ProviderCUDA,
		FaceModel:  face.ModelSFace,
		FaceActive: true,
		FaceRun:    vision.RunAuto,
	}

	remote := &vision.Model{Type: vision.ModelTypeLabels, Engine: "ollama", Name: "qwen3-vl:4b-instruct", Run: vision.RunOnSchedule}
	remote.ApplyEngineDefaults()

	t.Run("DefaultLabels", func(t *testing.T) {
		row := visionListRow(vision.NewLabelModel(classify.DefaultModelName()), settings)
		require.Len(t, row, len(visionListCols))
		assert.Equal(t, string(classify.DefaultModelName()), visionListValue(t, row, "Model"))
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, "cuda", visionListValue(t, row, "Provider"))
		assert.Equal(t, "224", visionListValue(t, row, "Resolution"))
		assert.Empty(t, visionListValue(t, row, "Options"))
		assert.Equal(t, "auto", visionListValue(t, row, "Schedule"))
		assert.Equal(t, report.Enabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.No, visionListValue(t, row, "Installed"))
	})
	t.Run("NamedLabels", func(t *testing.T) {
		row := visionListRow(&vision.Model{Type: vision.ModelTypeLabels, Name: string(classify.ModelEfficientFormerV2S1)}, settings)
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, "cuda", visionListValue(t, row, "Provider"))
		assert.Equal(t, "224", visionListValue(t, row, "Resolution"))
	})
	t.Run("NamedDetector", func(t *testing.T) {
		row := visionListRow(&vision.Model{Type: vision.ModelTypeNsfw, Name: string(nsfw.ModelFalconsai)}, settings)
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, "224", visionListValue(t, row, "Resolution"))
	})
	t.Run("DisabledByMode", func(t *testing.T) {
		model := vision.NewNsfwModel(nsfw.ModelYahoo)
		model.DisabledByMode = true
		assert.Equal(t, report.Disabled, visionListValue(t, visionListRow(model, settings), "Status"))
	})
	t.Run("Remote", func(t *testing.T) {
		row := visionListRow(remote, settings)
		assert.Equal(t, "qwen3-vl:4b-instruct", visionListValue(t, row, "Model"))
		assert.Equal(t, "ollama", visionListValue(t, row, "Engine"))
		assert.Empty(t, visionListValue(t, row, "Provider"))
		assert.Contains(t, visionListValue(t, row, "Endpoint"), "POST ")
		assert.Equal(t, vision.RunOnSchedule, visionListValue(t, row, "Schedule"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
	})
	t.Run("FaceModel", func(t *testing.T) {
		row := visionListRow(vision.FacenetModel.Clone(), settings)
		assert.Equal(t, face.ModelSFace, visionListValue(t, row, "Model"))
		assert.Equal(t, vision.EngineONNX, visionListValue(t, row, "Engine"))
		assert.Equal(t, "cuda", visionListValue(t, row, "Provider"))
		assert.Equal(t, "112", visionListValue(t, row, "Resolution"))
		assert.Empty(t, visionListValue(t, row, "Options"))
		assert.Equal(t, report.Enabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.No, visionListValue(t, row, "Installed"))
	})
	t.Run("FaceNone", func(t *testing.T) {
		none := settings
		none.FaceModel, none.FaceActive, none.FaceRun = face.ModelNone, false, vision.RunNever
		row := visionListRow(vision.FacenetModel.Clone(), none)
		assert.Equal(t, face.ModelNone, visionListValue(t, row, "Model"))
		assert.Empty(t, visionListValue(t, row, "Engine"))
		assert.Empty(t, visionListValue(t, row, "Provider"))
		assert.Empty(t, visionListValue(t, row, "Resolution"))
		assert.Equal(t, vision.RunNever, visionListValue(t, row, "Schedule"))
		assert.Equal(t, report.Disabled, visionListValue(t, row, "Status"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
	})
	t.Run("FaceNetEntry", func(t *testing.T) {
		facenet := settings
		facenet.FaceModel = face.ModelFaceNet
		model := vision.FacenetModel.Clone()
		row := visionListRow(model, facenet)
		assert.Equal(t, face.ModelFaceNet, visionListValue(t, row, "Model"))
		assert.Equal(t, vision.EngineTensorFlow, visionListValue(t, row, "Engine"))
		assert.Empty(t, visionListValue(t, row, "Provider"))
		assert.Equal(t, "160", visionListValue(t, row, "Resolution"))
		assert.Equal(t, `{"tags":"serve"}`, visionListValue(t, row, "Options"))
		assert.Equal(t, report.No, visionListValue(t, row, "Installed"))

		// A custom entry loads from its own path, which the registry knows nothing about.
		model.Name = "facenet_custom"
		row = visionListRow(model, facenet)
		assert.Equal(t, "facenet_custom", visionListValue(t, row, "Model"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
	})
	t.Run("FaceEntryDisabled", func(t *testing.T) {
		model := vision.FacenetModel.Clone()
		model.Disabled = true
		assert.Equal(t, report.Disabled, visionListValue(t, visionListRow(model, settings), "Status"))
	})
	t.Run("FaceEndpoint", func(t *testing.T) {
		model := &vision.Model{Type: vision.ModelTypeFace, Name: "facenet", Run: vision.RunManual,
			Service: vision.Service{Uri: "https://vision.example.com/api/v1/face", Method: "POST"}}
		row := visionListRow(model, settings)
		assert.Equal(t, "facenet", visionListValue(t, row, "Model"))
		assert.Equal(t, "POST https://vision.example.com/api/v1/face", visionListValue(t, row, "Endpoint"))
		assert.Equal(t, "auto", visionListValue(t, row, "Schedule"))
		assert.Equal(t, report.NotAssigned, visionListValue(t, row, "Installed"))
	})
}

// TestNewVisionListSettings verifies the settings follow the instance configuration.
func TestNewVisionListSettings(t *testing.T) {
	t.Run("FaceModelNone", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().FaceModel = face.ModelNone
		settings := newVisionListSettings(conf)
		assert.Equal(t, face.ModelNone, settings.FaceModel)
		assert.False(t, settings.FaceActive)
	})
	t.Run("FaceModelNotInstalled", func(t *testing.T) {
		conf := config.NewConfig(config.CliTestContext())
		conf.Options().ModelsPath = t.TempDir()
		conf.Options().FaceModel = face.ModelAuraFace
		settings := newVisionListSettings(conf)
		assert.Equal(t, face.ModelAuraFace, settings.FaceModel)
		assert.False(t, settings.FaceActive)
	})

	conf := get.Config()
	settings := newVisionListSettings(conf)
	assert.Equal(t, conf.ModelsPath(), settings.ModelsPath)
	assert.Equal(t, conf.OnnxProvider(), settings.Provider)
	assert.Equal(t, conf.FaceEngineRunType(), settings.FaceRun)

	disabled := conf.Options().DisableFaces
	t.Cleanup(func() { conf.Options().DisableFaces = disabled })
	conf.Options().DisableFaces = true
	settings = newVisionListSettings(conf)
	assert.False(t, settings.FaceActive)
	assert.Equal(t, vision.RunNever, settings.FaceRun)
}

// TestVisionRegisteredWidth verifies the input width of registered models named in vision.yml.
func TestVisionRegisteredWidth(t *testing.T) {
	assert.Equal(t, 224, visionRegisteredWidth(&vision.Model{Type: vision.ModelTypeLabels, Name: string(classify.ModelRepViTM10)}))
	assert.Equal(t, 224, visionRegisteredWidth(&vision.Model{Type: vision.ModelTypeNsfw, Name: string(nsfw.ModelYahoo)}))
	assert.Zero(t, visionRegisteredWidth(&vision.Model{Type: vision.ModelTypeLabels, Name: "custom_21k"}))
	assert.Zero(t, visionRegisteredWidth(&vision.Model{Type: vision.ModelTypeCaption, Name: string(nsfw.ModelYahoo)}))
}
