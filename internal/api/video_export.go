package api

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/proc"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// exportResponseWriter keeps native ServeContent errors out of shared caches.
type exportResponseWriter struct {
	http.ResponseWriter
}

// WriteHeader preserves the export cache policy on range and conditional responses.
func (w exportResponseWriter) WriteHeader(code int) {
	w.Header().Set("Cache-Control", "no-store")
	w.ResponseWriter.WriteHeader(code)
}

// GetVideoExport probes, prepares, or downloads a rendition of the selected file.
//
// @Summary probes, prepares, or downloads a source-specific media export
// @Id GetVideoExport
// @Tags Files, Videos
// @Produce json, video/mp4, video/quicktime
// @Param hash path string true "SHA1 source file hash"
// @Param token path string true "session-bound download token"
// @Param export query string true "export operation" Enums(probe, prepare, download)
// @Param kind query string false "rendition for prepare and download" Enums(avc, video-part)
// @Param filename query string false "sanitized attachment filename with the real output extension"
// @Success 200 {object} photoprism.ExportProbe
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 429 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Failure 504 {object} map[string]string
// @Failure 507 {object} map[string]string
// @Router /api/v1/videos/{hash}/{token}/export [get]
func GetVideoExport(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	// The path credential is authoritative; alternate query credentials cannot replace it.
	c.Set(downloadPathTokenKey, c.Param("token"))
	delete(c.Keys, downloadSessionKey)
	delete(c.Keys, downloadHeaderAuthKey)
	f, ok := videoExportSource(c)
	if !ok {
		return
	}
	mode, kind := c.Query("export"), c.Query("kind")
	if mode != "probe" && mode != "prepare" && mode != "download" || mode != "probe" && kind != "avc" && kind != "video-part" {
		videoExportError(c, http.StatusBadRequest)
		return
	}
	converter := get.Convert()
	switch mode {
	case "probe":
		probe, err := converter.ProbeExport(f)
		if err != nil {
			if errors.Is(err, photoprism.ErrExportBusy) {
				c.Header("Retry-After", "1")
				limiter.AbortJSON(c)
				return
			}
			videoExportError(c, http.StatusNotFound)
			return
		}
		if sess := DownloadSession(c); sess == nil || !sess.SeesFullDetail(acl.ResourceFiles) && !sess.SeesFullDetail(acl.ResourcePhotos) {
			probe.AppleKey = ""
		}
		c.JSON(http.StatusOK, probe)
	case "prepare":
		output, err := converter.PrepareExport(f, kind)
		if err != nil {
			code := http.StatusInternalServerError
			switch {
			case errors.Is(err, photoprism.ErrExportBusy):
				c.Header("Retry-After", "1")
				limiter.AbortJSON(c)
				return
			case errors.Is(err, status.ErrInsufficientStorage):
				code = http.StatusInsufficientStorage
			case errors.Is(err, proc.ErrTimeout):
				code = http.StatusGatewayTimeout
			case errors.Is(err, photoprism.ErrExportUnavailable):
				code = http.StatusUnprocessableEntity
			case errors.Is(err, photoprism.ErrExportSourceChanged):
				code = http.StatusNotFound
			}
			videoExportError(c, code)
			return
		}
		// Refresh admission after a long conversion before exposing the completed output.
		delete(c.Keys, downloadSessionKey)
		delete(c.Keys, downloadHeaderAuthKey)
		if _, ok = videoExportSource(c); !ok {
			return
		}
		c.JSON(http.StatusOK, output)
	case "download":
		file, offset, output, err := converter.ExportFile(f, kind)
		if err != nil {
			if errors.Is(err, photoprism.ErrExportBusy) {
				c.Header("Retry-After", "1")
				limiter.AbortJSON(c)
				return
			}
			videoExportError(c, http.StatusNotFound)
			return
		}
		defer file.Close()
		stat, err := file.Stat()
		if err != nil {
			videoExportError(c, http.StatusNotFound)
			return
		}
		base := strings.TrimSuffix(f.DownloadName(DownloadName(c), 0), filepath.Ext(f.FileName))
		base = clean.FileName(base)
		if base == "" {
			base = f.FileHash
		}
		fallback := base + "_" + kind + "." + output.Format
		c.Header("Content-Type", "video/"+map[string]string{"mp4": "mp4", "mov": "quicktime"}[output.Format])
		name := DownloadFileName(c, fallback, "."+output.Format)
		c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
		http.ServeContent(exportResponseWriter{c.Writer}, c.Request, name, stat.ModTime(), io.NewSectionReader(file, offset, output.Bytes))
	}
}

// videoExportSource authorizes and scopes the source before resolving rows or cache paths.
func videoExportSource(c *gin.Context) (*entity.File, bool) {
	sess, valid := AuthDownload(c, acl.Resources{acl.ResourceFiles, acl.ResourcePhotos})
	if !valid {
		if DownloadRateLimited(c) {
			limiter.AbortJSON(c)
		} else {
			videoExportError(c, http.StatusForbidden)
		}
		return nil, false
	}
	hash := c.Param("hash")
	if !rnd.IsSHA(hash) {
		videoExportError(c, http.StatusNotFound)
		return nil, false
	}
	if visible, err := search.FileDownloadable(hash, sess); err != nil || !visible {
		videoExportError(c, http.StatusNotFound)
		return nil, false
	}
	f, err := query.FileByHash(hash)
	if err != nil || !f.Exportable(sess) || f.FileMissing || f.FileError != "" {
		videoExportError(c, http.StatusNotFound)
		return nil, false
	}
	return f, true
}

// videoExportError returns uniform, path-free errors for media export requests.
func videoExportError(c *gin.Context, code int) {
	c.AbortWithStatusJSON(code, gin.H{"error": http.StatusText(code)})
}
