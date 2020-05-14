package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
)

func TestGetImportOptions(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetImportOptions(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/import")
		assert.True(t, gjson.Get(r.Body.String(), "dirs").IsArray())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestCancelImport(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CancelImport(router, conf)
		r := PerformRequest(app, "DELETE", "/api/v1/import")
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, "import canceled", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
