package server

import (
	"net/http"

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
				case "POST", "DELETE", "PUT", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK":
					log.Errorf("webdav-error: %s in %s %s", err, r.Method, r.URL)
				case "PROPFIND":
					log.Tracef("webdav-error: %s in %s %s", err, r.Method, r.URL)
				default:
					log.Debugf("webdav-error: %s in %s %s", err, r.Method, r.URL)
				}

			} else {
				switch r.Method {
				case "POST", "DELETE", "PUT", "COPY", "MOVE":
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

	router.Handle("OPTIONS", "/*path", handler)
	router.Handle("GET", "/*path", handler)
	router.Handle("HEAD", "/*path", handler)
	router.Handle("POST", "/*path", handler)
	router.Handle("DELETE", "/*path", handler)
	router.Handle("PUT", "/*path", handler)
	router.Handle("MKCOL", "/*path", handler)
	router.Handle("COPY", "/*path", handler)
	router.Handle("MOVE", "/*path", handler)
	router.Handle("LOCK", "/*path", handler)
	router.Handle("UNLOCK", "/*path", handler)
	router.Handle("PROPFIND", "/*path", handler)
	router.Handle("PROPPATCH", "/*path", handler)
}
