package vision

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/tensorflow"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
)

// TestService_UriMissing checks which services without a Uri cannot use the shared Vision API service.
func TestService_UriMissing(t *testing.T) {
	for _, tc := range []struct {
		format ApiFormat
		want   bool
	}{
		{"", false},
		{ApiFormatVision, false},
		{ApiFormatImages, false},
		{ApiFormatUrl, false},
		{ApiFormatOpenAI, true},
		{ApiFormatOllama, true},
		{" Vision ", false},
		{"OpenAI", true},
	} {
		format, want := tc.format, tc.want
		assert.Equal(t, want, (&Service{RequestFormat: format}).UriMissing(), format)
		assert.Equal(t, want, (&Service{RequestFormat: format}).UriUnresolved(), format)
		assert.False(t, (&Service{RequestFormat: format, Uri: "https://llm.example.com/v1/responses"}).UriMissing(), format)
		assert.False(t, (&Service{RequestFormat: format, Disabled: true}).UriMissing(), format)
	}

	var service *Service
	assert.False(t, service.UriMissing())
}

// TestModel_EndpointUriMissing checks that a model whose request format needs its own Uri is not sent to
// the shared service, and that its runs fail with an error naming the format.
func TestModel_EndpointUriMissing(t *testing.T) {
	images := Files{fs.Abs("./testdata/face_160x160.jpg")}

	t.Run("OpenAIFormat", func(t *testing.T) {
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)
		useSharedService(t, server.URL+"/api/v1/vision", "shared-vision-key")
		resetUnresolvedUriWarnings(t)
		hook := captureVisionLog(t)
		model := &Model{Type: ModelTypeCaption, Name: "gpt-5-mini", Service: Service{RequestFormat: ApiFormatOpenAI}}

		prevConfig := Config
		t.Cleanup(func() { Config = prevConfig })
		Config = &ConfigValues{Models: Models{model}, Thresholds: DefaultThresholds}

		uri, method := model.Endpoint()
		assert.Equal(t, "", uri)
		assert.Equal(t, "", method)
		assert.True(t, model.hasService())

		for i := 0; i < 2; i++ {
			_, _, err := captionInternal(images, media.SrcLocal)
			assert.EqualError(t, err, "caption model gpt-5-mini needs a service uri or an engine for the openai request format")
		}

		for _, run := range []func() error{
			func() error { _, err := labelsInternal(images, media.SrcLocal, entity.SrcImage); return err },
			func() error { _, err := nsfwInternalContext(images, media.SrcLocal, nsfwThresholdIndex); return err },
		} {
			Config.Models = Models{&Model{Type: ModelTypeLabels, Name: "x", Service: Service{RequestFormat: ApiFormatOllama}},
				&Model{Type: ModelTypeNsfw, Name: "x", Service: Service{RequestFormat: ApiFormatOllama}}}
			assert.Error(t, run())
		}

		assert.Zero(t, requests.Load())
		Config.Models = Models{model}

		warnings := warnMessages(hook.AllEntries())
		require.GreaterOrEqual(t, len(warnings), 2)
		assert.Equal(t, "audit: vision › caption model gpt-5-mini needs a service uri or an engine for the openai request format, so no service is used", warnings[0])
		assert.Equal(t, warnings[0], warnings[1])
	})
	t.Run("VisionFormat", func(t *testing.T) {
		useSharedService(t, "https://vision.example.com/api/v1/vision", "shared-vision-key")
		resetUnresolvedUriWarnings(t)

		for _, format := range []ApiFormat{"", ApiFormatVision, ApiFormatImages, ApiFormatUrl} {
			model := &Model{Type: ModelTypeLabels, Name: "remote", Service: Service{RequestFormat: format}}
			uri, _ := model.Endpoint()
			assert.Equal(t, "https://vision.example.com/api/v1/vision/labels", uri, format)
			assert.NoError(t, model.unresolvedUriErr(), format)
		}
	})
	t.Run("Engine", func(t *testing.T) {
		useSharedService(t, "https://vision.example.com/api/v1/vision", "shared-vision-key")
		resetUnresolvedUriWarnings(t)
		model := &Model{Type: ModelTypeLabels, Name: "qwen3-vl:8b", Engine: "ollama"}
		model.ApplyEngineDefaults()
		uri, _ := model.Endpoint()
		assert.NotEmpty(t, uri)
		assert.NotContains(t, uri, "vision.example.com")
	})
}

// TestModel_UriMissingConsumers checks that legacy entries keep their local mapping and that a model without
// the Uri its request format needs is not sent the shared service key.
func TestModel_UriMissingConsumers(t *testing.T) {
	t.Run("IsLegacy", func(t *testing.T) {
		// Legacy entries keep their local model, whatever request format they name.
		for _, format := range []ApiFormat{"", ApiFormatOllama, ApiFormatOpenAI} {
			model := &Model{Type: ModelTypeLabels, Name: "nasnet", Service: Service{RequestFormat: format}}
			assert.True(t, model.IsLegacy(), format)
		}
		model := &Model{Type: ModelTypeNsfw, Name: "nsfw", TensorFlow: &tensorflow.ModelInfo{}, Service: Service{RequestFormat: ApiFormatOllama}}
		assert.True(t, model.IsLegacy())
	})
	t.Run("EndpointKey", func(t *testing.T) {
		useSharedService(t, "https://vision.example.com/api/v1/vision", "shared-vision-key")
		model := &Model{Type: ModelTypeCaption, Name: "gpt-5-mini", Service: Service{RequestFormat: ApiFormatOpenAI}}
		assert.Equal(t, "", model.EndpointKey())
		model.Service.Key = "own-key"
		uri, _ := model.Endpoint()
		assert.Equal(t, "", uri)
		assert.Equal(t, "", model.EndpointKey())
	})
	t.Run("EngineError", func(t *testing.T) {
		model := &Model{Type: ModelTypeCaption, Name: "gpt-5-mini", Engine: "vision", Service: Service{RequestFormat: ApiFormatOpenAI}}
		assert.EqualError(t, model.unresolvedUriErrText(), "caption model gpt-5-mini needs a service uri for the openai request format")
	})
}

// TestConfigValues_LoadLegacyUriMissing checks that a legacy entry with a request format that needs a Uri is
// replaced by the default model, so its own service settings are discarded.
func TestConfigValues_LoadLegacyUriMissing(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "vision.yml")
	require.NoError(t, os.WriteFile(fileName, []byte("Models:\n- Type: nsfw\n  Name: nsfw\n  Service:\n    RequestFormat: ollama\n    Key: own-key\n"), fs.ModeFile))

	config := NewConfig()
	require.NoError(t, config.Load(fileName))

	model := config.Model(ModelTypeNsfw)
	require.NotNil(t, model)
	assert.Equal(t, Service{}, model.Service)

	useSharedService(t, "", "")
	uri, _ := model.Endpoint()
	assert.Equal(t, "", uri)

	useSharedService(t, "https://vision.example.com/api/v1/vision", "shared-vision-key")
	assert.Equal(t, "shared-vision-key", model.EndpointKey())
}
