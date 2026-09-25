package jwt

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/sync/singleflight"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

var (
	errKeyNotFound = errors.New("jwt: key not found")
)

// maxJWKSResponseBytes bounds how much of a JWKS response is read so a malicious
// or compromised IdP endpoint cannot exhaust memory; real key sets are a few KB.
const maxJWKSResponseBytes = 1 << 20 // 1 MiB

// VerifierStatus captures diagnostic information about a verifier's JWKS cache state.
type VerifierStatus struct {
	CacheURL        string    `json:"cacheUrl,omitempty"`
	CacheETag       string    `json:"cacheEtag,omitempty"`
	KeyIDs          []string  `json:"keyIds,omitempty"`
	KeyCount        int       `json:"keyCount"`
	CacheFetchedAt  time.Time `json:"cacheFetchedAt"`
	CacheAgeSeconds int64     `json:"cacheAgeSeconds"`
	CacheTTLSeconds int       `json:"cacheTtlSeconds"`
	CacheStale      bool      `json:"cacheStale"`
	CacheMaxAgeSecs int64     `json:"cacheMaxAgeSeconds"`
	CacheUsable     bool      `json:"cacheUsable"`
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
	// jwksFetchRetryAfter is how long a failed refresh is answered from memory.
	jwksFetchRetryAfter = 30 * time.Second
	// jwksForceRefreshAfter is the minimum interval between two forced refreshes of the same
	// endpoint, so a rotated key is discovered promptly while refresh work stays bounded per endpoint.
	jwksForceRefreshAfter = 30 * time.Second
	// defaultJWKSCacheTTL applies when no cache TTL is configured.
	defaultJWKSCacheTTL = 300 * time.Second
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

// Verifier validates Portal-issued JWTs on instances and services using JWKS with caching.
type Verifier struct {
	conf *config.Config

	mu        sync.Mutex
	cache     cacheEntry
	cachePath string

	failedURL string
	failedAt  int64
	failedErr error

	// forcedAt records the last forced refresh per configured JWKS endpoint.
	forcedAt map[string]int64

	// fetches coalesces concurrent refreshes of the same endpoint into one request.
	fetches singleflight.Group

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
		gojwt.WithIssuedAt(),
	)

	claims := &Claims{}
	keyFunc := func(token *gojwt.Token) (any, error) {
		if err := rejectCriticalHeaders(token); err != nil {
			return nil, err
		}

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

	if err := verifyTemporalClaims(claims); err != nil {
		return nil, err
	}

	scopeSet := map[string]struct{}{}

	for s := range strings.FieldsSeq(claims.Scope) {
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
		gojwt.WithIssuedAt(),
	}

	if iss := strings.TrimSpace(expected.Issuer); iss != "" {
		options = append(options, gojwt.WithIssuer(iss))
	}

	if aud := strings.TrimSpace(expected.Audience); aud != "" {
		options = append(options, gojwt.WithAudience(aud))
	}

	parser := gojwt.NewParser(options...)
	claims := &Claims{}
	keyFunc := func(token *gojwt.Token) (any, error) {
		if err := rejectCriticalHeaders(token); err != nil {
			return nil, err
		}
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

	if err := verifyTemporalClaims(claims); err != nil {
		return nil, err
	}

	if len(expected.Scope) > 0 {
		scopeSet := map[string]struct{}{}
		for s := range strings.FieldsSeq(claims.Scope) {
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

// verifyTemporalClaims checks the time claims the cluster contract requires: iat, nbf and exp
// present, a usable validity window, and a positive lifetime no longer than MaxTokenTTL.
// Whether iat and nbf have passed is decided by the parser, with the configured leeway.
func verifyTemporalClaims(claims *Claims) error {
	switch {
	case claims.IssuedAt == nil || claims.NotBefore == nil || claims.ExpiresAt == nil:
		return errors.New("jwt: missing temporal claims")
	case !claims.ExpiresAt.After(claims.NotBefore.Time):
		return errors.New("jwt: token expires before it is valid")
	}

	switch ttl := claims.ExpiresAt.Sub(claims.IssuedAt.Time); {
	case ttl <= 0:
		return errors.New("jwt: token expires before it was issued")
	case ttl > MaxTokenTTL:
		return errors.New("jwt: token ttl exceeds maximum")
	}

	return nil
}

// rejectCriticalHeaders refuses a token that names critical header extensions, since this
// verifier implements none (RFC 7515 §4.1.11).
func rejectCriticalHeaders(token *gojwt.Token) error {
	if token == nil {
		return errors.New("jwt: token is empty")
	} else if _, ok := token.Header["crit"]; ok {
		return errors.New("jwt: unsupported critical header")
	}

	return nil
}

// Status returns diagnostic information about the verifier's current JWKS cache.
func (v *Verifier) Status(ttl time.Duration) VerifierStatus {
	result := VerifierStatus{}

	if ttl > 0 {
		result.CacheTTLSeconds = int(ttl / time.Second)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	result.CacheURL = v.cache.URL
	result.CacheETag = v.cache.ETag
	result.JWKSURL = v.cache.URL
	result.KeyCount = len(v.cache.Keys)
	result.KeyIDs = make([]string, 0, len(v.cache.Keys))

	for _, key := range v.cache.Keys {
		result.KeyIDs = append(result.KeyIDs, key.Kid)
	}

	result.CachePath = v.cachePath

	// The age past which a refresh failure stops being survivable, so an operator can tell
	// "serving keys that are due a refresh" from "no longer accepting tokens".
	if ttl > 0 {
		result.CacheMaxAgeSecs = int64((ttl + RotationOverlap()).Seconds())
	}

	if v.cache.FetchedAt > 0 {
		fetched := time.Unix(v.cache.FetchedAt, 0).UTC()
		result.CacheFetchedAt = fetched
		age := v.now().UTC().Sub(fetched)
		result.CacheAgeSeconds = int64(age.Seconds())
		result.CacheUsable = len(v.cache.Keys) > 0 && v.staleKeysUsable(v.cache, ttl)
		if ttl > 0 && age > ttl {
			result.CacheStale = true
		}
	}

	return result
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

// keysForURL returns JWKS keys for the specified endpoint, reusing cache when possible. Concurrent
// callers for the same endpoint share one refresh.
func (v *Verifier) keysForURL(ctx context.Context, url string, force bool) ([]PublicJWK, error) {
	ttl := defaultJWKSCacheTTL

	if v.conf != nil && v.conf.JWKSCacheTTL() > 0 {
		ttl = time.Duration(v.conf.JWKSCacheTTL()) * time.Second
	}

	// A forced refresh runs at most once per interval per endpoint; beyond that the request is
	// answered from cache.
	if force && !v.forcedRefreshDue(url) {
		force = false
	}

	cached := v.snapshotCache()

	if keys, ok := v.cachedKeys(url, ttl, cached, force); ok {
		return keys, nil
	}

	if !force {
		if failure := v.recentFetchFailure(url); failure != nil {
			return v.keysAfterFailedRefresh(url, ttl, force, cached, failure)
		}
	}

	// Callers share the fetch, not the outcome: it is detached from any one request context, so a
	// caller that goes away neither cancels it nor decides what the others are served.
	ch := v.fetches.DoChan(url, func() (any, error) {
		return v.refreshKeys(context.WithoutCancel(ctx), url, ttl, force && v.takeForcedRefresh(url))
	})

	select {
	case res := <-ch:
		if res.Err != nil {
			return v.keysAfterFailedRefresh(url, ttl, force, v.snapshotCache(), res.Err)
		}

		keys, _ := res.Val.([]PublicJWK)

		return append([]PublicJWK(nil), keys...), nil
	case <-ctx.Done():
		return v.keysAfterFailedRefresh(url, ttl, force, v.snapshotCache(), ctx.Err())
	}
}

// keysAfterFailedRefresh applies this caller's own fallback when a refresh did not deliver keys, so a
// shared refresh cannot strip an ordinary verification of the cached keys it may still be served. A
// forced caller wants the error, since it is asking whether the endpoint has something new.
func (v *Verifier) keysAfterFailedRefresh(url string, ttl time.Duration, force bool, cached cacheEntry, err error) ([]PublicJWK, error) {
	if !force {
		if keys, ok := v.staleKeys(url, ttl, cached); ok {
			return keys, nil
		}
	}

	return nil, err
}

// forcedRefreshDue reports whether a forced refresh of the endpoint is due, without recording one.
func (v *Verifier) forcedRefreshDue(url string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.forcedRefreshDueLocked(url)
}

// takeForcedRefresh records a forced refresh and reports whether it was due. Only the goroutine that
// performs the fetch calls it, so a caller joining a refresh in flight does not consume the interval.
func (v *Verifier) takeForcedRefresh(url string) bool {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.forcedRefreshDueLocked(url) {
		return false
	}

	if v.forcedAt == nil {
		v.forcedAt = make(map[string]int64)
	}

	v.forcedAt[url] = v.now().Unix()

	return true
}

// forcedRefreshDueLocked reports whether the endpoint's forced refresh interval has elapsed; the
// caller must hold the mutex.
func (v *Verifier) forcedRefreshDueLocked(url string) bool {
	last, ok := v.forcedAt[url]

	if !ok {
		return true
	}

	age := v.now().Unix() - last

	return age < 0 || time.Duration(age)*time.Second >= jwksForceRefreshAfter
}

// refreshKeys fetches and caches the JWKS for the endpoint, retrying a transient error within the
// request that hit it. It runs inside the single-flight group, so only one call per endpoint is active.
func (v *Verifier) refreshKeys(ctx context.Context, url string, ttl time.Duration, force bool) ([]PublicJWK, error) {
	attempts := 0

	for {
		cached := v.snapshotCache()

		if keys, ok := v.cachedKeys(url, ttl, cached, force); ok {
			return keys, nil
		}

		// Only on entry: once a failure has been recorded, the retry attempts below must
		// still run so a transient blip is ridden out within the request that hit it. An explicit
		// refresh still tries, so priming is never blocked; allowForcedRefresh is what bounds how
		// often it may.
		if attempts == 0 && !force {
			if failure := v.recentFetchFailure(url); failure != nil {
				if keys, ok := v.staleKeys(url, ttl, cached); ok {
					return keys, nil
				}

				return nil, failure
			}
		}

		etag := ""
		if !force && cached.URL == url {
			etag = cached.ETag
		}

		result, err := v.fetchJWKS(ctx, url, etag)
		if err != nil {
			first := v.recordFetchFailure(url, err)

			if !force {
				if keys, ok := v.staleKeys(url, ttl, cached); ok {
					return keys, nil
				}
			}

			if first {
				log.Warnf("jwt: key set refresh failed and the cached keys are no longer usable, so tokens are denied")
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

		v.clearFetchFailure()

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

// staleKeys returns the cached keys for url when a refresh has failed but they are still
// inside the stale window.
func (v *Verifier) staleKeys(url string, ttl time.Duration, cache cacheEntry) ([]PublicJWK, bool) {
	if cache.URL != url || len(cache.Keys) == 0 || !v.staleKeysUsable(cache, ttl) {
		return nil, false
	}

	return append([]PublicJWK(nil), cache.Keys...), true
}

// recentFetchFailure returns the error from a recent failed refresh of url, or nil once the
// retry interval has passed.
func (v *Verifier) recentFetchFailure(url string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.failedURL != url || v.failedErr == nil {
		return nil
	}

	age := v.now().Unix() - v.failedAt

	if age < 0 || time.Duration(age)*time.Second > jwksFetchRetryAfter {
		return nil
	}

	return v.failedErr
}

// recordFetchFailure remembers a failed refresh and reports whether it opens a new retry
// interval, so an operator-visible warning is emitted once per interval rather than per request.
func (v *Verifier) recordFetchFailure(url string, err error) (first bool) {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := v.now().Unix()
	age := now - v.failedAt

	first = v.failedURL != url || v.failedErr == nil || age < 0 || time.Duration(age)*time.Second > jwksFetchRetryAfter

	v.failedURL = url
	v.failedAt = now
	v.failedErr = err

	return first
}

// clearFetchFailure drops the remembered failure after a successful refresh.
func (v *Verifier) clearFetchFailure() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.failedURL = ""
	v.failedAt = 0
	v.failedErr = nil
}

// staleKeysUsable reports whether cached keys may still be served after a refresh failure.
// Age is measured from the last successful fetch, and the window runs one cache TTL beyond
// the issuer's rotation overlap, which no live token can outlast.
func (v *Verifier) staleKeysUsable(cache cacheEntry, ttl time.Duration) bool {
	if cache.FetchedAt <= 0 {
		return false
	}

	age := v.now().Unix() - cache.FetchedAt

	if age < 0 {
		return false
	}

	return time.Duration(age)*time.Second <= ttl+RotationOverlap()
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

	// #nosec G704 JWKS URL is validated via config.SetJWKSUrl and verifier call paths.
	resp, err := v.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Debugf("jwt: %s (close JWKS response body)", closeErr)
		}
	}()

	switch resp.StatusCode {
	case http.StatusNotModified:
		return &jwksFetchResult{
			etag:        etag,
			fetchedAt:   v.now().Unix(),
			notModified: true,
		}, nil
	case http.StatusOK:
		var body JWKS
		if err := json.NewDecoder(io.LimitReader(resp.Body, maxJWKSResponseBytes)).Decode(&body); err != nil {
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

	base := min(jwksFetchBaseDelay<<(attempt-1), jwksFetchMaxDelay)

	jitterRange := base / 2

	if jitterRange > 0 {
		base += time.Duration(randInt63n(int64(jitterRange) + 1))
	}

	return base
}
