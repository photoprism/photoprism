package header

import "strings"

// Standard content request and response header names.
const (
	Accept             = "Accept"
	AcceptEncoding     = "Accept-Encoding"
	AcceptRanges       = "Accept-Ranges"
	ContentType        = "Content-Type"
	ContentDisposition = "Content-Disposition"
	ContentEncoding    = "Content-Encoding"
	ContentRange       = "Content-Range"
	Location           = "Location"
	Origin             = "Origin"
	Vary               = "Vary"
)

// Standard ContentType header values.
const (
	ContentTypeForm      = "application/x-www-form-urlencoded"
	ContentTypeMultipart = "multipart/form-data"
	ContentTypeJson      = "application/json"
	ContentTypeJsonUtf8  = "application/json; charset=utf-8"
)

// Vary response header defaults.
//
// Requests that include a standard authorization header should be automatically excluded
// from public caches: https://datatracker.ietf.org/doc/html/rfc7234#section-3
var (
	DefaultVaryHeaders = []string{AcceptEncoding, XAuthToken}
	DefaultVary        = strings.Join(DefaultVaryHeaders, ", ")
)
