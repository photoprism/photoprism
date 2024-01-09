package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/internal/auto"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/txt"
)

const WebDAVOriginals = "/originals"
const WebDAVImport = "/import"

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

var WebDAVHandler = func(c *gin.Context, router *gin.RouterGroup, srv *webdav.Handler) {
	srv.ServeHTTP(c.Writer, c.Request)
}

// WebDAV handles requests to the /originals and /import endpoints.
func WebDAV(filePath string, router *gin.RouterGroup, conf *config.Config) {
	if router == nil {
		log.Error("webdav: router is nil")
		return
	}

	if conf == nil {
		log.Error("webdav: conf is nil")
		return
	}

	// Native file system restricted to a specific directory.
	fileSystem := webdav.Dir(filePath)

	// Request logger function.
	loggerFunc := func(r *http.Request, err error) {
		if err != nil {
			switch r.Method {
			case MethodPut, MethodPost, MethodPatch, MethodDelete, MethodCopy, MethodMove:
				log.Errorf("webdav: %s in %s %s", clean.Error(err), clean.Log(r.Method), clean.Log(r.URL.String()))
			case MethodPropfind:
				log.Tracef("webdav: %s in %s %s", clean.Error(err), clean.Log(r.Method), clean.Log(r.URL.String()))
			default:
				log.Debugf("webdav: %s in %s %s", clean.Error(err), clean.Log(r.Method), clean.Log(r.URL.String()))
			}
		} else {
			// Handle optional upload request headers.
			if r.Method == MethodPut {
				var fileName string

				// Determine name and path of the uploaded file.
				if router.BasePath() == conf.BaseUri(WebDAVOriginals) {
					fileName = filepath.Join(conf.OriginalsPath(), strings.TrimPrefix(r.URL.Path, router.BasePath()))
				} else if router.BasePath() == conf.BaseUri(WebDAVImport) {
					fileName = filepath.Join(conf.ImportPath(), strings.TrimPrefix(r.URL.Path, router.BasePath()))
				}

				if fs.FileExists(fileName) {
					// Flag the uploaded file as favorite if the "X-Favorite" header is set to "1".
					if r.Header.Get(header.XFavorite) == "1" {
						FlagUploadAsFavorite(fileName)
					}

					// Set the file modification time based on the Unix timestamp found in the "X-OC-MTime" header.
					if mtimeUnix := txt.Int64(r.Header.Get(header.XModTime)); mtimeUnix <= 0 {
						// Ignore, as no Unix timestamp was provided.
					} else if mtime := time.Unix(mtimeUnix, 0); mtime.IsZero() || time.Now().Before(mtime) {
						log.Warnf("webdav: invalid modtime provided for %s", clean.Log(filepath.Base(fileName)))
					} else if mtimeErr := os.Chtimes(fileName, time.Time{}, mtime); mtimeErr != nil {
						log.Warnf("webdav: failed to set modtime for %s", clean.Log(filepath.Base(fileName)))
					} else {
						log.Infof("webdav: set modtime for %s", clean.Log(filepath.Base(fileName)))
					}
				}
			}

			switch r.Method {
			case MethodPut, MethodPost, MethodPatch, MethodDelete, MethodCopy, MethodMove:
				log.Infof("webdav: %s %s", clean.Log(r.Method), clean.Log(r.URL.String()))

				if router.BasePath() == conf.BaseUri(WebDAVOriginals) {
					auto.ShouldIndex()
				} else if router.BasePath() == conf.BaseUri(WebDAVImport) {
					auto.ShouldImport()
				}
			default:
				log.Tracef("webdav: %s %s", clean.Log(r.Method), clean.Log(r.URL.String()))
			}
		}
	}

	// WebDAV request handler.
	srv := &webdav.Handler{
		Prefix:     router.BasePath(),
		FileSystem: fileSystem,
		LockSystem: webdav.NewMemLS(),
		Logger:     loggerFunc,
	}

	// Request handler wrapper function.
	handlerFunc := func(c *gin.Context) {
		WebDAVHandler(c, router, srv)
	}

	// handleRead registers WebDAV methods used for browsing and downloading.
	handleRead := func(h func(*gin.Context)) {
		router.Handle(MethodHead, "/*path", h)
		router.Handle(MethodGet, "/*path", h)
		router.Handle(MethodOptions, "/*path", h)
		router.Handle(MethodLock, "/*path", h)
		router.Handle(MethodUnlock, "/*path", h)
		router.Handle(MethodPropfind, "/*path", h)
	}

	// handleWrite registers WebDAV methods to may modify the file system.
	handleWrite := func(h func(*gin.Context)) {
		router.Handle(MethodPut, "/*path", h)
		router.Handle(MethodPost, "/*path", h)
		router.Handle(MethodPatch, "/*path", h)
		router.Handle(MethodDelete, "/*path", h)
		router.Handle(MethodMkcol, "/*path", h)
		router.Handle(MethodCopy, "/*path", h)
		router.Handle(MethodMove, "/*path", h)
		router.Handle(MethodProppatch, "/*path", h)
	}

	// Handle supported WebDAV request methods.
	handleRead(handlerFunc)

	// Only supported with read-only mode disabled.
	if conf.ReadOnly() {
		handleWrite(func(c *gin.Context) {
			_ = c.AbortWithError(http.StatusForbidden, fmt.Errorf("forbidden in read-only mode"))
		})
	} else {
		handleWrite(handlerFunc)
	}
}

// FlagUploadAsFavorite adds the favorite flag to files uploaded via WebDAV.
func FlagUploadAsFavorite(fileName string) {
	yamlName := fs.AbsPrefix(fileName, false) + fs.ExtYAML

	// Abort if YAML file already exists to avoid overwriting metadata.
	if fs.FileExists(yamlName) {
		log.Warnf("webdav: %s already exists", clean.Log(filepath.Base(yamlName)))
		return
	}

	// Make sure directory exists.
	if err := os.MkdirAll(filepath.Dir(yamlName), fs.ModeDir); err != nil {
		log.Errorf("webdav: %s", err.Error())
		return
	}

	// Write YAML data to file.
	if err := os.WriteFile(yamlName, []byte("Favorite: true\n"), fs.ModeFile); err != nil {
		log.Errorf("webdav: %s", err.Error())
		return
	}

	// Log success.
	log.Infof("webdav: flagged %s as favorite", clean.Log(filepath.Base(fileName)))
}
