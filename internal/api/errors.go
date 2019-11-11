package api

import (
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"
)

var (
	ErrReadOnly = gin.H{"code": 403, "error": util.UcFirst(config.ErrReadOnly.Error())}
	ErrUnauthorized = gin.H{"code": 401, "error": util.UcFirst(config.ErrUnauthorized.Error())}
)
