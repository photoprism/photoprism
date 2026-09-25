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

	"github.com/photoprism/photoprism/internal/config"
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
}
