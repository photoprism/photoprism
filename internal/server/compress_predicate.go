package server

import (
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/http/proxy"
)

// GzipExcludedExtensions contains file extensions that should never be gzip-compressed.
// These formats are already compressed or typically served as large binary payloads.
var GzipExcludedExtensions = map[string]struct{}{
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

// ShouldExcludeGzipExt returns true if the given file extension should not be gzip-compressed.
func ShouldExcludeGzipExt(ext string) bool {
	_, ok := GzipExcludedExtensions[strings.ToLower(ext)]
	return ok
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
		return shouldCompressGzip(c, excludedFullPaths, excludedPrefixes, clusterThemePath, photoDlPrefix, sharePrefix)
	}
}

// shouldCompressGzip is the core decision logic for gzip compression.
// It is separated from NewGzipShouldCompressFn to enable unit testing.
func shouldCompressGzip(c *gin.Context, excludedFullPaths map[string]struct{}, excludedPrefixes []string, clusterThemePath, photoDlPrefix, sharePrefix string) bool {
	if c == nil || c.Request == nil {
		return false
	}

	// Only compress when the client explicitly accepts gzip and the connection is not upgraded.
	if !clientAcceptsGzip(c) {
		return false
	}
	if isConnectionUpgrade(c) {
		return false
	}

	path := c.Request.URL.Path
	if path == "" {
		return false
	}

	// Exclude known already-compressed/binary extensions.
	if ext := strings.ToLower(filepath.Ext(path)); ext != "" {
		if ShouldExcludeGzipExt(ext) {
			return false
		}
	}

	// Exclude configured prefix groups.
	if matchesPrefixExclusion(path, excludedPrefixes) {
		return false
	}

	// Exclude matched route patterns for dynamic endpoints.
	if full := c.FullPath(); full != "" {
		if _, ok := excludedFullPaths[full]; ok {
			return false
		}
	}

	// Fallback exclusions using raw path checks for robustness.
	if matchesFallbackExclusion(path, clusterThemePath, photoDlPrefix, sharePrefix) {
		return false
	}

	return true
}

// clientAcceptsGzip checks if the client accepts gzip encoding.
func clientAcceptsGzip(c *gin.Context) bool {
	return strings.Contains(strings.ToLower(c.GetHeader("Accept-Encoding")), "gzip")
}

// isConnectionUpgrade checks if the connection is being upgraded (e.g., WebSocket).
func isConnectionUpgrade(c *gin.Context) bool {
	return strings.Contains(strings.ToLower(c.GetHeader("Connection")), "upgrade")
}

// matchesPrefixExclusion checks if the path matches any excluded prefix.
func matchesPrefixExclusion(path string, excludedPrefixes []string) bool {
	for _, prefix := range excludedPrefixes {
		if prefix != "" && strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// matchesFallbackExclusion checks for fallback exclusions using raw path checks.
// Note: Keep the prefix guard here (not just HasSuffix), as the frontend SPA
// wildcard route may include paths ending in "/preview" (HTML) that should
// remain compressible (e.g., "/library/.../preview").
func matchesFallbackExclusion(path, clusterThemePath, photoDlPrefix, sharePrefix string) bool {
	if path == clusterThemePath {
		return true
	}
	if strings.HasPrefix(path, photoDlPrefix) && strings.HasSuffix(path, "/dl") {
		return true
	}
	if strings.HasPrefix(path, sharePrefix) && strings.HasSuffix(path, "/preview") {
		return true
	}
	return false
}
