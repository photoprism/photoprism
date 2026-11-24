package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/hub/places"
	"github.com/photoprism/photoprism/pkg/i18n"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// failPlacesRequests points the places service at an address that refuses connections for the duration of the test.
func failPlacesRequests(t *testing.T) {
	t.Helper()

	reverse, search, retries, delay := places.ReverseServiceUrls, places.SearchServiceUrls, places.Retries, places.RetryDelay
	places.ReverseServiceUrls = []string{"https://127.0.0.1:1/v1/reverse"}
	places.SearchServiceUrls = []string{"https://127.0.0.1:1/v1/search"}
	places.Retries, places.RetryDelay = 1, 0

	t.Cleanup(func() {
		places.ReverseServiceUrls, places.SearchServiceUrls, places.Retries, places.RetryDelay = reverse, search, retries, delay
	})
}

// browserLog runs fn and returns the messages it published on the log channel that browsers receive.
func browserLog(t *testing.T, fn func()) (messages []string) {
	t.Helper()

	sub := event.Subscribe("log.*")
	defer event.Unsubscribe(sub)

	fn()

	for {
		select {
		case msg := <-sub.Receiver:
			text, _ := msg.Fields["message"].(string)
			messages = append(messages, text)
		default:
			return messages
		}
	}
}

func TestGetPlacesReverse(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPlacesReverse(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/places/reverse?lat=52.5108869&lng=13.398947")

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "Berlin", gjson.Get(r.Body.String(), "name").String())
	})
	t.Run("Unauthorized", func(t *testing.T) {
		app, router, conf := NewApiTest()

		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPlacesReverse(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/places/reverse?lat=52.5108869&lng=13.398947")

		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("FeatureDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()

		conf.Options().DisablePlaces = true

		GetPlacesReverse(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/places/reverse?lat=52.5108869&lng=13.398947")

		assert.Equal(t, http.StatusForbidden, r.Code)

		conf.Options().DisablePlaces = false
	})
	t.Run("ServiceError", func(t *testing.T) {
		// A failed request is answered with a generic error.
		app, router, _ := NewApiTest()
		failPlacesRequests(t)
		appLog, sysLog := captureLog(t), captureSystemLog(t)

		GetPlacesReverse(router)

		var r *httptest.ResponseRecorder
		browser := browserLog(t, func() {
			r = PerformRequest(app, http.MethodGet, "/api/v1/places/reverse?lat=48.1371079&lng=11.5753822")
		})

		assert.Equal(t, http.StatusInternalServerError, r.Code)
		assert.Equal(t, i18n.Msg(i18n.ErrUnexpected), gjson.Get(r.Body.String(), "error").String())
		assert.NotContains(t, r.Body.String(), "127.0.0.1")
		assert.NotContains(t, r.Body.String(), "48.1371079")

		// The handler logs the query to the system log, not the package logger, and without the request URL.
		for _, entry := range appLog.AllEntries() {
			assert.NotContains(t, entry.Message, "48.1371079")
		}

		require.NotEmpty(t, sysLog.AllEntries())
		assert.Contains(t, sysLog.LastEntry().Message, "48.1371079")
		assert.NotContains(t, sysLog.LastEntry().Message, "127.0.0.1:1/v1")

		// The places client reports the failure on the browser log channel by host only.
		require.NotEmpty(t, browser)
		for _, msg := range browser {
			assert.NotContains(t, msg, "48.1371079")
		}
	})
	t.Run("LatitudeMissing", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPlacesReverse(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/places/reverse?lng=13.398947")

		assert.Equal(t, http.StatusBadRequest, r.Code)

	})
	t.Run("LongitudeMissing", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPlacesReverse(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/places/reverse?lat=52.5108869")

		assert.Equal(t, http.StatusBadRequest, r.Code)

	})
}

func TestGetPlacesSearch(t *testing.T) {
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPlacesSearch(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/places/search?q=Berlin&locale=en")

		assert.Equal(t, http.StatusOK, r.Code)
		assert.LessOrEqual(t, 1, int(gjson.Get(r.Body.String(), "#").Int()))
	})
	t.Run("Unauthorized", func(t *testing.T) {
		app, router, conf := NewApiTest()

		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)

		GetPlacesSearch(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/places/search?q=Berlin&locale=en")

		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("FeatureDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()

		conf.Options().DisablePlaces = true

		GetPlacesSearch(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/places/search?q=Berlin&locale=en")

		assert.Equal(t, http.StatusForbidden, r.Code)

		conf.Options().DisablePlaces = false
	})
	t.Run("ServiceError", func(t *testing.T) {
		// A failed request is answered with a generic error.
		app, router, _ := NewApiTest()
		failPlacesRequests(t)
		appLog, sysLog := captureLog(t), captureSystemLog(t)

		GetPlacesSearch(router)

		var r *httptest.ResponseRecorder
		browser := browserLog(t, func() {
			r = PerformRequest(app, http.MethodGet, "/api/v1/places/search?q=Neuschwanstein%20Castle&locale=en")
		})

		assert.Equal(t, http.StatusInternalServerError, r.Code)
		assert.Equal(t, i18n.Msg(i18n.ErrUnexpected), gjson.Get(r.Body.String(), "error").String())
		assert.NotContains(t, r.Body.String(), "127.0.0.1")
		assert.NotContains(t, r.Body.String(), "Neuschwanstein")

		// The handler logs the query to the system log, not the package logger, and without the request URL.
		for _, entry := range appLog.AllEntries() {
			assert.NotContains(t, entry.Message, "Neuschwanstein")
		}

		require.NotEmpty(t, sysLog.AllEntries())
		assert.Contains(t, sysLog.LastEntry().Message, "Neuschwanstein")
		assert.NotContains(t, sysLog.LastEntry().Message, "127.0.0.1:1/v1")

		// The places client reports the failure on the browser log channel by host only.
		require.NotEmpty(t, browser)
		for _, msg := range browser {
			assert.NotContains(t, msg, "Neuschwanstein")
		}
	})
	t.Run("EmptyQuery", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPlacesSearch(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/places/search?q=&locale=en")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
