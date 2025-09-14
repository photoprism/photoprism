package api

import (
	"archive/zip"
	gofs "io/fs"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

// ClusterGetTheme returns custom theme files as zip, if available.
//
//	@Summary	returns custom theme files as zip, if available
//	@Id			ClusterGetTheme
//	@Tags		Cluster
//	@Produce	application/zip
//	@Success	200				{file}		application/zip
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Router		/api/v1/cluster/theme [get]
func ClusterGetTheme(router *gin.RouterGroup) {
	router.GET("/cluster/theme", func(c *gin.Context) {
		// Check if client has cluster download privileges.
		s := Auth(c, acl.ResourceCluster, acl.ActionDownload)

		if s.Abort(c) {
			return
		}

		/*
			TODO - Consider the following optional hardening measures:
			  1. Track a hadError flag to log "partial success" if some files fail to zip.
			  2. Set limits (total size/entry count) in case theme directories grow unexpectedly.
			  3. Optionally, return a 404 or 204 error code when no files are added, though an empty zip file is acceptable.
		*/

		// Get app config.
		conf := get.Config()

		// Abort if this is not a portal server.
		if !conf.IsPortal() {
			AbortFeatureDisabled(c)
			return
		}

		clientIp := ClientIP(c)
		themePath := conf.PortalThemePath()

		// Resolve symbolic links.
		if resolved, err := filepath.EvalSymlinks(themePath); err != nil {
			event.AuditWarn([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "failed to resolve path"}, s.RefID, clean.Error(err))
			AbortNotFound(c)
			return
		} else {
			themePath = resolved
		}

		// Check if theme path exists.
		if !fs.PathExists(themePath) {
			event.AuditDebug([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "theme path not found"}, s.RefID)
			AbortNotFound(c)
			return
		} else {
			event.AuditDebug([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "creating theme archive from %s"}, s.RefID, clean.Log(themePath))
		}

		// Add response headers.
		AddDownloadHeader(c, "theme.zip")
		AddContentTypeHeader(c, header.ContentTypeZip)

		// Create zip writer to stream the theme files.
		zipWriter := zip.NewWriter(c.Writer)
		defer func(w *zip.Writer) {
			if closeErr := w.Close(); closeErr != nil {
				event.AuditWarn([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "failed to close", "%s"}, s.RefID, clean.Error(closeErr))
			}
		}(zipWriter)

		err := filepath.WalkDir(themePath, func(filePath string, info gofs.DirEntry, walkErr error) error {
			// Handle errors.
			if walkErr != nil {
				event.AuditWarn([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "failed to traverse theme path", "%s"}, s.RefID, clean.Error(walkErr))

				// If the error occurs on a directory, skip descending to avoid cascading errors.
				if info != nil && info.IsDir() {
					return gofs.SkipDir
				}

				return nil

			}

			// Get file base name.
			name := info.Name()

			// Skip any subdirectories to enhance security.
			if info.IsDir() {
				if filePath != themePath {
					return gofs.SkipDir
				}

				return nil
			}

			// Skip non-regular files and symlinks.
			if !info.Type().IsRegular() || info.Type()&gofs.ModeSymlink != 0 {
				return nil
			}

			// Skip hidden files by name.
			if fs.FileNameHidden(name) {
				return nil
			}

			// Get the relative file name to use as alias in the zip.
			alias := filepath.ToSlash(fs.RelName(filePath, themePath))

			event.AuditDebug([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "adding %s to archive"}, s.RefID, clean.Log(alias))

			// Stream zipped file contents.
			if zipErr := fs.ZipFile(zipWriter, filePath, alias, false); zipErr != nil {
				event.AuditWarn([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "failed to add %s", "%s"}, s.RefID, clean.Log(alias), clean.Error(zipErr))
			}

			return nil
		})

		// Log result.
		if err != nil {
			event.AuditErr([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", event.Failed, "%s"}, s.RefID, clean.Error(err))
		} else {
			event.AuditInfo([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", event.Succeeded}, s.RefID)
		}
	})
}
