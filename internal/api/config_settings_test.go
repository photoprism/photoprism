package api

import (
	iofs "io/fs"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestGetSettings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetSettings(router)

		r := PerformRequest(app, "GET", "/api/v1/settings")
		val := gjson.Get(r.Body.String(), "ui.theme")
		assert.NotEmpty(t, val.String())
		val2 := gjson.Get(r.Body.String(), "ui.language")
		assert.NotEmpty(t, val2.String())
		assert.Equal(t, http.StatusOK, r.Code)
	})
}

func TestSaveSettings(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetSettings(router)

		r := PerformRequest(app, "GET", "/api/v1/settings")
		val := gjson.Get(r.Body.String(), "ui.language")
		assert.Equal(t, "en", val.String())
		assert.Equal(t, http.StatusOK, r.Code)
		SaveSettings(router)
		r2 := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"language": "de"}}`)
		assert.Equal(t, http.StatusOK, r2.Code)
		r4 := PerformRequest(app, "GET", "/api/v1/settings")
		val2 := gjson.Get(r4.Body.String(), "ui.language")
		assert.Equal(t, "de", val2.String())
		r3 := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"language": "en"}}`)
		assert.Equal(t, http.StatusOK, r3.Code)
	})
	t.Run("TimeZone", func(t *testing.T) {
		app, router, conf := NewApiTest()
		GetSettings(router)
		SaveSettings(router)

		zone := conf.Settings().UI.TimeZone
		t.Cleanup(func() {
			r := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"timeZone":"`+zone+`"}}`)
			assert.Equal(t, http.StatusOK, r.Code)
		})

		r := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"timeZone":"Europe/Berlin"}}`)
		require.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "Europe/Berlin", gjson.Get(PerformRequest(app, "GET", "/api/v1/settings").Body.String(), "ui.timeZone").String())

		r = PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"timeZone":"UTC+99"}}`)
		require.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "Europe/Berlin", conf.Settings().UI.TimeZone)
		assert.Equal(t, "Europe/Berlin", gjson.Get(PerformRequest(app, "GET", "/api/v1/settings").Body.String(), "ui.timeZone").String())
	})
	t.Run("BadRequest", func(t *testing.T) {
		app, router, _ := NewApiTest()

		SaveSettings(router)

		r := PerformRequestWithBody(app, "POST", "/api/v1/settings", `{"ui":{"language":123}}`)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, _ := NewApiTest()

		SaveSettings(router)

		body := `{"ui":{"language":"` + strings.Repeat("a", 300*1024) + `"}}`
		r := PerformRequestWithBody(app, "POST", "/api/v1/settings", body)
		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
	// ClientWithoutUser pins that a session without a user gets its settings back without storing them.
	t.Run("ClientWithoutUser", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })

		SaveSettings(router)

		sess, err := entity.AddClientSession("settings-client", conf.SessionMaxAge(), "settings", authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })

		countOwnerless := func() (n int64) {
			require.NoError(t, entity.UnscopedDb().Model(&entity.UserSettings{}).
				Where("user_uid IS NULL OR user_uid = ''").Count(&n).Error)
			return n
		}

		before := countOwnerless()

		for range 2 {
			r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/settings", `{"ui":{"language":"de"}}`, sess.AuthToken())
			assert.Equal(t, http.StatusOK, r.Code)
			assert.Equal(t, "de", gjson.Get(r.Body.String(), "ui.language").String())
		}

		assert.Equal(t, before, countOwnerless())
	})
}

func TestSettingsErr(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		hook, console := captureLog(t), captureSystemLog(t)
		settingsErr(&entity.Session{RefID: "sessxyz123"}, "user", nil)
		assert.Empty(t, hook.AllEntries())
		assert.Empty(t, console.AllEntries())
	})
	t.Run("ReachesTheConsoleAndNotTheBrowser", func(t *testing.T) {
		hook, console := captureLog(t), captureSystemLog(t)

		sub := event.Subscribe("system.log.error")
		defer event.Unsubscribe(sub)

		settingsErr(&entity.Session{RefID: "sessxyz123"}, "user", entity.ErrInvalidUserUID)

		assert.Empty(t, hook.AllEntries())

		entries := console.AllEntries()
		require.Len(t, entries, 1)
		assert.Equal(t, "settings: session sessxyz123 › failed to save user settings › invalid user uid", entries[0].Message)

		select {
		case msg := <-sub.Receiver:
			assert.Contains(t, msg.Fields["message"], "sessxyz123")
			assert.Contains(t, msg.Fields["message"], "invalid user uid")
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for the system event")
		}
	})
	t.Run("AppSettings", func(t *testing.T) {
		hook, console := captureLog(t), captureSystemLog(t)

		settingsErr(&entity.Session{RefID: "sessxyz123"}, "app", &iofs.PathError{Op: "open", Path: "/photoprism/storage/config/settings.yml", Err: iofs.ErrPermission})

		assert.Empty(t, hook.AllEntries())

		entries := console.AllEntries()
		require.Len(t, entries, 1)
		assert.Equal(t, logrus.ErrorLevel, entries[0].Level)
		assert.Equal(t, "settings: session sessxyz123 › failed to save app settings › open /photoprism/storage/config/settings.yml: permission denied", entries[0].Message)
	})
}
