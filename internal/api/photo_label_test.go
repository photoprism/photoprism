package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestRemovePhotoLabel(t *testing.T) {
	t.Run("photo with label", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		RemovePhotoLabel(router, ctx)
		result := PerformRequest(app, "DELETE", "/api/v1/photos/pt9jtdre2lvl0yh7/label/1000001")
		assert.Equal(t, http.StatusOK, result.Code)
		val := gjson.Get(result.Body.String(), "Labels.#(LabelID==1000001).Uncertainty")
		assert.Equal(t, "100", val.String())
	})
	t.Run("try to remove wrong label", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		RemovePhotoLabel(router, ctx)
		result := PerformRequest(app, "DELETE", "/api/v1/photos/pt9jtdre2lvl0yh7/label/1000000")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
	t.Run("not existing photo", func(t *testing.T) {
		app, router, ctx := NewApiTest()
		RemovePhotoLabel(router, ctx)
		result := PerformRequest(app, "DELETE", "/api/v1/photos/xx/label/")
		assert.Equal(t, http.StatusNotFound, result.Code)
	})
}
