package jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestVerifierPrimeAndVerify(t *testing.T) {
	portalCfg := newTestConfig(t)
	clusterUUID := rnd.UUIDv7()
	portalCfg.Options().ClusterUUID = clusterUUID

	mgr, err := NewManager(portalCfg)
	require.NoError(t, err)
	mgr.now = func() time.Time { return time.Date(2025, 9, 24, 10, 30, 0, 0, time.UTC) }
	_, err = mgr.EnsureActiveKey()
	require.NoError(t, err)

	jwksBytes, err := json.Marshal(mgr.JWKS())
	require.NoError(t, err)

	etag := `"v1"`
	var requestCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", etag)
		w.Header().Set("Cache-Control", "max-age=300")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jwksBytes)
	}))
	defer server.Close()

	nodeCfg := newTestConfig(t)
	nodeCfg.SetJWKSUrl(server.URL + "/.well-known/jwks.json")
	nodeCfg.Options().ClusterUUID = clusterUUID
	nodeUUID := nodeCfg.NodeUUID()

	issuer := NewIssuer(mgr)
	issuer.now = func() time.Time { return time.Now().UTC() }

	spec := ClaimsSpec{
		Issuer:   fmt.Sprintf("portal:%s", clusterUUID),
		Subject:  "portal:client-test",
		Audience: fmt.Sprintf("node:%s", nodeUUID),
		Scope:    []string{"cluster", "vision"},
	}

	token, err := issuer.Issue(spec)
	require.NoError(t, err)

	verifier := NewVerifier(nodeCfg)
	ctx := context.Background()
	require.NoError(t, verifier.Prime(ctx, nodeCfg.JWKSUrl()))
	require.Equal(t, 1, requestCount)

	claims, err := verifier.VerifyToken(ctx, token, ExpectedClaims{
		Issuer:   spec.Issuer,
		Audience: spec.Audience,
		Scope:    []string{"cluster"},
		JWKSURL:  nodeCfg.JWKSUrl(),
	})
	require.NoError(t, err)
	require.Equal(t, spec.Subject, claims.Subject)
	require.Contains(t, claims.Scope, "cluster")

	// Force cache refresh by expiring entry and verify 304 handling.
	verifier.mu.Lock()
	verifier.cache.FetchedAt -= 1000
	verifier.mu.Unlock()

	_, err = verifier.VerifyToken(ctx, token, ExpectedClaims{
		Issuer:   spec.Issuer,
		Audience: spec.Audience,
		Scope:    []string{"cluster"},
		JWKSURL:  nodeCfg.JWKSUrl(),
	})
	require.NoError(t, err)
	require.Equal(t, 2, requestCount)

	// Missing scope should fail.
	_, err = verifier.VerifyToken(ctx, token, ExpectedClaims{
		Issuer:   spec.Issuer,
		Audience: spec.Audience,
		Scope:    []string{"cluster", "unknown"},
		JWKSURL:  nodeCfg.JWKSUrl(),
	})
	require.Error(t, err)
}

func TestVerifyTokenWithKeys(t *testing.T) {
	portalCfg := newTestConfig(t)
	clusterUUID := rnd.UUIDv7()
	portalCfg.Options().ClusterUUID = clusterUUID

	mgr, err := NewManager(portalCfg)
	require.NoError(t, err)
	mgr.now = func() time.Time { return time.Date(2025, 9, 24, 10, 30, 0, 0, time.UTC) }
	_, err = mgr.EnsureActiveKey()
	require.NoError(t, err)

	issuer := NewIssuer(mgr)
	issuer.now = func() time.Time { return time.Now().UTC() }

	spec := ClaimsSpec{
		Issuer:   fmt.Sprintf("portal:%s", clusterUUID),
		Subject:  "portal:client-test",
		Audience: "node:1234",
		Scope:    []string{"cluster"},
	}

	token, err := issuer.Issue(spec)
	require.NoError(t, err)

	keys := mgr.JWKS().Keys
	claims, err := VerifyTokenWithKeys(token, ExpectedClaims{
		Issuer:   spec.Issuer,
		Audience: spec.Audience,
		Scope:    []string{"cluster"},
	}, keys, 60*time.Second)
	require.NoError(t, err)
	require.Equal(t, spec.Subject, claims.Subject)

	// Ensure scope filtering is honored when expected scope is empty.
	claims, err = VerifyTokenWithKeys(token, ExpectedClaims{
		Issuer:   spec.Issuer,
		Audience: spec.Audience,
	}, keys, 60*time.Second)
	require.NoError(t, err)
	require.Equal(t, spec.Subject, claims.Subject)

	// Missing scope should fail when explicitly required.
	_, err = VerifyTokenWithKeys(token, ExpectedClaims{
		Issuer:   spec.Issuer,
		Audience: spec.Audience,
		Scope:    []string{"vision"},
	}, keys, 60*time.Second)
	require.Error(t, err)
}

func TestIssuerClampTTL(t *testing.T) {
	portalCfg := newTestConfig(t)
	mgr, err := NewManager(portalCfg)
	require.NoError(t, err)
	mgr.now = func() time.Time { return time.Unix(0, 0) }
	_, err = mgr.EnsureActiveKey()
	require.NoError(t, err)

	issuer := NewIssuer(mgr)
	issuer.now = func() time.Time { return time.Unix(1000, 0) }

	spec := ClaimsSpec{
		Issuer:   "portal:test",
		Subject:  "portal:client",
		Audience: "node:test",
		Scope:    []string{"cluster"},
		TTL:      7200 * time.Second,
	}

	token, err := issuer.Issue(spec)
	require.NoError(t, err)

	parsed := &Claims{}
	parser := gojwt.NewParser(gojwt.WithValidMethods([]string{gojwt.SigningMethodEdDSA.Alg()}), gojwt.WithoutClaimsValidation())
	_, err = parser.ParseWithClaims(token, parsed, func(token *gojwt.Token) (interface{}, error) {
		key, _ := mgr.ActiveKey()
		return key.PublicKey, nil
	})
	require.NoError(t, err)
	ttl := parsed.ExpiresAt.Sub(parsed.IssuedAt.Time)
	require.Equal(t, MaxTokenTTL, ttl)
}

func TestBackoffDuration(t *testing.T) {
	origRand := randInt63n
	randInt63n = func(n int64) int64 {
		if n <= 0 {
			return 0
		}
		return n - 1
	}
	t.Cleanup(func() { randInt63n = origRand })

	tests := []struct {
		name    string
		attempt int
		expect  time.Duration
	}{
		{"Attempt1", 1, 300 * time.Millisecond},
		{"Attempt2", 2, 600 * time.Millisecond},
		{"Attempt3", 3, 1200 * time.Millisecond},
		{"Attempt4", 4, 2400 * time.Millisecond},
		{"Attempt5", 5, 3 * time.Second},
		{"AttemptZero", 0, 300 * time.Millisecond},
	}

	for _, tt := range tests {
		if got := backoffDuration(tt.attempt); got != tt.expect {
			t.Errorf("%s: expected %s, got %s", tt.name, tt.expect, got)
		}
	}
}
