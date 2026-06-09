package api

import (
	"net/http"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
)

func TestGetFile(t *testing.T) {
	t.Run("SearchForExistingFile", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFile(router)
		r := PerformRequest(app, "GET", "/api/v1/files/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818")
		assert.Equal(t, http.StatusOK, r.Code)

		val := gjson.Get(r.Body.String(), "Name")
		assert.Equal(t, "2790/07/27900704_070228_D6D51B6C.jpg", val.String())
	})
	t.Run("SearchForNotExistingFile", func(t *testing.T) {
		app, router, _ := NewApiTest()
		GetFile(router)
		r := PerformRequest(app, "GET", "/api/v1/files/111")
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("SharedOnlySessionDenied", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetFile(router)

		// A shared-only (guest) session has no files access, so the endpoint denies the request
		// at the ACL check before the visibility gate. Per-photo file scope (the gate that returns
		// 404 for an out-of-scope file when the role does have files access, e.g. a viewer in
		// Plus/Pro) is covered by search.TestFileVisibleToSession.
		authToken := AuthenticateUser(app, router, "gandalf", "Gandalf123!")
		r := AuthenticatedRequest(app, "GET", "/api/v1/files/2cad9168fa6acc5c5c2965ddf6ec465ca42fd818", authToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
}
