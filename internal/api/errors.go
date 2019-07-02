package api

import (
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/util"
)

var (
	ErrReadOnly = gin.H{"error": util.UcFirst(config.ErrReadOnly.Error())}
)
