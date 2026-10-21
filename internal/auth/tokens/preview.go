package tokens

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// Preview signs the session-bound preview tokens of client sessions without an account. Its key and
// signature path are set by config.Propagate, as for Download.
var Preview = &Signer{}

// PreviewDay is the expiry window, in seconds, of preview tokens for sessions without an expiry.
const PreviewDay int64 = 86400

// previewParams returns the signed parameters of a preview token. The kind is signed but not
// transmitted, so download and preview tokens never verify as each other.
func previewParams(sessionID string) url.Values {
	return url.Values{"sid": {sessionID}, "kind": {PurposePreview}}
}

// PreviewExpires returns the expiry of a preview token for a session that expires at the given time:
// the session's own expiry, or the end of the next UTC day if it has none, so the value stays the same
// within a day and is valid for at least one more.
func PreviewExpires(sessionExpires, now time.Time) int64 {
	if !sessionExpires.IsZero() {
		return sessionExpires.Unix()
	}

	return (now.Unix()/PreviewDay + 2) * PreviewDay
}

// PreviewToken returns the signed preview token to deliver for a session that expires at the given time,
// or "" for an invalid session ID or an unconfigured signer.
func PreviewToken(sessionID string, sessionExpires time.Time) string {
	return SignPreview(sessionID, PreviewExpires(sessionExpires, time.Now()))
}

// SignPreview returns the compact preview token "<expires>.<sid>.<token>", or "" for an invalid session
// ID or an unconfigured signer.
func SignPreview(sessionID string, expires int64) string {
	if !rnd.IsSessionID(sessionID) {
		return ""
	}

	token := Preview.Sign(expires, previewParams(sessionID))

	if token == "" {
		return ""
	}

	return strconv.FormatInt(expires, 10) + "." + sessionID + "." + token
}

// VerifyPreview verifies a signed preview token and returns the session ID it is bound to, or ok=false
// when it is forged, expired, or the signer is unconfigured.
func VerifyPreview(expires int64, sid, token string) (sessionID string, ok bool) {
	if !rnd.IsSessionID(sid) || !Preview.Valid(expires, previewParams(sid), token) {
		return "", false
	}

	return sid, true
}

// ParseCompact splits a compact signed token "<expires>.<sid>.<token>" and reports whether it has that
// shape: exactly two dots, a positive expiry in canonical decimal form, and a valid session ID.
func ParseCompact(value string) (expires int64, sid, token string, ok bool) {
	if strings.Count(value, ".") != 2 {
		return 0, "", "", false
	}

	parts := strings.SplitN(value, ".", 3)

	if !rnd.IsSessionID(parts[1]) || parts[2] == "" {
		return 0, "", "", false
	}

	expires, ok = parseExpires(parts[0])

	if !ok {
		return 0, "", "", false
	}

	return expires, parts[1], parts[2], true
}
