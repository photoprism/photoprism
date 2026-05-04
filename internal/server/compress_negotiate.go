package server

import (
	"strconv"
	"strings"
)

// EncodingIdentity is the sentinel returned by NegotiateEncoding when no
// configured content-encoding is acceptable to the client.
const EncodingIdentity = "identity"

// EncodingGzip is the IANA token for the gzip content-encoding.
const EncodingGzip = "gzip"

// EncodingZstd is the IANA token for the zstd content-encoding (RFC 8478).
const EncodingZstd = "zstd"

// NegotiateEncoding selects the response content-encoding for a request given
// the operator's ordered preference list and the client's Accept-Encoding
// header. It returns one of the configured encodings (e.g. "zstd", "gzip") or
// EncodingIdentity ("identity") when no preference is acceptable.
//
// Negotiation follows RFC 9110 §12.5.3:
//
//   - A token with q=0 is an explicit refusal and disqualifies that encoding
//     even if the wildcard "*" would otherwise permit it.
//   - A "*" entry with q>0 acts as a fallback that accepts any server-offered
//     encoding not explicitly refused.
//   - A missing q parameter defaults to 1.
//   - An empty Accept-Encoding header means the client wants no encoding.
//
// The function never returns an encoding that does not appear in prefs.
func NegotiateEncoding(acceptEncoding string, prefs []string) string {
	if len(prefs) == 0 {
		return EncodingIdentity
	}

	accepted := parseAcceptEncoding(acceptEncoding)
	starQ, hasStar := accepted["*"]

	for _, enc := range prefs {
		enc = strings.ToLower(strings.TrimSpace(enc))
		if enc == "" {
			continue
		}
		if q, ok := accepted[enc]; ok {
			if q > 0 {
				return enc
			}
			// q=0 explicitly refuses this encoding; do not let "*" override it.
			continue
		}
		if hasStar && starQ > 0 {
			return enc
		}
	}
	return EncodingIdentity
}

// parseAcceptEncoding parses an HTTP Accept-Encoding header value into a map
// of lowercase encoding tokens to their q-values. Tokens without an explicit
// q parameter default to 1. Malformed q values are also treated as 1. Empty
// or missing tokens are silently dropped.
func parseAcceptEncoding(header string) map[string]float64 {
	result := make(map[string]float64)
	if header == "" {
		return result
	}
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		token := part
		q := 1.0
		if i := strings.IndexByte(part, ';'); i >= 0 {
			token = strings.TrimSpace(part[:i])
			for _, p := range strings.Split(part[i+1:], ";") {
				p = strings.TrimSpace(p)
				if !strings.HasPrefix(strings.ToLower(p), "q=") {
					continue
				}
				if v, err := strconv.ParseFloat(strings.TrimSpace(p[2:]), 64); err == nil {
					q = v
				}
			}
		}
		token = strings.ToLower(token)
		if token == "" {
			continue
		}
		result[token] = q
	}
	return result
}
