package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"net/http"
	"testing"
)

func TestCancelIndex(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		CancelIndexing(router)
		r := PerformRequest(app, "DELETE", "/api/v1/index")
		val := gjson.Get(r.Body.String(), "message")
		assert.Equal(t, "indexing canceled", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
