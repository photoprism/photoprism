package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUploadUserAvatar(t *testing.T) {
	t.Run("Forbidden", func(t *testing.T) {
		app, router, _ := NewApiTest()
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/avatar", adminUid)
		UploadUserAvatar(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
