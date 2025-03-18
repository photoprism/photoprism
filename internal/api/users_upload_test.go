package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUploadUserFiles(t *testing.T) {
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/upload/abc123456789", adminUid)
		// t.Logf("Request URL: %s", reqUrl)
		UploadUserFiles(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("ReadOnlyMode", func(t *testing.T) {
		app, router, config := NewApiTest()
		config.Options().ReadOnly = true
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/upload/abc123456789", adminUid)
		// t.Logf("Request URL: %s", reqUrl)
		UploadUserFiles(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusForbidden, r.Code)
		config.Options().ReadOnly = false
	})
	t.Run("QuotaExceeded", func(t *testing.T) {
		app, router, config := NewApiTest()
		config.Options().FilesQuota = 1
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/upload/abc123456789", adminUid)
		// t.Logf("Request URL: %s", reqUrl)
		UploadUserFiles(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusInsufficientStorage, r.Code)
		config.Options().FilesQuota = 0
	})
}
