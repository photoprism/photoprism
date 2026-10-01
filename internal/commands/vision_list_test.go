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
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/fs"
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
	assert.Equal(t, "no", visionInstalled(model, modelsPath))
	description := classify.DefaultModel()
	filename := description.ONNX.FilePath(filepath.Join(modelsPath, string(description.Name)))
	require.NoError(t, os.MkdirAll(filepath.Dir(filename), fs.ModeDir))
	require.NoError(t, os.WriteFile(filename, []byte("artifact"), fs.ModeFile))
	model.DisabledByMode = true
	assert.Equal(t, "yes", visionInstalled(model, modelsPath))
	detector := vision.NewNsfwModel(nsfw.ModelYahoo)
	assert.Equal(t, "no", visionInstalled(detector, modelsPath))
	custom := &vision.Model{Type: vision.ModelTypeNsfw, Name: "custom", Path: "custom/model.onnx"}
	assert.Equal(t, "no", visionInstalled(custom, modelsPath))
	require.NoError(t, os.MkdirAll(filepath.Join(modelsPath, "custom"), fs.ModeDir))
	require.NoError(t, os.WriteFile(filepath.Join(modelsPath, "custom/model.onnx"), []byte("artifact"), fs.ModeFile))
	assert.Equal(t, "yes", visionInstalled(custom, modelsPath))
	custom.Service = vision.Service{Uri: "https://example.com", Method: "POST"}
	assert.Equal(t, "n/a", visionInstalled(custom, modelsPath))
	assert.Equal(t, "n/a", visionInstalled(&vision.Model{Type: vision.ModelTypeCaption}, modelsPath))
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
			assert.Contains(t, row, "status")
		}
	})
}
