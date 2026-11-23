package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/service/hub"
)

// testHubRequest records a request received by the test hub.
type testHubRequest struct {
	Method  string
	Path    string
	Serial  string
	Token   string
	Runtime string
}

// testHub is a local hub service that issues a valid key and records the requests it receives.
type testHub struct {
	mu       sync.Mutex
	url      string
	requests []testHubRequest
	status   atomic.Int32
	sponsor  atomic.Bool
}

// testHubCredentials are recorded hub responses from the hub package tests; both sessions have expired.
var testHubCredentials = struct{ Member, Sponsor [3]string }{
	Member: [3]string{
		"b32e9ccdc90eb7c0f6f1b9fbc82b8a2b0e993304",
		"5991ea36a9611e9e00a8360c10b91567",
		"3ef5685c6391a568731c8fc94ccad82d92dea60476c8b672990047c822248f45366fc0e8e812ad15e0b5ae1eb20e866235c56b",
	},
	Sponsor: [3]string{
		"0e159b773c6fb779c3bf6c8ba6e322abf559dbaf",
		"23f0024975bd65ade06edcc8191f7fcc",
		"bde6d0cf514e5456591de5ae09d981056eb88dccf71ba268974bf2cc7b028545e7641c1dfbaa716e4d13f8b0e0d1863e64c16e1f0ac551fc85e1171a87cbda6608cbe330de9e0d5f5b0e14ff61f2ff08fee369",
	},
}

// Sponsor makes the test hub answer with the recorded credentials of a sponsor.
func (h *testHub) Sponsor() {
	h.sponsor.Store(true)
}

// Fail makes the test hub answer with the given HTTP status code, or with a key again if it is zero.
func (h *testHub) Fail(status int) {
	h.status.Store(int32(status)) //nolint:gosec // HTTP status codes fit
}

// Requests returns the requests received so far.
func (h *testHub) Requests() []testHubRequest {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]testHubRequest(nil), h.requests...)
}

// Methods returns the methods of the requests received so far.
func (h *testHub) Methods() (result []string) {
	for _, r := range h.Requests() {
		result = append(result, r.Method)
	}

	return result
}

// newTestHub starts a local hub service, routes hub requests to it, and restores the defaults when the test ends.
// Call enable after creating a test config, since that disables hub requests.
func newTestHub(t *testing.T) (h *testHub, enable func()) {
	t.Helper()

	h = &testHub{}

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req hub.Request
		_ = json.NewDecoder(r.Body).Decode(&req)

		h.mu.Lock()
		h.requests = append(h.requests, testHubRequest{Method: r.Method, Path: r.URL.Path, Serial: req.ClientSerial, Token: req.ApiToken, Runtime: req.ClientRuntime})
		h.mu.Unlock()

		if status := int(h.status.Load()); status != 0 {
			w.WriteHeader(status)
			return
		}

		creds, status := testHubCredentials.Member, "unregistered"

		if h.sponsor.Load() {
			creds, status = testHubCredentials.Sponsor, "sponsor"
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"key":     creds[0],
			"secret":  creds[1],
			"session": creds[2],
			"status":  status,
			"serial":  req.ClientSerial,
		})
	}))

	transport := http.DefaultTransport
	http.DefaultTransport = srv.Client().Transport
	h.url = srv.URL + "/v1/hello"

	t.Cleanup(func() {
		http.DefaultTransport = transport
		srv.Close()
		hub.ApplyTestConfig()
	})

	return h, func() { hub.SetBaseURL(h.url) }
}

// startHubTestConfig creates a config for dataPath like a starting instance and initializes its hub config.
func startHubTestConfig(t *testing.T, dataPath string, enable func(), setup ...func(c *Config)) *Config {
	t.Helper()

	return initHubTestConfig(t, dataPath, enable, true, setup...)
}

// initHubTestConfig creates a config for dataPath like the start command or, if start is false, another
// command, and initializes its hub config.
func initHubTestConfig(t *testing.T, dataPath string, enable func(), start bool, setup ...func(c *Config)) *Config {
	t.Helper()

	c := NewMinimalTestConfig(dataPath)

	for _, fn := range setup {
		fn(c)
	}

	require.NoError(t, c.CreateDirectories())
	require.NoError(t, c.InitSerial())

	// Only the local test hub may receive requests, also where PHOTOPRISM_TEST_HUB selects another endpoint.
	t.Cleanup(hub.ApplyTestConfig)
	hub.Disable()
	enable()
	c.SetStarting(start)
	c.initHub()
	t.Cleanup(c.stopHubTicker)

	return c
}

// ageHubFile sets the modification time of the hub config file of c to the given age.
func ageHubFile(t *testing.T, c *Config, age time.Duration) {
	t.Helper()

	mtime := time.Now().Add(-age)
	require.NoError(t, os.Chtimes(c.HubConfigFile(), mtime, mtime))
}

func TestNewConfig_Reborn(t *testing.T) {
	orig := hubWasReborn
	t.Cleanup(func() { hubWasReborn = orig })

	// The mark is read once, so clearing it later does not change what the config recorded.
	reborn := true
	hubWasReborn = func() bool { return reborn }
	c := NewConfig(CliTestContext())
	reborn = false
	assert.True(t, c.reborn)

	assert.False(t, NewConfig(CliTestContext()).reborn)
}

func TestConfig_RenewApiKeysWithToken_Log(t *testing.T) {
	// A request that cannot connect is logged with the service host, not the request URL.
	_, enable := newTestHub(t)
	c := startHubTestConfig(t, t.TempDir(), enable)
	key := c.Hub().Key
	require.NotEmpty(t, key)

	hub.SetBaseURL("https://127.0.0.1:1/v1/hello")
	hook := captureConfigLog(t)
	require.NoError(t, c.RenewApiKeysWithToken(""))

	var logged []string

	for _, entry := range hook.AllEntries() {
		logged = append(logged, entry.Message)
	}

	require.NotEmpty(t, logged)
	assert.True(t, namesPath(logged, "request to 127.0.0.1:1 failed"), "%v", logged)
	assert.False(t, namesPath(logged, "/v1/hello"), "%v", logged)
}

func TestHubRenewalDue(t *testing.T) {
	// Whole seconds, so that the limit is exact on file systems that store coarse modification times.
	now := time.Now().Truncate(time.Second)
	fileName := t.TempDir() + "/hub.yml"

	t.Run("Missing", func(t *testing.T) {
		assert.True(t, hubRenewalDue(fileName, now))
	})

	require.NoError(t, os.WriteFile(fileName, []byte("Key: \n"), 0o600))

	for _, tc := range []struct {
		name string
		age  time.Duration
		due  bool
	}{
		{"Recent", time.Minute, false},
		{"JustWritten", 0, false},
		{"BeforeLimit", hubRenewAfter - time.Second, false},
		{"AtLimit", hubRenewAfter, true},
		{"Old", 24 * time.Hour, true},
		{"Future", -time.Minute, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mtime := now.Add(-tc.age)
			require.NoError(t, os.Chtimes(fileName, mtime, mtime))
			assert.Equal(t, tc.due, hubRenewalDue(fileName, now))
		})
	}
}

func TestTouchHubFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		fileName := t.TempDir() + "/hub.yml"
		require.NoError(t, os.WriteFile(fileName, []byte("Key: abc\n"), 0o600))
		old := time.Now().Add(-time.Hour)
		require.NoError(t, os.Chtimes(fileName, old, old))

		touchHubFile(fileName)

		info, err := os.Stat(fileName)
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), info.ModTime(), time.Minute)
		data, err := os.ReadFile(fileName) //nolint:gosec // test file
		require.NoError(t, err)
		assert.Equal(t, "Key: abc\n", string(data))
	})
	t.Run("Missing", func(t *testing.T) {
		fileName := t.TempDir() + "/hub.yml"
		touchHubFile(fileName)
		assert.NoFileExists(t, fileName)
	})
}

func TestConfig_HubUpdate(t *testing.T) {
	t.Run("NoKey", func(t *testing.T) {
		// A recently written file without a key, e.g. one saved while hub requests were disabled, still activates.
		_, enable := newTestHub(t)
		c := NewMinimalTestConfig(t.TempDir())
		enable()
		require.NoError(t, os.MkdirAll(c.ConfigPath(), 0o700))
		require.NoError(t, os.WriteFile(c.HubConfigFile(), []byte("Key: \n"), 0o600))
		c.hub = hub.NewConfig("test", c.HubConfigFile(), "", "", "", "")
		c.SetStarting(true)
		assert.True(t, c.hubUpdate(nil))
		c.SetStarting(false)
		assert.True(t, c.hubUpdate(nil))
	})
	t.Run("Disabled", func(t *testing.T) {
		c := NewMinimalTestConfig(t.TempDir())
		t.Cleanup(hub.ApplyTestConfig)
		hub.Disable()
		c.hub = hub.NewConfig("test", c.HubConfigFile(), "", "", "", "")
		c.SetStarting(true)
		assert.False(t, c.hubUpdate(nil))
	})
	t.Run("StartAndCommands", func(t *testing.T) {
		_, enable := newTestHub(t)
		c := NewMinimalTestConfig(t.TempDir())
		enable()
		require.NoError(t, os.MkdirAll(c.ConfigPath(), 0o700))
		c.hub = hub.NewConfig("test", c.HubConfigFile(), "", "", "", "")
		c.hub.Key = "b32e9ccdc90eb7c0f6f1b9fbc82b8a2b0e993304"
		c.hub.Secret = "5991ea36a9611e9e00a8360c10b91567"
		require.NoError(t, os.WriteFile(c.HubConfigFile(), []byte("Key: \n"), 0o600))
		expired := fmt.Errorf("session expired")

		// A start always renews; other commands only an expired session, and only once the window has passed.
		c.SetStarting(true)
		assert.True(t, c.hubUpdate(nil))
		assert.True(t, c.hubUpdate(expired))
		c.SetStarting(false)
		assert.False(t, c.hubUpdate(nil))
		assert.False(t, c.hubUpdate(expired))

		ageHubFile(t, c, time.Hour)
		assert.False(t, c.hubUpdate(nil))
		assert.True(t, c.hubUpdate(expired))
	})
	t.Run("DetachServer", func(t *testing.T) {
		// A start that only spawns the daemon leaves the request to the daemon, also without a key.
		_, enable := newTestHub(t)
		c := NewMinimalTestConfig(t.TempDir())
		enable()
		c.hub = hub.NewConfig("test", c.HubConfigFile(), "", "", "", "")
		c.options.DetachServer = true
		c.SetStarting(true)

		c.reborn = false
		assert.False(t, c.hubUpdate(nil))
		assert.True(t, c.detachingServer())

		c.reborn = true
		assert.True(t, c.hubUpdate(nil))
		assert.False(t, c.detachingServer())

		// A config report with detach-server set sends nothing either, since it is no daemon.
		c.reborn = false
		c.startReport = true
		assert.False(t, c.hubUpdate(nil))
		c.startReport = false

		c.SetStarting(false)
		assert.False(t, c.detachingServer())
		assert.True(t, c.hubUpdate(nil), "other commands without a key activate")
	})
	t.Run("MismatchedSecret", func(t *testing.T) {
		// A key that does not match its secret is discarded, so the instance activates again.
		_, enable := newTestHub(t)
		c := NewMinimalTestConfig(t.TempDir())
		enable()
		require.NoError(t, os.MkdirAll(c.ConfigPath(), 0o700))
		require.NoError(t, os.WriteFile(c.HubConfigFile(), []byte("Key: \n"), 0o600))
		c.hub = hub.NewConfig("test", c.HubConfigFile(), "", "", "", "")
		c.hub.Key = "b32e9ccdc90eb7c0f6f1b9fbc82b8a2b0e993304"
		c.hub.Secret = "not-the-secret-of-this-key"
		c.SetStarting(true)
		assert.True(t, c.hubUpdate(nil))
		assert.Empty(t, c.hub.Key)
	})
}

func TestConfig_InitHub(t *testing.T) {
	t.Run("Restart", func(t *testing.T) {
		// Every start renews the stored key, so a membership change applies on restart.
		h, enable := newTestHub(t)
		dataPath := t.TempDir()

		first := startHubTestConfig(t, dataPath, enable)
		require.NotEmpty(t, first.Hub().Key)
		second := startHubTestConfig(t, dataPath, enable)

		require.Equal(t, []string{http.MethodPost, http.MethodPut}, h.Methods())
		assert.Equal(t, first.Hub().Key, second.Hub().Key)
		assert.Equal(t, first.Serial(), h.Requests()[1].Serial)
		assert.True(t, strings.HasSuffix(h.Requests()[1].Path, "/"+second.Hub().Key))
	})
	t.Run("FutureModTime", func(t *testing.T) {
		// A modification time in the future counts as due, so a wrong clock cannot suppress renewals.
		h, enable := newTestHub(t)
		dataPath := t.TempDir()

		first := startHubTestConfig(t, dataPath, enable)
		ageHubFile(t, first, -time.Hour)
		initHubTestConfig(t, dataPath, enable, false)

		assert.Equal(t, []string{http.MethodPost, http.MethodPut}, h.Methods())
	})
	t.Run("FailedRenewal", func(t *testing.T) {
		// Starts renew also after a refused request; the attempt starts the window of other commands.
		h, enable := newTestHub(t)
		dataPath := t.TempDir()

		first := startHubTestConfig(t, dataPath, enable)
		key := first.Hub().Key
		ageHubFile(t, first, time.Hour)
		h.Fail(http.StatusForbidden)

		startHubTestConfig(t, dataPath, enable)
		third := startHubTestConfig(t, dataPath, enable)
		initHubTestConfig(t, dataPath, enable, false)

		assert.Equal(t, []string{http.MethodPost, http.MethodPut, http.MethodPut}, h.Methods())
		assert.Equal(t, key, third.Hub().Key)
		assert.False(t, hubRenewalDue(third.HubConfigFile(), time.Now()))
	})
	t.Run("NoKey", func(t *testing.T) {
		// Without a key, every start activates, as the editions need the result before the database setup.
		h, enable := newTestHub(t)
		h.Fail(http.StatusForbidden)
		dataPath := t.TempDir()

		startHubTestConfig(t, dataPath, enable)
		c := startHubTestConfig(t, dataPath, enable)

		assert.Equal(t, []string{http.MethodPost, http.MethodPost}, h.Methods())
		assert.NoFileExists(t, c.HubConfigFile())
	})
	t.Run("DisabledCommand", func(t *testing.T) {
		// A command that disables hub requests does not rewrite the file, so it does not start the window.
		h, enable := newTestHub(t)
		dataPath := t.TempDir()

		first := startHubTestConfig(t, dataPath, enable)
		ageHubFile(t, first, time.Hour)
		initHubTestConfig(t, dataPath, func() {}, false)
		assert.True(t, hubRenewalDue(first.HubConfigFile(), time.Now()))

		initHubTestConfig(t, dataPath, enable, false)
		assert.Equal(t, []string{http.MethodPost, http.MethodPut}, h.Methods())
	})
	t.Run("ConnectTokenAtStart", func(t *testing.T) {
		// A token added before a restart is sent at that restart.
		h, enable := newTestHub(t)
		dataPath := t.TempDir()

		startHubTestConfig(t, dataPath, enable)
		t.Setenv("PHOTOPRISM_CONNECT", "test-connect-token")
		startHubTestConfig(t, dataPath, enable)

		require.Equal(t, []string{http.MethodPost, http.MethodPut}, h.Methods())
		assert.Equal(t, "test-connect-token", h.Requests()[1].Token)
	})
	t.Run("ConnectTokenEveryStart", func(t *testing.T) {
		// A token that does not activate is sent again at every start.
		for _, tc := range []struct {
			name   string
			status int
		}{{"Refused", http.StatusForbidden}, {"NotSponsor", 0}} {
			t.Run(tc.name, func(t *testing.T) {
				h, enable := newTestHub(t)
				dataPath := t.TempDir()

				startHubTestConfig(t, dataPath, enable)
				t.Setenv("PHOTOPRISM_CONNECT", "test-connect-token")
				h.Fail(tc.status)

				for range 3 {
					startHubTestConfig(t, dataPath, enable)
				}

				require.Equal(t, []string{http.MethodPost, http.MethodPut, http.MethodPut, http.MethodPut}, h.Methods())
				for _, r := range h.Requests()[1:] {
					assert.Equal(t, "test-connect-token", r.Token)
				}
			})
		}
	})
	t.Run("ExplicitConnectFailure", func(t *testing.T) {
		// A failed explicit connect leaves the file as it is, so the window of other commands is unchanged.
		h, enable := newTestHub(t)
		dataPath := t.TempDir()

		first := startHubTestConfig(t, dataPath, enable)
		ageHubFile(t, first, time.Hour)
		h.Fail(http.StatusForbidden)
		require.Error(t, first.RenewApiKeysWithToken("test-connect-token"))
		assert.True(t, hubRenewalDue(first.HubConfigFile(), time.Now()))

		h.Fail(0)
		initHubTestConfig(t, dataPath, enable, false)
		assert.Equal(t, []string{http.MethodPost, http.MethodPut, http.MethodPut}, h.Methods())
	})
	t.Run("SaveFails", func(t *testing.T) {
		// A file that cannot be saved still gets the time of the attempt, so other commands wait for the window.
		if os.Geteuid() == 0 {
			t.Skip("root can write any file")
		}

		h, enable := newTestHub(t)
		dataPath := t.TempDir()

		first := startHubTestConfig(t, dataPath, enable)
		ageHubFile(t, first, time.Hour)
		require.NoError(t, os.Chmod(first.HubConfigFile(), 0o444)) //nolint:gosec // read-only on purpose
		t.Cleanup(func() { _ = os.Chmod(first.HubConfigFile(), 0o600) })

		startHubTestConfig(t, dataPath, enable)
		initHubTestConfig(t, dataPath, enable, false)

		assert.Equal(t, []string{http.MethodPost, http.MethodPut}, h.Methods())
		assert.False(t, hubRenewalDue(first.HubConfigFile(), time.Now()))
	})
	t.Run("Ticker", func(t *testing.T) {
		// RenewApiKeys is what the daily tick calls: it renews regardless of the window and starts it again.
		h, enable := newTestHub(t)
		c := startHubTestConfig(t, t.TempDir(), enable)

		c.RenewApiKeys()
		assert.Equal(t, []string{http.MethodPost, http.MethodPut}, h.Methods())

		// A refused renewal starts the window as well.
		ageHubFile(t, c, time.Hour)
		h.Fail(http.StatusForbidden)
		c.RenewApiKeys()
		assert.Equal(t, []string{http.MethodPost, http.MethodPut, http.MethodPut}, h.Methods())
		assert.False(t, hubRenewalDue(c.HubConfigFile(), time.Now()))
		h.Fail(0)

		// While hub requests are disabled, it neither requests nor rewrites the file.
		ageHubFile(t, c, time.Hour)
		hub.Disable()
		c.RenewApiKeys()
		assert.Len(t, h.Methods(), 3)
		assert.True(t, hubRenewalDue(c.HubConfigFile(), time.Now()))
	})
	t.Run("RecentFileWithoutKey", func(t *testing.T) {
		// A recently written file without a key, e.g. one saved while hub requests were disabled, still activates.
		h, enable := newTestHub(t)
		dataPath := t.TempDir()
		c := NewMinimalTestConfig(dataPath)
		require.NoError(t, os.MkdirAll(c.ConfigPath(), 0o700))
		require.NoError(t, os.WriteFile(c.HubConfigFile(), []byte("Key: \n"), 0o600))

		startHubTestConfig(t, dataPath, enable)
		assert.Equal(t, []string{http.MethodPost}, h.Methods())
	})
	t.Run("ConnectTokenCommand", func(t *testing.T) {
		// Other commands do not renew a loaded key within the window, also with the token in their environment.
		h, enable := newTestHub(t)
		dataPath := t.TempDir()

		startHubTestConfig(t, dataPath, enable)
		t.Setenv("PHOTOPRISM_CONNECT", "test-connect-token")
		initHubTestConfig(t, dataPath, enable, false)

		assert.Equal(t, []string{http.MethodPost}, h.Methods())
	})
	t.Run("ConnectTokenWhenSponsor", func(t *testing.T) {
		// Once the membership is active, a start renews without sending the token again.
		h, enable := newTestHub(t)
		h.Sponsor()
		dataPath := t.TempDir()
		t.Setenv("PHOTOPRISM_CONNECT", "test-connect-token")

		first := startHubTestConfig(t, dataPath, enable)
		require.True(t, first.Hub().Sponsor())
		startHubTestConfig(t, dataPath, enable)

		require.Equal(t, []string{http.MethodPost, http.MethodPut}, h.Methods())
		assert.Equal(t, "test-connect-token", h.Requests()[0].Token)
		assert.Empty(t, h.Requests()[1].Token)
	})
	t.Run("CommandWithoutKey", func(t *testing.T) {
		h, enable := newTestHub(t)
		c := initHubTestConfig(t, t.TempDir(), enable, false)

		assert.Equal(t, []string{http.MethodPost}, h.Methods())
		assert.NotEmpty(t, c.Hub().Key)
		assert.FileExists(t, c.HubConfigFile())
	})
	t.Run("CommandWithExpiredSession", func(t *testing.T) {
		h, enable := newTestHub(t)
		dataPath := t.TempDir()

		first := startHubTestConfig(t, dataPath, enable)
		initHubTestConfig(t, dataPath, enable, false)
		require.Equal(t, []string{http.MethodPost}, h.Methods(), "no renewal within the window")

		ageHubFile(t, first, time.Hour)
		initHubTestConfig(t, dataPath, enable, false)
		assert.Equal(t, []string{http.MethodPost, http.MethodPut}, h.Methods())
	})
	t.Run("ConnectCommandWithinWindow", func(t *testing.T) {
		// Activation with a token is explicit and does not wait for the window.
		h, enable := newTestHub(t)
		c := startHubTestConfig(t, t.TempDir(), enable)

		require.NoError(t, c.RenewApiKeysWithToken("test-connect-token"))
		require.Equal(t, []string{http.MethodPost, http.MethodPut}, h.Methods())
		assert.Equal(t, "test-connect-token", h.Requests()[1].Token)
	})
	t.Run("DetachServer", func(t *testing.T) {
		// Only the daemon renews, so a start with detach-server sends one request instead of two.
		h, enable := newTestHub(t)
		dataPath := t.TempDir()
		parentOpts := func(c *Config) { c.options.DetachServer = true }
		childOpts := func(c *Config) { c.options.DetachServer, c.reborn = true, true }

		parent := startHubTestConfig(t, dataPath, enable, parentOpts)
		assert.Empty(t, h.Methods())
		assert.NoFileExists(t, parent.HubConfigFile())

		child := startHubTestConfig(t, dataPath, enable, childOpts)
		assert.Equal(t, []string{http.MethodPost}, h.Methods())
		assert.NotEmpty(t, child.Hub().Key)

		startHubTestConfig(t, dataPath, enable, parentOpts)
		startHubTestConfig(t, dataPath, enable, childOpts)
		assert.Equal(t, []string{http.MethodPost, http.MethodPut}, h.Methods())
	})
	t.Run("Runtime", func(t *testing.T) {
		h, enable := newTestHub(t)
		t.Setenv("KUBERNETES_SERVICE_HOST", "10.96.0.1")
		c := startHubTestConfig(t, t.TempDir(), enable)

		require.Len(t, h.Requests(), 1)
		assert.Equal(t, "kubernetes", h.Requests()[0].Runtime)
		assert.Equal(t, "kubernetes", c.Hub().Runtime)
	})
	t.Run("ReadOnly", func(t *testing.T) {
		h, enable := newTestHub(t)
		c := startHubTestConfig(t, t.TempDir(), enable, func(c *Config) { c.options.ReadOnly = true })

		assert.True(t, c.ReadOnly())
		assert.Equal(t, []string{http.MethodPost}, h.Methods())
		assert.FileExists(t, c.HubConfigFile())
	})
}
