package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"golang.org/x/time/rate"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// TestWebSocket_Config checks which client config the connection handshake returns for a session.
func TestWebSocket_Config(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	userToken := AuthenticateUser(app, router, "alice", "Alice123!")
	require.NotEmpty(t, userToken)

	WebSocket(router)

	srv := httptest.NewServer(app)
	defer srv.Close()

	// handshake connects, sends the session token, and returns the config of the first config update.
	handshake := func(t *testing.T, token string) gjson.Result {
		t.Helper()

		ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/api/v1/ws", http.Header{"Origin": {srv.URL}})
		require.NoError(t, err)
		defer ws.Close()

		require.NoError(t, ws.WriteJSON(map[string]string{"session": token}))
		require.NoError(t, ws.SetReadDeadline(time.Now().Add(10*time.Second)))

		for {
			_, msg, readErr := ws.ReadMessage()
			require.NoError(t, readErr)

			if gjson.GetBytes(msg, "event").String() == "config.updated" {
				return gjson.GetBytes(msg, "data.config")
			}
		}
	}

	t.Run("UserSession", func(t *testing.T) {
		assert.NotEmpty(t, handshake(t, userToken).Get("previewToken").String())
	})
	t.Run("ClientOutsideScope", func(t *testing.T) {
		_, sess := newOwnedClientSession(t, "metrics")
		assert.Empty(t, handshake(t, sess.AuthToken()).Get("previewToken").String())
	})
	t.Run("ExpiredAccount", func(t *testing.T) {
		u := *entity.UserFixtures.Pointer("bob")
		expired := time.Now().Add(-time.Hour)
		u.ExpiresAt = &expired
		sess := entity.NewSession(3600, 0).SetUser(&u)
		require.NoError(t, sess.Create())
		t.Cleanup(func() { _ = sess.Delete() })

		ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/api/v1/ws", http.Header{"Origin": {srv.URL}})
		require.NoError(t, err)
		defer ws.Close()

		require.NoError(t, ws.WriteJSON(map[string]string{"session": sess.AuthToken()}))
		require.NoError(t, ws.SetReadDeadline(time.Now().Add(10*time.Second)))

		for {
			_, msg, readErr := ws.ReadMessage()
			require.NoError(t, readErr)

			if gjson.GetBytes(msg, "event").String() == "config.updated" {
				assert.Empty(t, gjson.GetBytes(msg, "data.config.previewToken").String())
				break
			}
		}

		// The connection receives events as the unknown user while the account is expired.
		var bound []string

		wsAuth.mutex.Lock()
		for connId, sid := range wsAuth.sid {
			if sid == sess.ID {
				bound = append(bound, wsAuth.user[connId].UserUID)
			}
		}
		wsAuth.mutex.Unlock()

		assert.Equal(t, []string{entity.UnknownUser.UserUID}, bound)
	})
}

// TestWebSocket_AuthLimit checks that failed session tokens sent over a WebSocket count against the client's authentication rate limit.
func TestWebSocket_AuthLimit(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() { conf.SetAuthMode(config.AuthModePublic) })

	origLimit := limiter.Auth
	t.Cleanup(func() { limiter.Auth = origLimit })
	limiter.Auth = limiter.NewLimit(rate.Every(24*time.Hour), 3)

	WebSocket(router)

	srv := httptest.NewServer(app)
	defer srv.Close()

	for range 3 {
		ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/api/v1/ws", http.Header{"Origin": {srv.URL}})
		require.NoError(t, err)
		require.NoError(t, ws.WriteJSON(map[string]string{"session": rnd.AuthToken()}))
		_ = ws.Close()
	}

	assert.Eventually(t, func() bool { return limiter.Auth.Reject("127.0.0.1") }, 5*time.Second, 20*time.Millisecond)

	// A client over the limit is told so and the connection is closed with "try again later".
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(srv.URL, "http")+"/api/v1/ws", http.Header{"Origin": {srv.URL}})
	require.NoError(t, err)
	defer ws.Close()
	require.NoError(t, ws.WriteJSON(map[string]string{"session": rnd.AuthToken()}))
	require.NoError(t, ws.SetReadDeadline(time.Now().Add(10*time.Second)))

	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, wsRateLimitedEvent, gjson.GetBytes(msg, "event").String())
	assert.Equal(t, int64(http.StatusTooManyRequests), gjson.GetBytes(msg, "data.code").Int())

	_, _, err = ws.ReadMessage()
	assert.True(t, websocket.IsCloseError(err, websocket.CloseTryAgainLater), "%v", err)
}
