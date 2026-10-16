package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media/video"
	"github.com/photoprism/photoprism/pkg/proc"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// GetVideo returns video content or source-specific preparation metadata.
//
//	@Summary		returns video content or source-specific preparation metadata
//	@Description	Returns video bytes for playback and download, ExportProbe JSON for action=probe, or ExportOutput JSON for action=prepare.
//	@Id				GetVideo
//	@Produce		json, video/mp4, video/quicktime
//	@Tags			Files, Videos
//	@Param			hash							path	string	true	"SHA1 source file hash"
//	@Param			token							path	string	true	"preview token for playback; download token for source-specific actions"
//	@Param			format							path	string	true	"playback format; actions accept avc or native (embedded video)"
//	@Param			action							query	string	false	"source-specific operation without playback fallback"	Enums(probe, prepare, download)
//	@Param			download						query	string	false	"serve playback as an attachment when nonempty"
//	@Param			name							query	string	false	"download naming convention"	Enums(file, share, original)
//	@Param			filename						query	string	false	"sanitized attachment filename with the real output extension"
//	@Success		200								"Video content or action-specific JSON metadata"
//	@Failure		400,403,404,422,429,500,504,507	{object}	map[string]string
//	@Router			/api/v1/videos/{hash}/{token}/{format} [get]
func GetVideo(router *gin.RouterGroup) {
	router.GET("/videos/:hash/:token/:format", func(c *gin.Context) {
		action := c.Query("action")
		if action != "" {
			c.Header("Cache-Control", "no-store")
		}
		f, ok := videoSource(c)
		if !ok {
			return
		}
		formatName := clean.Token(c.Param("format"))
		format := video.Types[formatName]
		kind := "avc"
		if formatName == "native" {
			kind = "video-part"
		}
		converter := get.Convert()
		switch action {
		case "probe":
			probe, err := converter.ProbeExport(f)
			if err != nil {
				videoError(c, err, http.StatusNotFound)
				return
			}
			if sess := DownloadSession(c); sess == nil || !sess.SeesFullDetail(acl.ResourceFiles) && !sess.SeesFullDetail(acl.ResourcePhotos) {
				probe.AppleKey = ""
			}
			c.JSON(http.StatusOK, probe)
			return
		case "prepare":
			output, err := converter.PrepareExport(f, kind)
			if err != nil {
				videoError(c, err, http.StatusInternalServerError)
				return
			}
			// Reauthorize after a long conversion before exposing the completed output.
			if _, ok = videoSource(c); ok {
				c.JSON(http.StatusOK, output)
			}
			return
		}
		var content *videoContent
		var err error
		if action == "download" {
			file, offset, output, openErr := converter.ExportFile(f, kind)
			err = openErr
			if err == nil {
				content = &videoContent{file: file, offset: offset, size: output.Bytes, extension: "." + output.Format, contentType: "video/" + map[string]string{"mp4": "mp4", "mov": "quicktime"}[output.Format]}
			}
		} else {
			content, err = videoPlayback(f, format, c.GetHeader("Range") != "")
		}
		if err != nil {
			videoError(c, err, http.StatusNotFound)
			return
		}
		defer content.file.Close()
		content.serve(c, f, kind)
	})
}

// videoSource authorizes and resolves playback or an exact, download-scoped source.
func videoSource(c *gin.Context) (*entity.File, bool) {
	exact := c.Query("action") != ""
	var sess *entity.Session
	if exact {
		c.Set(downloadPathTokenKey, c.Param("token"))
		delete(c.Keys, downloadSessionKey)
		delete(c.Keys, downloadHeaderAuthKey)
		var valid bool
		sess, valid = AuthDownload(c, acl.Resources{acl.ResourceFiles, acl.ResourcePhotos})
		if !valid {
			if DownloadRateLimited(c) {
				limiter.AbortJSON(c)
			} else {
				videoError(c, nil, http.StatusForbidden)
			}
			return nil, false
		}
	} else if InvalidPreviewToken(c) {
		videoError(c, nil, http.StatusForbidden)
		return nil, false
	}
	action, format := c.Query("action"), clean.Token(c.Param("format"))
	_, supported := video.Types[format]
	if !exact && !supported || exact && (format != "avc" && format != "native" || action != "probe" && action != "prepare" && action != "download") {
		videoError(c, nil, http.StatusBadRequest)
		return nil, false
	}
	hash := c.Param("hash")
	if !exact {
		hash = clean.Token(hash)
	}
	if !rnd.IsSHA(hash) {
		videoError(c, nil, http.StatusNotFound)
		return nil, false
	}
	if exact {
		if visible, err := search.FileDownloadable(hash, sess); err != nil || !visible {
			videoError(c, nil, http.StatusNotFound)
			return nil, false
		}
	}
	f, err := query.FileByHash(hash)
	if err == nil && !exact && !f.FileVideo && f.MediaType != entity.MediaLive && (!fs.IsAnimatedImage(f.FileName) || f.MediaType != entity.MediaAnimated && !f.IsAnimated()) {
		f, err = query.VideoByPhotoUID(f.PhotoUID)
	}
	if err != nil || f.FileError != "" || f.FileHash == "" || exact && (!f.Exportable(sess) || f.FileMissing) {
		videoError(c, nil, http.StatusNotFound)
		return nil, false
	}
	return f, true
}

// videoError preserves playback fallbacks and returns path-free errors for preparation actions.
func videoError(c *gin.Context, err error, code int) {
	if err != nil {
		log.Debugf("video: %s", clean.Error(err))
	}
	if errors.Is(err, photoprism.ErrExportBusy) {
		c.Header("Retry-After", "1")
		limiter.AbortJSON(c)
		return
	}
	switch {
	case errors.Is(err, status.ErrInsufficientStorage):
		code = http.StatusInsufficientStorage
	case errors.Is(err, proc.ErrTimeout):
		code = http.StatusGatewayTimeout
	case c.Query("action") == "prepare" && errors.Is(err, photoprism.ErrExportUnavailable):
		code = http.StatusUnprocessableEntity
	case errors.Is(err, photoprism.ErrExportSourceChanged):
		code = http.StatusNotFound
	}
	switch {
	case c.Query("action") != "":
		c.AbortWithStatusJSON(code, gin.H{"error": http.StatusText(code)})
	case code == http.StatusForbidden || code == http.StatusBadRequest:
		c.Data(code, "image/svg+xml", videoIconSvg)
	case code == http.StatusInsufficientStorage:
		Abort(c, code, i18n.ErrInsufficientStorage)
	default:
		AbortVideo(c)
	}
}
