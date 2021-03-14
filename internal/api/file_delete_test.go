package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteFile(t *testing.T) {
	t.Run("delete not existing file", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/5678/files/23456hbg")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("delete primary file", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/pt9jtdre2lvl0yh7/files/ft8es39w45bnlqdw")
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})
	t.Run("try to delete file", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/pt9jtdre2lvl0yh8/files/ft9es39w45bnlqdw")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
