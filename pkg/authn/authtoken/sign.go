package authtoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// Sign returns a bunny.net-compatible Advanced token authenticating a request. The signature covers
// the message signaturePath + expires + signingData + clientIP with HMAC-SHA256, base64url-encoded
// (no padding) behind the HS256- prefix, exactly as the bunny.net edge computes it. signaturePath is
// the URL path or a path prefix (a directory token); clientIP is empty unless IP locking is used.
func Sign(key []byte, signaturePath string, expires int64, params url.Values, clientIP string) string {
	msg := signaturePath + strconv.FormatInt(expires, 10) + SigningData(params) + clientIP

	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(msg))

	return Prefix + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// SigningData renders the signed query parameters the way bunny.net expects: every parameter except
// "token" and "expires", sorted by key, joined as key=value pairs with "&". Returns an empty string
// when no signable parameters are present.
func SigningData(params url.Values) string {
	if len(params) == 0 {
		return ""
	}

	keys := make([]string, 0, len(params))

	for k := range params {
		if k == "token" || k == "expires" {
			continue
		}
		keys = append(keys, k)
	}

	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))

	for _, k := range keys {
		pairs = append(pairs, k+"="+params.Get(k))
	}

	return strings.Join(pairs, "&")
}
