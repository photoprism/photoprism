//go:build debug
// +build debug

package api

import (
	"path/filepath"

	"github.com/photoprism/photoprism/pkg/fs"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	files "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

// GetDocs registers the Swagger API documentation endpoints.
func GetDocs(router *gin.RouterGroup) {
	conf := get.Config()
	swaggerFile := filepath.Join(conf.AssetsPath(), "docs/api/v1/swagger.json")

	if !fs.FileExistsNotEmpty(swaggerFile) {
		return
	}

	router.StaticFile("/swagger.json", swaggerFile)
	handler := swagger.WrapHandler(files.Handler, swagger.URL(conf.ApiUri()+"/swagger.json"))
	router.GET("/docs", handler)
	router.GET("/docs/*any", handler)
}
