package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
)

func TestPostVisionFace(t *testing.T) {
	t.Run("GenerateFaceEmbeddings", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionFace(router)

		files := vision.Files{
			fs.Abs("./testdata/face_160x160.jpg"),
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

		req, err := vision.NewApiRequestImages(files, scheme.Data, media.SrcLocal)

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

		req, err := vision.NewApiRequestImages(files, scheme.Data, media.SrcLocal)

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

		req, err := vision.NewApiRequestImages(files, scheme.Data, media.SrcLocal)

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

		// t.Logf("error: %s", apiResponse.Err())

		assert.Error(t, apiResponse.Err())
		assert.False(t, apiResponse.HasResult())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("InvalidReference", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionFace(router)

		// A raw local path is not an https/data URL and must be rejected with 400,
		// consistent with the labels endpoint, rather than a 200 with empty embeddings.
		body := `{"images":["/photoprism/originals/pp-july/peach_pi-1280x720.jpg"]}`
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/face", body)

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NotAnImage", func(t *testing.T) {
		// A well-formed data URL that does not contain a decodable image must be rejected
		// with 400 like the labels and nsfw endpoints, not answered with empty embeddings.
		refs := map[string]string{
			"PlainText": "data:text/plain;base64," + media.EncodeBase64String([]byte("not an image")),
			"Svg":       "data:image/svg+xml;base64," + media.EncodeBase64String([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect width="10" height="10"/></svg>`)),
			"Html":      "data:text/html;base64," + media.EncodeBase64String([]byte("<html><body>hi</body></html>")),
		}
		for name, ref := range refs {
			t.Run(name, func(t *testing.T) {
				app, router, _ := NewApiTest()
				PostVisionFace(router)
				r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/face", `{"images":["`+ref+`"]}`)
				assert.Equal(t, http.StatusBadRequest, r.Code)
			})
		}
	})
	t.Run("NoBody", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionFace(router)
		r := PerformRequest(app, http.MethodPost, "/api/v1/vision/face")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionFace(router)

		body := `{"images":["data:image/jpeg;base64,` + strings.Repeat("a", int(MaxVisionRequestBytes)) + `"]}`
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/face", body)

		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
}
