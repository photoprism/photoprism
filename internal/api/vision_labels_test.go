package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/media"
)

func TestPostVisionLabels(t *testing.T) {
	t.Run("OneImage", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionLabels(router)

		files := vision.Files{
			fs.Abs("./testdata/cat_224x224.jpg"),
		}

		req, err := vision.NewApiRequestImages(files, scheme.Data, media.SrcLocal)

		if err != nil {
			t.Fatal(err)
		}

		jsonReq, jsonErr := req.JSON()

		if jsonErr != nil {
			t.Fatal(err)
		}

		// t.Logf("request: %s", string(jsonReq))

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/labels", string(jsonReq))

		apiResponse := &vision.ApiResponse{}

		if apiJson, apiErr := io.ReadAll(r.Body); apiErr != nil {
			t.Fatal(apiErr)
		} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
			t.Fatal(apiErr)
		}

		assert.Len(t, apiResponse.Result.Labels, 1)
		assert.Equal(t, vision.ModelTypeLabels, apiResponse.Model.Type)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("TwoImages", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionLabels(router)

		files := vision.Files{
			fs.Abs("./testdata/cat_224x224.jpg"),
			fs.Abs("./testdata/green_224x224.jpg"),
		}

		req, err := vision.NewApiRequestImages(files, scheme.Data, media.SrcLocal)

		if err != nil {
			t.Fatal(err)
		}

		jsonReq, jsonErr := req.JSON()

		if jsonErr != nil {
			t.Fatal(err)
		}

		// t.Logf("request: %s", string(jsonReq))

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/labels", string(jsonReq))

		apiResponse := &vision.ApiResponse{}

		if apiJson, apiErr := io.ReadAll(r.Body); apiErr != nil {
			t.Fatal(apiErr)
		} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
			t.Fatal(apiErr)
		}

		assert.Len(t, apiResponse.Result.Labels, 2)
		assert.Equal(t, vision.ModelTypeLabels, apiResponse.Model.Type)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NoImages", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionLabels(router)

		files := vision.Files{}

		req, err := vision.NewApiRequestImages(files, scheme.Data, media.SrcLocal)

		if err != nil {
			t.Fatal(err)
		}

		jsonReq, jsonErr := req.JSON()

		if jsonErr != nil {
			t.Fatal(err)
		}

		t.Logf("request: %s", string(jsonReq))

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/labels", string(jsonReq))

		apiResponse := &vision.ApiResponse{}

		if apiJson, apiErr := io.ReadAll(r.Body); apiErr != nil {
			t.Fatal(apiErr)
		} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
			t.Fatal(apiErr)
		}

		t.Logf("error: %s", apiResponse.Err())

		assert.Error(t, apiResponse.Err())
		assert.False(t, apiResponse.HasResult())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NoBody", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionLabels(router)
		r := PerformRequest(app, http.MethodPost, "/api/v1/vision/labels")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionLabels(router)

		body := `{"images":["data:image/jpeg;base64,` + strings.Repeat("a", int(MaxVisionRequestBytes)) + `"]}`
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/labels", body)

		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
}

// TestVisionApiDisabled checks that each Vision API endpoint answers a disabled API with exactly one JSON body.
func TestVisionApiDisabled(t *testing.T) {
	conf := get.Config()
	orig := conf.Options().VisionApi
	conf.Options().VisionApi = false
	t.Cleanup(func() { conf.Options().VisionApi = orig })

	endpoints := []struct {
		name     string
		register func(*gin.RouterGroup)
	}{
		{"labels", PostVisionLabels},
		{"caption", PostVisionCaption},
		{"face", PostVisionFace},
		{"nsfw", PostVisionNsfw},
	}

	for _, tc := range endpoints {
		t.Run(tc.name, func(t *testing.T) {
			app, router, _ := NewApiTest()
			tc.register(router)

			r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/"+tc.name, `{"id":"3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8"}`)
			assert.Equal(t, http.StatusForbidden, r.Code)

			dec := json.NewDecoder(r.Body)
			var resp vision.ApiResponse
			require.NoError(t, dec.Decode(&resp))
			assert.Equal(t, "3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8", resp.Id)
			assert.Equal(t, http.StatusForbidden, resp.Code)
			assert.Equal(t, "Forbidden", resp.Error)
			assert.Equal(t, io.EOF, dec.Decode(&json.RawMessage{}), "unexpected second body")
		})
	}
	t.Run("Unauthorized", func(t *testing.T) {
		// A request without permission is refused before the API is checked, with its own single response.
		for _, tc := range endpoints {
			app, router, conf := NewApiTest()
			conf.SetAuthMode(config.AuthModePasswd)
			tc.register(router)

			r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/"+tc.name, `{}`)
			conf.SetAuthMode(config.AuthModePublic)
			assert.Equal(t, http.StatusUnauthorized, r.Code, tc.name)

			dec := json.NewDecoder(r.Body)
			var resp i18n.Response
			require.NoError(t, dec.Decode(&resp), tc.name)
			assert.Equal(t, http.StatusUnauthorized, resp.Code, tc.name)
			assert.Equal(t, io.EOF, dec.Decode(&json.RawMessage{}), tc.name)
		}
	})
}
