//go:build debug
// +build debug

package api

import (
	"bytes"
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
	files "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/header"
)

//go:embed swagger.json
var swaggerJSON []byte

// GetDocs registers the Swagger API documentation endpoints.
func GetDocs(router *gin.RouterGroup) {
	// Get global configuration.
	conf := get.Config()

	// Serve swagger.json, with the default host "demo.photoprism.app" being replaced by the configured hostname.
	router.GET("swagger.json", func(c *gin.Context) {
		c.Data(http.StatusOK, header.ContentTypeJson, bytes.ReplaceAll(swaggerJSON, []byte("demo.photoprism.app"), []byte(conf.SiteHost())))
	})

	// Serve Swagger UI.
	if handler := swagger.WrapHandler(files.Handler, swagger.URL(conf.ApiUri()+"/swagger.json")); handler != nil {
		router.GET("/docs", handler)
		router.GET("/docs/*any", handler)
	}
}
