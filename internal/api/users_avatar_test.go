package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/header"
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
	t.Run("BobCannotChangeAlice", func(t *testing.T) {
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
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, conf := NewApiTest()
		adminUid := entity.Admin.UserUID

		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		reqUrl := fmt.Sprintf("/api/v1/users/%s/avatar", adminUid)
		UploadUserAvatar(router)
		authToken := AuthenticateAdmin(app, router)

		tooLarge := bytes.Repeat([]byte("A"), 22*1024*1024)
		body, ctype, err := buildMultipart(map[string][]byte{"avatar.jpg": tooLarge})
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, reqUrl, body)
		req.Header.Set("Content-Type", ctype)
		header.SetAuthorization(req, authToken)

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})
}
