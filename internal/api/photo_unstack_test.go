package api

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestPhotoUnstack(t *testing.T) {
	t.Run("UnstackXmpSidecarFile", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/files/fs6sg6bw45bnlqdw/unstack")
		// Sidecar files can not be unstacked.
		assert.Equal(t, http.StatusBadRequest, r.Code)
		// t.Logf("RESP: %s", r.Body.String())
	})
	t.Run("UnstackBridge3Jpg", func(t *testing.T) {
		app, router, c := NewApiTest()
		PhotoUnstack(router)
		require.NoError(t, fs.Copy("./testdata/london_160x160.jpg", filepath.Join(c.Options().OriginalsPath, "London", "bridge3.jpg"), true))
		require.NoError(t, fs.Copy("./testdata/face_160x160.jpg", filepath.Join(c.Options().OriginalsPath, "1990", "04", "bridge2.jpg"), true))
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/files/fs6sg6bwhhbnlqdn/unstack")
		assert.Equal(t, http.StatusOK, r.Code)
		// t.Logf("RESP: %s", r.Body.String())
	})
	t.Run("NotExistingFile", func(t *testing.T) {
		app, router, _ := NewApiTest()
		PhotoUnstack(router)
		r := PerformRequest(app, "POST", "/api/v1/photos/ps6sg6be2lvl0yh7/files/xxx/unstack")
		assert.Equal(t, http.StatusNotFound, r.Code)
		// t.Logf("RESP: %s", r.Body.String())
	})
}
