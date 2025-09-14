package vision

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/service/http/scheme"
)

func TestModel(t *testing.T) {
	t.Run("Nasnet", func(t *testing.T) {
		ServiceUri = "https://app.localssl.dev/api/v1/vision"
		uri, method := NasnetModel.Endpoint()
		ServiceUri = ""
		assert.Equal(t, "https://app.localssl.dev/api/v1/vision/labels", uri)
		assert.Equal(t, http.MethodPost, method)

		uri, method = NasnetModel.Endpoint()
		assert.Equal(t, "", uri)
		assert.Equal(t, "", method)
	})
	t.Run("Caption", func(t *testing.T) {
		uri, method := CaptionModel.Endpoint()
		assert.Equal(t, "http://ollama:11434/api/generate", uri)
		assert.Equal(t, "POST", method)

		model, name, version := CaptionModel.Model()
		assert.Equal(t, "gemma3:latest", model)
		assert.Equal(t, "gemma3", name)
		assert.Equal(t, "latest", version)
	})
	t.Run("ParseName", func(t *testing.T) {
		m := &Model{
			Type:       ModelTypeCaption,
			Name:       "deepseek-r1:1.5b",
			Version:    "",
			Resolution: 720,
			Prompt:     CaptionPromptDefault,
			Service: Service{
				Uri:            "http://foo:bar@photoprism-vision:5000/api/v1/vision/caption",
				FileScheme:     scheme.Data,
				RequestFormat:  ApiFormatVision,
				ResponseFormat: ApiFormatVision,
			},
		}

		uri, method := m.Endpoint()
		assert.Equal(t, "http://foo:bar@photoprism-vision:5000/api/v1/vision/caption", uri)
		assert.Equal(t, "POST", method)

		model, name, version := m.Model()
		assert.Equal(t, "deepseek-r1:1.5b", model)
		assert.Equal(t, "deepseek-r1", name)
		assert.Equal(t, "1.5b", version)
	})
}

func TestParseTypes(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		result := ParseTypes("nsfw, labels, Caption")
		assert.Equal(t, ModelTypes{"nsfw", "labels", "caption"}, result)
	})
	t.Run("None", func(t *testing.T) {
		result := ParseTypes("")
		assert.Equal(t, ModelTypes{}, result)
	})
	t.Run("Invalid", func(t *testing.T) {
		result := ParseTypes("foo, captions")
		assert.Equal(t, ModelTypes{}, result)
	})
}
