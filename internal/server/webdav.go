package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/photoprism/photoprism/internal/config"
	"golang.org/x/net/webdav"
)

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
		Prefix: router.BasePath(),
		FileSystem: f,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("webdav: %s %s, ERROR: %s\n", r.Method, r.URL, err)
			} else {
				log.Printf("webdav: %s %s \n", r.Method, r.URL)
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
