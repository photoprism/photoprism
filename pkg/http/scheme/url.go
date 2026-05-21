package scheme

import (
	"net/url"
	"strings"
	"sync"
)

// NormalizeBaseURLMaxLen caps the input length eligible for caching; longer
// strings still normalize correctly but bypass the cache so a single oversized
// input cannot bloat the cache map.
const NormalizeBaseURLMaxLen = 2048

// NormalizeBaseURLCacheEntries bounds the in-memory cache. In practice only a
// handful of distinct base URLs (SiteUrl, CdnUrl, possibly an empty string)
// are normalized over the process lifetime; the cap is a safety net so a
// runaway caller cannot grow the map without limit.
const NormalizeBaseURLCacheEntries = 128

var (
	normalizeBaseURLCache   = make(map[string]string, NormalizeBaseURLCacheEntries)
	normalizeBaseURLCacheMu sync.RWMutex
)

// NormalizeBaseURL returns a base URL with a trailing slash if s is not empty
// or does not contain only whitespace. The default port, query strings,
// and fragments are omitted, but Userinfo is preserved.
// If parsing fails, NormalizeBaseURL returns s with a trailing slash appended.
//
// Results are cached by raw input string. The output is a pure function of the
// input, so cache entries never need to be invalidated; the only bounds are
// the input length cap and the maximum number of cached entries.
func NormalizeBaseURL(s string) string {
	if len(s) > NormalizeBaseURLMaxLen {
		return normalizeBaseURL(s)
	}

	normalizeBaseURLCacheMu.RLock()
	if out, ok := normalizeBaseURLCache[s]; ok {
		normalizeBaseURLCacheMu.RUnlock()
		return out
	}
	normalizeBaseURLCacheMu.RUnlock()

	out := normalizeBaseURL(s)

	normalizeBaseURLCacheMu.Lock()
	if len(normalizeBaseURLCache) < NormalizeBaseURLCacheEntries {
		normalizeBaseURLCache[s] = out
	}
	normalizeBaseURLCacheMu.Unlock()

	return out
}

// normalizeBaseURL implements the uncached normalization used by NormalizeBaseURL.
func normalizeBaseURL(s string) string {
	s = strings.TrimSpace(s)

	if s == "" {
		return ""
	}

	u, err := url.Parse(s)

	if err != nil {
		return strings.TrimRight(s, "/") + "/"
	}

	StripDefaultPort(u)
	u.RawQuery = ""
	u.ForceQuery = false
	u.Fragment = ""
	u.RawFragment = ""
	u.Path = strings.TrimRight(u.Path, "/") + "/"

	return u.String()
}
