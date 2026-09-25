package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/pkg/http/header"
)

// ErrUnsupportedContentType reports a request body in an encoding the endpoint does not accept.
var ErrUnsupportedContentType = errors.New("unsupported content type")

// BindOAuthRequest binds an OAuth2 request body with the JSON, form, or multipart
// decoder. These are the encodings the token and revoke endpoints accept: RFC 6749
// specifies form, and the Web UI sends JSON. Any other encoding is refused, so that
// a body the endpoint cannot read is reported rather than left silently unbound.
func BindOAuthRequest(c *gin.Context, frm any) error {
	switch {
	case header.HasContentType(&c.Request.Header, header.ContentTypeJson):
		return c.ShouldBindWith(frm, binding.JSON)
	case header.HasContentType(&c.Request.Header, header.ContentTypeMultipart):
		return c.ShouldBindWith(frm, binding.FormMultipart)
	case c.ContentType() == "", IsOAuthFormRequest(c):
		// Values are read from the request body, so that a credential is never
		// taken from a query string, where it would reach logs along the URL.
		return c.ShouldBindWith(frm, binding.FormPost)
	default:
		return ErrUnsupportedContentType
	}
}

// IsOAuthFormRequest reports whether an OAuth2 request carries a form-encoded body.
// The handlers share this with BindOAuthRequest so that the encoding a route gates
// on and the one it decodes are a single decision.
func IsOAuthFormRequest(c *gin.Context) bool {
	return header.HasContentType(&c.Request.Header, header.ContentTypeForm)
}
