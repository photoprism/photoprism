package tokens

import (
	"crypto/subtle"
	"net/url"
	"strconv"
	"time"

	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// PublicToken is the placeholder download token delivered while the instance runs in public mode, where
// no token is verified because all access is unauthenticated.
const PublicToken = "public"

// Download signs the session-bound tokens carried by the header-less download endpoints. Its key and
// signature path (the API base path, a directory token authorizing every download URL) are set by
// config.Propagate.
var Download = &Signer{}

var (
	// PublicMode reports whether the instance runs without authentication, in which case download tokens
	// are delivered as the PublicToken placeholder and never verified. Set by config.Propagate.
	PublicMode bool
	// CoarseDownload is the single instance-wide download token — the admin-configured static token or,
	// when none is set, one auto-generated random value — delivered to the sessionless public/share
	// client configs and accepted as an unscoped (not session-bound) capability. Set by config.Propagate.
	CoarseDownload string
	// DownloadStatic reports whether CoarseDownload is an admin-configured static token (PHOTOPRISM_DOWNLOAD_TOKEN).
	// When true it is delivered to every client so permanent, cacheable download URLs keep working; when
	// false authenticated sessions instead receive a signed, session-scoped token. Set by config.Propagate.
	DownloadStatic bool
)

// DownloadToken returns the "?t=" download token value to deliver for the given session: the PublicToken
// placeholder in public mode, the configured static token when one is set (so permanent-cache setups keep
// working for every client), otherwise a signed, session-scoped token — falling back to the coarse
// instance token for sessionless callers (the public/share client configs, which have no session to sign).
func DownloadToken(sessionID string) string {
	switch {
	case PublicMode:
		return PublicToken
	case DownloadStatic:
		return CoarseDownload
	default:
		if signed := SignDownload(sessionID); signed != "" {
			return signed
		}
		return CoarseDownload
	}
}

// IsCoarseDownload reports whether token equals the instance's coarse (non-session) download token,
// compared in constant time. The coarse token is an unscoped capability, so a match grants the by-design
// broad access of the public/share configs rather than a session-bound scope.
func IsCoarseDownload(token string) bool {
	return CoarseDownload != "" && subtle.ConstantTimeCompare([]byte(token), []byte(CoarseDownload)) == 1
}

// SignDownload returns the compact, session-bound download token value clients pass as the "?t="
// parameter: "<expires>.<sid>.<token>", where <token> is a Download signature over the signature path
// with the session ID as the signed "sid" parameter and an expiry from ttl.DownloadToken. Returns ""
// for an invalid session ID or an unconfigured signer (e.g. a crypto/rand failure), so a missing key
// never degrades to a forgeable empty-key HMAC.
func SignDownload(sessionID string) string {
	if !rnd.IsSessionID(sessionID) {
		return ""
	}

	expires := time.Now().Add(time.Duration(ttl.DownloadToken.Int()) * time.Second).Unix()

	token := Download.Sign(expires, url.Values{"sid": {sessionID}})
	if token == "" {
		return ""
	}

	return strconv.FormatInt(expires, 10) + "." + sessionID + "." + token
}

// VerifyDownload verifies the signed download token parameters taken from a request and returns the
// session ID it is bound to, or ok=false when the token is missing, forged, expired, or the signer is
// unconfigured.
func VerifyDownload(expires int64, sid, token string) (sessionID string, ok bool) {
	if !rnd.IsSessionID(sid) || !Download.Valid(expires, url.Values{"sid": {sid}}, token) {
		return "", false
	}

	return sid, true
}
