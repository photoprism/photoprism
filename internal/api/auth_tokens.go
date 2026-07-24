package api

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// InvalidPreviewToken checks if the token found in the request is valid for image thumbnails and video
// streams. A valid preview token is accepted, and so is the coarse download token (cross-acceptance: the
// higher-value download token also grants preview access — a signed download token is longer than 64
// characters, so clean.UrlToken trims it and it never matches here).
func InvalidPreviewToken(c *gin.Context) bool {
	token := clean.UrlToken(c.Param("token"))

	if token == "" {
		token = clean.UrlToken(c.Query("t"))
	}

	return entity.InvalidPreviewToken(token) && !tokens.IsCoarseDownload(token)
}

// AuthDownload authorizes a file-download request and returns the session it is scoped to together with
// whether the request is authorized. The session is nil for a coarse capability — a configured static
// token, or an unauthenticated public-mode request — in which case the handler applies the by-design
// broad (public, non-private) access. It merges the token gate and the session lookup, so a caller need
// not call InvalidDownloadToken and DownloadSession separately.
func AuthDownload(c *gin.Context) (sess *entity.Session, valid bool) {
	if sess = DownloadSession(c); sess != nil {
		return sess, true
	}

	// No bound session: a coarse (configured static or auto-generated instance) token still authorizes
	// an unscoped download.
	if tokens.IsCoarseDownload(clean.UrlToken(c.Query("t"))) {
		return nil, true
	}

	// Audit the unauthorized request centrally, mirroring AuthAny, so the endpoints do not each log it.
	event.AuditWarn([]string{ClientIP(c), string(acl.ResourceFiles), "download", status.Denied})

	return nil, false
}

// InvalidDownloadToken checks if the request is not authorized for a file download. It is a thin wrapper
// around AuthDownload for callers that only need the yes/no gate; prefer AuthDownload when the resolved
// session is needed to scope the response.
func InvalidDownloadToken(c *gin.Context) bool {
	_, valid := AuthDownload(c)
	return !valid
}

// downloadSessionKey is the gin context key under which the resolved download session is memoized.
const downloadSessionKey = "download_session"

// DownloadSession returns the session the request is bound to (a signed "?t=" token or a Portal JWT
// header), the shared public session in public mode, or nil for a coarse/forged/expired token. The
// result is memoized on the request context so the gate and the handler resolve it once.
func DownloadSession(c *gin.Context) *entity.Session {
	if v, ok := c.Get(downloadSessionKey); ok {
		sess, _ := v.(*entity.Session)
		return sess
	}

	sess := resolveDownloadSession(c)
	c.Set(downloadSessionKey, sess)

	return sess
}

// resolveDownloadSession does the actual token → session resolution for DownloadSession.
func resolveDownloadSession(c *gin.Context) *entity.Session {
	if get.Config().Public() {
		return get.Session().Public()
	}

	// The Portal authorizes a download with its cluster JWT in a request header (a transient JWT session
	// can't back a "?t=" token). Restricted to JWTs (authAnyJWT rejects non-JWT tokens) granting access to
	// all files, so only a trusted full-access principal qualifies; every other client presents a "?t=".
	if AuthToken(c) != "" {
		if s := authAnyJWT(c, ClientIP(c), AuthToken(c), acl.ResourceFiles, acl.Permissions{acl.AccessAll}); s != nil && s.Valid() {
			return s
		}
		// Not an all-photos cluster JWT: fall through to the "?t=" token path so a non-JWT header (a
		// client that also sends a bearer) does not shadow a valid download token.
	}

	// Header-less browser context: a signed, session-bound "?t=" token (compact `<expires>.<sid>.<token>`
	// or verbose `token=…`) binds the request to its session.
	if sessionID, ok := signedDownloadSession(c); ok {
		if sess, err := entity.FindSession(sessionID); err == nil {
			return sess
		}
	}

	// A coarse (static/instance) token or an unknown value resolves to no session: the handlers treat a
	// nil session as an unscoped coarse capability (public/share access) or reject it as invalid.
	return nil
}

// signedDownloadSession verifies a signed download token in the request and returns the bound session
// ID. It accepts the compact origin form (`?t=<expires>.<sid>.<token>`) used by download URLs and the
// verbose bunny.net edge form (`?token=…&expires=…&sid=…`), returning ok=false when neither is a valid
// signed token.
func signedDownloadSession(c *gin.Context) (sessionID string, ok bool) {
	// Compact origin form: t=<expires>.<sid>.<token>.
	if t := c.Query("t"); strings.Count(t, ".") == 2 {
		parts := strings.SplitN(t, ".", 3)

		if expires, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
			if id, valid := tokens.VerifyDownload(expires, parts[1], parts[2]); valid {
				return id, true
			}
		}
	}

	// bunny.net edge form: token=…&expires=…&sid=….
	if token := c.Query("token"); token != "" {
		if expires, err := strconv.ParseInt(c.Query("expires"), 10, 64); err == nil {
			return tokens.VerifyDownload(expires, c.Query("sid"), token)
		}
	}

	return "", false
}
