package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/txt"
)

var (
	ErrUnauthorized     = gin.H{"code": http.StatusUnauthorized, "error": txt.UcFirst(config.ErrUnauthorized.Error())}
	ErrReadOnly         = gin.H{"code": http.StatusForbidden, "error": txt.UcFirst(config.ErrReadOnly.Error())}
	ErrUploadNSFW       = gin.H{"code": http.StatusForbidden, "error": txt.UcFirst(config.ErrUploadNSFW.Error())}
	ErrPublic           = gin.H{"code": http.StatusForbidden, "error": "Not available in public mode"}
	ErrAccountNotFound  = gin.H{"code": http.StatusNotFound, "error": "Account not found"}
	ErrConnectionFailed = gin.H{"code": http.StatusConflict, "error": "Failed to connect"}
	ErrAlbumNotFound    = gin.H{"code": http.StatusNotFound, "error": "Album not found"}
	ErrPhotoNotFound    = gin.H{"code": http.StatusNotFound, "error": "Photo not found"}
	ErrLabelNotFound    = gin.H{"code": http.StatusNotFound, "error": "Label not found"}
	ErrFileNotFound     = gin.H{"code": http.StatusNotFound, "error": "File not found"}
	ErrSessionNotFound  = gin.H{"code": http.StatusNotFound, "error": "Session not found"}
	ErrUnexpectedError  = gin.H{"code": http.StatusInternalServerError, "error": "Unexpected error"}
	ErrSaveFailed       = gin.H{"code": http.StatusInternalServerError, "error": "Changes could not be saved"}
	ErrDeleteFailed     = gin.H{"code": http.StatusInternalServerError, "error": "Changes could not be saved"}
	ErrFormInvalid      = gin.H{"code": http.StatusBadRequest, "error": "Changes could not be saved"}
	ErrFeatureDisabled  = gin.H{"code": http.StatusForbidden, "error": "Feature disabled"}
	ErrNotFound         = gin.H{"code": http.StatusNotFound, "error": "Not found"}
	ErrInvalidPassword  = gin.H{"code": http.StatusBadRequest, "error": "Invalid password, please try again"}
)

func GetErrors(router *gin.RouterGroup) {
	router.GET("/errors", func(c *gin.Context) {
		s := Auth(SessionID(c), acl.ResourceLogs, acl.ActionSearch)

		if s.Invalid() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrUnauthorized)
			return
		}

		limit := txt.Int(c.Query("count"))
		offset := txt.Int(c.Query("offset"))

		if resp, err := query.Errors(limit, offset); err != nil {
			c.AbortWithStatusJSON(400, gin.H{"error": txt.UcFirst(err.Error())})
			return
		} else {
			c.Header("X-Count", strconv.Itoa(len(resp)))
			c.Header("X-Limit", strconv.Itoa(limit))
			c.Header("X-Offset", strconv.Itoa(offset))

			c.JSON(http.StatusOK, resp)
		}
	})
}
