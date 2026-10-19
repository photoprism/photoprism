package oauth

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/safe"
)

const testMetadataURL = "https://claude.ai/oauth/mcp-client.json"

// testMetadata returns a valid client metadata document for the test URL.
func testMetadata() string {
	return `{"client_id":"` + testMetadataURL + `","client_name":"Claude","redirect_uris":["https://claude.ai/api/mcp/auth_callback"],"token_endpoint_auth_method":"none","grant_types":["authorization_code","refresh_token"],"response_types":["code"]}`
}

// jsonHeader returns the response header of a JSON document.
func jsonHeader() http.Header {
	return http.Header{"Content-Type": {"application/json; charset=utf-8"}}
}

// stubMetadata replaces the metadata fetcher for a test and counts its calls.
func stubMetadata(t *testing.T, body string, header http.Header, err error) *int {
	t.Helper()

	calls := 0
	prev := FetchMetadata
	FetchMetadata = func(string) ([]byte, http.Header, error) {
		calls++
		return []byte(body), header, err
	}

	FlushMetadataCache()
	t.Cleanup(func() {
		FetchMetadata = prev
		FlushMetadataCache()
	})

	return &calls
}

func TestParseClientMetadata(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		c, err := ParseClientMetadata(testMetadataURL, []byte(testMetadata()))
		require.NoError(t, err)
		assert.Equal(t, testMetadataURL, c.ID)
		assert.Equal(t, "Claude", c.Name)
		assert.True(t, c.Metadata)
		assert.Equal(t, []string{"https://claude.ai/api/mcp/auth_callback"}, c.RedirectURIs)
		assert.Equal(t, JoinScopes(AppScopes), c.Scope)
	})
	cases := map[string]string{
		"OtherClientID":   strings.Replace(testMetadata(), `"client_id":"`+testMetadataURL, `"client_id":"https://evil.example.com/c.json`, 1),
		"NoRedirectURIs":  `{"client_id":"` + testMetadataURL + `","redirect_uris":[]}`,
		"ScriptRedirect":  `{"client_id":"` + testMetadataURL + `","redirect_uris":["javascript:alert(1)"]}`,
		"NoCodeGrant":     `{"client_id":"` + testMetadataURL + `","redirect_uris":["https://claude.ai/cb"],"grant_types":["client_credentials"]}`,
		"NoCodeResponse":  `{"client_id":"` + testMetadataURL + `","redirect_uris":["https://claude.ai/cb"],"response_types":["token"]}`,
		"InvalidJSON":     `{"client_id":`,
		"TooManyRedirect": `{"client_id":"` + testMetadataURL + `","redirect_uris":[` + strings.Repeat(`"https://claude.ai/cb",`, MetadataMaxRedirects) + `"https://claude.ai/cb"]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ParseClientMetadata(testMetadataURL, []byte(body))
			assert.ErrorIs(t, err, ErrInvalidMetadata)
		})
	}
	t.Run("MaxRedirects", func(t *testing.T) {
		body := `{"client_id":"` + testMetadataURL + `","redirect_uris":[` + strings.Repeat(`"https://claude.ai/cb",`, MetadataMaxRedirects-1) + `"https://claude.ai/cb"]}`
		_, err := ParseClientMetadata(testMetadataURL, []byte(body))
		assert.NoError(t, err)
	})
	t.Run("RedirectURITooLong", func(t *testing.T) {
		long := "https://claude.ai/" + strings.Repeat("x", MaxRedirectURILength)
		_, err := ParseClientMetadata(testMetadataURL, []byte(`{"client_id":"`+testMetadataURL+`","redirect_uris":["`+long+`"]}`))
		assert.ErrorIs(t, err, ErrInvalidMetadata)
	})
	t.Run("ConfidentialClient", func(t *testing.T) {
		_, err := ParseClientMetadata(testMetadataURL, []byte(`{"client_id":"`+testMetadataURL+`","redirect_uris":["https://claude.ai/cb"],"token_endpoint_auth_method":"private_key_jwt"}`))
		assert.ErrorIs(t, err, ErrInvalidClient)
	})
	t.Run("Scope", func(t *testing.T) {
		c, err := ParseClientMetadata(testMetadataURL, []byte(`{"client_id":"`+testMetadataURL+`","redirect_uris":["https://claude.ai/cb"],"scope":"mcp photos password"}`))
		require.NoError(t, err)
		assert.Equal(t, "mcp photos", c.Scope)
	})
}

func TestMetadataClientName(t *testing.T) {
	assert.Equal(t, "Claude", metadataClientName("  Claude ", testMetadataURL))
	assert.Equal(t, "Clau de", metadataClientName("Clau\u202ede ", testMetadataURL))
	assert.Equal(t, "claude.ai", metadataClientName("", testMetadataURL))
	assert.Equal(t, "claude.ai", metadataClientName("\u3164\u2800 \uFFA0", testMetadataURL))
	assert.Equal(t, "claude.ai", metadataClientName("--", testMetadataURL))
	assert.Equal(t, "claude.ai", metadataClientName("Claude https://claude.ai", testMetadataURL))
	assert.Equal(t, "Claude 3", metadataClientName("Claude\u3164 3", testMetadataURL))
	assert.Len(t, []rune(metadataClientName(strings.Repeat("x", 300), testMetadataURL)), 200)
}

func TestMetadataClientScope(t *testing.T) {
	assert.Equal(t, JoinScopes(AppScopes), metadataClientScope(""))
	assert.Equal(t, "mcp albums", metadataClientScope("mcp albums albums users"))
	assert.Equal(t, "", metadataClientScope("users"))
}

func TestMetadataCacheAge(t *testing.T) {
	assert.Equal(t, MetadataCacheDefaultAge, metadataCacheAge(http.Header{}))
	assert.Equal(t, 2*time.Hour, metadataCacheAge(http.Header{"Cache-Control": {"public, max-age=7200"}}))
	assert.Equal(t, MetadataCacheMinAge, metadataCacheAge(http.Header{"Cache-Control": {"max-age=1"}}))
	assert.Equal(t, MetadataCacheMaxAge, metadataCacheAge(http.Header{"Cache-Control": {"max-age=99999999"}}))
	assert.Equal(t, MetadataCacheMinAge, metadataCacheAge(http.Header{"Cache-Control": {"no-store"}}))
	assert.Equal(t, MetadataCacheMinAge, metadataCacheAge(http.Header{"Cache-Control": {"no-store, max-age=86400"}}))
	assert.Equal(t, MetadataCacheMinAge, metadataCacheAge(http.Header{"Cache-Control": {"max-age=86400, no-cache"}}))
}

func TestMetadataFetchOptions(t *testing.T) {
	opt := metadataFetchOptions()
	assert.Equal(t, MetadataTimeout, opt.Timeout)
	assert.Equal(t, int64(MetadataMaxBytes), opt.MaxSizeBytes)
	assert.False(t, opt.AllowPrivate)
	assert.True(t, opt.NoRedirects)
	assert.Equal(t, "application/json", opt.Accept)
}

func TestFetchMetadata(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(testMetadata()))
	}))
	t.Cleanup(ts.Close)

	_, _, err := FetchMetadata(ts.URL + "/client.json")
	assert.ErrorIs(t, err, safe.ErrPrivateIP)
}

// metadataCacheIDs returns whether each of the client IDs is cached.
func metadataCacheIDs(ids ...string) []bool {
	metadataCache.Lock()
	defer metadataCache.Unlock()

	result := make([]bool, len(ids))

	for i, id := range ids {
		_, result[i] = metadataCache.items[id]
	}

	return result
}

func TestCacheMetadataClient(t *testing.T) {
	t.Run("EvictsEarliestExpiry", func(t *testing.T) {
		FlushMetadataCache()
		t.Cleanup(FlushMetadataCache)

		now := time.Now()

		// The first entry expires last, so insertion order cannot explain which entry is evicted.
		cacheMetadataClient("https://claude.ai/latest", &Client{}, now.Add(24*time.Hour))

		for i := 1; i < MetadataCacheMaxItems; i++ {
			cacheMetadataClient("https://claude.ai/"+time.Duration(i).String(), &Client{}, now.Add(time.Hour+time.Duration(i)*time.Minute))
		}

		cacheMetadataClient("https://claude.ai/new", &Client{}, now.Add(time.Hour))

		assert.Len(t, metadataCache.items, MetadataCacheMaxItems)
		assert.Equal(t, []bool{true, false, true, true}, metadataCacheIDs("https://claude.ai/latest", "https://claude.ai/1ns", "https://claude.ai/2ns", "https://claude.ai/new"))
	})
	t.Run("RemovesPastStaleAge", func(t *testing.T) {
		FlushMetadataCache()
		t.Cleanup(FlushMetadataCache)

		now := time.Now()

		for i := 0; i < MetadataCacheMaxItems; i++ {
			cacheMetadataClient("https://claude.ai/"+time.Duration(i).String(), &Client{}, now.Add(time.Hour+time.Duration(i)*time.Minute))
		}

		// Two entries past their stale age are removed, so the full cache shrinks by one with the new entry.
		metadataCache.Lock()
		metadataCache.items["https://claude.ai/0s"] = metadataEntry{client: &Client{}, expires: now.Add(-MetadataCacheStaleAge)}
		metadataCache.items["https://claude.ai/1ns"] = metadataEntry{client: &Client{}, expires: now.Add(-MetadataCacheStaleAge - time.Hour)}
		metadataCache.Unlock()

		cacheMetadataClient("https://claude.ai/new", &Client{}, now.Add(time.Hour))

		assert.Len(t, metadataCache.items, MetadataCacheMaxItems-1)
		assert.Equal(t, []bool{false, false, true, true}, metadataCacheIDs("https://claude.ai/0s", "https://claude.ai/1ns", "https://claude.ai/2ns", "https://claude.ai/new"))
	})
	t.Run("KeepsStale", func(t *testing.T) {
		FlushMetadataCache()
		t.Cleanup(FlushMetadataCache)

		now := time.Now()

		cacheMetadataClient("https://claude.ai/stale", &Client{}, now.Add(-time.Hour))
		cacheMetadataClient("https://claude.ai/new", &Client{}, now.Add(time.Hour))

		assert.Equal(t, []bool{true, true}, metadataCacheIDs("https://claude.ai/stale", "https://claude.ai/new"))
	})
}

func TestFindMetadataClient(t *testing.T) {
	t.Run("Cached", func(t *testing.T) {
		calls := stubMetadata(t, testMetadata(), jsonHeader(), nil)

		c, err := FindMetadataClient(testMetadataURL, ClientLookup{})
		require.NoError(t, err)
		assert.Equal(t, "Claude", c.Name)

		_, err = FindMetadataClient(testMetadataURL, ClientLookup{})
		require.NoError(t, err)
		assert.Equal(t, 1, *calls)
	})
	t.Run("FetchError", func(t *testing.T) {
		stubMetadata(t, "", nil, errors.New("connection to private or loopback address not allowed"))
		_, err := FindMetadataClient(testMetadataURL, ClientLookup{})
		assert.ErrorIs(t, err, ErrMetadataUnavailable)
	})
	t.Run("Expired", func(t *testing.T) {
		calls := stubMetadata(t, `{}`, jsonHeader(), nil)
		cacheMetadataClient(testMetadataURL, &Client{ID: testMetadataURL}, time.Now().Add(-time.Second))

		_, err := FindMetadataClient(testMetadataURL, ClientLookup{})
		assert.ErrorIs(t, err, ErrInvalidMetadata)
		assert.Equal(t, 1, *calls)
		assert.Equal(t, []bool{false}, metadataCacheIDs(testMetadataURL))
	})
	t.Run("StaleWhileUnavailable", func(t *testing.T) {
		calls := stubMetadata(t, "", nil, errors.New("timeout"))
		cacheMetadataClient(testMetadataURL, &Client{ID: testMetadataURL}, time.Now().Add(-time.Second))
		stale := ClientLookup{AllowStale: true}

		_, err := FindMetadataClient(testMetadataURL, ClientLookup{})
		assert.ErrorIs(t, err, ErrMetadataUnavailable)

		c, err := FindMetadataClient(testMetadataURL, stale)
		require.NoError(t, err)
		assert.Equal(t, testMetadataURL, c.ID)
		assert.Equal(t, 2, *calls)

		c, err = FindMetadataClient(testMetadataURL, ClientLookup{AllowFetch: func() bool { return false }, AllowStale: true})
		require.NoError(t, err)
		assert.Equal(t, testMetadataURL, c.ID)
		assert.Equal(t, 2, *calls)

		metadataCache.Lock()
		metadataCache.items[testMetadataURL] = metadataEntry{client: &Client{}, expires: time.Now().Add(-MetadataCacheStaleAge)}
		metadataCache.Unlock()

		_, err = FindMetadataClient(testMetadataURL, stale)
		assert.ErrorIs(t, err, ErrMetadataUnavailable)
	})
	t.Run("Withdrawn", func(t *testing.T) {
		for _, code := range []int{http.StatusNotFound, http.StatusGone, http.StatusForbidden} {
			stubMetadata(t, "", nil, &safe.StatusError{Code: code})
			cacheMetadataClient(testMetadataURL, &Client{ID: testMetadataURL}, time.Now().Add(-time.Second))

			_, err := FindMetadataClient(testMetadataURL, ClientLookup{AllowStale: true})
			assert.ErrorIs(t, err, ErrMetadataUnavailable)
			assert.Equal(t, []bool{false}, metadataCacheIDs(testMetadataURL))
		}
	})
	t.Run("ServerError", func(t *testing.T) {
		for _, code := range []int{http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusBadGateway} {
			stubMetadata(t, "", nil, &safe.StatusError{Code: code})
			cacheMetadataClient(testMetadataURL, &Client{ID: testMetadataURL}, time.Now().Add(-time.Second))

			_, err := FindMetadataClient(testMetadataURL, ClientLookup{AllowStale: true})
			assert.NoError(t, err)
		}
	})
	t.Run("ContentType", func(t *testing.T) {
		stubMetadata(t, testMetadata(), http.Header{"Content-Type": {"text/plain"}}, nil)
		_, err := FindMetadataClient(testMetadataURL, ClientLookup{})
		assert.ErrorIs(t, err, ErrInvalidMetadata)

		stubMetadata(t, testMetadata(), http.Header{"Content-Type": {"application/client-metadata+json"}}, nil)
		_, err = FindMetadataClient(testMetadataURL, ClientLookup{})
		assert.NoError(t, err)
	})
	t.Run("AllowFetch", func(t *testing.T) {
		calls := stubMetadata(t, testMetadata(), jsonHeader(), nil)
		asked := 0
		allow := func() bool { asked++; return asked == 1 }

		_, err := FindMetadataClient(testMetadataURL, ClientLookup{AllowFetch: allow})
		require.NoError(t, err)
		_, err = FindMetadataClient(testMetadataURL, ClientLookup{AllowFetch: allow})
		require.NoError(t, err)
		assert.Equal(t, 1, asked)

		FlushMetadataCache()
		_, err = FindMetadataClient(testMetadataURL, ClientLookup{AllowFetch: allow})
		assert.ErrorIs(t, err, ErrMetadataRateLimited)
		assert.Equal(t, 1, *calls)
	})
	t.Run("ErrorsNotCached", func(t *testing.T) {
		calls := stubMetadata(t, `{}`, jsonHeader(), nil)
		_, err := FindMetadataClient(testMetadataURL, ClientLookup{})
		assert.Error(t, err)
		_, err = FindMetadataClient(testMetadataURL, ClientLookup{})
		assert.Error(t, err)
		assert.Equal(t, 2, *calls)
	})
}

func TestFindClient_Metadata(t *testing.T) {
	conf := testAppConfig(t)

	t.Run("RegisteredRefusesWithoutFetch", func(t *testing.T) {
		calls := stubMetadata(t, testMetadata(), jsonHeader(), nil)
		conf.Options().OAuthApps = config.OAuthAppsRegistered
		_, err := FindClient(conf, testMetadataURL)
		assert.ErrorIs(t, err, ErrInvalidClient)
		assert.Equal(t, 0, *calls)
	})
	t.Run("Any", func(t *testing.T) {
		calls := stubMetadata(t, testMetadata(), jsonHeader(), nil)
		conf.Options().OAuthApps = config.OAuthAppsAny
		c, err := FindClient(conf, testMetadataURL)
		require.NoError(t, err)
		assert.True(t, c.Metadata)
		assert.Equal(t, 1, *calls)
	})
	t.Run("OriginNotListed", func(t *testing.T) {
		calls := stubMetadata(t, testMetadata(), jsonHeader(), nil)
		conf.Options().OAuthApps = "https://chatgpt.com"
		_, err := FindClient(conf, testMetadataURL)
		assert.ErrorIs(t, err, ErrInvalidClient)
		assert.Equal(t, 0, *calls)
	})
	t.Run("OriginListed", func(t *testing.T) {
		stubMetadata(t, testMetadata(), jsonHeader(), nil)
		conf.Options().OAuthApps = "https://claude.ai"
		_, err := FindClient(conf, testMetadataURL)
		assert.NoError(t, err)
	})
}

func TestClientOrigin(t *testing.T) {
	assert.Equal(t, "https://claude.ai", ClientOrigin(testMetadataURL, "https://other.example.com/cb"))
	assert.Equal(t, "https://other.example.com", ClientOrigin("cs5gfen1bgxz7s9i", "https://other.example.com/cb"))
}

func TestFlushMetadataCache(t *testing.T) {
	cacheMetadataClient(testMetadataURL, &Client{}, time.Now().Add(time.Hour))
	FlushMetadataCache()
	metadataCache.Lock()
	n := len(metadataCache.items)
	metadataCache.Unlock()
	assert.Zero(t, n)
}

func TestWithdrawnMetadata(t *testing.T) {
	assert.True(t, withdrawnMetadata(&safe.StatusError{Code: http.StatusNotFound}))
	assert.True(t, withdrawnMetadata(fmt.Errorf("fetch: %w", &safe.StatusError{Code: http.StatusGone})))
	assert.False(t, withdrawnMetadata(&safe.StatusError{Code: http.StatusTooManyRequests}))
	assert.False(t, withdrawnMetadata(&safe.StatusError{Code: http.StatusRequestTimeout}))
	assert.False(t, withdrawnMetadata(&safe.StatusError{Code: http.StatusServiceUnavailable}))
	assert.False(t, withdrawnMetadata(errors.New("timeout")))
}

func TestUncacheMetadataClient(t *testing.T) {
	FlushMetadataCache()
	t.Cleanup(FlushMetadataCache)

	cacheMetadataClient(testMetadataURL, &Client{}, time.Now().Add(time.Hour))
	uncacheMetadataClient(testMetadataURL)
	uncacheMetadataClient(testMetadataURL)
	assert.Equal(t, []bool{false}, metadataCacheIDs(testMetadataURL))
}
