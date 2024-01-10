package server

import (
	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
)

const (
	WebDAVOriginals = "/originals"
	WebDAVImport    = "/import"
)

// registerWebDAVRoutes configures the built-in WebDAV server.
func registerWebDAVRoutes(router *gin.Engine, conf *config.Config) {
	if conf.DisableWebDAV() {
		log.Info("webdav: disabled")
	} else {
		var info string
		if conf.ReadOnly() {
			info = " in read-only mode"
		} else {
			info = ""
		}

		WebDAV(conf.OriginalsPath(), router.Group(conf.BaseUri(WebDAVOriginals), WebDAVAuth(conf)), conf)
		log.Infof("webdav: shared %s/%s", conf.BaseUri(WebDAVOriginals), info)

		if conf.ImportPath() != "" {
			WebDAV(conf.ImportPath(), router.Group(conf.BaseUri(WebDAVImport), WebDAVAuth(conf)), conf)
			log.Infof("webdav: shared %s/%s", conf.BaseUri(WebDAVImport), info)
		}
	}
}
