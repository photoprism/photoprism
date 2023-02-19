package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestGetSettings(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetSettings(router)
		r := PerformRequest(app, "GET", "/api/v1/settings")
		val := gjson.Get(r.Body.String(), "ui.theme")
		assert.NotEmpty(t, val.String())
		val2 := gjson.Get(r.Body.String(), "ui.language")
		assert.NotEmpty(t, val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestSaveSettings(t *testing.T) {
	t.Run("Ok", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetSettings(router)
		r := PerformRequest(app, "GET", "/api/v1/settings")
		val := gjson.Get(r.Body.String(), "ui.language")
		assert.Equal(t, "en", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
		SaveSettings(router)
		r2 := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"language": "de"}}`)
		assert.Equal(t, http.StatusOK, r2.Code)
		r4 := PerformRequest(app, "GET", "/api/v1/settings")
		val2 := gjson.Get(r4.Body.String(), "ui.language")
		assert.Equal(t, "de", val2.String())
		r3 := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"language": "en"}}`)
		assert.Equal(t, http.StatusOK, r3.Code)
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SaveSettings(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"language":123}}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
