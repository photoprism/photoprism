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

func TestPostVisionNsfw(t *testing.T) {
	t.Run("OneImage", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionNsfw(router)

		files := vision.Files{
			fs.Abs("./testdata/nsfw_224x224.jpg"),
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
		} else if nsfw := apiResponse.Result.Nsfw[0]; !nsfw.IsNsfw(0.6) {
			t.Fatalf("image should not be safe for work: %#v", nsfw)
		} else {
			// Drawing:7.547473e-05, Hentai:0.19912475, Neutral:0.00097554235, Porn:0.67095983, Sexy:0.12886441
			assert.InDelta(t, nsfw.Drawing, 0.01, 0.2)
			assert.InDelta(t, nsfw.Hentai, 0.2, 0.2)
			assert.InDelta(t, nsfw.Porn, 0.7, 0.2)
			assert.InDelta(t, nsfw.Sexy, 0.1, 0.2)
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

		req, err := vision.NewApiRequestImages(files, scheme.Data)

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

		req, err := vision.NewApiRequestImages(files, scheme.Data)

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
		PostVisionNsfw(router)
		r := PerformRequest(app, http.MethodPost, "/api/v1/vision/nsfw")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
