package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/authn"
)

// holdPeopleLock takes the people update lock for the rest of the test, as a concurrent edit would.
func holdPeopleLock(t *testing.T) {
	t.Helper()

	require.NoError(t, mutex.UpdatePeople.Start())
	t.Cleanup(mutex.UpdatePeople.Stop)
}

// TestPeopleLockOrder pins that the people edit handlers authorize before they take the people
// update lock, so a request that is refused never holds it, while concurrent authorized edits
// are still answered busy.
func TestPeopleLockOrder(t *testing.T) {
	app, router, conf := NewApiTest()

	UpdateMarker(router)
	ClearMarkerSubject(router)
	UpdateSubject(router)

	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(prevAuthMode) })

	adminToken := AuthenticateAdmin(app, router)
	require.NotEmpty(t, adminToken)

	// A client session whose scope grants neither files nor people.
	denied, err := entity.AddClientSession("people-lock-test", conf.SessionMaxAge(), "metrics", authn.GrantClientCredentials, nil)
	require.NoError(t, err)
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(denied) })

	// Unknown UIDs, so an authorized request ends at the lookup after the lock and changes no fixture.
	endpoints := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"UpdateMarker", http.MethodPut, "/api/v1/markers/ms6sg6b1zzzzzz01", `{"MarkerName":"Lock Test"}`},
		{"ClearMarkerSubject", http.MethodDelete, "/api/v1/markers/ms6sg6b1zzzzzz01/subject", ""},
		{"UpdateSubject", http.MethodPut, "/api/v1/subjects/js6sg6b1zzzzzz01", `{"Name":"Lock Test"}`},
	}

	for _, e := range endpoints {
		t.Run(e.name, func(t *testing.T) {
			t.Run("NoTokenWhileHeld", func(t *testing.T) {
				holdPeopleLock(t)

				r := PerformRequestWithBody(app, e.method, e.path, e.body)
				assert.Equal(t, http.StatusUnauthorized, r.Code, r.Body.String())
				assert.True(t, mutex.UpdatePeople.Running(), "the lock stays with its holder")
			})
			t.Run("InsufficientScopeWhileHeld", func(t *testing.T) {
				holdPeopleLock(t)

				r := AuthenticatedRequestWithBody(app, e.method, e.path, e.body, denied.AuthToken())
				assert.Equal(t, http.StatusForbidden, r.Code, r.Body.String())
				assert.True(t, mutex.UpdatePeople.Running(), "the lock stays with its holder")
			})
			t.Run("AuthorizedWhileHeld", func(t *testing.T) {
				holdPeopleLock(t)

				r := AuthenticatedRequestWithBody(app, e.method, e.path, e.body, adminToken)
				assert.Equal(t, http.StatusTooManyRequests, r.Code, r.Body.String())
			})
			t.Run("AuthorizedReleasesTheLock", func(t *testing.T) {
				require.False(t, mutex.UpdatePeople.Running())
				t.Cleanup(mutex.UpdatePeople.Stop)

				// The lookup after the lock finds nothing, and the lock is released on that path too.
				r := AuthenticatedRequestWithBody(app, e.method, e.path, e.body, adminToken)
				assert.Equal(t, http.StatusNotFound, r.Code, r.Body.String())
				assert.False(t, mutex.UpdatePeople.Running())
			})
			t.Run("RefusedLeavesTheLockFree", func(t *testing.T) {
				require.False(t, mutex.UpdatePeople.Running())

				r := PerformRequestWithBody(app, e.method, e.path, e.body)
				assert.Equal(t, http.StatusUnauthorized, r.Code, r.Body.String())
				assert.False(t, mutex.UpdatePeople.Running())

				r = AuthenticatedRequestWithBody(app, e.method, e.path, e.body, denied.AuthToken())
				assert.Equal(t, http.StatusForbidden, r.Code, r.Body.String())
				assert.False(t, mutex.UpdatePeople.Running())
			})
		})
	}

	t.Run("WakeupIntervalAfterAuthorization", func(t *testing.T) {
		// The interval is reported only to an authorized caller.
		prevInterval, prevUnsafe := conf.Options().WakeupInterval, conf.Options().Unsafe
		conf.Options().WakeupInterval = 2 * time.Hour
		conf.Options().Unsafe = false
		t.Cleanup(func() {
			conf.Options().WakeupInterval = prevInterval
			conf.Options().Unsafe = prevUnsafe
		})

		for _, e := range endpoints[:2] {
			r := PerformRequestWithBody(app, e.method, e.path, e.body)
			assert.Equal(t, http.StatusUnauthorized, r.Code, e.name)

			r = AuthenticatedRequestWithBody(app, e.method, e.path, e.body, adminToken)
			assert.Equal(t, http.StatusForbidden, r.Code, e.name)
			assert.False(t, mutex.UpdatePeople.Running())
		}
	})
}
