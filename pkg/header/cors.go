package header

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
	DefaultAccessControlAllowHeaders     = "Origin, Accept, Accept-Ranges, Content-Range"
	DefaultAccessControlAllowMethods     = "GET, HEAD, OPTIONS"
	DefaultAccessControlMaxAge           = "3600"
)
