package server

import (
	"net/http"

	"github.com/photoprism/photoprism/pkg/txt"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/auto"
	"github.com/photoprism/photoprism/internal/config"
	"golang.org/x/net/webdav"
)

const WebDAVOriginals = "/originals"
const WebDAVImport = "/import"

// ANY /webdav/*
func WebDAV(path string, router *gin.RouterGroup, conf *config.Config) {
	if router == nil {
		log.Error("webdav: router is nil")
		return
	}

	if conf == nil {
		log.Error("webdav: conf is nil")
		return
	}

	f := webdav.Dir(path)

	srv := &webdav.Handler{
		Prefix:     router.BasePath(),
		FileSystem: f,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				switch r.Method {
				case MethodPut, MethodPost, MethodPatch, MethodDelete, MethodCopy, MethodMove:
					log.Errorf("webdav: %s in %s %s", txt.Quote(err.Error()), r.Method, r.URL)
				case MethodPropfind:
					log.Tracef("webdav: %s in %s %s", txt.Quote(err.Error()), r.Method, r.URL)
				default:
					log.Debugf("webdav: %s in %s %s", txt.Quote(err.Error()), r.Method, r.URL)
				}

			} else {
				switch r.Method {
				case MethodPut, MethodPost, MethodPatch, MethodDelete, MethodCopy, MethodMove:
					log.Infof("webdav: %s %s", r.Method, r.URL)

					if router.BasePath() == WebDAVOriginals {
						auto.ShouldIndex()
					} else if router.BasePath() == WebDAVImport {
						auto.ShouldImport()
					}
				default:
					log.Tracef("webdav: %s %s", r.Method, r.URL)
				}
			}
		},
	}

	handler := func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		srv.ServeHTTP(w, r)
	}

	router.Handle(MethodHead, "/*path", handler)
	router.Handle(MethodGet, "/*path", handler)
	router.Handle(MethodPut, "/*path", handler)
	router.Handle(MethodPost, "/*path", handler)
	router.Handle(MethodPatch, "/*path", handler)
	router.Handle(MethodDelete, "/*path", handler)
	router.Handle(MethodOptions, "/*path", handler)
	router.Handle(MethodMkcol, "/*path", handler)
	router.Handle(MethodCopy, "/*path", handler)
	router.Handle(MethodMove, "/*path", handler)
	router.Handle(MethodLock, "/*path", handler)
	router.Handle(MethodUnlock, "/*path", handler)
	router.Handle(MethodPropfind, "/*path", handler)
	router.Handle(MethodProppatch, "/*path", handler)
}
