package jwt

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

var (
	errKeyNotFound = errors.New("jwt: key not found")
)

// VerifierStatus captures diagnostic information about a verifier's JWKS cache state.
type VerifierStatus struct {
	CacheURL        string    `json:"cacheUrl,omitempty"`
	CacheETag       string    `json:"cacheEtag,omitempty"`
	KeyIDs          []string  `json:"keyIds,omitempty"`
	KeyCount        int       `json:"keyCount"`
	CacheFetchedAt  time.Time `json:"cacheFetchedAt,omitempty"`
	CacheAgeSeconds int64     `json:"cacheAgeSeconds"`
	CacheTTLSeconds int       `json:"cacheTtlSeconds"`
	CacheStale      bool      `json:"cacheStale"`
	CachePath       string    `json:"cachePath,omitempty"`
	JWKSURL         string    `json:"jwksUrl,omitempty"`
}

const (
	// jwksFetchMaxRetries caps the number of immediate retry attempts after a fetch error.
	jwksFetchMaxRetries = 3
	// jwksFetchBaseDelay is the initial retry delay (with jitter) applied after the first failure.
	jwksFetchBaseDelay = 200 * time.Millisecond
	// jwksFetchMaxDelay is the upper bound for retry delays to prevent unbounded backoff.
	jwksFetchMaxDelay = 2 * time.Second
)

// randInt63n is defined for deterministic testing of jitter (overridable in tests).
var randInt63n = rand.Int64N

// cacheEntry stores the JWKS material cached on disk and in memory.
type cacheEntry struct {
	URL       string      `json:"url"`
	ETag      string      `json:"etag,omitempty"`
	Keys      []PublicJWK `json:"keys"`
	FetchedAt int64       `json:"fetchedAt"`
}

// Verifier validates Portal-issued JWTs on Nodes using JWKS with caching.
type Verifier struct {
	conf *config.Config

	mu        sync.Mutex
	cache     cacheEntry
	cachePath string

	httpClient *http.Client
	now        func() time.Time
}

// ExpectedClaims describes the constraints that must hold for a token.
type ExpectedClaims struct {
	Issuer   string
	Audience string
	Scope    []string
	JWKSURL  string
}

// NewVerifier instantiates a verifier with sane defaults.
func NewVerifier(conf *config.Config) *Verifier {
	v := &Verifier{
		conf:       conf,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		now:        time.Now,
	}
	if conf != nil {
		v.cachePath = filepath.Join(conf.ConfigPath(), "jwks-cache.json")
	}
	_ = v.loadCache()
	return v
}

// Prime ensures JWKS material is cached locally.
func (v *Verifier) Prime(ctx context.Context, jwksURL string) error {
	_, err := v.keysForURL(ctx, jwksURL, true)
	return err
}

// VerifyToken validates a JWT against the expected claims and returns decoded claims.
func (v *Verifier) VerifyToken(ctx context.Context, tokenString string, expected ExpectedClaims) (*Claims, error) {
	if v == nil {
		return nil, errors.New("jwt: verifier not initialized")
	}
	if strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("jwt: token is empty")
	}
	if strings.TrimSpace(expected.Issuer) == "" {
		return nil, errors.New("jwt: expected issuer required")
	}
	if strings.TrimSpace(expected.Audience) == "" {
		return nil, errors.New("jwt: expected audience required")
	}

	jwksUrl := strings.TrimSpace(expected.JWKSURL)

	if jwksUrl == "" && v.conf != nil {
		jwksUrl = strings.TrimSpace(v.conf.JWKSUrl())
	}

	if jwksUrl == "" {
		return nil, errors.New("jwt: jwks url not configured")
	}

	leeway := 60 * time.Second
	if v.conf != nil && v.conf.JWTLeeway() > 0 {
		leeway = time.Duration(v.conf.JWTLeeway()) * time.Second
	}

	parser := gojwt.NewParser(
		gojwt.WithLeeway(leeway),
		gojwt.WithValidMethods([]string{gojwt.SigningMethodEdDSA.Alg()}),
		gojwt.WithIssuer(expected.Issuer),
		gojwt.WithAudience(expected.Audience),
	)

	claims := &Claims{}
	keyFunc := func(token *gojwt.Token) (interface{}, error) {
		kid, _ := token.Header["kid"].(string)

		if kid == "" {
			return nil, errors.New("jwt: missing kid header")
		}

		pk, err := v.publicKeyForKid(ctx, jwksUrl, kid, false)

		if errors.Is(err, errKeyNotFound) {
			pk, err = v.publicKeyForKid(ctx, jwksUrl, kid, true)
		}

		if err != nil {
			return nil, err
		}

		return pk, nil
	}

	if _, err := parser.ParseWithClaims(tokenString, claims, keyFunc); err != nil {
		return nil, err
	}

	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return nil, errors.New("jwt: missing temporal claims")
	}

	if ttl := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); ttl > MaxTokenTTL {
		return nil, errors.New("jwt: token ttl exceeds maximum")
	}

	scopeSet := map[string]struct{}{}

	for _, s := range strings.Fields(claims.Scope) {
		scopeSet[s] = struct{}{}
	}

	for _, req := range expected.Scope {
		if _, ok := scopeSet[req]; !ok {
			return nil, fmt.Errorf("jwt: missing scope %s", req)
		}
	}

	return claims, nil
}

// VerifyTokenWithKeys verifies a token using the provided JWKS keys without performing HTTP fetches.
func VerifyTokenWithKeys(tokenString string, expected ExpectedClaims, keys []PublicJWK, leeway time.Duration) (*Claims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("jwt: token is empty")
	}

	if len(keys) == 0 {
		return nil, errors.New("jwt: no jwks keys provided")
	}

	if leeway <= 0 {
		leeway = 60 * time.Second
	}

	keyMap := make(map[string]ed25519.PublicKey, len(keys))

	for _, jwk := range keys {
		if jwk.Kid == "" {
			continue
		}
		raw, err := base64.RawURLEncoding.DecodeString(jwk.X)
		if err != nil {
			return nil, err
		}
		if len(raw) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("jwt: invalid public key length %d", len(raw))
		}
		pk := make(ed25519.PublicKey, ed25519.PublicKeySize)
		copy(pk, raw)
		keyMap[jwk.Kid] = pk
	}

	if len(keyMap) == 0 {
		return nil, errors.New("jwt: no valid jwks keys provided")
	}

	options := []gojwt.ParserOption{
		gojwt.WithLeeway(leeway),
		gojwt.WithValidMethods([]string{gojwt.SigningMethodEdDSA.Alg()}),
	}

	if iss := strings.TrimSpace(expected.Issuer); iss != "" {
		options = append(options, gojwt.WithIssuer(iss))
	}

	if aud := strings.TrimSpace(expected.Audience); aud != "" {
		options = append(options, gojwt.WithAudience(aud))
	}

	parser := gojwt.NewParser(options...)
	claims := &Claims{}
	keyFunc := func(token *gojwt.Token) (interface{}, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("jwt: missing kid header")
		}
		pk, ok := keyMap[kid]
		if !ok {
			return nil, errKeyNotFound
		}
		return pk, nil
	}

	if _, err := parser.ParseWithClaims(tokenString, claims, keyFunc); err != nil {
		return nil, err
	}

	if claims.IssuedAt == nil || claims.ExpiresAt == nil {
		return nil, errors.New("jwt: missing temporal claims")
	}

	if ttl := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time); ttl > MaxTokenTTL {
		return nil, errors.New("jwt: token ttl exceeds maximum")
	}

	if len(expected.Scope) > 0 {
		scopeSet := map[string]struct{}{}
		for _, s := range strings.Fields(claims.Scope) {
			scopeSet[s] = struct{}{}
		}
		for _, req := range expected.Scope {
			if _, ok := scopeSet[req]; !ok {
				return nil, fmt.Errorf("jwt: missing scope %s", req)
			}
		}
	}

	return claims, nil
}

// Status returns diagnostic information about the verifier's current JWKS cache.
func (v *Verifier) Status(ttl time.Duration) VerifierStatus {
	status := VerifierStatus{}

	if ttl > 0 {
		status.CacheTTLSeconds = int(ttl / time.Second)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	status.CacheURL = v.cache.URL
	status.CacheETag = v.cache.ETag
	status.JWKSURL = v.cache.URL
	status.KeyCount = len(v.cache.Keys)
	status.KeyIDs = make([]string, 0, len(v.cache.Keys))

	for _, key := range v.cache.Keys {
		status.KeyIDs = append(status.KeyIDs, key.Kid)
	}

	status.CachePath = v.cachePath

	if v.cache.FetchedAt > 0 {
		fetched := time.Unix(v.cache.FetchedAt, 0).UTC()
		status.CacheFetchedAt = fetched
		age := time.Since(fetched)
		status.CacheAgeSeconds = int64(age.Seconds())
		if ttl > 0 && age > ttl {
			status.CacheStale = true
		}
	}

	return status
}

// publicKeyForKid resolves the public key for the given key ID, fetching JWKS data if needed.
func (v *Verifier) publicKeyForKid(ctx context.Context, url, kid string, force bool) (ed25519.PublicKey, error) {
	keys, err := v.keysForURL(ctx, url, force)

	if err != nil {
		return nil, err
	}

	for _, k := range keys {
		if k.Kid != kid {
			continue
		}
		raw, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil {
			return nil, err
		}
		if len(raw) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("jwt: invalid public key length %d", len(raw))
		}
		pk := make(ed25519.PublicKey, ed25519.PublicKeySize)
		copy(pk, raw)
		return pk, nil
	}

	return nil, errKeyNotFound
}

// keysForURL returns JWKS keys for the specified endpoint, reusing cache when possible.
func (v *Verifier) keysForURL(ctx context.Context, url string, force bool) ([]PublicJWK, error) {
	ttl := 300 * time.Second

	if v.conf != nil && v.conf.JWKSCacheTTL() > 0 {
		ttl = time.Duration(v.conf.JWKSCacheTTL()) * time.Second
	}

	attempts := 0

	for {
		cached := v.snapshotCache()

		if keys, ok := v.cachedKeys(url, ttl, cached, force); ok {
			return keys, nil
		}

		etag := ""
		if !force && cached.URL == url {
			etag = cached.ETag
		}

		result, err := v.fetchJWKS(ctx, url, etag)
		if err != nil {
			if !force && cached.URL == url && len(cached.Keys) > 0 {
				return append([]PublicJWK(nil), cached.Keys...), nil
			}

			attempts++
			if attempts >= jwksFetchMaxRetries {
				return nil, err
			}

			delay := backoffDuration(attempts)
			log.Debugf("jwt: jwks fetch retry %d for %s in %s (%s)", attempts, url, delay, err)

			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		if keys, ok := v.updateCache(url, result); ok {
			return keys, nil
		}
		// Cache changed by another goroutine between snapshot and update; retry.
	}
}

// snapshotCache returns the current JWKS cache entry under lock for safe reading.
func (v *Verifier) snapshotCache() cacheEntry {
	v.mu.Lock()
	defer v.mu.Unlock()
	cache := v.cache
	return cache
}

// cachedKeys returns cached JWKS keys if they are fresh enough and match the target URL.
func (v *Verifier) cachedKeys(url string, ttl time.Duration, cache cacheEntry, force bool) ([]PublicJWK, bool) {
	if force || cache.URL != url || len(cache.Keys) == 0 {
		return nil, false
	}

	age := v.now().Unix() - cache.FetchedAt
	if age < 0 {
		return nil, false
	}

	if time.Duration(age)*time.Second > ttl {
		return nil, false
	}

	return append([]PublicJWK(nil), cache.Keys...), true
}

type jwksFetchResult struct {
	keys        []PublicJWK
	etag        string
	fetchedAt   int64
	notModified bool
}

// fetchJWKS downloads the JWKS document (respecting conditional requests) and returns the parsed keys.
func (v *Verifier) fetchJWKS(ctx context.Context, url, etag string) (*jwksFetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := v.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotModified:
		return &jwksFetchResult{
			etag:        etag,
			fetchedAt:   v.now().Unix(),
			notModified: true,
		}, nil
	case http.StatusOK:
		var body JWKS
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return nil, err
		}
		if len(body.Keys) == 0 {
			return nil, errors.New("jwt: jwks contains no keys")
		}
		return &jwksFetchResult{
			keys:      append([]PublicJWK(nil), body.Keys...),
			etag:      resp.Header.Get("ETag"),
			fetchedAt: v.now().Unix(),
		}, nil
	default:
		return nil, fmt.Errorf("jwt: jwks fetch failed: %s", resp.Status)
	}
}

// updateCache stores the JWKS fetch result on success and returns the fresh keys.
func (v *Verifier) updateCache(url string, result *jwksFetchResult) ([]PublicJWK, bool) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if result.notModified {
		if v.cache.URL != url {
			return nil, false
		}
		v.cache.FetchedAt = result.fetchedAt
		if result.etag != "" {
			v.cache.ETag = result.etag
		}
		_ = v.saveCacheLocked()
		return append([]PublicJWK(nil), v.cache.Keys...), true
	}

	v.cache = cacheEntry{
		URL:       url,
		ETag:      result.etag,
		Keys:      append([]PublicJWK(nil), result.keys...),
		FetchedAt: result.fetchedAt,
	}

	_ = v.saveCacheLocked()
	return append([]PublicJWK(nil), v.cache.Keys...), true
}

// loadCache restores a previously persisted JWKS cache entry from disk.
func (v *Verifier) loadCache() error {
	if v.cachePath == "" || !fs.FileExists(v.cachePath) {
		return nil
	}

	b, err := os.ReadFile(v.cachePath)
	if err != nil || len(b) == 0 {
		return err
	}

	var entry cacheEntry
	if err = json.Unmarshal(b, &entry); err != nil {
		return err
	}

	v.cache = entry
	return nil
}

// saveCacheLocked persists the current cache entry to disk; caller must hold the mutex.
func (v *Verifier) saveCacheLocked() error {
	if v.cachePath == "" {
		return nil
	}

	if err := fs.MkdirAll(filepath.Dir(v.cachePath)); err != nil {
		return err
	}

	data, err := json.Marshal(v.cache)

	if err != nil {
		return err
	}

	return os.WriteFile(v.cachePath, data, fs.ModeSecretFile)
}

// backoffDuration returns the retry delay for the given fetch attempt, adding jitter.
func backoffDuration(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}

	base := jwksFetchBaseDelay << (attempt - 1)

	if base > jwksFetchMaxDelay {
		base = jwksFetchMaxDelay
	}

	jitterRange := base / 2

	if jitterRange > 0 {
		base += time.Duration(randInt63n(int64(jitterRange) + 1))
	}

	return base
}
