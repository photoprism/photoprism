package server

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"

	"github.com/photoprism/photoprism/internal/api"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/workers/auto"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/txt"
)

// WebDAVHandler wraps the http request handler so that it can be customized.
var WebDAVHandler = func(c *gin.Context, router *gin.RouterGroup, srv *webdav.Handler) {
	ServeWebDAV(c.Writer, c.Request, srv)
}

// WebDAVWriteMethod returns true for methods that modify WebDAV state.
func WebDAVWriteMethod(method string) bool {
	switch method {
	case header.MethodPut, header.MethodMkcol, header.MethodDelete, header.MethodMove, header.MethodCopy, header.MethodProppatch, header.MethodLock, header.MethodUnlock:
		return true
	default:
		return false
	}
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
			// Route WebDAV request errors to the console-only system log, not log.*.
			// x/net/webdav embeds absolute originals/import paths in its messages,
			// which must stay out of the browser log viewer and the persisted errors
			// table; operators still see full detail in the server console.
			switch {
			case request.Method == header.MethodMkcol && errors.Is(err, os.ErrExist):
				// MKCOL on an existing collection is a benign probe: sync clients such as
				// PhotoSync test for a directory before creating it — expected, not a failure.
				event.SystemDebug([]string{"webdav", "collection %s already exists"}, clean.Log(request.URL.String()))
			case WebDAVWriteMethod(request.Method):
				event.SystemError([]string{"webdav", "%s in %s %s"}, clean.Error(err), clean.Log(request.Method), clean.Log(request.URL.String()))
			default:
				event.SystemDebug([]string{"webdav", "%s in %s %s"}, clean.Error(err), clean.Log(request.Method), clean.Log(request.URL.String()))
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
			case header.MethodPut, header.MethodMkcol, header.MethodDelete, header.MethodMove, header.MethodCopy, header.MethodProppatch:
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

	// Create WebDAV request handler.
	srv := &webdav.Handler{
		Prefix:     router.BasePath(),
		FileSystem: fileSystem,
		LockSystem: lockSystem,
		Logger:     loggerFunc,
	}

	// Wrap handler to check quota and permissions.
	handlerFunc := func(c *gin.Context) {
		// PATCH is intentionally not supported by the x/net/webdav handler in
		// PhotoPrism and must return 405 instead of a generic 400 response.
		if c.Request.Method == header.MethodPatch {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}

		// Abort PUT and COPY requests if there
		// is not enough free storage to upload new files.
		switch c.Request.Method {
		case header.MethodPut, header.MethodCopy:
			if conf.InsufficientStorage() {
				c.AbortWithStatus(http.StatusInsufficientStorage)
				return
			}
		}

		// Bound an uploaded file to the configured originals size limit (when set) so a single
		// PUT cannot stream an unbounded body to disk; the free-storage check above only catches
		// the next request. No-op when no originals limit is configured.
		if c.Request.Method == header.MethodPut {
			if limit := conf.OriginalsLimitBytes(); limit > 0 {
				api.LimitRequestBodyBytes(c, limit)
			}
		}

		// Invoke handler callback.
		WebDAVHandler(c, router, srv)
	}

	// handleRead registers WebDAV methods used for browsing and downloading.
	handleRead := func(path string, h func(*gin.Context)) {
		router.Handle(header.MethodHead, path, h)
		router.Handle(header.MethodGet, path, h)
		router.Handle(header.MethodPost, path, h)
		router.Handle(header.MethodOptions, path, h)
		router.Handle(header.MethodPropfind, path, h)
	}

	// handleWrite registers WebDAV methods to may modify the file system.
	handleWrite := func(path string, h func(*gin.Context)) {
		router.Handle(header.MethodPut, path, h)
		router.Handle(header.MethodDelete, path, h)
		router.Handle(header.MethodMkcol, path, h)
		router.Handle(header.MethodCopy, path, h)
		router.Handle(header.MethodMove, path, h)
		router.Handle(header.MethodLock, path, h)
		router.Handle(header.MethodUnlock, path, h)
		router.Handle(header.MethodProppatch, path, h)
	}

	// handleUnsupported registers methods that should always return 405.
	handleUnsupported := func(path string, h func(*gin.Context)) {
		router.Handle(header.MethodPatch, path, h)
	}

	// Register both base and wildcard routes to avoid automatic slash redirects
	// on collection roots such as "/originals" and "/import".
	for _, route := range []string{"", "/*path"} {
		handleRead(route, handlerFunc)
		handleUnsupported(route, func(c *gin.Context) {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
		})

		// Only supported with read-only mode disabled.
		if conf.ReadOnly() {
			handleWrite(route, func(c *gin.Context) {
				c.AbortWithStatus(http.StatusMethodNotAllowed)
			})
		} else {
			handleWrite(route, handlerFunc)
		}
	}
}

// WebDAVFileName determines the name and path of an uploaded file and returns its name if it exists.
func WebDAVFileName(request *http.Request, router *gin.RouterGroup, conf *config.Config) (fileName string) {
	// Check if this is a PUT request, as used for file uploads.
	if request.Method != header.MethodPut {
		return ""
	}

	basePath := router.BasePath()

	// Determine the absolute file path based on the request URL and the configuration.
	switch basePath {
	case conf.BaseUri(WebDAVOriginals):
		// Resolve the requested path safely under OriginalsPath.
		rel := strings.TrimPrefix(request.URL.Path, basePath)
		// Make relative if a leading slash remains after trimming the base.
		rel = strings.TrimLeft(rel, "/\\")
		if name, err := joinUnderBase(conf.OriginalsPath(), rel); err == nil {
			fileName = name
		} else {
			return ""
		}
	case conf.BaseUri(WebDAVImport):
		// Resolve the requested path safely under ImportPath.
		rel := strings.TrimPrefix(request.URL.Path, basePath)
		rel = strings.TrimLeft(rel, "/\\")
		if name, err := joinUnderBase(conf.ImportPath(), rel); err == nil {
			fileName = name
		} else {
			return ""
		}
	default:
		return ""
	}

	// Check if the file actually exists and return an empty string otherwise.
	if !fs.FileExists(fileName) {
		return ""
	}

	return fileName
}

// joinUnderBase resolves a relative request name under baseDir using the shared
// safe-join, so WebDAV uploads cannot escape the originals or import directory.
// Absolute paths, Windows-style volume names, drive-letter prefixes, and
// parent-directory traversal are rejected, with containment verified via
// filepath.Rel rather than a string prefix.
func joinUnderBase(baseDir, rel string) (string, error) {
	return fs.SafeJoin(baseDir, rel)
}

// WebDAVSetFavoriteFlag adds the favorite flag to files uploaded via WebDAV.
func WebDAVSetFavoriteFlag(fileName string) {
	yamlName := fs.AbsPrefix(fileName, false) + fs.ExtYml

	// Abort if YAML file already exists to avoid overwriting metadata.
	if fs.FileExists(yamlName) {
		log.Warnf("webdav: %s already exists", clean.Log(filepath.Base(yamlName)))
		return
	}

	// Make sure directory exists.
	if err := fs.MkdirAll(filepath.Dir(yamlName)); err != nil {
		// Console-only: the error embeds the absolute sidecar path (see loggerFunc).
		event.SystemError([]string{"webdav", "%s"}, clean.Error(err))
		return
	}

	// Write YAML data to file.
	if err := fs.WriteFile(yamlName, []byte("Favorite: true\n"), fs.ModeConfigFile); err != nil {
		// Console-only: the error embeds the absolute sidecar path (see loggerFunc).
		event.SystemError([]string{"webdav", "%s"}, clean.Error(err))
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
