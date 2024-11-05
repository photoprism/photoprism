package header

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
	ContentTypeHtml      = "text/html; charset=utf-8"
	ContentTypeText      = "text/plain; charset=utf-8"
	ContentTypeSVG       = "image/svg+xml"
	ContentTypeAVC       = "video/mp4; codecs=\"avc1\""
)
