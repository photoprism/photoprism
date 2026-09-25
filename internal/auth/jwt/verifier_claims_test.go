package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// claimsTestSigner returns a manager with one active key plus the JWKS it publishes, so a
// case can sign a token whose claims or headers the issuer would never emit.
func claimsTestSigner(t *testing.T) (*Manager, []PublicJWK) {
	t.Helper()

	mgr, err := NewManager(newTestConfig(t))
	require.NoError(t, err)

	_, err = mgr.EnsureActiveKey()
	require.NoError(t, err)

	return mgr, mgr.JWKS().Keys
}

// signClaims signs claims with the manager's active key, adding the given extra headers.
func signClaims(t *testing.T, mgr *Manager, claims *Claims, headers map[string]any) string {
	t.Helper()

	key, err := mgr.ActiveKey()
	require.NoError(t, err)

	token := gojwt.NewWithClaims(gojwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = key.Kid
	token.Header["typ"] = "JWT"

	maps.Copy(token.Header, headers)

	signed, err := token.SignedString(key.PrivateKey)
	require.NoError(t, err)

	return signed
}

// primedVerifier returns a verifier primed against a local JWKS server publishing the
// manager's keys, so a case can exercise the request path rather than only the key-list form.
func primedVerifier(t *testing.T, mgr *Manager) (*Verifier, string) {
	t.Helper()

	jwksBytes, err := json.Marshal(mgr.JWKS())
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jwksBytes)
	}))
	t.Cleanup(server.Close)

	cfg := newTestConfig(t)
	url := server.URL + "/.well-known/jwks.json"
	cfg.SetJWKSUrl(url)

	v := NewVerifier(cfg)
	require.NoError(t, v.Prime(context.Background(), url))

	return v, url
}

func TestVerifyTemporalClaims(t *testing.T) {
	const issuer = "portal:temporal"
	const audience = "node:temporal"

	mgr, keys := claimsTestSigner(t)
	now := time.Now().UTC()

	registered := func(iat, nbf, exp time.Time) gojwt.RegisteredClaims {
		c := gojwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   "portal:client",
			Audience:  gojwt.ClaimStrings{audience},
			ExpiresAt: gojwt.NewNumericDate(exp),
		}
		if !iat.IsZero() {
			c.IssuedAt = gojwt.NewNumericDate(iat)
		}
		if !nbf.IsZero() {
			c.NotBefore = gojwt.NewNumericDate(nbf)
		}
		return c
	}

	verifier, jwksURL := primedVerifier(t, mgr)

	// Run every case through both entry points: the JWKS-backed verifier that serves API
	// requests, and the key-list form the CLI uses.
	verify := func(claims gojwt.RegisteredClaims) error {
		token := signClaims(t, mgr, &Claims{Scope: "cluster", RegisteredClaims: claims}, nil)

		_, withKeys := VerifyTokenWithKeys(token, ExpectedClaims{Issuer: issuer, Audience: audience}, keys, 60*time.Second)
		_, withJWKS := verifier.VerifyToken(context.Background(), token,
			ExpectedClaims{Issuer: issuer, Audience: audience, JWKSURL: jwksURL})

		if (withKeys == nil) != (withJWKS == nil) {
			t.Errorf("entry points disagree: keys=%v jwks=%v", withKeys, withJWKS)
		}

		return withKeys
	}

	t.Run("Valid", func(t *testing.T) {
		assert.NoError(t, verify(registered(now, now, now.Add(5*time.Minute))))
	})
	t.Run("AtTheMaximumLifetime", func(t *testing.T) {
		assert.NoError(t, verify(registered(now, now, now.Add(MaxTokenTTL))))
	})
	t.Run("FutureIssuedAt", func(t *testing.T) {
		// nbf has passed and the window is usable, so only issued-at validation refuses this.
		future := now.Add(30 * 24 * time.Hour)
		assert.Error(t, verify(registered(future, now, future.Add(5*time.Minute))))
	})
	t.Run("MissingNotBefore", func(t *testing.T) {
		assert.Error(t, verify(registered(now, time.Time{}, now.Add(5*time.Minute))))
	})
	t.Run("MissingIssuedAt", func(t *testing.T) {
		assert.Error(t, verify(registered(time.Time{}, now, now.Add(5*time.Minute))))
	})
	t.Run("ExpiresBeforeItIsValid", func(t *testing.T) {
		// Every timestamp is inside the leeway, so only the usable-window check refuses this.
		assert.Error(t, verify(registered(now.Add(-5*time.Minute), now.Add(-10*time.Second), now.Add(-30*time.Second))))
	})
	t.Run("ExpiresBeforeIssued", func(t *testing.T) {
		assert.Error(t, verify(registered(now, now, now.Add(-time.Second))))
	})
	t.Run("LifetimeAboveTheMaximum", func(t *testing.T) {
		assert.Error(t, verify(registered(now, now, now.Add(MaxTokenTTL+time.Second))))
	})
}

func TestRejectCriticalHeaders(t *testing.T) {
	const issuer = "portal:crit"
	const audience = "node:crit"

	mgr, keys := claimsTestSigner(t)
	now := time.Now().UTC()

	claims := &Claims{
		Scope: "cluster",
		RegisteredClaims: gojwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   "portal:client",
			Audience:  gojwt.ClaimStrings{audience},
			IssuedAt:  gojwt.NewNumericDate(now),
			NotBefore: gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(now.Add(5 * time.Minute)),
		},
	}

	expected := ExpectedClaims{Issuer: issuer, Audience: audience}
	verifier, jwksURL := primedVerifier(t, mgr)

	// Both entry points, so the API path and the CLI path cannot drift apart.
	verify := func(headers map[string]any) (error, error) {
		token := signClaims(t, mgr, claims, headers)
		_, withKeys := VerifyTokenWithKeys(token, expected, keys, 60*time.Second)
		_, withJWKS := verifier.VerifyToken(context.Background(), token,
			ExpectedClaims{Issuer: issuer, Audience: audience, JWKSURL: jwksURL})
		return withKeys, withJWKS
	}

	t.Run("NoCriticalHeader", func(t *testing.T) {
		withKeys, withJWKS := verify(nil)
		assert.NoError(t, withKeys)
		assert.NoError(t, withJWKS)
	})
	t.Run("UnknownExtension", func(t *testing.T) {
		withKeys, withJWKS := verify(map[string]any{"crit": []string{"pp-unknown"}, "pp-unknown": "x"})
		assert.Error(t, withKeys)
		assert.Error(t, withJWKS)
	})
	t.Run("EmptyList", func(t *testing.T) {
		withKeys, withJWKS := verify(map[string]any{"crit": []string{}})
		assert.Error(t, withKeys)
		assert.Error(t, withJWKS)
	})
	t.Run("NilHeader", func(t *testing.T) {
		assert.Error(t, rejectCriticalHeaders(nil))
	})
}

// TestVerifierStaleKeysAreBounded covers the trust a verifier keeps when it cannot refresh:
// cached keys stay usable for a finite window measured from the last successful fetch, and
// failed attempts do not extend it.
func TestVerifierStaleKeysAreBounded(t *testing.T) {
	portalCfg := newTestConfig(t)
	clusterUUID := rnd.UUIDv7()
	portalCfg.Options().ClusterUUID = clusterUUID

	mgr, err := NewManager(portalCfg)
	require.NoError(t, err)
	_, err = mgr.EnsureActiveKey()
	require.NoError(t, err)

	jwksBytes, err := json.Marshal(mgr.JWKS())
	require.NoError(t, err)

	var serving bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if !serving {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jwksBytes)
	}))
	defer server.Close()

	nodeCfg := newTestConfig(t)
	nodeCfg.SetJWKSUrl(server.URL + "/.well-known/jwks.json")
	nodeCfg.Options().ClusterUUID = clusterUUID

	issuer := NewIssuer(mgr)
	spec := ClaimsSpec{
		Issuer:   fmt.Sprintf("portal:%s", clusterUUID),
		Subject:  "portal:client-test",
		Audience: fmt.Sprintf("node:%s", nodeCfg.NodeUUID()),
		Scope:    []string{"cluster"},
	}

	verifier := NewVerifier(nodeCfg)
	ctx := context.Background()

	serving = true
	require.NoError(t, verifier.Prime(ctx, nodeCfg.JWKSUrl()))
	serving = false

	expected := ExpectedClaims{Issuer: spec.Issuer, Audience: spec.Audience, JWKSURL: nodeCfg.JWKSUrl()}

	// Age the cache without touching the clock the issuer signs with, so each token is
	// current while the key material behind it is not.
	ageCache := func(d time.Duration) {
		verifier.mu.Lock()
		verifier.cache.FetchedAt -= int64(d / time.Second)
		verifier.mu.Unlock()
	}

	ttl := time.Duration(nodeCfg.JWKSCacheTTL()) * time.Second
	require.Positive(t, ttl)

	t.Run("WithinTheWindow", func(t *testing.T) {
		ageCache(ttl + time.Minute)

		verifier.mu.Lock()
		before := verifier.cache.FetchedAt
		verifier.mu.Unlock()

		token, err := issuer.Issue(spec)
		require.NoError(t, err)

		_, err = verifier.VerifyToken(ctx, token, expected)
		assert.NoError(t, err, "keys within the stale window must still verify")

		verifier.mu.Lock()
		after := verifier.cache.FetchedAt
		verifier.mu.Unlock()

		assert.Equal(t, before, after, "a failed refresh must not extend the window")
	})
	t.Run("PastTheWindow", func(t *testing.T) {
		ageCache(ttl + RotationOverlap() + time.Minute)

		token, err := issuer.Issue(spec)
		require.NoError(t, err)

		_, err = verifier.VerifyToken(ctx, token, expected)
		assert.Error(t, err, "keys past the stale window must be refused")
	})
	t.Run("FutureTimestamp", func(t *testing.T) {
		verifier.mu.Lock()
		verifier.cache.FetchedAt = time.Now().Add(time.Hour).Unix()
		verifier.mu.Unlock()

		token, err := issuer.Issue(spec)
		require.NoError(t, err)

		_, err = verifier.VerifyToken(ctx, token, expected)
		assert.Error(t, err, "a cache timestamp in the future must not be trusted")
	})
	t.Run("RecoversAfterASuccessfulRefresh", func(t *testing.T) {
		// Poison the cache locally, so this case does not depend on what ran before it.
		verifier.mu.Lock()
		verifier.cache.FetchedAt = time.Now().Add(-(ttl + RotationOverlap() + time.Hour)).Unix()
		verifier.failedURL, verifier.failedAt, verifier.failedErr = "", 0, nil
		verifier.mu.Unlock()

		serving = true

		token, err := issuer.Issue(spec)
		require.NoError(t, err)

		_, err = verifier.VerifyToken(ctx, token, expected)
		require.NoError(t, err)

		assert.Less(t, verifier.Status(ttl).CacheAgeSeconds, int64(60), "a successful refresh must reset the age")
	})
}

func TestStaleKeysUsable(t *testing.T) {
	v := &Verifier{now: time.Now}
	ttl := 300 * time.Second
	window := ttl + RotationOverlap()

	entry := func(fetchedAt int64) cacheEntry {
		return cacheEntry{URL: "https://portal.example.com/jwks.json", Keys: []PublicJWK{{Kid: "k1"}}, FetchedAt: fetchedAt}
	}

	t.Run("Fresh", func(t *testing.T) {
		assert.True(t, v.staleKeysUsable(entry(time.Now().Unix()), ttl))
	})
	t.Run("AtTheBoundary", func(t *testing.T) {
		assert.True(t, v.staleKeysUsable(entry(time.Now().Add(-window).Unix()), ttl))
	})
	t.Run("PastTheBoundary", func(t *testing.T) {
		assert.False(t, v.staleKeysUsable(entry(time.Now().Add(-window-2*time.Second).Unix()), ttl))
	})
	t.Run("NeverFetched", func(t *testing.T) {
		assert.False(t, v.staleKeysUsable(entry(0), ttl))
	})
	t.Run("NegativeTimestamp", func(t *testing.T) {
		assert.False(t, v.staleKeysUsable(entry(-1), ttl))
	})
	t.Run("FutureTimestamp", func(t *testing.T) {
		assert.False(t, v.staleKeysUsable(entry(time.Now().Add(time.Hour).Unix()), ttl))
	})
}

// TestVerifierFetchFailureIsRemembered covers the cost of a refresh outage: the first request
// pays the retry attempts, and repeats inside the interval are answered from memory.
func TestVerifierFetchFailureIsRemembered(t *testing.T) {
	mgr, _ := claimsTestSigner(t)

	var requests int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cfg := newTestConfig(t)
	url := server.URL + "/.well-known/jwks.json"
	cfg.SetJWKSUrl(url)

	v := NewVerifier(cfg)
	ctx := context.Background()

	_, err := v.keysForURL(ctx, url, false)
	require.Error(t, err)
	require.Equal(t, jwksFetchMaxRetries, requests, "the first attempt runs the retry ladder")

	for range 5 {
		_, err = v.keysForURL(ctx, url, false)
		assert.Error(t, err)
	}

	assert.Equal(t, jwksFetchMaxRetries, requests, "repeats inside the interval must not fetch again")

	// An explicit refresh always tries, so priming is never blocked by a remembered failure.
	_, err = v.keysForURL(ctx, url, true)
	require.Error(t, err)
	assert.Greater(t, requests, jwksFetchMaxRetries)

	assert.NotNil(t, v.recentFetchFailure(url))
	assert.Nil(t, v.recentFetchFailure("https://other.example.com/jwks.json"))

	v.clearFetchFailure()
	assert.Nil(t, v.recentFetchFailure(url))

	_ = mgr
}
