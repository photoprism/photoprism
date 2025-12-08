package api

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

func TestPostVisionLabels(t *testing.T) {
	t.Run("OneImage", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionLabels(router)

		files := vision.Files{
			fs.Abs("./testdata/cat_224x224.jpg"),
		}

		req, err := vision.NewApiRequestImages(files, scheme.Data)

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

		req, err := vision.NewApiRequestImages(files, scheme.Data)

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

		req, err := vision.NewApiRequestImages(files, scheme.Data)

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
}
