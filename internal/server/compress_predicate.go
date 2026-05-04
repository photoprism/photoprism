package server

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/proxy"
)

// excludedExtensions lists file extensions whose responses should never be
// HTTP-compressed. These formats are already compressed or typically served
// as large binary payloads where gzip/zstd add CPU cost without saving bytes.
var excludedExtensions = map[string]struct{}{
	".png":  {},
	".gif":  {},
	".jpeg": {},
	".jpg":  {},
	".webp": {},
	".mp3":  {},
	".mp4":  {},
	".zip":  {},
	".gz":   {},
}

// NewShouldCompressFn returns an encoding-agnostic predicate that decides
// whether a given request is eligible for HTTP response compression. The
// predicate inspects the request URL path, file extension, and matched Gin
// route pattern; it deliberately does not look at Accept-Encoding or
// Connection headers (the middleware owns those checks). When conf is nil it
// returns a predicate that always declines.
func NewShouldCompressFn(conf *config.Config) func(c *gin.Context) bool {
	if conf == nil {
		return func(*gin.Context) bool { return false }
	}

	apiBase := conf.BaseUri(config.ApiUri)

	// Raw path fallbacks for dynamic exclusions in case FullPath is unavailable.
	sharePrefix := conf.BaseUri("/s/")
	photoDlPrefix := apiBase + "/photos/"
	clusterThemePath := apiBase + "/cluster/theme"

	// FullPath patterns (exact match) for dynamic routes that should bypass compression.
	excludedFullPaths := map[string]struct{}{
		apiBase + "/photos/:uid/dl":               {},
		apiBase + "/cluster/theme":                {},
		conf.BaseUri("/s/:token/:shared/preview"): {},
	}

	// Path prefixes that should bypass compression (prefix match on raw URL path).
	excludedPrefixes := []string{
		// Health endpoints are small and frequently polled; compression would add overhead.
		conf.BaseUri("/livez"),
		conf.BaseUri("/health"),
		conf.BaseUri("/readyz"),
		conf.BaseUri(config.ApiUri + "/t"),
		conf.BaseUri(config.ApiUri + "/folders/t"),
		conf.BaseUri(config.ApiUri + "/dl"),
		conf.BaseUri(config.ApiUri + "/zip"),
		conf.BaseUri(config.ApiUri + "/albums"),
		conf.BaseUri(config.ApiUri + "/labels"),
		conf.BaseUri(config.ApiUri + "/videos"),
		conf.BaseUri(proxy.PathPrefix),
		// Bundled and custom static assets are served with precompressed
		// .zst / .gz siblings via PrecompressedStatic; bypass the runtime
		// encoder so it never re-encodes an already-encoded body and so
		// PHOTOPRISM_HTTP_COMPRESSION=none consistently disables every
		// encoded code path on these routes.
		conf.BaseUri(config.StaticUri),
		conf.BaseUri(config.CustomStaticUri),
	}

	return func(c *gin.Context) bool {
		if c == nil || c.Request == nil {
			return false
		}

		path := c.Request.URL.Path
		if path == "" {
			return false
		}

		// Exclude known already-compressed/binary extensions.
		if ext := strings.ToLower(filepath.Ext(path)); ext != "" {
			if _, ok := excludedExtensions[ext]; ok {
				return false
			}
		}

		// Exclude configured prefix groups.
		for _, prefix := range excludedPrefixes {
			if prefix != "" && strings.HasPrefix(path, prefix) {
				return false
			}
		}

		// Exclude matched route patterns for dynamic endpoints.
		if full := c.FullPath(); full != "" {
			if _, ok := excludedFullPaths[full]; ok {
				return false
			}
		}

		// Fallback exclusions using raw path checks for robustness.
		// Note: Keep the prefix guard here (not just HasSuffix), as the frontend SPA
		// wildcard route may include paths ending in "/preview" (HTML) that should
		// remain compressible (e.g., "/library/.../preview").
		if path == clusterThemePath {
			return false
		}
		if strings.HasPrefix(path, photoDlPrefix) && strings.HasSuffix(path, "/dl") {
			return false
		}
		if strings.HasPrefix(path, sharePrefix) && strings.HasSuffix(path, "/preview") {
			return false
		}

		return true
	}
}
