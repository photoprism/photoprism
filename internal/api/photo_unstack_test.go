package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPhotoUnstack(t *testing.T) {
	t.Run("unstack xmp sidecar file", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/files/fs6sg6bw45bnlqdw/unstack")
		// Sidecar files can not be unstacked.
		assert.Equal(t, http.StatusBadRequest, r.Code)
		// t.Logf("RESP: %s", r.Body.String())
	})

	t.Run("unstack bridge3.jpg", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/files/fs6sg6bwhhbnlqdn/unstack")
		// TODO: Have a real file in place for testing the success case. This file does not exist, so it cannot be unstacked.
		assert.Equal(t, http.StatusNotFound, r.Code)
		// t.Logf("RESP: %s", r.Body.String())
	})

	t.Run("not existing file", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/files/xxx/unstack")
		assert.Equal(t, http.StatusNotFound, r.Code)
		// t.Logf("RESP: %s", r.Body.String())
	})
}
