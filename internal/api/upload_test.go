package api

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestUpload(t *testing.T) {
	t.Run("forbidden", func(t *testing.T) {
		app, router, _ := NewApiTest()
		Upload(router)
		r := PerformRequest(app, "POST", "/api/v1/upload/xxx")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
