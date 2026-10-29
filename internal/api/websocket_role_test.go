package api

import (
	"encoding/json"
	"errors"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// wsConnect opens a websocket to the test server, authenticates it with the token and waits for the
// client config the handshake sends, so the connection is known to be authenticated before an event
// is published.
func wsConnect(t *testing.T, srv *httptest.Server, authToken string) *websocket.Conn {
	t.Helper()

	uri := "ws" + strings.TrimPrefix(srv.URL, "http") + "/api/v1/ws"
	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	require.NoError(t, err)

	t.Cleanup(func() { _ = conn.Close() })

	require.NoError(t, conn.WriteJSON(map[string]string{"session": authToken}))
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(10*time.Second)))

	for {
		var msg struct {
			Event string `json:"event"`
		}

		require.NoError(t, conn.ReadJSON(&msg), "handshake sent no client config")

		if msg.Event == "config.updated" {
			return conn
		}
	}
}

// wsReceivesMarker reports whether the connection is sent the log event carrying the marker within
// the wait, ignoring any other traffic on the socket.
func wsReceivesMarker(t *testing.T, conn *websocket.Conn, marker string, wait time.Duration) bool {
	t.Helper()

	deadline := time.Now().Add(wait)
	require.NoError(t, conn.SetReadDeadline(deadline))

	for {
		_, raw, err := conn.ReadMessage()

		if err != nil {
			return false
		}

		var msg struct {
			Event string          `json:"event"`
			Data  json.RawMessage `json:"data"`
		}

		if json.Unmarshal(raw, &msg) == nil && msg.Event == "log.info" && strings.Contains(string(msg.Data), marker) {
			return true
		}

		if time.Now().After(deadline) {
			return false
		}
	}
}

// wsRefusalGrace is how long a refused connection is watched after the same event reached the admin
// connection; Publish has queued it for every subscriber by then, so a longer wait proves nothing more.
const wsRefusalGrace = 250 * time.Millisecond

// wsRefused publishes a log event and reports whether the connection did not receive it, after the admin
// connection confirmed delivery of the same event. The watch must end in a read timeout, not a closed socket.
func wsRefused(t *testing.T, admin, conn *websocket.Conn, publish func() string) bool {
	t.Helper()

	marker := publish()
	require.True(t, wsReceivesMarker(t, admin, marker, 5*time.Second), "admin connection was not sent the event")
	require.NoError(t, conn.SetReadDeadline(time.Now().Add(wsRefusalGrace)))

	for {
		_, raw, err := conn.ReadMessage()

		if err != nil {
			var netErr net.Error
			require.True(t, errors.As(err, &netErr) && netErr.Timeout(), "connection closed instead of staying open: %v", err)

			return true
		}

		if strings.Contains(string(raw), marker) {
			return false
		}
	}
}

// TestWebSocket_EffectiveRole covers what the production handshake and delivery loop send a client
// session acting for a privileged account.
func TestWebSocket_EffectiveRole(t *testing.T) {
	app, router, conf := NewApiTest()
	conf.SetAuthMode(config.AuthModePasswd)
	defer conf.SetAuthMode(config.AuthModePublic)

	WebSocket(router)
	adminToken := AuthenticateAdmin(app, router)
	require.NotEmpty(t, adminToken)

	srv := httptest.NewServer(app)
	defer srv.Close()

	publishMarker := func() string {
		marker := rnd.UUID()
		event.Publish("log.info", event.Data{"level": "info", "message": marker})

		return marker
	}

	// The admin connection is the positive control for every refusal below: each one first confirms that
	// the event was delivered, so a refusal cannot follow from the socket never delivering anything.
	admin := wsConnect(t, srv, adminToken)

	t.Run("AdminSessionReceivesLogEvents", func(t *testing.T) {
		assert.True(t, wsReceivesMarker(t, admin, publishMarker(), 5*time.Second))
	})
	t.Run("InstanceClientForAdminAccountRefused", func(t *testing.T) {
		conn := wsConnect(t, srv, mixedPrincipalToken(t, acl.RoleInstance, "alice"))
		assert.True(t, wsRefused(t, admin, conn, publishMarker))
	})
	t.Run("ServiceClientForAdminAccountRefused", func(t *testing.T) {
		conn := wsConnect(t, srv, mixedPrincipalToken(t, acl.RoleService, "alice"))
		assert.True(t, wsRefused(t, admin, conn, publishMarker))
	})
	t.Run("DefaultClientForAdminAccountRefused", func(t *testing.T) {
		// A registered client is refused whichever role it holds, the default one included.
		conn := wsConnect(t, srv, mixedPrincipalToken(t, acl.RoleClient, "alice"))
		assert.True(t, wsRefused(t, admin, conn, publishMarker))
	})
	t.Run("AppPasswordForAdminAccountRefused", func(t *testing.T) {
		// An app password authenticates a client or an agent, so it is refused like one even though it
		// records no client of its own.
		sess, err := entity.AddClientSession("app-password", 3600, "*", authn.GrantPassword, entity.FindUserByName("alice"))
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })

		entity.FlushSessionCache()

		conn := wsConnect(t, srv, sess.AuthToken())
		assert.True(t, wsRefused(t, admin, conn, publishMarker))
	})
	t.Run("InstanceClientRefusedAfterAReload", func(t *testing.T) {
		token := providerClientToken(t, acl.RoleInstance, authn.ProviderLocal, "alice")

		// Reload the session the way a server that did not just mint it does, so the handshake resolves
		// the client record rather than reading one already attached.
		entity.FlushSessionCache()

		conn := wsConnect(t, srv, token)
		assert.True(t, wsRefused(t, admin, conn, publishMarker))
	})
	t.Run("ClientWithAnUnlistedRoleRefused", func(t *testing.T) {
		// A role another edition registers, reaching this build through a restored database. It maps to
		// no role here, and the session still authenticates a client.
		token := unlistedRoleToken(t, "manager", "alice")
		conn := wsConnect(t, srv, token)

		assert.True(t, wsRefused(t, admin, conn, publishMarker))
	})
}

// unlistedRoleToken registers a client whose role string this edition does not map to a role, opens a
// session for it and returns the token.
func unlistedRoleToken(t *testing.T, role, userName string) string {
	t.Helper()

	user := entity.FindUserByName(userName)
	require.NotNil(t, user)
	require.NotContains(t, acl.ClientRoles, role, "this edition maps the role, so the case tests nothing")

	client := &entity.Client{
		ClientUID:    rnd.GenerateUID(entity.ClientUID),
		UserUID:      user.UserUID,
		UserName:     user.UserName,
		ClientName:   "role-test-" + role,
		ClientRole:   role,
		ClientType:   authn.ClientConfidential,
		AuthProvider: authn.ProviderClient.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "*",
		AuthExpires:  3600,
		AuthTokens:   5,
		AuthEnabled:  true,
	}

	require.NoError(t, client.Create())
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(client).Error })

	sess := entity.NewSession(client.AuthExpires, 0).SetClient(client).SetGrantType(authn.GrantClientCredentials)
	sess.SetUser(user)
	require.NoError(t, sess.Create())
	t.Cleanup(func() { _ = sess.Delete() })

	require.True(t, sess.IsClient())
	require.Equal(t, acl.RoleNone, sess.GetClientRole())

	return sess.AuthToken()
}

// providerClientToken registers a client with the given role and auth provider owned by the named
// user, opens a session for it and returns the token.
func providerClientToken(t *testing.T, role acl.Role, provider authn.ProviderType, userName string) string {
	t.Helper()

	user := entity.FindUserByName(userName)
	require.NotNil(t, user)

	client := &entity.Client{
		ClientUID:    rnd.GenerateUID(entity.ClientUID),
		UserUID:      user.UserUID,
		UserName:     user.UserName,
		ClientName:   "role-test-" + role.String() + "-" + provider.String(),
		ClientRole:   role.String(),
		ClientType:   authn.ClientConfidential,
		AuthProvider: provider.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "*",
		AuthExpires:  3600,
		AuthTokens:   5,
		AuthEnabled:  true,
	}

	require.NoError(t, client.Create())
	t.Cleanup(func() { _ = entity.UnscopedDb().Delete(client).Error })

	sess := entity.NewSession(client.AuthExpires, 0).SetClient(client).SetGrantType(authn.GrantClientCredentials)
	sess.SetUser(user)
	sess.SetProvider(authn.ProviderClient)
	require.NoError(t, sess.Create())
	t.Cleanup(func() { _ = sess.Delete() })

	require.Equal(t, role, sess.GetClientRole())

	return sess.AuthToken()
}
