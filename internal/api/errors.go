package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt"
)

var (
	ErrUnauthorized    = gin.H{"code": http.StatusUnauthorized, "error": txt.UcFirst(config.ErrUnauthorized.Error())}
	ErrReadOnly        = gin.H{"code": http.StatusForbidden, "error": txt.UcFirst(config.ErrReadOnly.Error())}
	ErrUploadNSFW      = gin.H{"code": http.StatusForbidden, "error": txt.UcFirst(config.ErrUploadNSFW.Error())}
	ErrAlbumNotFound   = gin.H{"code": http.StatusNotFound, "error": "Album not found"}
	ErrPhotoNotFound   = gin.H{"code": http.StatusNotFound, "error": "Photo not found"}
	ErrUnexpectedError = gin.H{"code": http.StatusInternalServerError, "error": "Unexpected error"}
)
