package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/auth/acl"
)

func TestSearchPhotoTimeline(t *testing.T) {
	t.Run("MissingBucketRejected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?uid=ps6sg6be2lvl0yh0")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("SuccessWithoutCount", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0")
		body := r.Body.String()

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "month", gjson.Get(body, "bucket").String())
		assert.Equal(t, "desc", gjson.Get(body, "order").String())
		assert.Equal(t, int64(1), gjson.Get(body, "photoCount").Int())
		assert.Equal(t, int64(0), gjson.Get(body, "unknownDateCount").Int())
		assert.Equal(t, int64(1), gjson.Get(body, "buckets.#").Int())
		assert.Equal(t, "1990-04", gjson.Get(body, "buckets.0.key").String())
		assert.Equal(t, int64(1), gjson.Get(body, "buckets.0.photoCount").Int())
	})

	t.Run("CalendarPermissionRequired", func(t *testing.T) {
		withTimelineAclRule(t, acl.ResourceCalendar, acl.Roles{})

		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0")

		assert.Equal(t, http.StatusForbidden, r.Code)
	})

	t.Run("PhotoPermissionRequired", func(t *testing.T) {
		withTimelineAclRule(t, acl.ResourcePhotos, acl.Roles{})

		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0")

		assert.Equal(t, http.StatusForbidden, r.Code)
	})

	t.Run("CalendarFeatureRequired", func(t *testing.T) {
		_, _, conf := NewApiTest()
		settings := conf.Settings()
		enabled := settings.Features.Calendar
		settings.Features.Calendar = false
		t.Cleanup(func() {
			settings.Features.Calendar = enabled
		})

		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0")

		assert.Equal(t, http.StatusForbidden, r.Code)
		assert.Equal(t, "Feature disabled", gjson.Get(r.Body.String(), "error").String())
	})

	t.Run("DayBucket", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=day&uid=ps6sg6be2lvl0yh0")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("CountOffsetIgnored", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0&count=1&offset=1")
		body := r.Body.String()

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, int64(1), gjson.Get(body, "photoCount").Int())
		assert.Equal(t, int64(1), gjson.Get(body, "buckets.#").Int())
		assert.Equal(t, "1990-04", gjson.Get(body, "buckets.0.key").String())
	})

	t.Run("InvalidShapingParamsIgnored", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0&count=bad&offset=bad&reverse=bad&merged=bad")
		body := r.Body.String()

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, int64(1), gjson.Get(body, "photoCount").Int())
		assert.Equal(t, "1990-04", gjson.Get(body, "buckets.0.key").String())
	})

	t.Run("ReverseMergedIgnored", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0&reverse=true&merged=true")
		body := r.Body.String()

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "desc", gjson.Get(body, "order").String())
		assert.Equal(t, "1990-04", gjson.Get(body, "buckets.0.key").String())
	})

	t.Run("OrderFilterAccepted", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&order=similar")
		body := r.Body.String()

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "desc", gjson.Get(body, "order").String())
	})

	t.Run("InvalidOrderRejected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0&order=asc")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("EmptyOrderIgnored", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0&order=")
		body := r.Body.String()

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "desc", gjson.Get(body, "order").String())
		assert.Equal(t, int64(1), gjson.Get(body, "photoCount").Int())
	})

	t.Run("InvalidFilterRejected", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&added=not-a-date")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("InvalidBucket", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=year&uid=ps6sg6be2lvl0yh0")

		assert.Equal(t, http.StatusBadRequest, r.Code)
	})

	t.Run("NoMatchReturnsEmptyBuckets", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&label=totally-unknown-label")
		body := r.Body.String()

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, int64(0), gjson.Get(body, "photoCount").Int())
		assert.Equal(t, int64(0), gjson.Get(body, "unknownDateCount").Int())
		assert.Contains(t, strings.ReplaceAll(body, " ", ""), `"buckets":[]`)
	})

	t.Run("DoesNotMutateRequestQuery", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)

		req := httptest.NewRequest("GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0&count=bad&offset=bad", nil)
		rawQuery := req.URL.RawQuery
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, rawQuery, req.URL.RawQuery)
	})

	t.Run("BeforePhotoUidRoute", func(t *testing.T) {
		app, router, _ := NewApiTest()
		SearchPhotoTimeline(router)
		GetPhoto(router)

		r := PerformRequest(app, "GET", "/api/v1/photos/timeline?bucket=month&uid=ps6sg6be2lvl0yh0")
		body := r.Body.String()

		assert.Equal(t, http.StatusOK, r.Code)
		assert.Equal(t, "month", gjson.Get(body, "bucket").String())
	})
}

func withTimelineAclRule(t *testing.T, resource acl.Resource, roles acl.Roles) {
	t.Helper()

	acl.RulesMutex.Lock()
	previous := acl.Rules[resource]
	acl.Rules[resource] = roles
	acl.RulesMutex.Unlock()

	t.Cleanup(func() {
		acl.RulesMutex.Lock()
		acl.Rules[resource] = previous
		acl.RulesMutex.Unlock()
	})
}
