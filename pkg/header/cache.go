package header

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// The CacheControl request and response header field contains directives (instructions)
	// that control caching in browsers and shared caches (e.g. proxies, CDNs).
	// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
	CacheControl = "Cache-Control"

	// CacheControlDefault indicates that the response remains valid for 604800 seconds (7 days) after it
	// has been generated. Note that max-age is not the time that has elapsed since the response was received,
	// but the time that has elapsed since the response was created on the origin server.
	CacheControlDefault = "max-age=604800"

	// CacheControlNoStore indicates that caches of any kind (private or shared) should not store the response.
	CacheControlNoStore = "no-store"

	// CacheControlNoCache indicates that the response can be stored in caches, but must be validated with
	// the origin server before each reuse, even when the cache is not connected to the origin server.
	CacheControlNoCache = "no-cache"

	// CacheControlPublic indicates that the response can be stored in a shared cache.
	// Responses to requests with Authorization header fields must not be stored in a shared cache;
	// however, the public directive causes such responses to be stored in a shared cache.
	CacheControlPublic = "public"

	// CacheControlPrivate indicates that the response can only be stored in a private cache (e.g. browsers).
	// You should add the private directive for personalized content, especially for responses sent after login.
	CacheControlPrivate = "private"

	// CacheControlImmutable indicates that the response will not be updated while it's fresh.
	CacheControlImmutable = "immutable"
)

// CacheControl defaults.
var (
	CacheControlPublicDefault  = CacheControlPublic + ", " + CacheControlDefault  // public, max-age=604800
	CacheControlPrivateDefault = CacheControlPrivate + ", " + CacheControlDefault // private, max-age=604800
)

// CacheControlMaxAge returns a CacheControl header value based on the specified
// duration in seconds or the defaults if duration is not a positive number.
func CacheControlMaxAge(duration int, public bool) string {
	if duration < 0 {
		return CacheControlNoCache
	} else if duration > DurationYear {
		duration = DurationYear
	}

	switch {
	case duration > 0 && public:
		return "public, max-age=" + strconv.Itoa(duration)
	case duration > 0:
		return "private, max-age=" + strconv.Itoa(duration)
	case public:
		return CacheControlPublicDefault
	default:
		return CacheControlPrivateDefault
	}
}

// SetCacheControl adds a CacheControl header to the response based on the specified parameters.
// If maxAge is 0, the defaults will be used.
func SetCacheControl(c *gin.Context, duration int, public bool) {
	if c == nil {
		return
	} else if c.Writer == nil {
		return
	}

	c.Header(CacheControl, CacheControlMaxAge(duration, public))
}

// SetCacheControlImmutable adds a CacheControl header to the response based on the specified parameters
// and with the immutable directive set. If maxAge is 0, the defaults will be used.
func SetCacheControlImmutable(c *gin.Context, maxAge int, public bool) {
	if c == nil {
		return
	} else if c.Writer == nil {
		return
	}

	if maxAge < 0 {
		c.Header(CacheControl, CacheControlNoCache)
		return
	}

	c.Header(CacheControl, CacheControlMaxAge(maxAge, public)+", "+CacheControlImmutable)
}
