package api

import (
	_ "embed" // required for go:embed video placeholder
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/i18n"
)

//go:embed embed/video.mp4
var brokenVideo []byte

// Abort writes a localized error response and stops further handler processing.
func Abort(c *gin.Context, code int, id i18n.Message, params ...interface{}) {
	resp := i18n.NewResponse(code, id, params...)

	if code >= 400 {
		log.Debugf("api: aborting request with error code %d (%s)", code, strings.ToLower(resp.String()))
	} else {
		log.Debugf("api: aborting request with response code %d (%s)", code, strings.ToLower(resp.String()))
	}

	c.AbortWithStatusJSON(code, resp)
}

// Error aborts the request while attaching error details to the response payload.
func Error(c *gin.Context, code int, err error, id i18n.Message, params ...interface{}) {
	resp := i18n.NewResponse(code, id, params...)

	if err != nil {
		resp.Details = err.Error()

		if reqPath := c.FullPath(); reqPath == "" {
			log.Errorf("api: error %d %s (%s)", code, clean.Error(err), strings.ToLower(resp.String()))
		} else {
			log.Errorf("api: error %d %s in %s (%s)", code, clean.Error(err), clean.Log(reqPath), strings.ToLower(resp.String()))
		}
	}

	c.AbortWithStatusJSON(code, resp)
}

// AbortNotFound renders a "404 Not Found" error page or JSON response.
var AbortNotFound = func(c *gin.Context) {
	conf := get.Config()

	switch c.NegotiateFormat(gin.MIMEHTML, gin.MIMEJSON) {
	case gin.MIMEJSON:
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.Msg(i18n.ErrNotFound)})
	default:
		var redirect string

		// Redirect to site root if current path is different.
		if root, path := conf.BaseUri("/"), c.Request.URL.Path; path != "" && path != root {
			redirect = root
		}

		values := gin.H{
			"signUp":   config.SignUp,
			"config":   conf.ClientPublic(),
			"error":    i18n.Msg(i18n.ErrNotFound),
			"code":     http.StatusNotFound,
			"redirect": redirect,
		}

		c.HTML(http.StatusNotFound, "404.gohtml", values)
	}

	c.Abort()
}

// AbortUnauthorized aborts with status code 401.
func AbortUnauthorized(c *gin.Context) {
	Abort(c, http.StatusUnauthorized, i18n.ErrUnauthorized)
}

// AbortPaymentRequired aborts with status code 402.
func AbortPaymentRequired(c *gin.Context) {
	Abort(c, http.StatusPaymentRequired, i18n.ErrPaymentRequired)
}

// AbortForbidden aborts with status code 403.
func AbortForbidden(c *gin.Context) {
	Abort(c, http.StatusForbidden, i18n.ErrForbidden)
}

// AbortEntityNotFound aborts with status code 404.
func AbortEntityNotFound(c *gin.Context) {
	Abort(c, http.StatusNotFound, i18n.ErrEntityNotFound)
}

// AbortAlbumNotFound aborts with status code 404.
func AbortAlbumNotFound(c *gin.Context) {
	Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
}

// AbortSaveFailed aborts with a generic 500 "save failed" error.
func AbortSaveFailed(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
}

// AbortDeleteFailed aborts with a generic 500 "delete failed" error.
func AbortDeleteFailed(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, i18n.ErrDeleteFailed)
}

// AbortNotImplemented aborts with status 501 when a feature is unavailable.
func AbortNotImplemented(c *gin.Context) {
	Abort(c, http.StatusNotImplemented, i18n.ErrUnsupported)
}

// AbortUnexpectedError aborts with a generic 500 error.
func AbortUnexpectedError(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, i18n.ErrUnexpected)
}

// AbortBadRequest attaches validation details and responds with status 400.
func AbortBadRequest(c *gin.Context, errs ...error) {
	// Log and attach validation errors to the context.
	for _, err := range errs {
		if err != nil {
			// Add error message to the debug logs.
			log.Debugf("api: %s", err)

			// Attach error to the current context
			_ = c.Error(err)
		}
	}

	// Abort request with error 400.
	Abort(c, http.StatusBadRequest, i18n.ErrBadRequest)
}

// AbortFeatureDisabled aborts with a forbidden response when a feature is disabled.
func AbortFeatureDisabled(c *gin.Context) {
	Abort(c, http.StatusForbidden, i18n.ErrFeatureDisabled)
}

// AbortQuotaExceeded aborts with a forbidden response when quotas are exhausted.
func AbortQuotaExceeded(c *gin.Context) {
	Abort(c, http.StatusForbidden, i18n.ErrQuotaExceeded)
}

// AbortBusy responds with HTTP 429 to signal temporary overload.
func AbortBusy(c *gin.Context) {
	Abort(c, http.StatusTooManyRequests, i18n.ErrBusy)
}

// AbortInvalidName aborts with HTTP 400 when a name fails validation.
func AbortInvalidName(c *gin.Context) {
	Abort(c, http.StatusBadRequest, i18n.ErrInvalidName)
}

// AbortInvalidCredentials responds with HTTP 401 for failed authentication attempts.
func AbortInvalidCredentials(c *gin.Context) {
	if c != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": authn.ErrInvalidCredentials.Error(), "code": i18n.ErrInvalidCredentials, "message": i18n.Msg(i18n.ErrInvalidCredentials)})
	}
}

// AbortVideo writes a placeholder MP4 response for video errors.
func AbortVideo(c *gin.Context) {
	if c != nil {
		AbortVideoWithStatus(c, http.StatusOK)
	}
}

// AbortVideoWithStatus writes the placeholder MP4 response using the provided status code.
func AbortVideoWithStatus(c *gin.Context, code int) {
	if c != nil {
		c.Data(code, header.ContentTypeMp4AvcMain, brokenVideo)
	}
}
