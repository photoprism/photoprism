package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetErrors(t *testing.T) {
	//TODO add error fixtures
	t.Run("successful request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetErrors(router)
		r := PerformRequest(app, "GET", "/api/v1/errors")
		//val := gjson.Get(r.Body.String(), "flags")
		//assert.Equal(t, "public debug settings", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}
