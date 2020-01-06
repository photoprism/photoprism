package api

import (
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/ling"
)

var (
	ErrReadOnly     = gin.H{"code": 403, "error": ling.UcFirst(config.ErrReadOnly.Error())}
	ErrUnauthorized = gin.H{"code": 401, "error": ling.UcFirst(config.ErrUnauthorized.Error())}
	ErrUploadNSFW   = gin.H{"code": 403, "error": ling.UcFirst(config.ErrUploadNSFW.Error())}
)
