package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestSearchServices(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchServices(router)
		sess := AuthenticateAdmin(app, router)
		r := AuthenticatedRequest(app, "GET", "/api/v1/services?count=10", sess)
		val := gjson.Get(r.Body.String(), "#(AccName=\"Test Account\").AccURL")
		count := gjson.Get(r.Body.String(), "#")
		assert.LessOrEqual(t, int64(1), count.Int())
		assert.Equal(t, "http://dummy-webdav/", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchServices(router)
		r := PerformRequest(app, "GET", "/api/v1/services?xxx=10")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
