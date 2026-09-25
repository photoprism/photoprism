package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUpdateLens(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateLens(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/lenses/1000002", `{"Make": "Tamron", "Model": "Tamron SP AF 24-135mm F3.5-5.6 AD AL (190D)"}`)
		val := gjson.Get(r.Body.String(), "Name")
		assert.Equal(t, "Tamron SP AF 24-135mm F3.5-5.6 AD AL (190D)", val.String())
		val2 := gjson.Get(r.Body.String(), "Model")
		assert.Equal(t, "SP AF 24-135mm F3.5-5.6 AD AL (190D)", val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateLens(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/lenses/1000002", `{"Make": 123, "Model": ""}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Unable to do that", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("BadModel", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateLens(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/lenses/1000002", `{"Make": "123", "Model": ""}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Invalid name", val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateLens(router)
		r := PerformRequestWithBody(app, "PUT", "/api/v1/lenses/199000002", `{"Make": "Pentax", "Model": "Tamron SP AF 24-135mm F3.5-5.6 AD AL (190D)"}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, "Lens not found", val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("UnknownLens", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UpdateLens(router)
		r := PerformRequestWithBody(app, "PUT", fmt.Sprintf("/api/v1/lenses/%d", entity.UnknownLens.ID), `{"Make": "Example", "Model": "Example"}`)
		assert.Equal(t, http.StatusForbidden, r.Code)

		// The shared placeholder must keep its name.
		found := entity.Lens{}
		assert.NoError(t, entity.Db().First(&found, "id = ?", entity.UnknownLens.ID).Error)
		assert.Equal(t, entity.UnknownLens.LensName, found.LensName)
	})
}
