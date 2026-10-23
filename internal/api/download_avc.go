package api

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// DownloadFormatAvc requests the AVC rendition of a video, live photo, or animated image.
const DownloadFormatAvc = "avc"

// avcCodecPattern matches the codec names the web client treats as AVC.
var avcCodecPattern = regexp.MustCompile(`(?i)^(avc[1-4]?|h\.?264|x264)([ /._-]|$)`)

// avcDownloadable reports whether an AVC rendition of f may be downloaded, which matches the files the
// web client offers one for: originals that are animated, live, or videos not already encoded in AVC.
func avcDownloadable(f *entity.File) bool {
	if f == nil || f.FileRoot != entity.RootOriginals || f.FileSidecar || f.FileMissing || f.FileError != "" || f.FileHash == "" {
		return false
	}

	mediaType := media.Type(f.MediaType)

	if mediaType == media.Unknown {
		if mediaType = media.Formats[fs.Type(strings.ToLower(f.FileType))]; mediaType == media.Unknown && f.FileVideo {
			mediaType = media.Video
		}
	}

	switch mediaType {
	case media.Animated, media.Live:
		return true
	case media.Image:
		return f.FileFrames > 1 || f.FileDuration > 0
	case media.Video:
		return insta360Capture(f) || !avcCodecPattern.MatchString(f.FileCodec)
	default:
		return false
	}
}

// avcDownloadName returns the attachment name of an AVC rendition of f.
func avcDownloadName(f *entity.File, n customize.DownloadName) string {
	return clean.AttachmentName(fs.StripExt(f.DownloadName(n, 0)) + fs.ExtMp4)
}

// sendAvcDownload sends the AVC rendition of an original as an attachment. Every refusal returns the same
// 404 as an unknown file.
func sendAvcDownload(c *gin.Context, f *entity.File, sess *entity.Session) {
	// The file row must belong to a picture the session may download.
	if visible, err := search.PhotoDownloadable(f.PhotoUID, sess); err != nil || !visible || !avcDownloadable(f) {
		c.Data(http.StatusNotFound, "image/svg+xml", brokenIconSvg)
		return
	}

	conf := get.Config()
	src := newVideoSource(f)

	if f.MediaType == entity.MediaLive {
		info, reader, err := openEmbeddedVideo(f, src.FileName)

		if err != nil {
			c.Data(http.StatusNotFound, "image/svg+xml", brokenIconSvg)
			return
		}

		defer reader.Close()

		if src, err = cacheEmbeddedVideo(conf, f, info, reader); err != nil {
			c.Data(http.StatusNotFound, "image/svg+xml", brokenIconSvg)
			return
		}
	}

	fileName, contentType, err := resolveVideo(conf, f, src, video.Avc, true)

	if errors.Is(err, status.ErrInsufficientStorage) {
		Abort(c, http.StatusInsufficientStorage, i18n.ErrInsufficientStorage)
		return
	} else if err != nil {
		c.Data(http.StatusNotFound, "image/svg+xml", brokenIconSvg)
		return
	}

	AddContentTypeHeader(c, contentType)
	c.FileAttachment(fileName, avcDownloadName(f, DownloadName(c)))
}
