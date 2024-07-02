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

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/workers/auto"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/txt"
)

// WebDAVHandler wraps the http request handler so that it can be customized.
var WebDAVHandler = func(c *gin.Context, router *gin.RouterGroup, srv *webdav.Handler) {
	srv.ServeHTTP(c.Writer, c.Request)
}

// WebDAV handles requests to the "/originals" and "/import" endpoints.
func WebDAV(dir string, router *gin.RouterGroup, conf *config.Config) {
	if router == nil {
		log.Error("webdav: router is nil")
		return
	}

	if conf == nil {
		log.Error("webdav: conf is nil")
		return
	}

	// Native file system restricted to a specific directory.
	fileSystem := webdav.Dir(dir)
	lockSystem := mutex.WebDAV(dir)

	// Request logger function.
	loggerFunc := func(request *http.Request, err error) {
		if err != nil {
			switch request.Method {
			case MethodPut, MethodPost, MethodPatch, MethodDelete, MethodCopy, MethodMove:
				log.Errorf("webdav: %s in %s %s", clean.Error(err), clean.Log(request.Method), clean.Log(request.URL.String()))
			case MethodPropfind:
				log.Tracef("webdav: %s in %s %s", clean.Error(err), clean.Log(request.Method), clean.Log(request.URL.String()))
			default:
				log.Debugf("webdav: %s in %s %s", clean.Error(err), clean.Log(request.Method), clean.Log(request.URL.String()))
			}
		} else {
			// Determine the filename if it is an uploaded file and process custom request headers, if any.
			if fileName := WebDAVFileName(request, router, conf); fileName != "" {
				// Flag the uploaded file as favorite if the "X-Favorite" header is set to "1".
				if request.Header.Get(header.XFavorite) == "1" {
					WebDAVSetFavoriteFlag(fileName)
				}

				// Set the file modification time based on the Unix timestamp found in the "X-OC-MTime" header.
				if fileMtime := txt.Int64(request.Header.Get(header.XModTime)); fileMtime > 0 {
					WebDAVSetFileMtime(fileName, fileMtime)
				}
			}

			switch request.Method {
			case MethodPut, MethodPost, MethodPatch, MethodDelete, MethodCopy, MethodMove:
				log.Infof("webdav: %s %s", clean.Log(request.Method), clean.Log(request.URL.String()))

				if router.BasePath() == conf.BaseUri(WebDAVOriginals) {
					auto.ShouldIndex()
				} else if router.BasePath() == conf.BaseUri(WebDAVImport) {
					auto.ShouldImport()
				}
			default:
				log.Tracef("webdav: %s %s", clean.Log(request.Method), clean.Log(request.URL.String()))
			}
		}
	}

	// WebDAV request handler.
	srv := &webdav.Handler{
		Prefix:     router.BasePath(),
		FileSystem: fileSystem,
		LockSystem: lockSystem,
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

// WebDAVFileName determines the name and path of an uploaded file and returns its name if it exists.
func WebDAVFileName(request *http.Request, router *gin.RouterGroup, conf *config.Config) (fileName string) {
	// Check if this is a PUT request, as used for file uploads.
	if request.Method != MethodPut {
		return ""
	}

	basePath := router.BasePath()

	// Determine the absolute file path based on the request URL and the configuration.
	switch basePath {
	case conf.BaseUri(WebDAVOriginals):
		fileName = filepath.Join(conf.OriginalsPath(), strings.TrimPrefix(request.URL.Path, basePath))
	case conf.BaseUri(WebDAVImport):
		fileName = filepath.Join(conf.ImportPath(), strings.TrimPrefix(request.URL.Path, basePath))
	default:
		return ""
	}

	// Check if the file actually exists and return an empty string otherwise.
	if !fs.FileExists(fileName) {
		return ""
	}

	return fileName
}

// WebDAVSetFavoriteFlag adds the favorite flag to files uploaded via WebDAV.
func WebDAVSetFavoriteFlag(fileName string) {
	yamlName := fs.AbsPrefix(fileName, false) + fs.ExtYAML

	// Abort if YAML file already exists to avoid overwriting metadata.
	if fs.FileExists(yamlName) {
		log.Warnf("webdav: %s already exists", clean.Log(filepath.Base(yamlName)))
		return
	}

	// Make sure directory exists.
	if err := fs.MkdirAll(filepath.Dir(yamlName)); err != nil {
		log.Errorf("webdav: %s", err.Error())
		return
	}

	// Write YAML data to file.
	if err := fs.WriteFile(yamlName, []byte("Favorite: true\n")); err != nil {
		log.Errorf("webdav: %s", err.Error())
		return
	}

	// Log success.
	log.Infof("webdav: flagged %s as favorite", clean.Log(filepath.Base(fileName)))
}

// WebDAVSetFileMtime updaters the file modification time based on a Unix timestamp string.
func WebDAVSetFileMtime(fileName string, mtimeUnix int64) {
	if mtime := time.Unix(mtimeUnix, 0); mtimeUnix <= 0 || mtime.IsZero() || time.Now().Before(mtime) {
		log.Warnf("webdav: invalid mtime provided for %s", clean.Log(filepath.Base(fileName)))
	} else if mtimeErr := os.Chtimes(fileName, time.Time{}, mtime); mtimeErr != nil {
		log.Warnf("webdav: failed to set mtime for %s", clean.Log(filepath.Base(fileName)))
	} else {
		log.Infof("webdav: set mtime for %s", clean.Log(filepath.Base(fileName)))
	}
}
