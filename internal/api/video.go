package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/video"
)

// GetVideo streams video content.
//
// GET /api/v1/videos/:hash/:token/:type
//
// Parameters:
//
//	hash: string The photo or video file hash as returned by the search API
//	type: string Video format
func GetVideo(router *gin.RouterGroup) {
	router.GET("/videos/:hash/:token/:format", func(c *gin.Context) {
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", brokenIconSvg)
			return
		}

		fileHash := clean.Token(c.Param("hash"))
		formatName := clean.Token(c.Param("format"))

		format, ok := video.Types[formatName]

		if !ok {
			log.Errorf("video: invalid format %s", clean.Log(formatName))
			c.Data(http.StatusOK, "image/svg+xml", videoIconSvg)
			return
		}

		f, err := query.FileByHash(fileHash)

		if err != nil {
			log.Errorf("video: requested file not found (%s)", err)
			c.Data(http.StatusOK, "image/svg+xml", videoIconSvg)
			return
		}

		if !f.FileVideo {
			f, err = query.VideoByPhotoUID(f.PhotoUID)

			if err != nil {
				log.Errorf("video: no playable file found (%s)", err)
				c.Data(http.StatusOK, "image/svg+xml", videoIconSvg)
				return
			}
		}

		if f.FileError != "" {
			log.Errorf("video: file has error %s", f.FileError)
			c.Data(http.StatusOK, "image/svg+xml", videoIconSvg)
			return
		}

		conf := get.Config()
		fileName := photoprism.FileName(f.FileRoot, f.FileName)

		// If file is not a video, try to find and stream embedded video data.
		if f.MediaType != entity.MediaVideo {
			if info, videoErr := video.ProbeFile(fileName); info.VideoOffset < 0 || !info.Compatible || videoErr != nil {
				logError("video", videoErr)
				log.Warnf("video: no data found in %s", clean.Log(f.FileName))
				AddContentTypeHeader(c, video.ContentTypeAVC)
				c.File(get.Config().StaticFile("video/404.mp4"))
			} else if reader, readErr := video.NewReader(fileName, info.VideoOffset); readErr != nil {
				log.Errorf("video: failed to read data from %s (%s)", clean.Log(f.FileName), readErr)
				AddContentTypeHeader(c, video.ContentTypeAVC)
				c.File(get.Config().StaticFile("video/404.mp4"))
			} else {
				defer reader.Close()
				AddVideoCacheHeader(c, conf.CdnVideo())
				c.DataFromReader(http.StatusOK, info.VideoSize(), info.VideoContentType(), reader, nil)
			}

			return
		}

		fileBitrate := f.Bitrate()

		// File format supported by the client/browser?
		supported := f.FileCodec != "" && f.FileCodec == string(format.Codec) || format.Codec == video.CodecUnknown && f.FileType == string(format.FileType)

		// File bitrate too high (for streaming)?
		transcode := !supported || conf.FFmpegEnabled() && conf.FFmpegBitrateExceeded(fileBitrate)

		if mf, err := photoprism.NewMediaFile(fileName); err != nil {
			// Set missing flag so that the file doesn't show up in search results anymore.
			logError("video", f.Update("FileMissing", true))

			// Log error and default to 404.mp4
			log.Errorf("video: file %s is missing", clean.Log(f.FileName))
			fileName = get.Config().StaticFile("video/404.mp4")
			AddContentTypeHeader(c, video.ContentTypeAVC)
		} else if transcode {
			if f.FileCodec != "" {
				log.Debugf("video: %s is %s compressed and cannot be streamed directly, average bitrate %.1f MBit/s", clean.Log(f.FileName), clean.Log(strings.ToUpper(f.FileCodec)), fileBitrate)
			} else {
				log.Debugf("video: %s cannot be streamed directly, average bitrate %.1f MBit/s", clean.Log(f.FileName), fileBitrate)
			}

			conv := get.Convert()

			if avcFile, avcErr := conv.ToAvc(mf, get.Config().FFmpegEncoder(), false, false); avcFile != nil && avcErr == nil {
				fileName = avcFile.FileName()
			} else {
				// Log error and default to 404.mp4
				log.Errorf("video: failed to transcode %s", clean.Log(f.FileName))
				fileName = get.Config().StaticFile("video/404.mp4")
			}

			AddContentTypeHeader(c, video.ContentTypeAVC)
		} else {
			if f.FileCodec != "" && f.FileCodec != f.FileType {
				log.Debugf("video: %s is %s compressed and requires no transcoding, average bitrate %.1f MBit/s", clean.Log(f.FileName), clean.Log(strings.ToUpper(f.FileCodec)), fileBitrate)
				AddContentTypeHeader(c, fmt.Sprintf("%s; codecs=\"%s\"", f.FileMime, clean.Codec(f.FileCodec)))
			} else {
				log.Debugf("video: %s is streamed directly, average bitrate %.1f MBit/s", clean.Log(f.FileName), fileBitrate)
				AddContentTypeHeader(c, f.FileMime)
			}
		}

		// Add HTTP cache header.
		AddVideoCacheHeader(c, conf.CdnVideo())

		// Return requested content.
		if c.Query("download") != "" {
			c.FileAttachment(fileName, f.DownloadName(DownloadName(c), 0))
		} else {
			c.File(fileName)
		}

		return
	})
}
