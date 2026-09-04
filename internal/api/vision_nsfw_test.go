package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/media"
)

func TestPostVisionNsfw(t *testing.T) {
	t.Run("OneImage", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionNsfw(router)

		files := vision.Files{
			fs.Abs("./testdata/nsfw_224x224.jpg"),
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

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/nsfw", string(jsonReq))

		apiResponse := &vision.ApiResponse{}

		if apiJson, apiErr := io.ReadAll(r.Body); apiErr != nil {
			t.Fatal(apiErr)
		} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
			t.Fatal(apiErr)
		}

		// t.Logf("response: %#v", apiResponse)

		assert.Len(t, apiResponse.Result.Nsfw, 1)

		if len(apiResponse.Result.Nsfw) != 1 {
			t.Fatal("one nsfw result expected")
		} else if result := apiResponse.Result.Nsfw[0]; !result.Decide(0.6).IsUnsafe() {
			t.Fatalf("image should not be safe for work: %#v", result)
		} else {
			// Drawing:7.547473e-05, Hentai:0.19912475, Neutral:0.00097554235, Porn:0.67095983, Sexy:0.12886441
			assert.InDelta(t, result.Drawing, 0.01, 0.2)
			assert.InDelta(t, result.Hentai, 0.2, 0.2)
			assert.InDelta(t, result.Porn, 0.7, 0.2)
			assert.InDelta(t, result.Sexy, 0.1, 0.2)

			// The service decides against its own threshold and records which one, so a client
			// with a stricter policy can re-decide the same scores without asking again.
			assert.True(t, result.IsSafe())
			assert.InDelta(t, nsfw.DefaultThreshold, result.Threshold, 1e-6)
			assert.InDelta(t, result.Porn, result.Score, 1e-6)
		}

		assert.Equal(t, vision.ModelTypeNsfw, apiResponse.Model.Type)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("TwoImages", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionNsfw(router)

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

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/nsfw", string(jsonReq))

		apiResponse := &vision.ApiResponse{}

		if apiJson, apiErr := io.ReadAll(r.Body); apiErr != nil {
			t.Fatal(apiErr)
		} else if apiErr = json.Unmarshal(apiJson, apiResponse); apiErr != nil {
			t.Fatal(apiErr)
		}

		assert.Len(t, apiResponse.Result.Nsfw, 2)
		assert.Equal(t, vision.ModelTypeNsfw, apiResponse.Model.Type)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("NoImages", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionNsfw(router)

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

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/nsfw", string(jsonReq))

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
		PostVisionNsfw(router)

		// A raw local path is not an https/data URL and must be rejected with 400,
		// consistent with the labels endpoint, rather than a 200 with an empty result.
		body := `{"images":["/photoprism/originals/pp-july/peach_pi-1280x720.jpg"]}`
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/nsfw", body)

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NoBody", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionNsfw(router)
		r := PerformRequest(app, http.MethodPost, "/api/v1/vision/nsfw")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionNsfw(router)

		body := `{"images":["data:image/jpeg;base64,` + strings.Repeat("a", int(MaxVisionRequestBytes)) + `"]}`
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/vision/nsfw", body)

		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
}
