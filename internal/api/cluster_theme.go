package api

import (
	"archive/zip"
	gofs "io/fs"
	"net"
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
		// Get app config and client IP.
		conf := get.Config()
		clientIp := ClientIP(c)

		// Optional IP-based allowance via ClusterCIDR.
		refID := "-"
		if cidr := conf.ClusterCIDR(); cidr != "" {
			if _, ipnet, err := net.ParseCIDR(cidr); err == nil {
				if ip := net.ParseIP(clientIp); ip != nil && ipnet.Contains(ip) {
					// Allowed by CIDR; proceed without session.
					refID = "cidr"
				}
			}
		}

		// If not allowed by CIDR, require regular auth.
		if refID == "-" {
			s := Auth(c, acl.ResourceCluster, acl.ActionDownload)
			if s.Abort(c) {
				return
			}
			refID = s.RefID
		}

		/*
			TODO - Consider the following optional hardening measures:
			  1. Track a hadError flag to log "partial success" if some files fail to zip.
			  2. Set limits (total size/entry count) in case theme directories grow unexpectedly.
			  3. Optionally, return a 404 or 204 error code when no files are added, though an empty zip file is acceptable.
		*/

		// Abort if this is not a portal server.
		if !conf.Portal() {
			AbortFeatureDisabled(c)
			return
		}
		themePath := conf.ThemePath()

		// Resolve symbolic links.
		if resolved, err := filepath.EvalSymlinks(themePath); err != nil {
			event.AuditWarn([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "failed to resolve path"}, refID, clean.Error(err))
			AbortNotFound(c)
			return
		} else {
			themePath = resolved
		}

		// Check if theme path exists.
		if !fs.PathExists(themePath) {
			event.AuditDebug([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "theme path not found"}, refID)
			AbortNotFound(c)
			return
		}

		// Require a non-empty app.js file to avoid distributing empty themes.
		// This aligns with bootstrap behavior, which only installs a theme when
		// app.js exists locally or can be fetched from the Portal.
		if !fs.FileExistsNotEmpty(filepath.Join(themePath, "app.js")) {
			event.AuditDebug([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "app.js missing or empty"}, refID)
			AbortNotFound(c)
			return
		}

		event.AuditDebug([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "creating theme archive from %s"}, refID, clean.Log(themePath))

		// Add response headers.
		AddDownloadHeader(c, "theme.zip")
		AddContentTypeHeader(c, header.ContentTypeZip)

		// Create zip writer to stream the theme files.
		zipWriter := zip.NewWriter(c.Writer)
		defer func(w *zip.Writer) {
			if closeErr := w.Close(); closeErr != nil {
				event.AuditWarn([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "failed to close", "%s"}, refID, clean.Error(closeErr))
			}
		}(zipWriter)

		err := filepath.WalkDir(themePath, func(filePath string, info gofs.DirEntry, walkErr error) error {
			// Handle errors.
			if walkErr != nil {
				event.AuditWarn([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "failed to traverse theme path", "%s"}, refID, clean.Error(walkErr))

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

			event.AuditDebug([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "adding %s to archive"}, refID, clean.Log(alias))

			// Stream zipped file contents.
			if zipErr := fs.ZipFile(zipWriter, filePath, alias, false); zipErr != nil {
				event.AuditWarn([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", "failed to add %s", "%s"}, refID, clean.Log(alias), clean.Error(zipErr))
			}

			return nil
		})

		// Log result.
		if err != nil {
			event.AuditErr([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", event.Failed, "%s"}, refID, clean.Error(err))
		} else {
			event.AuditInfo([]string{clientIp, "session %s", string(acl.ResourceCluster), "theme", "download", event.Succeeded}, refID)
		}
	})
}
