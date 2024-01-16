package header

import "strings"

// Content header names.
const (
	Accept             = "Accept"
	AcceptEncoding     = "Accept-Encoding"
	AcceptRanges       = "Accept-Ranges"
	ContentType        = "Content-Type"
	ContentTypeForm    = "application/x-www-form-urlencoded"
	ContentTypeJson    = "application/json"
	ContentDisposition = "Content-Disposition"
	ContentEncoding    = "Content-Encoding"
	ContentRange       = "Content-Range"
	Location           = "Location"
	Origin             = "Origin"
	Vary               = "Vary"
)

// Content header defaults.
var (
	DefaultVaryHeaders = []string{XAuthToken, AcceptEncoding}
	DefaultVary        = strings.Join(DefaultVaryHeaders, ", ")
)
