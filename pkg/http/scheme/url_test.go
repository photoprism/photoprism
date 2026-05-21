package scheme

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeBaseURL(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		// Empty values.
		{"Empty", "", ""},
		{"Spaces", "             ", ""},
		{"Tabs", "\t\t\t\t ", ""},

		// Trailing-slash policy.
		{"AlreadyNormalized", "https://example.com/", "https://example.com/"},
		{"NoTrailingSlash", "https://example.com", "https://example.com/"},
		{"ExtraTrailingSlashes", "https://example.com:443////", "https://example.com/"},

		// Default-port stripping.
		{"HttpsDefaultPort", "https://example.com:443/", "https://example.com/"},
		{"HttpDefaultPort", "http://example.com:80/sub", "http://example.com/sub/"},
		{"NonDefaultPortPreserved", "https://example.com:8443/", "https://example.com:8443/"},
		{"MismatchedScheme", "http://example.com:443/", "http://example.com:443/"},

		// Uncommon but well-formed inputs.
		{"IPv6DefaultPort", "https://[::1]:443/", "https://[::1]/"},
		{"IPv6NonDefaultPort", "https://[2001:db8::1]:8443/path", "https://[2001:db8::1]:8443/path/"},
		{"PathPreserved", "https://example.com:443/i/pro-1/", "https://example.com/i/pro-1/"},
		{"QueryStripped", "https://example.com:443/i/?lang=de&page=2", "https://example.com/i/"},
		{"ForceQueryStripped", "https://example.com/?", "https://example.com/"},
		{"FragmentStripped", "https://example.com/library/#photo123", "https://example.com/library/"},

		// Policy: userinfo is preserved verbatim.
		{"UserinfoPreserved", "https://user:secret@example.com:443/", "https://user:secret@example.com/"},

		// Unix-socket schemes: port stripping must stay a no-op.
		{"UnixScheme", "unix:///var/run/photoprism.sock", "unix:///var/run/photoprism.sock/"},
		{"HttpUnixScheme", "http+unix:///var/run/photoprism.sock", "http+unix:///var/run/photoprism.sock/"},

		// Parse failure falls back to TrimRight + "/".
		{"ParseError", ":foo", ":foo/"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, NormalizeBaseURL(tc.in))
		})
	}
}

func TestNormalizeBaseURL_CacheHit(t *testing.T) {
	in := "https://cache-hit.example.com:443/"
	want := "https://cache-hit.example.com/"

	first := NormalizeBaseURL(in)
	second := NormalizeBaseURL(in)

	assert.Equal(t, want, first)
	assert.Equal(t, want, second)

	normalizeBaseURLCacheMu.RLock()
	_, ok := normalizeBaseURLCache[in]
	normalizeBaseURLCacheMu.RUnlock()
	assert.True(t, ok, "cache must retain the entry after first call")
}

func TestNormalizeBaseURL_OversizedBypassesCache(t *testing.T) {
	in := "https://bypass.example.com/" + strings.Repeat("x", NormalizeBaseURLMaxLen)
	out := NormalizeBaseURL(in)

	assert.True(t, strings.HasPrefix(out, "https://bypass.example.com/"))

	normalizeBaseURLCacheMu.RLock()
	_, ok := normalizeBaseURLCache[in]
	normalizeBaseURLCacheMu.RUnlock()
	assert.False(t, ok, "inputs longer than NormalizeBaseURLMaxLen must not be cached")
}

func BenchmarkNormalizeBaseURL_Cached(b *testing.B) {
	in := "https://bench.example.com:443/library/"
	NormalizeBaseURL(in) // warm the cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NormalizeBaseURL(in)
	}
}

func BenchmarkNormalizeBaseURL_Uncached(b *testing.B) {
	in := "https://bench.example.com:443/library/"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeBaseURL(in)
	}
}

func TestNormalizeBaseURL_ConcurrentReaders(t *testing.T) {
	in := "https://concurrent.example.com:443/path/"
	want := "https://concurrent.example.com/path/"

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 256; j++ {
				assert.Equal(t, want, NormalizeBaseURL(in))
			}
		}()
	}
	wg.Wait()
}
