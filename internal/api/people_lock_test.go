package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// slowBody signals its first read and blocks it until released, as a body that arrives slowly would.
type slowBody struct {
	started chan struct{}
	release chan struct{}
	data    *strings.Reader
	once    sync.Once
}

// newSlowBody returns a slowBody that eventually yields data.
func newSlowBody(data string) *slowBody {
	return &slowBody{started: make(chan struct{}), release: make(chan struct{}), data: strings.NewReader(data)}
}

// Read blocks until the body is released, then returns its data.
func (b *slowBody) Read(p []byte) (int, error) {
	b.once.Do(func() { close(b.started) })
	<-b.release

	return b.data.Read(p)
}

// Close does nothing.
func (b *slowBody) Close() error { return nil }

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

	t.Run("BodyReadBeforeTheLock", func(t *testing.T) {
		// While the body is still arriving, the lock is free for other edits. The body is invalid
		// JSON, so the request ends at the form and changes no fixture.
		for _, e := range []struct {
			name string
			path string
		}{
			{"UpdateMarker", "/api/v1/markers/ms6sg6b14ahkyd24"},
			{"UpdateSubject", "/api/v1/subjects/js6sg6b1qekk9jx8"},
		} {
			t.Run(e.name, func(t *testing.T) {
				require.False(t, mutex.UpdatePeople.Running())
				t.Cleanup(mutex.UpdatePeople.Stop)

				body := newSlowBody("{")
				released := false
				t.Cleanup(func() {
					if !released {
						close(body.release)
					}
				})

				req := httptest.NewRequest(http.MethodPut, e.path, body)
				req.Header.Set(header.XAuthToken, adminToken)
				w := httptest.NewRecorder()
				done := make(chan struct{})

				go func() {
					app.ServeHTTP(w, req)
					close(done)
				}()

				select {
				case <-body.started:
				case <-time.After(10 * time.Second):
					t.Fatal("the handler never read the body")
				}

				assert.False(t, mutex.UpdatePeople.Running(), "the lock is not held while the body arrives")

				released = true
				close(body.release)
				<-done

				assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
				assert.False(t, mutex.UpdatePeople.Running())
			})
		}
	})
	t.Run("OversizedBodyBeforeTheLock", func(t *testing.T) {
		// The body is bounded and read before the lock, so an oversized one is refused even while
		// another edit holds it.
		holdPeopleLock(t)

		oversized := `{"Name":"` + strings.Repeat("x", int(MaxMutationRequestBytes)) + `"}`

		// Exactly at the bound, the same body passes the read and meets the held lock instead.
		within := `{"Name":"` + strings.Repeat("x", int(MaxMutationRequestBytes)-11) + `"}`
		require.Len(t, within, int(MaxMutationRequestBytes))

		for _, path := range []string{"/api/v1/markers/ms6sg6b14ahkyd24", "/api/v1/subjects/js6sg6b1qekk9jx8"} {
			r := AuthenticatedRequestWithBody(app, http.MethodPut, path, oversized, adminToken)
			assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code, path)
			assert.True(t, mutex.UpdatePeople.Running(), "the lock stays with its holder")

			r = AuthenticatedRequestWithBody(app, http.MethodPut, path, within, adminToken)
			assert.Equal(t, http.StatusTooManyRequests, r.Code, path)
		}
	})
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
