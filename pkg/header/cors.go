package header

import (
	"net/http"
	"strings"
)

// Cross-Origin Resource Sharing (CORS) headers.
const (
	AccessControlAllowOrigin  = "Access-Control-Allow-Origin"  // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	AccessControlAllowHeaders = "Access-Control-Allow-Headers" // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
	AccessControlAllowMethods = "Access-Control-Allow-Methods" // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
	AccessControlMaxAge       = "Access-Control-Max-Age"       // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age
)

// CORS header defaults.
var (
	DefaultAccessControlAllowOrigin  = ""
	CorsHeaders                      = []string{Accept, AcceptRanges, ContentDisposition, ContentEncoding, ContentRange, Location}
	DefaultAccessControlAllowHeaders = strings.Join(CorsHeaders, ", ")
	CorsMethods                      = []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	DefaultAccessControlAllowMethods = strings.Join(CorsMethods, ", ")
	DefaultAccessControlMaxAge       = "3600"
	CorsExt                          = map[string]bool{".eot": true, ".ttf": true, ".woff": true, ".woff2": true, ".css": true}
)

// AllowCORS checks if CORS headers can be safely used based on a request's file path.
// See: https://www.w3.org/TR/css-fonts-3/#font-fetching-requirements
func AllowCORS(path string) bool {
	// Return false if path is empty.
	if path == "" {
		return false
	}

	// Extract extension from path.
	var ext string
	l := len(path) - 1
	for i := l; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			ext = path[i:]
			if l-len(ext) < 0 {
				// Return false if there is no filename.
				return false
			} else if r := path[i-1]; (r < '0' || r > '9') && (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
				// Return false if the filename is invalid.
				return false
			}
			break
		}
	}

	// Return false if path does not include an extension.
	if ext == "" {
		return false
	}

	// Check list of allowed extensions.
	return CorsExt[ext]
}
