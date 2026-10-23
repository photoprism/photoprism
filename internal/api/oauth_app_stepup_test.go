package api

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// challengeScopes returns the scopes named by an insufficient_scope challenge.
func challengeScopes(t *testing.T, challenge string) []string {
	t.Helper()

	m := regexp.MustCompile(`scope="([^"]*)"`).FindStringSubmatch(challenge)
	require.Len(t, m, 2, challenge)

	return strings.Fields(m[1])
}

// appCode authorizes the client for the scopes, approves them with the admin token, and returns the authorization code.
func appCode(t *testing.T, app *gin.Engine, conf *config.Config, adminToken, clientID string, scopes []string) string {
	t.Helper()

	q := testAuthorizeQuery(conf, clientID)
	q.Set("scope", strings.Join(scopes, " "))
	id := authorizeRequestID(t, app, q)

	body, err := json.Marshal(map[string]any{"request": id, "scopes": scopes})
	require.NoError(t, err)

	r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", string(body), adminToken)
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())

	loc, err := url.Parse(gjson.Get(r.Body.String(), "redirectUri").String())
	require.NoError(t, err)

	return loc.Query().Get("code")
}

// appTokens authorizes the client for the scopes, approves them with the admin token, and returns the access and refresh tokens.
func appTokens(t *testing.T, app *gin.Engine, conf *config.Config, adminToken, clientID string, scopes []string) (string, string) {
	t.Helper()

	r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(clientID, appCode(t, app, conf, adminToken, clientID, scopes)), "", "")
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())
	assert.Equal(t, strings.Join(scopes, " "), gjson.Get(r.Body.String(), "scope").String())

	return gjson.Get(r.Body.String(), "access_token").String(), gjson.Get(r.Body.String(), "refresh_token").String()
}

func TestOAuthApp_StepUpWithDefaultClient(t *testing.T) {
	app, router, conf, _ := prepareAppTokenTest(t)

	// Register the app client as "photoprism clients add --name ... --redirect-uri ..." does, without a scope.
	flags := flag.NewFlagSet("test", flag.ContinueOnError)
	flags.String("name", "", "Usage")
	flags.String("scope", "", "Usage")
	flags.Var(cli.NewStringSlice(), "redirect-uri", "Usage")
	require.NoError(t, flags.Parse([]string{"--name=Default App", "--redirect-uri=" + testAppCallback}))

	f := form.AddClientFromCli(cli.NewContext(cli.NewApp(), flags, nil))
	require.Equal(t, "*", f.Scope())

	client, err := entity.AddClient(f)
	require.NoError(t, err)
	t.Cleanup(func() {
		entity.DeleteClientGrants(client.ClientUID)
		_ = client.Delete()
	})

	adminToken := AuthenticateAdmin(app, router)
	call := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search_media","arguments":{"limit":2}}}`

	// Connected with mcp only, the app is asked for the scopes the tool needs.
	token, _ := appTokens(t, app, conf, adminToken, client.ClientUID, []string{"mcp"})
	w := mcpPost(app, call, token, mcpInitSession(t, app, token))
	require.Equal(t, http.StatusForbidden, w.Code)

	challenge := w.Header().Get(header.WWWAuthenticate)
	assert.Contains(t, challenge, `error="insufficient_scope"`)
	scopes := challengeScopes(t, challenge)
	assert.Equal(t, []string{"mcp", "photos"}, scopes)

	// After approving the challenged scopes, the call succeeds in a new MCP session.
	token, refreshToken := appTokens(t, app, conf, adminToken, client.ClientUID, scopes)
	out := mcpCallTool(t, app, token, mcpInitSession(t, app, token), "search_media", map[string]any{"limit": 2})
	assert.Len(t, out.Get("items").Array(), 2)

	// A refresh keeps the approved scopes.
	r := postTokenForm(app, "/api/v1/oauth/token", refreshForm(client.ClientUID, refreshToken), "", "")
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())
	assert.Equal(t, "mcp photos", gjson.Get(r.Body.String(), "scope").String())
	token = gjson.Get(r.Body.String(), "access_token").String()
	out = mcpCallTool(t, app, token, mcpInitSession(t, app, token), "search_media", map[string]any{"limit": 2})
	assert.Len(t, out.Get("items").Array(), 2)
}

func TestOAuthAppToken_NarrowedClientScope(t *testing.T) {
	app, router, conf, client := prepareAppTokenTest(t)
	require.NoError(t, client.SetScope("*").Save())
	adminToken := AuthenticateAdmin(app, router)

	t.Run("Narrowed", func(t *testing.T) {
		code := appCode(t, app, conf, adminToken, client.ClientUID, []string{"mcp", "photos"})
		require.NoError(t, client.SetScope("mcp").Save())
		t.Cleanup(func() { _ = client.SetScope("*").Save() })

		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		assert.Equal(t, "mcp", gjson.Get(r.Body.String(), "scope").String())
	})
	t.Run("WithoutMCP", func(t *testing.T) {
		// A refused code keeps the user's current grant.
		appTokens(t, app, conf, adminToken, client.ClientUID, []string{"mcp"})
		admin := entity.FindUserByName("admin")
		require.NotNil(t, admin)
		prev := entity.FindUserClientGrant(admin.UserUID, client.ClientUID)
		require.NotNil(t, prev)

		code := appCode(t, app, conf, adminToken, client.ClientUID, []string{"mcp", "photos"})
		require.NoError(t, client.SetScope("photos").Save())
		t.Cleanup(func() { _ = client.SetScope("*").Save() })

		r := postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, code), "", "")
		assert.Equal(t, http.StatusBadRequest, r.Code)
		assert.Equal(t, "invalid_grant", gjson.Get(r.Body.String(), "error").String())
		assert.Equal(t, "scope no longer allowed", gjson.Get(r.Body.String(), "error_description").String())

		current := entity.FindUserClientGrant(admin.UserUID, client.ClientUID)
		require.NotNil(t, current)
		assert.Equal(t, prev.GrantUID, current.GrantUID)
	})
}

func TestChallengeScopes(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, []string{"mcp", "photos"}, challengeScopes(t, `Bearer error="insufficient_scope", scope="mcp photos"`))
	})
}
