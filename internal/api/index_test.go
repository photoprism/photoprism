package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
)

func TestGetIndexingOptions(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetIndexingOptions(router, conf)
		r := PerformRequest(app, "GET", "/api/v1/index")
		assert.True(t, gjson.Get(r.Body.String(), "dirs").IsArray())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestCancelIndex(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, conf := NewApiTest()
		CancelIndexing(router, conf)
		r := PerformRequest(app, "DELETE", "/api/v1/index")
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, "indexing canceled", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
