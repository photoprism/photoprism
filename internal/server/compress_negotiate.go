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

// NegotiateEncoding picks an encoding from prefs that the client accepts per
// RFC 9110 §12.5.3 (q=0 refusal beats "*", missing q defaults to 1), returning
// EncodingIdentity when none is acceptable. Never returns an encoding outside prefs.
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

// parseAcceptEncoding parses an Accept-Encoding header into a map of lowercase
// tokens to q-values; missing or malformed q defaults to 1.
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
