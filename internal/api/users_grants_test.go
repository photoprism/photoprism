package api

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// allowSessionManagement grants admins the management of all sessions, as in Pro and Portal, for the test.
func allowSessionManagement(t *testing.T) {
	t.Helper()

	prev := acl.Rules[acl.ResourceSessions][acl.RoleAdmin]
	acl.Rules[acl.ResourceSessions][acl.RoleAdmin] = acl.GrantFullAccess
	t.Cleanup(func() { acl.Rules[acl.ResourceSessions][acl.RoleAdmin] = prev })
}

func TestFindUserGrants(t *testing.T) {
	app, router, conf, client := prepareAppTokenTest(t)
	FindUserGrants(router)
	DeleteUserGrant(router)
	tokenPair(t, app, router, conf, client)

	adminToken := signIn(app, "admin", "photoprism")

	t.Run("Own", func(t *testing.T) {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/users/"+entity.Admin.UserUID+"/grants", adminToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, client.ClientUID, gjson.Get(r.Body.String(), "0.ClientID").String())
		assert.Equal(t, "Claude", gjson.Get(r.Body.String(), "0.ClientName").String())
		assert.Equal(t, "https://claude.ai", gjson.Get(r.Body.String(), "0.Origin").String())
		assert.Equal(t, "mcp", gjson.Get(r.Body.String(), "0.Scope").String())

		var keys []string
		gjson.Get(r.Body.String(), "0").ForEach(func(key, _ gjson.Result) bool {
			keys = append(keys, key.String())
			return true
		})
		assert.ElementsMatch(t, []string{"UID", "ClientID", "ClientName", "Origin", "Scope", "CreatedAt", "UsedAt"}, keys)
	})
	t.Run("OtherUserAsGuest", func(t *testing.T) {
		guestToken := signIn(app, "gandalf", "Gandalf123!")
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/users/"+entity.Admin.UserUID+"/grants", guestToken)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("OtherUserAsAdmin", func(t *testing.T) {
		alice := entity.FindUserByName("alice")
		require.NotNil(t, alice)

		// In CE, admins manage only their own sessions.
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/users/"+alice.UserUID+"/grants", adminToken)
		assert.Equal(t, http.StatusForbidden, r.Code)

		allowSessionManagement(t)
		r = AuthenticatedRequest(app, http.MethodGet, "/api/v1/users/"+alice.UserUID+"/grants", adminToken)
		assert.Equal(t, http.StatusOK, r.Code)
	})
	t.Run("InvalidUID", func(t *testing.T) {
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/users/abc/grants", adminToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("Unauthenticated", func(t *testing.T) {
		r := PerformRequest(app, http.MethodGet, "/api/v1/users/"+entity.Admin.UserUID+"/grants")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
}

func TestDeleteUserGrant(t *testing.T) {
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`

	t.Run("EndsTokensAndMCPSessions", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		FindUserGrants(router)
		DeleteUserGrant(router)
		accessToken, refreshToken := tokenPair(t, app, router, conf, client)
		adminToken := signIn(app, "admin", "photoprism")

		w := mcpPost(app, initBody, accessToken, "")
		require.Equal(t, http.StatusOK, w.Code)
		sessionID := w.Header().Get("Mcp-Session-Id")
		require.NotEmpty(t, sessionID)

		grant := entity.FindUserClientGrant(entity.Admin.UserUID, client.ClientUID)
		require.NotNil(t, grant)

		// Another user cannot disconnect the app.
		guestToken := signIn(app, "gandalf", "Gandalf123!")
		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/users/"+entity.Admin.UserUID+"/grants/"+grant.GrantUID, guestToken)
		assert.Equal(t, http.StatusForbidden, r.Code)

		r = AuthenticatedRequest(app, http.MethodDelete, "/api/v1/users/"+entity.Admin.UserUID+"/grants/"+grant.GrantUID, adminToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Nil(t, entity.FindGrant(grant.GrantUID))

		// The access token, the refresh token and the MCP session have ended.
		assert.Equal(t, http.StatusUnauthorized, mcpPost(app, initBody, accessToken, "").Code)
		callBody := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_config_keys","arguments":{"query":"database","limit":5}}}`
		assert.Equal(t, http.StatusNotFound, mcpPost(app, callBody, adminToken, sessionID).Code)

		r = postTokenForm(app, "/api/v1/oauth/token", refreshForm(client.ClientUID, refreshToken), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("NotFound", func(t *testing.T) {
		app, router, _, _ := prepareAppTokenTest(t)
		DeleteUserGrant(router)
		adminToken := AuthenticateAdmin(app, router)
		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/users/"+entity.Admin.UserUID+"/grants/"+rnd.GenerateUID(entity.GrantUID), adminToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("OtherUsersGrant", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		DeleteUserGrant(router)
		tokenPair(t, app, router, conf, client)
		adminToken := signIn(app, "admin", "photoprism")
		grant := entity.FindUserClientGrant(entity.Admin.UserUID, client.ClientUID)
		require.NotNil(t, grant)

		alice := entity.FindUserByName("alice")
		allowSessionManagement(t)
		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/users/"+alice.UserUID+"/grants/"+grant.GrantUID, adminToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))
	})
}

func TestGrantLifecycle(t *testing.T) {
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`

	t.Run("PastSessionMaxAge", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		_, refreshToken := tokenPair(t, app, router, conf, client)
		grant := entity.FindUserClientGrant(entity.Admin.UserUID, client.ClientUID)
		require.NotNil(t, grant)

		created := time.Now().Add(-time.Duration(conf.SessionMaxAge()+86400) * time.Second)
		require.NoError(t, entity.UnscopedDb().Model(&entity.Grant{}).Where("grant_uid = ?", grant.GrantUID).UpdateColumn("created_at", created).Error)

		r := postTokenForm(app, "/api/v1/oauth/token", refreshForm(client.ClientUID, refreshToken), "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, http.StatusOK, mcpPost(app, initBody, gjson.Get(r.Body.String(), "access_token").String(), "").Code)
	})
	t.Run("RoleChangedAtSignIn", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		accessToken, refreshToken := tokenPair(t, app, router, conf, client)
		grant := entity.FindUserClientGrant(entity.Admin.UserUID, client.ClientUID)
		require.NotNil(t, grant)

		// A role applied without revoking derived sessions, as the OIDC and LDAP role sync does,
		// leaves the grant in place, and the next request is limited by the new role.
		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)
		role, superAdmin := admin.UserRole, admin.SuperAdmin
		require.NoError(t, admin.Updates(entity.Values{"UserRole": "guest", "SuperAdmin": false}))
		t.Cleanup(func() { _ = admin.Updates(entity.Values{"UserRole": role, "SuperAdmin": superAdmin}) })
		entity.FlushSessionCache()

		assert.Equal(t, http.StatusForbidden, mcpPost(app, initBody, accessToken, "").Code)
		assert.NotNil(t, entity.FindGrant(grant.GrantUID))

		require.NoError(t, admin.Updates(entity.Values{"UserRole": role, "SuperAdmin": superAdmin}))
		entity.FlushSessionCache()

		r := postTokenForm(app, "/api/v1/oauth/token", refreshForm(client.ClientUID, refreshToken), "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
	})
	t.Run("DerivedSessionsRevoked", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		accessToken, refreshToken := tokenPair(t, app, router, conf, client)

		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)
		admin.RevokeDerivedSessions(nil)

		assert.Equal(t, http.StatusUnauthorized, mcpPost(app, initBody, accessToken, "").Code)
		r := postTokenForm(app, "/api/v1/oauth/token", refreshForm(client.ClientUID, refreshToken), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("CodeIssuedBeforeRevocation", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		code := approvedCode(t, app, router, conf, client)

		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)
		admin.RevokeDerivedSessions(nil)

		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("AdminDeletesAppSession", func(t *testing.T) {
		app, router, conf, client := prepareAppTokenTest(t)
		DeleteSession(router)
		accessToken, _ := tokenPair(t, app, router, conf, client)
		adminToken := signIn(app, "admin", "photoprism")
		sess, err := entity.FindAnySession(rnd.SessionID(accessToken))
		require.NoError(t, err)
		grant := entity.FindGrant(sess.AuthID)
		require.NotNil(t, grant)

		allowSessionManagement(t)
		r := AuthenticatedRequest(app, http.MethodDelete, "/api/v1/sessions/"+sess.RefID, adminToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Nil(t, entity.FindGrant(grant.GrantUID))
	})
}

func TestDeleteUserGrant_KeepsOtherMCPSessions(t *testing.T) {
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`
	callBody := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_config_keys","arguments":{"query":"database","limit":5}}}`

	app, router, conf, client := prepareAppTokenTest(t)
	DeleteUserGrant(router)
	other, err := entity.AddClient(entityAppForm("Other App"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = other.Delete() })

	firstToken, _ := tokenPair(t, app, router, conf, client)

	q := testAuthorizeQuery(conf, other.ClientUID)
	id := authorizeRequestID(t, app, q)
	adminToken := signIn(app, "admin", "photoprism")
	r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":["mcp"]}`, adminToken)
	require.Equal(t, http.StatusOK, r.Code)
	loc, err := url.Parse(gjson.Get(r.Body.String(), "redirectUri").String())
	require.NoError(t, err)
	r = postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(other.ClientUID, loc.Query().Get("code")), "", "")
	require.Equal(t, http.StatusOK, r.Code)
	secondToken := gjson.Get(r.Body.String(), "access_token").String()

	w := mcpPost(app, initBody, firstToken, "")
	require.Equal(t, http.StatusOK, w.Code)
	firstSession := w.Header().Get("Mcp-Session-Id")
	w = mcpPost(app, initBody, secondToken, "")
	require.Equal(t, http.StatusOK, w.Code)
	secondSession := w.Header().Get("Mcp-Session-Id")

	grant := entity.FindUserClientGrant(entity.Admin.UserUID, client.ClientUID)
	require.NotNil(t, grant)
	r = AuthenticatedRequest(app, http.MethodDelete, "/api/v1/users/"+entity.Admin.UserUID+"/grants/"+grant.GrantUID, adminToken)
	require.Equal(t, http.StatusOK, r.Code)

	assert.Equal(t, http.StatusNotFound, mcpPost(app, callBody, adminToken, firstSession).Code)
	assert.Equal(t, http.StatusOK, mcpPost(app, callBody, secondToken, secondSession).Code)
}

func TestNewUserGrant(t *testing.T) {
	g := NewUserGrant(entity.Grant{GrantUID: "gs5gfen1bgxz7s9i", ClientID: "cs5gfen1bgxz7s9i", ClientName: "Claude",
		RedirectURI: "https://claude.ai/api/mcp/auth_callback", Scope: "mcp", Resource: "https://photos.example.com/api/v1/mcp"})
	assert.Equal(t, "gs5gfen1bgxz7s9i", g.UID)
	assert.Equal(t, "https://claude.ai", g.Origin)
	assert.Equal(t, "mcp", g.Scope)
}

func TestUserGrantsAuth(t *testing.T) {
	app, router, conf, _ := prepareAppTokenTest(t)
	FindUserGrants(router)
	adminToken := AuthenticateAdmin(app, router)

	t.Run("PublicMode", func(t *testing.T) {
		conf.Options().Public = true
		t.Cleanup(func() { conf.Options().Public = false })
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/users/"+entity.Admin.UserUID+"/grants", adminToken)
		assert.Equal(t, http.StatusNotFound, r.Code)
	})
	t.Run("OwnWithViewPermission", func(t *testing.T) {
		guestToken := signIn(app, "gandalf", "Gandalf123!")
		gandalf := entity.FindUserByName("gandalf")
		require.NotNil(t, gandalf)
		r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/users/"+gandalf.UserUID+"/grants", guestToken)
		assert.Equal(t, http.StatusOK, r.Code, r.Body.String())
	})
}

func TestMcpGrantSessions_GrantRevokedMeanwhile(t *testing.T) {
	initBody := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}`
	callBody := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_config_keys","arguments":{"query":"database","limit":5}}}`

	app, router, conf, client := prepareAppTokenTest(t)
	accessToken, _ := tokenPair(t, app, router, conf, client)
	adminToken := signIn(app, "admin", "photoprism")

	grant := entity.FindUserClientGrant(entity.Admin.UserUID, client.ClientUID)
	require.NotNil(t, grant)

	// Remove only the grant row, as if it was revoked while the session is being opened.
	require.NoError(t, entity.UnscopedDb().Where("grant_uid = ?", grant.GrantUID).Delete(&entity.Grant{}).Error)

	w := mcpPost(app, initBody, accessToken, "")
	require.Equal(t, http.StatusOK, w.Code)
	sessionID := w.Header().Get("Mcp-Session-Id")
	require.NotEmpty(t, sessionID)

	assert.Equal(t, http.StatusNotFound, mcpPost(app, callBody, adminToken, sessionID).Code)
}
