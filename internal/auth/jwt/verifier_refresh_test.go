package jwt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jwksServer serves the given key set and counts the requests it answers.
func jwksServer(t *testing.T, keys []PublicJWK, requests *atomic.Int64) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(JWKS{Keys: keys}))
	}))

	t.Cleanup(server.Close)

	return server
}

// refreshTestVerifier returns a verifier for the endpoint, with a clock the test controls.
func refreshTestVerifier(t *testing.T, url string, now func() time.Time) *Verifier {
	t.Helper()

	cfg := newTestConfig(t)
	cfg.SetJWKSUrl(url)

	v := NewVerifier(cfg)
	v.now = now

	return v
}

func TestVerifierForcedRefreshIsBounded(t *testing.T) {
	_, keys := claimsTestSigner(t)

	t.Run("UnknownKeyIdsShareOneRefresh", func(t *testing.T) {
		var requests atomic.Int64
		server := jwksServer(t, keys, &requests)
		url := server.URL + "/.well-known/jwks.json"

		clock := time.Now()
		v := refreshTestVerifier(t, url, func() time.Time { return clock })
		ctx := context.Background()

		// Priming uses this interval's forced refresh, so the unknown key IDs presented after it are
		// answered from the cache.
		require.NoError(t, v.Prime(ctx, url))
		primed := requests.Load()

		for range 25 {
			_, err := v.publicKeyForKid(ctx, url, "kid-that-does-not-exist", true)
			assert.ErrorIs(t, err, errKeyNotFound)
		}

		assert.Equal(t, primed, requests.Load(), "unknown key ids must not each cost a fetch")

		// The endpoint is still reachable once the interval has passed.
		clock = clock.Add(jwksForceRefreshAfter + time.Second)
		_, err := v.publicKeyForKid(ctx, url, "kid-that-does-not-exist", true)
		assert.ErrorIs(t, err, errKeyNotFound)
		assert.Equal(t, primed+1, requests.Load())
	})
	t.Run("RotationIsDiscoveredAfterTheInterval", func(t *testing.T) {
		_, rotated := claimsTestSigner(t)
		require.NotEqual(t, keys[0].Kid, rotated[0].Kid)

		var requests atomic.Int64
		var served atomic.Bool

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			set := keys

			if served.Load() {
				set = rotated
			}

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(JWKS{Keys: set}))
		}))
		t.Cleanup(server.Close)

		url := server.URL + "/.well-known/jwks.json"
		clock := time.Now()
		v := refreshTestVerifier(t, url, func() time.Time { return clock })
		ctx := context.Background()

		require.NoError(t, v.Prime(ctx, url))
		before := requests.Load()

		// The Portal rotates: the new key id is not in the cached set yet.
		served.Store(true)

		// Inside the interval the endpoint is left alone and the new key stays unknown.
		clock = clock.Add(jwksForceRefreshAfter - time.Second)
		_, err := v.publicKeyForKid(ctx, url, rotated[0].Kid, true)
		assert.ErrorIs(t, err, errKeyNotFound)
		assert.Equal(t, before, requests.Load())

		// Past it, the rotated key is fetched and resolves.
		clock = clock.Add(2 * time.Second)
		pk, err := v.publicKeyForKid(ctx, url, rotated[0].Kid, true)
		assert.NoError(t, err)
		assert.NotEmpty(t, pk)
		assert.Equal(t, before+1, requests.Load())
	})
	t.Run("UnreachableEndpointRetriesOncePerInterval", func(t *testing.T) {
		var requests atomic.Int64

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		t.Cleanup(server.Close)

		url := server.URL + "/.well-known/jwks.json"
		clock := time.Now()
		v := refreshTestVerifier(t, url, func() time.Time { return clock })
		ctx := context.Background()

		for range 10 {
			_, err := v.publicKeyForKid(ctx, url, "unknown", true)
			assert.Error(t, err)
		}

		// One forced refresh runs its retry ladder; the rest are answered from the remembered failure.
		assert.Equal(t, int64(jwksFetchMaxRetries), requests.Load())

		// The endpoint is contacted again once the interval reopens, so a recovered Portal is found.
		clock = clock.Add(jwksForceRefreshAfter + jwksFetchRetryAfter + time.Second)
		_, err := v.publicKeyForKid(ctx, url, "unknown", true)
		assert.Error(t, err)
		assert.Equal(t, int64(2*jwksFetchMaxRetries), requests.Load())
	})
}

func TestVerifierSharedRefreshKeepsCachedKeysUsable(t *testing.T) {
	_, keys := claimsTestSigner(t)

	var requests atomic.Int64
	block := make(chan struct{})

	// The endpoint is primed once, then hangs: the state a Portal outage leaves behind.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) > 1 {
			<-block
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(JWKS{Keys: keys}))
	}))
	t.Cleanup(server.Close)

	url := server.URL + "/.well-known/jwks.json"
	clock := time.Now()
	v := refreshTestVerifier(t, url, func() time.Time { return clock })

	primed, err := v.keysForURL(context.Background(), url, false)
	require.NoError(t, err)
	require.NotEmpty(t, primed)

	// Age the cache past its TTL but well inside the window in which it stays usable.
	clock = clock.Add(defaultJWKSCacheTTL + time.Minute)

	// A forced refresh is in flight and will fail. An ordinary verification arriving behind it must
	// still be served the cached keys rather than the forced caller's error.
	forced := make(chan struct{})

	go func() {
		defer close(forced)
		_, _ = v.keysForURL(context.Background(), url, true)
	}()

	time.Sleep(100 * time.Millisecond)

	got, err := v.keysForURL(context.Background(), url, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, got, "an ordinary verification keeps its cached keys while a shared refresh fails")

	close(block)
	<-forced
}

func TestVerifierJoiningARefreshKeepsTheForcedSlot(t *testing.T) {
	_, keys := claimsTestSigner(t)

	var requests atomic.Int64
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		<-release
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(JWKS{Keys: keys}))
	}))
	t.Cleanup(server.Close)

	url := server.URL + "/.well-known/jwks.json"
	v := refreshTestVerifier(t, url, time.Now)

	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })

	// An ordinary refresh is in flight.
	ordinary := make(chan struct{})

	go func() {
		defer close(ordinary)
		_, _ = v.keysForURL(context.Background(), url, false)
	}()

	time.Sleep(100 * time.Millisecond)

	// A forced caller joins it rather than performing its own fetch, so the interval is still open
	// for the refresh that would actually look for a rotated key.
	joined := make(chan struct{})

	go func() {
		defer close(joined)
		_, _ = v.keysForURL(context.Background(), url, true)
	}()

	time.Sleep(100 * time.Millisecond)
	once.Do(func() { close(release) })
	<-ordinary
	<-joined

	assert.True(t, v.forcedRefreshDue(url), "joining a refresh in flight must not use the interval")
	assert.Equal(t, int64(1), requests.Load())
}

func TestVerifierCoalescesConcurrentRefreshes(t *testing.T) {
	_, keys := claimsTestSigner(t)

	var requests atomic.Int64
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		<-release
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(JWKS{Keys: keys}))
	}))
	t.Cleanup(server.Close)

	url := server.URL + "/.well-known/jwks.json"
	v := refreshTestVerifier(t, url, time.Now)

	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })

	var wg sync.WaitGroup
	errs := make([]error, 16)

	for i := range errs {
		wg.Add(1)

		go func() {
			defer wg.Done()
			_, errs[i] = v.keysForURL(context.Background(), url, false)
		}()
	}

	// Let the goroutines pile up on the in-flight request before it answers.
	time.Sleep(100 * time.Millisecond)
	once.Do(func() { close(release) })
	wg.Wait()

	for _, err := range errs {
		assert.NoError(t, err)
	}

	assert.Equal(t, int64(1), requests.Load(), "a burst must produce one request, not one per caller")
}

func TestVerifierRefreshSurvivesCallerCancellation(t *testing.T) {
	_, keys := claimsTestSigner(t)

	var requests atomic.Int64
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		<-release
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(JWKS{Keys: keys}))
	}))
	t.Cleanup(server.Close)

	url := server.URL + "/.well-known/jwks.json"
	v := refreshTestVerifier(t, url, time.Now)

	// Always let the handler finish, so a failing assertion cannot leave a goroutine blocked on it.
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)

	// One caller starts the refresh and then goes away. The fetch is detached from its context, so
	// the verification waiting on the same refresh is still served.
	leaving, cancel := context.WithCancel(context.Background())
	left := make(chan error, 1)

	go func() {
		_, err := v.keysForURL(leaving, url, false)
		left <- err
	}()

	time.Sleep(100 * time.Millisecond)

	staying := make(chan []PublicJWK, 1)

	go func() {
		got, _ := v.keysForURL(context.Background(), url, false)
		staying <- got
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	// The caller that left is released without waiting for the fetch to finish.
	select {
	case err := <-left:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("a cancelled caller must not be held for the detached fetch")
	}

	unblock()

	assert.NotEmpty(t, <-staying, "the remaining caller is still served")
	assert.Equal(t, int64(1), requests.Load())
}

func TestVerifierTakeForcedRefresh(t *testing.T) {
	clock := time.Now()
	v := &Verifier{now: func() time.Time { return clock }}

	t.Run("FirstIsAllowed", func(t *testing.T) {
		assert.True(t, v.takeForcedRefresh("https://portal.example.com/jwks.json"))
	})
	t.Run("RepeatInsideTheIntervalIsRefused", func(t *testing.T) {
		assert.False(t, v.takeForcedRefresh("https://portal.example.com/jwks.json"))
	})
	t.Run("OtherEndpointIsIndependent", func(t *testing.T) {
		assert.True(t, v.takeForcedRefresh("https://other.example.com/jwks.json"))
	})
	t.Run("AllowedAgainAfterTheInterval", func(t *testing.T) {
		clock = clock.Add(jwksForceRefreshAfter + time.Second)
		assert.True(t, v.takeForcedRefresh("https://portal.example.com/jwks.json"))
	})
}
