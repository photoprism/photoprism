package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUpdateCamera(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		defer func() {
			entity.FlushCameraCache()
			assert.NoError(t, entity.UnscopedDb().Save(entity.CameraFixtures.Pointer("canon-eos-7d")).Error)
		}()
		app, router, _ := NewApiTest()
		UpdateCamera(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/cameras/1000002", `{"Make": "Pentax", "Model": "K-1"}`)
		val := gjson.Get(r.Body.String(), "Name")
		assert.Equal(t, "PENTAX K-1", val.String())
		val2 := gjson.Get(r.Body.String(), "Model")
		assert.Equal(t, "K-1", val2.String())
		val3 := gjson.Get(r.Body.String(), "Make")
		assert.Equal(t, "PENTAX", val3.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateCamera(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/cameras/1000002", `{"Make": 123, "Model": ""}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Unable to do that", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("BadModel", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateCamera(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/cameras/1000002", `{"Make": "123", "Model": ""}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Invalid name", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateCamera(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/cameras/199000002", `{"Make": "Pentax", "Model": "K-1"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Camera not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
