package header

import (
	"net/http"
	"strings"
)

// Cross-Origin Resource Sharing (CORS) headers.
const (
	AccessControlAllowOrigin      = "Access-Control-Allow-Origin"      // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	AccessControlAllowCredentials = "Access-Control-Allow-Credentials" // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
	AccessControlAllowHeaders     = "Access-Control-Allow-Headers"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
	AccessControlAllowMethods     = "Access-Control-Allow-Methods"     // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
	AccessControlMaxAge           = "Access-Control-Max-Age"           // https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age
)

// CORS header defaults.
var (
	DefaultAccessControlAllowOrigin      = ""
	DefaultAccessControlAllowCredentials = ""
	SafeHeaders                          = []string{Accept, AcceptRanges, ContentDisposition, ContentEncoding, ContentRange, Location, Vary}
	DefaultAccessControlAllowHeaders     = strings.Join(SafeHeaders, ", ")
	SafeMethods                          = []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	DefaultAccessControlAllowMethods     = strings.Join(SafeMethods, ", ")
	DefaultAccessControlMaxAge           = "3600"
)
