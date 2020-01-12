package api

import (
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt"
)

var (
	ErrReadOnly     = gin.H{"code": 403, "error": txt.UcFirst(config.ErrReadOnly.Error())}
	ErrUnauthorized = gin.H{"code": 401, "error": txt.UcFirst(config.ErrUnauthorized.Error())}
	ErrUploadNSFW   = gin.H{"code": 403, "error": txt.UcFirst(config.ErrUploadNSFW.Error())}
)
