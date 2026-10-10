package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/server/limiter"
)

func TestCreateSession_RateLimitExceeded(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	CreateSession(router)

	// Tighten rate limits and do repeated bad logins from UnknownIP
	oldLogin, oldAuth := limiter.Login, limiter.Auth
	defer func() { limiter.Login, limiter.Auth = oldLogin, oldAuth }()
	limiter.Login = limiter.NewLimit(rate.Every(24*time.Hour), 3)
	limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 3)

	for range 3 {
		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "admin", "password": "wrong"}`)
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	}
	// Next attempt should be 429
	r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{"username": "admin", "password": "wrong"}`)
	assert.Equal(t, http.StatusTooManyRequests, r.Code)
}

func TestCreateSession_MissingFields(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	CreateSession(router)
	// Empty object -> unauthorized (invalid credentials)
	r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/session", `{}`)
	assert.Equal(t, http.StatusUnauthorized, r.Code)
}

// TestCreateSession_RateLimitIPv6Network checks that failed logins from one IPv6 /64 share a limit while
// audit lines keep the full client address.
func TestCreateSession_RateLimitIPv6Network(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)
	CreateSession(router)

	oldLogin, oldAuth := limiter.Login, limiter.Auth
	defer func() { limiter.Login, limiter.Auth = oldLogin, oldAuth }()
	limiter.Login = limiter.NewLimit(rate.Every(24*time.Hour), 3)
	limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 3)

	origAudit := event.AuditLog
	auditLogger, auditHook := test.NewNullLogger()
	auditLogger.SetLevel(logrus.TraceLevel)
	event.AuditLog = auditLogger
	t.Cleanup(func() { event.AuditLog = origAudit })

	// login sends a failed login from the specified address and returns the status.
	login := func(remoteAddr string) int {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/session", strings.NewReader(`{"username": "admin", "password": "wrong"}`))
		req.RemoteAddr = remoteAddr
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		return w.Code
	}

	for i := range 3 {
		assert.Equal(t, http.StatusUnauthorized, login(fmt.Sprintf("[2001:db8:1:2::%x]:1234", i+1)))
	}

	assert.Equal(t, http.StatusTooManyRequests, login("[2001:db8:1:2:ffff::1]:1234"))
	assert.Equal(t, http.StatusUnauthorized, login("[2001:db8:1:3::1]:1234"))

	var found bool

	for _, entry := range auditHook.AllEntries() {
		if strings.Contains(entry.Message, "2001:db8:1:2::2 ") {
			found = true
		}
		assert.NotContains(t, entry.Message, "/64")
	}

	assert.True(t, found, "audit lines must name the full client address")
}
