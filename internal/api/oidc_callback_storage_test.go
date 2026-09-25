package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
)

// authCallback is the script the auth template rendered, with what the client config resolved for it.
type authCallback struct {
	script    string
	namespace string
	keys      []string
	loginUri  string
}

// callbackStores holds the browser storage the rendered callback script left behind.
type callbackStores struct {
	Local    map[string]string `json:"localStorage"`
	Session  map[string]string `json:"sessionStorage"`
	Location string            `json:"location"`
}

// renderAuthCallback renders the auth template for the given status and returns its script together
// with the storage namespace, the key names the script clears, and the page it sends the browser to.
func renderAuthCallback(t *testing.T, status Event) authCallback {
	t.Helper()

	conf := config.TestConfig()
	app := gin.New()
	app.LoadHTMLFiles(conf.TemplateFiles()...)
	app.GET("/oidc-auth-template", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth.gohtml", gin.H{
			"status":       status,
			"session_id":   "sess1example",
			"access_token": "token1example",
			"provider":     "oidc",
			"error":        "invalid credentials",
			"user":         gin.H{"ID": 1, "Name": "alice", "DisplayName": "Alice"},
			"config":       conf.ClientPublic(),
		})
	})

	r := PerformRequest(app, http.MethodGet, "/oidc-auth-template")
	require.Equal(t, http.StatusOK, r.Code)

	body := r.Body.String()
	start := strings.LastIndex(body, "<script>")
	require.NotEqual(t, -1, start)
	end := strings.LastIndex(body, "</script>")
	require.Greater(t, end, start)

	return authCallback{
		script:    body[start+len("<script>") : end],
		namespace: conf.ClientPublic().StorageNamespace,
		keys:      sessionStorageKeys(t),
		loginUri:  conf.ClientPublic().LoginUri,
	}
}

// sessionStorageKeys returns the storage keys a signed-in app holds, read from session.js so the
// list cannot be satisfied by whatever the template names. The failed-sign-in keys are added
// because session.js reads those through the notification layer.
func sessionStorageKeys(t *testing.T) []string {
	t.Helper()

	src, err := os.ReadFile("../../frontend/src/common/session.js")
	require.NoError(t, err)

	keys := append([]string{"session.error", "session.messageId", "session.messageParams"}, adoptedStorageKeys(t)...)

	for _, m := range regexp.MustCompile(`storageKey \+ "\.([A-Za-z0-9_]+)"`).FindAllStringSubmatch(string(src), -1) {
		if key := "session." + m[1]; !slices.Contains(keys, key) {
			keys = append(keys, key)
		}
	}

	require.Contains(t, keys, "session.token")
	require.Contains(t, keys, "session.scope")

	return keys
}

// runAuthCallback executes the rendered callback script against the given storage contents and
// returns what the two stores hold afterwards. It skips the test where node is unavailable, since
// the script is browser code and nothing in the Go build interprets it.
func runAuthCallback(t *testing.T, script string, preset map[string]map[string]string) callbackStores {
	t.Helper()

	node, err := exec.LookPath("node")

	if err != nil {
		t.Skip("node is required to execute the rendered callback script")
	}

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "callback.js")
	presetPath := filepath.Join(dir, "preset.json")
	require.NoError(t, os.WriteFile(scriptPath, []byte(script), 0o600))
	presetJSON, err := json.Marshal(preset)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(presetPath, presetJSON, 0o600))

	cmd := exec.Command(node, "testdata/auth_callback_driver.js", scriptPath, presetPath) //nolint:gosec // fixed driver, test-owned arguments
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	require.NoError(t, err, "%s", stderr.String())

	var stores callbackStores
	require.NoError(t, json.Unmarshal(out, &stores))

	return stores
}

// callbackPreset returns one store holding a stale value under every key the callback clears, in
// both spellings, plus a setting and another instance's namespace that it must leave alone.
func callbackPreset(c authCallback) map[string]string {
	prefix := "pp:" + c.namespace + ":"
	preset := map[string]string{
		prefix + "settings":           "keep-me",
		"pp:othernamespace:authToken": "other-instance-token", //nolint:gosec // synthetic test value
	}

	for _, key := range c.keys {
		preset[key] = "stale-" + key
		preset[prefix+key] = "stale-" + prefix + key
	}

	return preset
}

// assertCallbackCleared checks that no stale value survived in either store and that the keys the
// app adopts unprefixed are gone from both.
func assertCallbackCleared(t *testing.T, c authCallback, stores callbackStores) {
	t.Helper()

	prefix := "pp:" + c.namespace + ":"

	for _, store := range []map[string]string{stores.Local, stores.Session} {
		for _, key := range c.keys {
			assert.NotContains(t, store, key, "the unprefixed spelling must be cleared")
			assert.NotEqual(t, "stale-"+prefix+key, store[prefix+key], "no stale value may survive")
		}

		assert.Equal(t, "keep-me", store[prefix+"settings"], "an unrelated key stays")
		assert.Equal(t, "other-instance-token", store["pp:othernamespace:authToken"], "another namespace stays")
	}
}

func TestAuthCallbackClearsBrowserStorage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c := renderAuthCallback(t, StatusSuccess)
		prefix := "pp:" + c.namespace + ":"
		stores := runAuthCallback(t, c.script, map[string]map[string]string{
			"localStorage":   callbackPreset(c),
			"sessionStorage": callbackPreset(c),
		})

		assertCallbackCleared(t, c, stores)

		// The new session replaces the stale one, in the store the preference selects.
		assert.Equal(t, "sess1example", stores.Local[prefix+"session.id"])
		assert.Equal(t, "token1example", stores.Local[prefix+"session.token"])
		assert.Equal(t, "oidc", stores.Local[prefix+"session.provider"])
		assert.NotContains(t, stores.Local, prefix+"session.error", "a stale error is gone")
		assert.NotContains(t, stores.Session, prefix+"session.id", "the session lands in one store only")
		assert.Equal(t, c.loginUri, stores.Location)
	})
	t.Run("Failed", func(t *testing.T) {
		c := renderAuthCallback(t, StatusFailed)
		prefix := "pp:" + c.namespace + ":"
		stores := runAuthCallback(t, c.script, map[string]map[string]string{
			"localStorage":   callbackPreset(c),
			"sessionStorage": callbackPreset(c),
		})

		assertCallbackCleared(t, c, stores)
		assert.NotContains(t, stores.Local, prefix+"session.token", "a failed sign-in leaves no token")
		assert.Equal(t, "invalid credentials", stores.Local[prefix+"session.error"])
		assert.Equal(t, c.loginUri, stores.Location)
	})
	t.Run("SessionStoragePreference", func(t *testing.T) {
		c := renderAuthCallback(t, StatusSuccess)
		prefix := "pp:" + c.namespace + ":"
		local := callbackPreset(c)
		local[prefix+"session"] = "true"
		stores := runAuthCallback(t, c.script, map[string]map[string]string{
			"localStorage":   local,
			"sessionStorage": callbackPreset(c),
		})

		assertCallbackCleared(t, c, stores)
		assert.Equal(t, "true", stores.Local[prefix+"session"], "the storage preference is preserved")
		assert.Equal(t, "sess1example", stores.Session[prefix+"session.id"], "the session follows the preference")
		assert.NotContains(t, stores.Local, prefix+"session.id")
	})
}
