package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/config"
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

	for i := 0; i < 3; i++ {
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
