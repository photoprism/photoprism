package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// OIDCSessionCookie is the name of the narrowly-scoped cookie that lets the
// Portal OIDC OP authenticate a browser on a top-level navigation to
// /api/v1/oauth/authorize, which carries no Authorization or X-Auth-Token header.
const OIDCSessionCookie = "oidc_session"

// OIDCSessionCookieTTL bounds how long the OP session-signal cookie stays valid.
// It is short because the cookie only needs to bridge the brief window between a
// header-less top-level navigation to /api/v1/oauth/authorize and the SPA
// re-presenting its session; it is refreshed on every authenticated session GET.
const OIDCSessionCookieTTL = 10 * time.Minute

const (
	oidcSessionKeyFile = "oidc_session.key"
	oidcSessionKeyLen  = 32
)

var (
	oidcSessionKeyOnce sync.Once
	oidcSessionKey     []byte
)

// OIDCSessionCookiePath returns the URL path the OP session cookie is scoped to,
// derived from where the OAuth endpoints are mounted (the APIv1 base path plus
// "/oauth"). Scoping to the OAuth subtree keeps the cookie off the general API
// surface, so it adds no CSRF vector to state-changing endpoints, while still
// covering the /oauth/authorize endpoint a top-level browser navigation hits.
func OIDCSessionCookiePath(conf *config.Config) string {
	if conf == nil {
		return config.ApiUri + "/oauth"
	}
	return conf.BaseUri(config.ApiUri) + "/oauth"
}

// OIDCSessionCookieClearPath returns the path at which the OP session cookie is
// cleared on logout, or "" when this node has no OP cookie to clear. On the Portal
// it is the local OP path; on an instance whose OIDC OP is the Portal it is derived
// from the issuer (the cookie lives at the Portal's shared-domain path, not the
// instance's /i/<tenant> base), so a cluster-wide Sign-Out works from any node.
func OIDCSessionCookieClearPath(conf *config.Config) string {
	if conf == nil {
		return ""
	}
	if conf.Portal() {
		return OIDCSessionCookiePath(conf)
	}
	if conf.OIDCIssuerOnSiteDomain() {
		return strings.TrimRight(conf.OIDCUri().Path, "/") + config.ApiUri + "/oauth"
	}
	return ""
}

// oidcSessionSignalKey loads (or generates and persists) the HMAC key used to
// sign the OP session-signal cookie, kept under the Portal config keys
// directory. It falls back to a process-local random key when persistence is
// unavailable so sign and verify still agree within a single process.
func oidcSessionSignalKey() []byte {
	oidcSessionKeyOnce.Do(func() {
		if key := loadOrCreateOIDCSessionKey(); len(key) >= oidcSessionKeyLen {
			oidcSessionKey = key
			return
		}
		key := make([]byte, oidcSessionKeyLen)
		_, _ = rand.Read(key)
		oidcSessionKey = key
		// Console-only: identity-provider key management must not reach the browser log viewer.
		event.SystemWarn([]string{"oidc", "session signal key", "persist", "using a process-local key, so OP session cookies will not verify across restarts or replicas"})
	})
	return oidcSessionKey
}

// loadOrCreateOIDCSessionKey reads the persisted HMAC key, creating it on first
// use. Returns nil when no config or writable key path is available.
func loadOrCreateOIDCSessionKey() []byte {
	conf := get.Config()
	if conf == nil {
		return nil
	}
	dir := filepath.Join(conf.PortalConfigPath(), "keys")
	path := filepath.Join(dir, oidcSessionKeyFile)
	if b, err := os.ReadFile(path); err == nil && len(b) >= oidcSessionKeyLen { //nolint:gosec // path is the portal config keys directory plus a fixed filename, not user input
		return b
	}
	if err := fs.MkdirAll(dir); err != nil {
		return nil
	}
	key := make([]byte, oidcSessionKeyLen)
	if _, err := rand.Read(key); err != nil {
		return nil
	}
	if err := os.WriteFile(path, key, fs.ModeSecretFile); err != nil {
		return nil
	}
	return key
}

// signOIDCSession returns a URL-safe, HMAC-signed value binding sessionID to
// exp. The payload is authenticated but not secret: it carries the session id
// (a hash of the bearer token, not the token itself), so the value is opaque to
// the browser and cannot be replayed against any other endpoint.
func signOIDCSession(sessionID string, exp time.Time) string {
	payload := sessionID + "|" + strconv.FormatInt(exp.Unix(), 10)
	mac := hmac.New(sha256.New, oidcSessionSignalKey())
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." +
		base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// parseOIDCSession verifies a signed value and returns the embedded session id
// when the signature is valid and the value has not expired. The HMAC check is
// constant-time, so a forged or tampered value cannot be distinguished by
// response timing.
func parseOIDCSession(value string) (sessionID string, ok bool) {
	encPayload, encSig, found := strings.Cut(value, ".")
	if !found {
		return "", false
	}
	payload, err := base64.RawURLEncoding.DecodeString(encPayload)
	if err != nil {
		return "", false
	}
	sig, err := base64.RawURLEncoding.DecodeString(encSig)
	if err != nil {
		return "", false
	}
	mac := hmac.New(sha256.New, oidcSessionSignalKey())
	mac.Write(payload)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return "", false
	}
	id, expStr, found := strings.Cut(string(payload), "|")
	if !found {
		return "", false
	}
	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || time.Now().After(time.Unix(expUnix, 0)) {
		return "", false
	}
	if !rnd.IsSessionID(id) {
		return "", false
	}
	return id, true
}

// SetOIDCSessionCookie writes the OP session-signal cookie for sess, which the
// /api/v1/oauth/authorize endpoint reads to resume a browser on a top-level
// navigation that carries no Authorization header. The value is an HMAC-signed
// reference to the session id, not the bearer token, and it expires on its own.
func SetOIDCSessionCookie(c *gin.Context, sess *entity.Session, cookiePath string, secure bool) {
	// Only a session the OP admits gets a cookie, so a cookie cannot widen admission.
	if c == nil || sess == nil || !rnd.IsSessionID(sess.ID) || !OIDCSessionEligible(sess) {
		return
	}

	if cookiePath == "" {
		cookiePath = config.ApiUri + "/oauth"
	}

	value := signOIDCSession(sess.ID, time.Now().Add(OIDCSessionCookieTTL))

	http.SetCookie(c.Writer, &http.Cookie{ //nolint:gosec // HttpOnly and SameSite are set; Secure is caller-controlled (HTTP is allowed only for loopback/cluster-internal advertise URLs)
		Name:     OIDCSessionCookie,
		Value:    value,
		Path:     cookiePath,
		MaxAge:   int(OIDCSessionCookieTTL.Seconds()),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearOIDCSessionCookie removes the OP session cookie, e.g. on logout. The
// cookiePath must match the value used by SetOIDCSessionCookie so the browser
// overwrites the same cookie.
func ClearOIDCSessionCookie(c *gin.Context, cookiePath string, secure bool) {
	if c == nil {
		return
	}

	if cookiePath == "" {
		cookiePath = config.ApiUri + "/oauth"
	}

	http.SetCookie(c.Writer, &http.Cookie{ //nolint:gosec // HttpOnly and SameSite are set; Secure is caller-controlled (HTTP is allowed only for loopback/cluster-internal advertise URLs)
		Name:     OIDCSessionCookie,
		Value:    "",
		Path:     cookiePath,
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// OIDCSessionCookieSession resolves the Portal session referenced by the OP session
// cookie, or nil when it is absent, invalid, expired, or no longer maps to a session
// the OP admits; an unresolvable cookie is cleared. Admission is rechecked here, not
// only where the cookie is written. Only the OIDC OP authorize handler may call it.
func OIDCSessionCookieSession(c *gin.Context) *entity.Session {
	sess := oidcSessionCookieSession(c)

	if sess == nil {
		return nil
	} else if !OIDCSessionEligible(sess) {
		clearOIDCSessionCookieForConfig(c)
		return nil
	}

	return sess
}

// OIDCSessionCookieEndSession resolves the referenced session for the end-session
// endpoint, without the admission policy. Ending a session must not be narrower than
// starting one, or the sessions the policy distrusts are the ones logout cannot reach.
func OIDCSessionCookieEndSession(c *gin.Context) *entity.Session {
	return oidcSessionCookieSession(c)
}

// oidcSessionCookieSession verifies the OP session cookie and resolves the session it
// references, clearing a cookie that resolves to nothing.
func oidcSessionCookieSession(c *gin.Context) *entity.Session {
	if c == nil {
		return nil
	}

	raw, err := c.Cookie(OIDCSessionCookie)
	if err != nil || raw == "" {
		return nil
	}

	sessionID, ok := parseOIDCSession(raw)
	if !ok {
		clearOIDCSessionCookieForConfig(c)
		return nil
	}

	sess, err := entity.FindSession(sessionID)
	if err != nil || sess == nil || sess.Invalid() || sess.NoUser() {
		clearOIDCSessionCookieForConfig(c)
		return nil
	}

	return sess
}

// clearOIDCSessionCookieForConfig clears the OP session cookie at the path and security
// level the current configuration writes it with, so the browser overwrites the same one.
func clearOIDCSessionCookieForConfig(c *gin.Context) {
	cookiePath := config.ApiUri + "/oauth"
	secure := false

	if conf := get.Config(); conf != nil {
		cookiePath = OIDCSessionCookiePath(conf)
		secure = conf.SiteHttps()
	}

	ClearOIDCSessionCookie(c, cookiePath, secure)
}
