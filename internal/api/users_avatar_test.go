package api

import (
	"fmt"
	"github.com/photoprism/photoprism/internal/config"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestUploadUserAvatar(t *testing.T) {
	t.Run("InvalidRequestBody", func(t *testing.T) {
		app, router, _ := NewApiTest()
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/avatar", adminUid)
		UploadUserAvatar(router)
		r := PerformRequestWithBody(app, "POST", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("SettingsDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().DisableSettings = true

		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s/avatar", adminUid)
		UploadUserAvatar(router)

		r := PerformRequestWithBody(app, "POST", reqUrl, "{}")
		assert.Equal(t, http.StatusForbidden, r.Code)
		conf.Options().DisableSettings = false
	})
	t.Run("bobCannotChangeAlice", func(t *testing.T) {
		app, router, conf := NewApiTest()
		adminUid := entity.Admin.UserUID

		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		reqUrl := fmt.Sprintf("/api/v1/users/%s/avatar", adminUid)
		UploadUserAvatar(router)

		authToken := AuthenticateUser(app, router, "bob", "Bobbob123!")

		r := AuthenticatedRequestWithBody(app, http.MethodPost, reqUrl, `{}`, authToken)

		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
