package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// unreadBody is a request body that records whether it was read.
type unreadBody struct {
	read bool
}

// Read marks the body as read and returns a short JSON body.
func (b *unreadBody) Read(p []byte) (int, error) {
	if b.read {
		return 0, io.EOF
	}

	b.read = true
	return copy(p, `{"id":"3487da77-246e-4b4d-b1b2-2b5d5ee7b5a8","images":[]}`), nil
}

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

// TestVisionApiDisabled checks that each Vision API endpoint answers a disabled API with exactly one JSON body,
// without reading the request.
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

			for _, contentType := range []string{header.ContentTypeJson, header.ContentTypeMultipart} {
				body := &unreadBody{}
				req := httptest.NewRequest(http.MethodPost, "/api/v1/vision/"+tc.name, body)
				req.Header.Set(header.ContentType, contentType)
				r := httptest.NewRecorder()
				app.ServeHTTP(r, req)

				assert.False(t, body.read, "request body was read")
				assert.Equal(t, http.StatusForbidden, r.Code, contentType)

				dec := json.NewDecoder(r.Body)
				var resp vision.ApiResponse
				require.NoError(t, dec.Decode(&resp))
				assert.True(t, rnd.IsUUID(resp.Id), resp.Id)
				assert.Equal(t, http.StatusForbidden, resp.Code)
				assert.Equal(t, "Forbidden", resp.Error)
				assert.Equal(t, io.EOF, dec.Decode(&json.RawMessage{}), "unexpected second body")
			}
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

// TestVisionAuth checks that service-key and session callers reach the Vision API in password auth mode.
func TestVisionAuth(t *testing.T) {
	conf := get.Config()
	origApi, origKey, origVisionApi, origAuthMode := vision.ServiceApi, vision.ServiceKey, conf.Options().VisionApi, conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	vision.ServiceApi = true
	vision.ServiceKey = "vision-service-key-abc123"
	t.Cleanup(func() {
		conf.SetAuthMode(origAuthMode)
		vision.ServiceApi, vision.ServiceKey, conf.Options().VisionApi = origApi, origKey, origVisionApi
	})

	req, err := vision.NewApiRequestImages(vision.Files{fs.Abs("./testdata/cat_224x224.jpg")}, scheme.Data, media.SrcLocal)
	require.NoError(t, err)
	body, err := req.JSON()
	require.NoError(t, err)

	post := func(token string) *httptest.ResponseRecorder {
		app, router, _ := NewApiTest()
		PostVisionLabels(router)
		return AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/vision/labels", string(body), token)
	}

	t.Run("ServiceKey", func(t *testing.T) {
		r := post(vision.ServiceKey)
		assert.Equal(t, http.StatusOK, r.Code)

		var resp vision.ApiResponse
		require.NoError(t, json.Unmarshal(r.Body.Bytes(), &resp))
		assert.Len(t, resp.Result.Labels, 1)
	})
	t.Run("WrongServiceKey", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, post("vision-service-key-abc124").Code)
		assert.Equal(t, http.StatusUnauthorized, post("vision-service-key-abc12").Code)
		assert.Equal(t, http.StatusUnauthorized, post("vision-service-key-abc1234").Code)
		assert.Equal(t, http.StatusUnauthorized, post("VISION-SERVICE-KEY-ABC123").Code)
	})
	t.Run("NoToken", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, post("").Code)
	})
	t.Run("EmptyServiceKey", func(t *testing.T) {
		vision.ServiceKey = ""
		t.Cleanup(func() { vision.ServiceKey = "vision-service-key-abc123" })
		assert.Equal(t, http.StatusUnauthorized, post("").Code)
	})
	t.Run("ServiceApiOff", func(t *testing.T) {
		vision.ServiceApi = false
		t.Cleanup(func() { vision.ServiceApi = true })
		assert.Equal(t, http.StatusUnauthorized, post(vision.ServiceKey).Code)
	})
	t.Run("Session", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PostVisionLabels(router)
		token := AuthenticateAdmin(app, router)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/vision/labels", string(body), token)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("Disabled", func(t *testing.T) {
		// Config.Propagate derives vision.ServiceApi from VisionApi().
		conf.Options().VisionApi = false
		vision.ServiceApi = false
		t.Cleanup(func() { conf.Options().VisionApi, vision.ServiceApi = origVisionApi, true })
		assert.Equal(t, http.StatusUnauthorized, post(vision.ServiceKey).Code)
		assert.Equal(t, http.StatusUnauthorized, post("vision-service-key-abc124").Code)

		app, router, _ := NewApiTest()
		PostVisionLabels(router)
		token := AuthenticateAdmin(app, router)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/vision/labels", string(body), token)
		assert.Equal(t, http.StatusForbidden, r.Code)

		var resp vision.ApiResponse
		require.NoError(t, json.Unmarshal(r.Body.Bytes(), &resp))
		assert.True(t, rnd.IsUUID(resp.Id), resp.Id)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	})
}

// TestVisionMultipart checks that each Vision API endpoint refuses a multipart request without reading it.
func TestVisionMultipart(t *testing.T) {
	conf := get.Config()
	orig := conf.Options().VisionApi
	conf.Options().VisionApi = true
	t.Cleanup(func() { conf.Options().VisionApi = orig })

	for _, tc := range []struct {
		name     string
		register func(*gin.RouterGroup)
	}{
		{"labels", PostVisionLabels},
		{"caption", PostVisionCaption},
		{"face", PostVisionFace},
		{"nsfw", PostVisionNsfw},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app, router, _ := NewApiTest()
			tc.register(router)

			body := &unreadBody{}
			req := httptest.NewRequest(http.MethodPost, "/api/v1/vision/"+tc.name, body)
			req.Header.Set(header.ContentType, header.ContentTypeMultipart)
			r := httptest.NewRecorder()
			app.ServeHTTP(r, req)

			assert.False(t, body.read, "request body was read")
			assert.Equal(t, http.StatusBadRequest, r.Code)

			var resp vision.ApiResponse
			require.NoError(t, json.Unmarshal(r.Body.Bytes(), &resp))
			assert.Equal(t, http.StatusBadRequest, resp.Code)
			assert.Equal(t, "Bad Request", resp.Error)
		})
	}
}

// propagateVision applies the specified vision options through Config.Propagate and restores them on cleanup.
func propagateVision(t *testing.T, conf *config.Config, enabled bool, uri, key string) {
	t.Helper()

	opt := conf.Options()
	origApi, origUri, origKey := opt.VisionApi, opt.VisionUri, opt.VisionKey
	t.Cleanup(func() {
		opt.VisionApi, opt.VisionUri, opt.VisionKey = origApi, origUri, origKey
		conf.Propagate()
	})

	opt.VisionApi, opt.VisionUri, opt.VisionKey = enabled, uri, key
	conf.Propagate()
}

// TestVisionServiceKeyDisabled checks that only an instance that serves the Vision API accepts the service key.
func TestVisionServiceKeyDisabled(t *testing.T) {
	const serviceKey = "vision-service-key-abc123"

	conf := get.Config()
	origAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() {
		conf.SetAuthMode(origAuthMode)
		conf.Propagate()
	})

	// Returns the response to a labels request with the specified token and the audit messages it logged.
	post := func(t *testing.T, token string) (*httptest.ResponseRecorder, []string) {
		t.Helper()

		orig := event.AuditLog
		logger, hook := logtest.NewNullLogger()
		logger.SetLevel(logrus.TraceLevel)
		event.AuditLog = logger
		t.Cleanup(func() { event.AuditLog = orig })

		app, router, _ := NewApiTest()
		PostVisionLabels(router)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/vision/labels", `{}`, token)

		var messages []string

		for _, entry := range hook.AllEntries() {
			messages = append(messages, entry.Message)
		}

		return r, messages
	}

	// Checks that the service key is refused exactly like a wrong key.
	refused := func(t *testing.T) {
		t.Helper()

		r, audit := post(t, serviceKey)
		wrong, wrongAudit := post(t, "vision-service-key-abc124")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
		assert.Equal(t, wrong.Code, r.Code)
		assert.Equal(t, wrong.Body.String(), r.Body.String())
		assert.Equal(t, wrongAudit, audit)
		assert.NotEmpty(t, audit)

		for _, msg := range audit {
			assert.NotContains(t, msg, status.Granted, msg)
		}
	}

	t.Run("Enabled", func(t *testing.T) {
		propagateVision(t, conf, true, "", serviceKey)

		r, audit := post(t, serviceKey)
		assert.NotEqual(t, http.StatusUnauthorized, r.Code)
		assert.True(t, slices.ContainsFunc(audit, func(msg string) bool { return strings.Contains(msg, status.Granted) }), audit)
	})
	t.Run("Disabled", func(t *testing.T) {
		propagateVision(t, conf, false, "", serviceKey)
		refused(t)
	})
	t.Run("ClientOnly", func(t *testing.T) {
		propagateVision(t, conf, false, "https://vision.example.com/api/v1/vision", serviceKey)
		refused(t)
	})
}
