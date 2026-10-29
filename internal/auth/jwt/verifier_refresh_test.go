package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/photoprism/photoprism/pkg/rnd"
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
	arrived := make(chan struct{})
	block := make(chan struct{})

	var release sync.Once
	unblock := func() { release.Do(func() { close(block) }) }

	// The endpoint is primed once, then stalls until released and fails: the state a Portal outage
	// leaves behind. The second request signals that the forced refresh is in flight.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if n := requests.Add(1); n > 1 {
			if n == 2 {
				close(arrived)
			}

			<-block
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(JWKS{Keys: keys}))
	}))
	t.Cleanup(server.Close)
	t.Cleanup(unblock)

	url := server.URL + "/.well-known/jwks.json"
	clock := time.Now()
	v := refreshTestVerifier(t, url, func() time.Time { return clock })

	// A long stale wait keeps the ordinary call on the failure fallback this test is about.
	prevWait := jwksStaleWait
	jwksStaleWait = time.Minute
	t.Cleanup(func() { jwksStaleWait = prevWait })

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

	select {
	case <-arrived:
	case <-time.After(5 * time.Second):
		t.Fatal("forced refresh did not reach the endpoint")
	}

	type result struct {
		keys []PublicJWK
		err  error
	}

	ordinary := make(chan result, 1)

	go func() {
		got, getErr := v.keysForURL(context.Background(), url, false)
		ordinary <- result{got, getErr}
	}()

	// The refresh stays stalled long enough for the ordinary call to join it, then fails at once.
	time.Sleep(100 * time.Millisecond)
	unblock()

	select {
	case r := <-ordinary:
		assert.NoError(t, r.err)
		assert.NotEmpty(t, r.keys, "an ordinary verification keeps its cached keys while a shared refresh fails")
	case <-time.After(10 * time.Second):
		t.Fatal("ordinary verification did not return")
	}

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

// TestVerifierStaleWait checks how long an ordinary verification waits for a refresh while its cached
// keys are past their TTL: briefly while they are still usable, and for the refresh once they are not.
func TestVerifierStaleWait(t *testing.T) {
	prev := jwksStaleWait
	jwksStaleWait = 50 * time.Millisecond
	t.Cleanup(func() { jwksStaleWait = prev })

	// stalledServer serves keys once, then holds every further request until released and fails it.
	stalledServer := func(t *testing.T, keys []PublicJWK) (url string, release func()) {
		t.Helper()

		var requests atomic.Int64
		block := make(chan struct{})

		var once sync.Once
		release = func() { once.Do(func() { close(block) }) }

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
		t.Cleanup(release)

		return server.URL + "/.well-known/jwks.json", release
	}

	t.Run("UsableKeysServedWhileRefreshStalls", func(t *testing.T) {
		_, keys := claimsTestSigner(t)
		url, release := stalledServer(t, keys)
		clock := time.Now()
		v := refreshTestVerifier(t, url, func() time.Time { return clock })

		_, err := v.keysForURL(context.Background(), url, false)
		require.NoError(t, err)

		clock = clock.Add(defaultJWKSCacheTTL + time.Minute)

		start := time.Now()
		got, err := v.keysForURL(context.Background(), url, false)

		assert.NoError(t, err)
		assert.Equal(t, keys, got)
		assert.Less(t, time.Since(start), 5*time.Second, "the stalled refresh must not hold the verification")

		// The refresh keeps running after the early return and records its failure once released.
		release()
		assert.Eventually(t, func() bool { return v.recentFetchFailure(url) != nil }, 10*time.Second, 10*time.Millisecond)
	})
	t.Run("RefreshInTimeDecides", func(t *testing.T) {
		_, oldKeys := claimsTestSigner(t)
		_, newKeys := claimsTestSigner(t)
		require.NotEqual(t, oldKeys[0].Kid, newKeys[0].Kid)

		var requests atomic.Int64

		// The first fetch serves the old key, every later one only the new key.
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			keys := newKeys

			if requests.Add(1) == 1 {
				keys = oldKeys
			}

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(JWKS{Keys: keys}))
		}))
		t.Cleanup(server.Close)

		url := server.URL + "/.well-known/jwks.json"
		clock := time.Now()
		v := refreshTestVerifier(t, url, func() time.Time { return clock })

		// The wait must not cut a refresh short that answers in time.
		prevWait := jwksStaleWait
		jwksStaleWait = 5 * time.Second
		t.Cleanup(func() { jwksStaleWait = prevWait })

		_, err := v.keysForURL(context.Background(), url, false)
		require.NoError(t, err)

		clock = clock.Add(defaultJWKSCacheTTL + time.Minute)

		got, err := v.keysForURL(context.Background(), url, false)

		assert.NoError(t, err)
		assert.Equal(t, newKeys, got, "a key the refresh removed is not served")
	})
	t.Run("ForcedCallerWaitsForRefresh", func(t *testing.T) {
		// A caller looking for a key the cache lacks waits for the refresh in flight, also when the
		// forced interval is used up and its own call counts as ordinary.
		_, keys := claimsTestSigner(t)
		url, release := stalledServer(t, keys)
		clock := time.Now()
		v := refreshTestVerifier(t, url, func() time.Time { return clock })

		_, err := v.keysForURL(context.Background(), url, false)
		require.NoError(t, err)

		clock = clock.Add(defaultJWKSCacheTTL + time.Minute)

		first := make(chan struct{})

		go func() {
			defer close(first)
			_, _ = v.keysForURL(context.Background(), url, true)
		}()

		require.Eventually(t, func() bool { return !v.forcedRefreshDue(url) }, 5*time.Second, 5*time.Millisecond)

		done := make(chan error, 1)

		go func() {
			_, getErr := v.keysForURL(context.Background(), url, true)
			done <- getErr
		}()

		select {
		case <-done:
			t.Fatal("returned before the refresh in flight finished")
		case <-time.After(10 * jwksStaleWait):
		}

		release()

		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("verification did not return")
		}

		<-first
	})
	t.Run("UnusableKeysWaitForRefresh", func(t *testing.T) {
		_, keys := claimsTestSigner(t)
		url, release := stalledServer(t, keys)
		clock := time.Now()
		v := refreshTestVerifier(t, url, func() time.Time { return clock })

		_, err := v.keysForURL(context.Background(), url, false)
		require.NoError(t, err)

		// Past the window in which cached keys may still be served.
		clock = clock.Add(defaultJWKSCacheTTL + RotationOverlap() + time.Minute)

		done := make(chan error, 1)

		go func() {
			_, getErr := v.keysForURL(context.Background(), url, false)
			done <- getErr
		}()

		select {
		case <-done:
			t.Fatal("returned before the refresh finished")
		case <-time.After(10 * jwksStaleWait):
		}

		release()

		select {
		case getErr := <-done:
			assert.Error(t, getErr)
		case <-time.After(10 * time.Second):
			t.Fatal("verification did not return")
		}
	})
}

func TestHasKid(t *testing.T) {
	keys := []PublicJWK{{Kid: "a"}, {Kid: "b"}}

	t.Run("Found", func(t *testing.T) {
		assert.True(t, hasKid(keys, "b"))
	})
	t.Run("Missing", func(t *testing.T) {
		assert.False(t, hasKid(keys, "c"))
		assert.False(t, hasKid(nil, "a"))
	})
	t.Run("NoKid", func(t *testing.T) {
		assert.True(t, hasKid(keys, ""))
		assert.True(t, hasKid(nil, ""))
	})
}

// TestVerifierNewKeyWaitsForRefresh checks that a token signed with a key the cache lacks waits for the
// refresh in flight rather than being refused when the stale wait ends.
func TestVerifierNewKeyWaitsForRefresh(t *testing.T) {
	prev := jwksStaleWait
	jwksStaleWait = 50 * time.Millisecond
	t.Cleanup(func() { jwksStaleWait = prev })

	clusterUUID := rnd.UUIDv7()

	// newSigner returns a key manager with its own active key for the cluster.
	newSigner := func() *Manager {
		cfg := newTestConfig(t)
		cfg.Options().ClusterUUID = clusterUUID
		mgr, err := NewManager(cfg)
		require.NoError(t, err)
		_, err = mgr.EnsureActiveKey()
		require.NoError(t, err)
		return mgr
	}

	oldSigner, newSigner2 := newSigner(), newSigner()
	require.NotEqual(t, oldSigner.JWKS().Keys[0].Kid, newSigner2.JWKS().Keys[0].Kid)

	var requests atomic.Int64
	block := make(chan struct{})

	var once sync.Once
	release := func() { once.Do(func() { close(block) }) }

	// The first fetch serves the old key; later ones wait until released, then serve the new key.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jwks := oldSigner.JWKS()

		if requests.Add(1) > 1 {
			<-block
			jwks = newSigner2.JWKS()
		}

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(jwks))
	}))
	t.Cleanup(server.Close)
	t.Cleanup(release)

	nodeCfg := newTestConfig(t)
	nodeCfg.SetJWKSUrl(server.URL + "/.well-known/jwks.json")
	nodeCfg.Options().ClusterUUID = clusterUUID

	spec := ClaimsSpec{
		Issuer:   fmt.Sprintf("portal:%s", clusterUUID),
		Subject:  "portal:client-test",
		Audience: fmt.Sprintf("node:%s", nodeCfg.NodeUUID()),
		Scope:    []string{"cluster"},
	}
	expected := ExpectedClaims{Issuer: spec.Issuer, Audience: spec.Audience, JWKSURL: nodeCfg.JWKSUrl()}

	ctx := context.Background()
	v := NewVerifier(nodeCfg)
	require.NoError(t, v.Prime(ctx, nodeCfg.JWKSUrl()))

	// Age the cache past its TTL but inside the window in which its keys stay usable.
	v.mu.Lock()
	v.cache.FetchedAt -= int64((time.Duration(nodeCfg.JWKSCacheTTL())*time.Second + time.Minute) / time.Second)
	v.mu.Unlock()

	// A forced refresh is in flight and holds the forced interval.
	forced := make(chan struct{})

	go func() {
		defer close(forced)
		_, _ = v.keysForURL(ctx, nodeCfg.JWKSUrl(), true)
	}()

	require.Eventually(t, func() bool { return !v.forcedRefreshDue(nodeCfg.JWKSUrl()) }, 5*time.Second, 5*time.Millisecond)

	token, err := NewIssuer(newSigner2).Issue(spec)
	require.NoError(t, err)

	done := make(chan error, 1)

	go func() {
		_, verifyErr := v.VerifyToken(ctx, token, expected)
		done <- verifyErr
	}()

	select {
	case verifyErr := <-done:
		t.Fatalf("returned before the refresh in flight finished: %v", verifyErr)
	case <-time.After(10 * jwksStaleWait):
	}

	release()

	select {
	case verifyErr := <-done:
		assert.NoError(t, verifyErr, "the refreshed key verifies the token")
	case <-time.After(10 * time.Second):
		t.Fatal("verification did not return")
	}

	<-forced
}

func TestVerifier_keysForKid(t *testing.T) {
	prev := jwksStaleWait
	jwksStaleWait = 50 * time.Millisecond
	t.Cleanup(func() { jwksStaleWait = prev })

	_, keys := claimsTestSigner(t)
	kid := keys[0].Kid

	var requests atomic.Int64
	block := make(chan struct{})

	var once sync.Once
	release := func() { once.Do(func() { close(block) }) }

	// The first fetch serves the keys; later ones wait until released and fail.
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
	t.Cleanup(release)

	url := server.URL + "/.well-known/jwks.json"

	var clock atomic.Int64
	clock.Store(time.Now().UnixNano())
	v := refreshTestVerifier(t, url, func() time.Time { return time.Unix(0, clock.Load()) })

	t.Run("CachedWithinTTL", func(t *testing.T) {
		got, err := v.keysForKid(context.Background(), url, "unknown", false)
		require.NoError(t, err)
		assert.Equal(t, keys, got)
	})

	// Past the TTL, inside the window in which the cached keys stay usable.
	clock.Store(time.Unix(0, clock.Load()).Add(defaultJWKSCacheTTL + time.Minute).UnixNano())

	t.Run("KnownKidAfterStaleWait", func(t *testing.T) {
		got, err := v.keysForKid(context.Background(), url, kid, false)
		require.NoError(t, err)
		assert.Equal(t, keys, got)
	})
	t.Run("UnknownKidWaitsForRefresh", func(t *testing.T) {
		done := make(chan struct{})

		go func() {
			defer close(done)
			_, _ = v.keysForKid(context.Background(), url, "unknown", false)
		}()

		select {
		case <-done:
			t.Fatal("returned before the refresh finished")
		case <-time.After(10 * jwksStaleWait):
		}

		release()

		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("did not return")
		}
	})
}
