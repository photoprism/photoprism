package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

func TestUpdateUser(t *testing.T) {
	t.Run("Alice", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		UpdateUser(router)
		sessId := AuthenticateUser(app, router, "alice", "Alice123!")
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s", adminUid)
		t.Logf("Request URL: %s", reqUrl)
		r := AuthenticatedRequestWithBody(app, "PUT", reqUrl, "{Email:\"admin@example.com\",Details:{Location:\"WebStorm\"}}", sessId)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("Forbidden", func(t *testing.T) {
		app, router, _ := NewApiTest()
		adminUid := entity.Admin.UserUID
		reqUrl := fmt.Sprintf("/api/v1/users/%s", adminUid)
		UpdateUser(router)
		r := PerformRequestWithBody(app, "PUT", reqUrl, "{foo:123}")
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
