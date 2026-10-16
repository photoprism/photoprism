package api

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// videoContent retains the open file and bounded segment used for playback or download.
type videoContent struct {
	file        *os.File
	offset      int64
	size        int64
	extension   string
	contentType string
}

// videoResponseWriter retains the preparation cache policy on native HTTP errors.
type videoResponseWriter struct {
	http.ResponseWriter
}

// WriteHeader preserves no-store on range and conditional responses.
func (w videoResponseWriter) WriteHeader(code int) {
	w.Header().Set("Cache-Control", "no-store")
	w.ResponseWriter.WriteHeader(code)
}

// serve delivers the same open video segment for streaming and attachment requests.
func (v *videoContent) serve(c *gin.Context, f *entity.File, kind string) {
	stat, err := v.file.Stat()
	if err != nil {
		videoError(c, err, http.StatusNotFound)
		return
	}
	name := f.DownloadName(DownloadName(c), 0)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	exact := c.Query("action") != ""
	if exact {
		name = clean.FileName(name)
		if name == "" {
			name = f.FileHash
		}
		name += "_" + kind
	} else {
		AddVideoCacheHeader(c, get.Config().CdnVideo())
	}
	name = DownloadFileName(c, name+v.extension, v.extension)
	c.Header("Content-Type", v.contentType)
	if c.Query("download") != "" || c.Query("action") == "download" {
		c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	}
	var writer http.ResponseWriter = c.Writer
	if exact {
		writer = videoResponseWriter{writer}
	}
	http.ServeContent(writer, c.Request, name, stat.ModTime(), io.NewSectionReader(v.file, v.offset, v.size))
}

// videoPlayback selects a compatible playback file while preserving indexed animation metadata.
func videoPlayback(f *entity.File, format video.Type, ranged bool) (*videoContent, error) {
	conf := get.Config()
	name := photoprism.FileName(f.FileRoot, f.FileName)
	contentType, bitrate := f.ContentType(), f.Bitrate()
	extension := filepath.Ext(name)
	var offset, size int64
	if f.MediaType == entity.MediaLive {
		info, err := video.ProbeFile(name)
		if err != nil || !info.Compatible || info.VideoOffset < 0 {
			return nil, photoprism.ErrExportUnavailable
		}
		contentType, bitrate, extension = info.VideoContentType(), info.VideoBitrate(), info.VideoFileExt()
		if info.VideoCodec == format.Codec && (!ranged || !conf.FFmpegEnabled() || !conf.FFmpegBitrateExceeded(bitrate)) {
			offset, size = info.VideoOffset, info.VideoSize()
		} else {
			reader, readErr := video.NewReader(name, info.VideoOffset)
			if readErr != nil {
				return nil, readErr
			}
			defer reader.Close()
			name, err = fs.CacheFileFromReader(filepath.Join(conf.MediaFileCachePath(f.FileHash), f.FileHash+extension), reader)
			if err != nil {
				return nil, err
			}
		}
	}
	if size == 0 {
		mediaFile, err := photoprism.NewMediaFile(name)
		if err != nil {
			logErr("video", f.Update("FileMissing", true))
			return nil, err
		}
		if f.MediaType == entity.MediaAnimated || f.IsAnimated() {
			mediaFile.SetAnimationInfo(f.FileFrames, f.FileDuration)
		}
		playable, ready := photoprism.Insta360PlaybackFile(mediaFile)
		if !ready {
			return nil, photoprism.ErrExportNotReady
		}
		if playable != mediaFile {
			mediaFile, name = playable, playable.FileName()
			contentType, bitrate = playable.ContentType(), playable.VideoInfo().VideoBitrate()
			extension = fs.ExtMp4
		}
		if !video.Compatible(contentType, format.ContentType) || conf.FFmpegEnabled() && conf.FFmpegBitrateExceeded(bitrate) {
			converted, convertErr := get.Convert().ToAvc(mediaFile, conf.FFmpegEncoder(), false, false)
			if convertErr != nil {
				return nil, convertErr
			}
			if converted == nil {
				return nil, photoprism.ErrExportUnavailable
			}
			name, contentType, extension = converted.FileName(), header.ContentTypeMp4AvcMain, fs.ExtMp4
		}
	}
	file, err := os.Open(name) //nolint:gosec // name derives from an indexed source or standard conversion output
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil || !stat.Mode().IsRegular() {
		_ = file.Close()
		return nil, photoprism.ErrExportUnavailable
	}
	if size == 0 {
		size = stat.Size()
	}
	return &videoContent{file: file, offset: offset, size: size, extension: extension, contentType: contentType}, nil
}
