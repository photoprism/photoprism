package api

import (
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestUploadToService(t *testing.T) {
	t.Run("invalid request", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UploadToService(router)
		r := PerformRequest(app, "POST", "/api/v1/services/1000000/upload")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrBadRequest), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("account not found", func(t *testing.T) {
		app, router, _ := NewApiTest()
		UploadToService(router)
		r := PerformRequest(app, "POST", "/api/v1/services/999000/upload")
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrAccountNotFound), val.String())
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
