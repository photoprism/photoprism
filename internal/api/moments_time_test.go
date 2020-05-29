package api

import (
	"github.com/tidwall/gjson"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMomentsTime(t *testing.T) {
	t.Run("get moments time", func(t *testing.T) {
		app, router, conf := NewApiTest()

		GetMomentsTime(router, conf)

		r := PerformRequest(app, "GET", "/api/v1/moments/time")
		val := gjson.Get(r.Body.String(), `#(Year=="2790").Count`)
		assert.LessOrEqual(t, val.Int(), int64(2))
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
