package vision

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/service/http/scheme"
)

func TestNewApiRequest(t *testing.T) {
	t.Run("Data", func(t *testing.T) {
		thumbnails := Files{examplesPath + "/chameleon_lime.jpg"}
		result, err := NewApiRequestImages(thumbnails, scheme.Data)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// t.Logf("request: %#v", result)

		if result != nil {
			json, jsonErr := result.JSON()
			assert.NoError(t, jsonErr)
			assert.NotEmpty(t, json)
			// t.Logf("json: %s", json)
		}
	})
	t.Run("Https", func(t *testing.T) {
		thumbnails := Files{examplesPath + "/chameleon_lime.jpg"}
		result, err := NewApiRequestImages(thumbnails, scheme.Https)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// t.Logf("request: %#v", result)
		if result != nil {
			json, jsonErr := result.JSON()
			assert.NoError(t, jsonErr)
			assert.NotEmpty(t, json)
			t.Logf("json: %s", json)
		}
	})
}

func TestPerformApiRequestOllama(t *testing.T) {
	t.Run("Labels", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req ApiRequest
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			assert.Equal(t, FormatJSON, req.Format)
			assert.NoError(t, json.NewEncoder(w).Encode(ApiResponseOllama{
				Model:    "qwen2.5vl:latest",
				Response: `{"labels":[{"name":"test","confidence":0.9,"topicality":0.8}]}`,
			}))
		}))
		defer server.Close()

		apiRequest := &ApiRequest{
			Id:             "test",
			Model:          "qwen2.5vl:latest",
			Format:         FormatJSON,
			Images:         []string{"data:image/jpeg;base64,AA=="},
			ResponseFormat: ApiFormatOllama,
		}

		resp, err := PerformApiRequest(apiRequest, server.URL, http.MethodPost, "")
		assert.NoError(t, err)
		assert.Len(t, resp.Result.Labels, 1)
		assert.Equal(t, "test", resp.Result.Labels[0].Name)
		assert.Nil(t, resp.Result.Caption)
	})
	t.Run("CaptionFallback", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NoError(t, json.NewEncoder(w).Encode(ApiResponseOllama{
				Model:    "qwen2.5vl:latest",
				Response: "plain text",
			}))
		}))
		defer server.Close()

		apiRequest := &ApiRequest{
			Id:             "test2",
			Model:          "qwen2.5vl:latest",
			Format:         FormatJSON,
			Images:         []string{"data:image/jpeg;base64,AA=="},
			ResponseFormat: ApiFormatOllama,
		}

		resp, err := PerformApiRequest(apiRequest, server.URL, http.MethodPost, "")
		assert.NoError(t, err)
		assert.Len(t, resp.Result.Labels, 0)
		if assert.NotNil(t, resp.Result.Caption) {
			assert.Equal(t, "plain text", resp.Result.Caption.Text)
		}
	})
}
