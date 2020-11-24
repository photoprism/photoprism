package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"
)

func TestGetMomentsTime(t *testing.T) {
	t.Run("get moments time", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetMomentsTime(router)

		r := PerformRequest(app, "GET", "/api/v1/moments/time")
		val := gjson.Get(r.Body.String(), `#(Year=="2790").Count`)
		assert.LessOrEqual(t, val.Int(), int64(2))
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
