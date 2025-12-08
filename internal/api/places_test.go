package api

import (
	"net/http"
	"testing"

	"github.com/photoprism/photoprism/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

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
	t.Run("EmptyQuery", func(t *testing.T) {
		app, router, _ := NewApiTest()

		GetPlacesSearch(router)

		r := PerformRequest(app, http.MethodGet, "/api/v1/places/search?q=&locale=en")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
}
