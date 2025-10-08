package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteFile(t *testing.T) {
	t.Run("DeleteNotExistingFile", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/5678/files/23456hbg")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("DeletePrimaryFile", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh7/files/fs6sg6bw45bnlqdw")
		assert.Equal(t, http.StatusInternalServerError, r.Code)
	})
	t.Run("TryToDeleteFile", func(t *testing.T) {
		app, router, _ := NewApiTest()

		DeleteFile(router)

		r := PerformRequest(app, "DELETE", "/api/v1/photos/ps6sg6be2lvl0yh8/files/fs6sg6bw45bn0001")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
}
