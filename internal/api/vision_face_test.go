package api

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/service/http/scheme"
)

func TestPostVisionFace(t *testing.T) {
	t.Run("Face", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionFace(router)

		files := vision.Files{
			fs.Abs("./testdata/face_160x160.jpg"),
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

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/face", string(jsonReq))

		apiResponse := &vision.ApiResponse{}

		if apiJson, apiErr := io.ReadAll(r.Body); apiErr != nil {
			t.Fatal(apiErr)
		} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
			t.Fatal(apiErr)
		}

		// t.Logf("response: %#v", apiResponse)

		assert.Len(t, apiResponse.Result.Embeddings, 1)

		if len(apiResponse.Result.Embeddings) != 1 {
			t.Fatal("one nsfw result expected")
		}

		assert.Equal(t, vision.ModelTypeFace, apiResponse.Model.Type)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("London", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionFace(router)

		files := vision.Files{
			fs.Abs("./testdata/london_160x160.jpg"),
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

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/face", string(jsonReq))

		apiResponse := &vision.ApiResponse{}

		if apiJson, apiErr := io.ReadAll(r.Body); apiErr != nil {
			t.Fatal(apiErr)
		} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
			t.Fatal(apiErr)
		}

		assert.Len(t, apiResponse.Result.Embeddings, 1)
		assert.Equal(t, vision.ModelTypeFace, apiResponse.Model.Type)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("WrongResolution", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionFace(router)

		files := vision.Files{
			fs.Abs("./testdata/face_320x320.jpg"),
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

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/face", string(jsonReq))

		apiResponse := &vision.ApiResponse{}

		if apiJson, apiErr := io.ReadAll(r.Body); apiErr != nil {
			t.Fatal(apiErr)
		} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
			t.Fatal(apiErr)
		}

		// t.Logf("response: %#v", apiResponse)

		assert.Len(t, apiResponse.Result.Embeddings, 1)

		if len(apiResponse.Result.Embeddings) != 1 {
			t.Fatal("one nsfw result expected")
		}

		assert.Equal(t, vision.ModelTypeFace, apiResponse.Model.Type)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NoImages", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionFace(router)

		files := vision.Files{}

		req, err := vision.NewApiRequestImages(files, scheme.Data)

		if err != nil {
			t.Fatal(err)
		}

		jsonReq, jsonErr := req.JSON()

		if jsonErr != nil {
			t.Fatal(err)
		}

		// t.Logf("request: %s", string(jsonReq))

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/face", string(jsonReq))

		apiResponse := &vision.ApiResponse{}

		if apiJson, apiErr := io.ReadAll(r.Body); apiErr != nil {
			t.Fatal(apiErr)
		} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
			t.Fatal(apiErr)
		}

		if apiResponse == nil {
			t.Fatal("api response expected")
		}

		// t.Logf("error: %s", apiResponse.Err())

		assert.Error(t, apiResponse.Err())
		assert.False(t, apiResponse.HasResult())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NoBody", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionFace(router)
		r := PerformRequest(app, http.MethodPost, "/api/v1/vision/face")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
