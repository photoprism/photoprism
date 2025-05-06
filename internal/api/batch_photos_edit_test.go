package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/i18n"
)

func TestBatchPhotosEdit(t *testing.T) {
	t.Run("ReturnValues", func(t *testing.T) {
		app, router, _ := NewApiTest()

		BatchPhotosEdit(router)

		response := PerformRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			`{"photos": ["ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"]}`,
		)

		body := response.Body.String()

		assert.NotEmpty(t, body)
		assert.True(t, strings.HasPrefix(body, `{"values":{"`), "unexpected response")

		// fmt.Println(body)
		/* photos := gjson.Get(body, "photos")
		values := gjson.Get(body, "values")
		t.Logf("photos: %#v", photos)
		t.Logf("values: %#v", values) */

		assert.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("ReturnPhotosAndValues", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		authToken := AuthenticateUser(app, router, "alice", "Alice123!")

		BatchPhotosEdit(router)

		response := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/batch/photos/edit",
			`{"photos": ["ps6sg6be2lvl0yh7","ps6sg6be2lvl0yh8","ps6sg6byk7wrbk47","ps6sg6be2lvl0yh0"], "return": true, "values": {}}`,
			authToken)

		body := response.Body.String()

		assert.NotEmpty(t, body)
		assert.True(t, strings.HasPrefix(body, `{"photos":[{"ID"`), "unexpected response")

		fmt.Println(body)
		/* photos := gjson.Get(body, "photos")
		values := gjson.Get(body, "values")
		t.Logf("photos: %#v", photos)
		t.Logf("values: %#v", values) */

		assert.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("MissingSelection", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosEdit(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/edit", `{"photos": [], "return": true}`)
		val := gjson.Get(r.Body.String(), "error")
		assert.Equal(t, i18n.Msg(i18n.ErrNoItemsSelected), val.String())
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()
		BatchPhotosEdit(router)
		r := PerformRequestWithBody(app, "POST", "/api/v1/batch/photos/edit", `{"photos": 123, "return": true}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("ReturnValuesAsAdmin", func(t *testing.T) {
		app, router, conf := NewApiTest()

		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		BatchPhotosEdit(router)

		sessId := AuthenticateUser(app, router, "alice", "Alice123!")

		response := AuthenticatedRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			`{"photos": ["ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"]}`,
			sessId,
		)

		body := response.Body.String()

		assert.NotEmpty(t, body)
		assert.True(t, strings.HasPrefix(body, `{"values":{"`), "unexpected response")

		assert.Equal(t, http.StatusOK, response.Code)
	})
	t.Run("ReturnValuesAsGuest", func(t *testing.T) {
		app, router, conf := NewApiTest()

		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		BatchPhotosEdit(router)

		sessId := AuthenticateUser(app, router, "gandalf", "Gandalf123!")

		response := AuthenticatedRequestWithBody(app,
			"POST", "/api/v1/batch/photos/edit",
			`{"photos": ["ps6sg6be2lvl0yh7", "ps6sg6be2lvl0yh8"]}`,
			sessId,
		)

		if response.Code != http.StatusForbidden {
			t.Fatal(response.Body.String())
		}

		val := gjson.Get(response.Body.String(), "error")
		assert.Equal(t, "Permission denied", val.String())
	})
}
