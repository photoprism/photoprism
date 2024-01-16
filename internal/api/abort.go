package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/pkg/clean"
)

func Abort(c *gin.Context, code int, id i18n.Message, params ...interface{}) {
	resp := i18n.NewResponse(code, id, params...)

	log.Debugf("api-v1: abort %s with code %d (%s)", clean.Log(c.FullPath()), code, strings.ToLower(resp.String()))

	c.AbortWithStatusJSON(code, resp)
}

func Error(c *gin.Context, code int, err error, id i18n.Message, params ...interface{}) {
	resp := i18n.NewResponse(code, id, params...)

	if err != nil {
		resp.Details = err.Error()
		log.Errorf("api-v1: error %s with code %d in %s (%s)", clean.Error(err), code, clean.Log(c.FullPath()), strings.ToLower(resp.String()))
	}

	c.AbortWithStatusJSON(code, resp)
}

// AbortUnauthorized aborts with status code 401.
func AbortUnauthorized(c *gin.Context) {
	Abort(c, http.StatusUnauthorized, i18n.ErrUnauthorized)
}

// AbortForbidden aborts with status code 403.
func AbortForbidden(c *gin.Context) {
	Abort(c, http.StatusForbidden, i18n.ErrForbidden)
}

// AbortNotFound aborts with status code 404.
func AbortNotFound(c *gin.Context) {
	Abort(c, http.StatusNotFound, i18n.ErrNotFound)
}

// AbortEntityNotFound aborts with status code 404.
func AbortEntityNotFound(c *gin.Context) {
	Abort(c, http.StatusNotFound, i18n.ErrEntityNotFound)
}

// AbortAlbumNotFound aborts with status code 404.
func AbortAlbumNotFound(c *gin.Context) {
	Abort(c, http.StatusNotFound, i18n.ErrAlbumNotFound)
}

func AbortSaveFailed(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, i18n.ErrSaveFailed)
}

func AbortDeleteFailed(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, i18n.ErrDeleteFailed)
}

func AbortUnexpectedError(c *gin.Context) {
	Abort(c, http.StatusInternalServerError, i18n.ErrUnexpected)
}

func AbortBadRequest(c *gin.Context) {
	Abort(c, http.StatusBadRequest, i18n.ErrBadRequest)
}

func AbortFeatureDisabled(c *gin.Context) {
	Abort(c, http.StatusForbidden, i18n.ErrFeatureDisabled)
}

func AbortBusy(c *gin.Context) {
	Abort(c, http.StatusTooManyRequests, i18n.ErrBusy)
}
