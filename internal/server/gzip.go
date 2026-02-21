package server

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/proxy"
)

// gzipExcludedExtensions contains file extensions that should never be gzip-compressed.
// These formats are already compressed or typically served as large binary payloads.
var gzipExcludedExtensions = map[string]struct{}{
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

// NewGzipShouldCompressFn returns a high-performance gzip decision function for PhotoPrism.
// It mirrors the legacy exclusion rules (extensions and path prefixes) and adds targeted
// route exclusions for binary/streaming endpoints that must not be compressed.
func NewGzipShouldCompressFn(conf *config.Config) func(c *gin.Context) bool {
	if conf == nil {
		return func(*gin.Context) bool { return false }
	}

	apiBase := conf.BaseUri(config.ApiUri)

	// Raw path fallbacks for dynamic exclusions in case FullPath is unavailable.
	sharePrefix := conf.BaseUri("/s/")
	photoDlPrefix := apiBase + "/photos/"
	clusterThemePath := apiBase + "/cluster/theme"

	// FullPath patterns (exact match) for dynamic routes that should bypass gzip.
	excludedFullPaths := map[string]struct{}{
		apiBase + "/photos/:uid/dl":               {},
		apiBase + "/cluster/theme":                {},
		conf.BaseUri("/s/:token/:shared/preview"): {},
	}

	// Path prefixes that should bypass gzip (prefix match on raw URL path).
	excludedPrefixes := []string{
		// Health endpoints are small and frequently polled; gzip would add overhead.
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
	}

	return func(c *gin.Context) bool {
		if c == nil || c.Request == nil {
			return false
		}

		// Only compress when the client explicitly accepts gzip and the connection is not upgraded.
		if !strings.Contains(strings.ToLower(c.GetHeader("Accept-Encoding")), "gzip") {
			return false
		}
		if strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade") {
			return false
		}

		path := c.Request.URL.Path
		if path == "" {
			return false
		}

		// Exclude known already-compressed/binary extensions.
		if ext := strings.ToLower(filepath.Ext(path)); ext != "" {
			if _, ok := gzipExcludedExtensions[ext]; ok {
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
