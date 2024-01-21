package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/video"
)

// GetVideo streams video content.
//
// The request parameters are:
//
//   - hash: string The photo or video file hash as returned by the search API
//   - type: string Video format
//
// GET /api/v1/videos/:hash/:token/:type
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
		} else if f.FileHash == "" {
			log.Errorf("video: file hash missing in index")
			c.Data(http.StatusOK, "image/svg+xml", videoIconSvg)
			return
		}

		// Get app config.
		conf := get.Config()

		// Get video bitrate, codec, and file type.
		videoBitrate := f.Bitrate()
		videoCodec := f.FileCodec
		videoFileType := f.FileType
		videoFileName := photoprism.FileName(f.FileRoot, f.FileName)

		// If the file has a hybrid photo/video format, try to find and send the embedded video data.
		if f.MediaType == entity.MediaLive {
			if info, videoErr := video.ProbeFile(videoFileName); info.VideoOffset < 0 || !info.Compatible || videoErr != nil {
				logErr("video", videoErr)
				log.Warnf("video: no embedded media found in %s", clean.Log(f.FileName))
				AddContentTypeHeader(c, video.ContentTypeAVC)
				c.File(get.Config().StaticFile("video/404.mp4"))
				return
			} else if reader, readErr := video.NewReader(videoFileName, info.VideoOffset); readErr != nil {
				log.Errorf("video: failed to read media embedded in %s (%s)", clean.Log(f.FileName), readErr)
				AddContentTypeHeader(c, video.ContentTypeAVC)
				c.File(get.Config().StaticFile("video/404.mp4"))
				return
			} else if c.Request.Header.Get("Range") == "" && info.VideoCodec == format.Codec {
				defer reader.Close()
				AddVideoCacheHeader(c, conf.CdnVideo())
				c.DataFromReader(http.StatusOK, info.VideoSize(), info.VideoContentType(), reader, nil)
				return
			} else if cacheName, cacheErr := fs.CacheFileFromReader(filepath.Join(conf.MediaFileCachePath(f.FileHash), f.FileHash+info.VideoFileExt()), reader); cacheErr != nil {
				log.Errorf("video: failed to cache %s embedded in %s (%s)", strings.ToUpper(videoFileType), clean.Log(f.FileName), cacheErr)
				AddContentTypeHeader(c, video.ContentTypeAVC)
				c.File(get.Config().StaticFile("video/404.mp4"))
				return
			} else {
				// Serve embedded videos from cache to allow streaming and transcoding.
				videoBitrate = info.VideoBitrate()
				videoCodec = info.VideoCodec.String()
				videoFileType = info.VideoFileType().String()
				videoFileName = cacheName
				log.Debugf("video: streaming %s encoded %s in %s from cache", strings.ToUpper(videoCodec), strings.ToUpper(videoFileType), clean.Log(f.FileName))
			}
		}

		// Check video format support.
		supported := videoCodec != "" && videoCodec == format.Codec.String() || format.Codec == video.CodecUnknown && videoFileType == format.FileType.String()

		// Check video bitrate against the configured limit.
		transcode := !supported || conf.FFmpegEnabled() && conf.FFmpegBitrateExceeded(videoBitrate)

		if mediaFile, mediaErr := photoprism.NewMediaFile(videoFileName); mediaErr != nil {
			// Set missing flag so that the file doesn't show up in search results anymore.
			logErr("video", f.Update("FileMissing", true))

			// Log error and default to 404.mp4
			log.Errorf("video: file %s is missing", clean.Log(f.FileName))
			videoFileName = get.Config().StaticFile("video/404.mp4")
			AddContentTypeHeader(c, video.ContentTypeAVC)
		} else if transcode {
			if videoCodec != "" {
				log.Debugf("video: %s is %s encoded and cannot be streamed directly, average bitrate %.1f MBit/s", clean.Log(f.FileName), strings.ToUpper(videoCodec), videoBitrate)
			} else {
				log.Debugf("video: %s cannot be streamed directly, average bitrate %.1f MBit/s", clean.Log(f.FileName), videoBitrate)
			}

			conv := get.Convert()

			if avcFile, avcErr := conv.ToAvc(mediaFile, get.Config().FFmpegEncoder(), false, false); avcFile != nil && avcErr == nil {
				videoFileName = avcFile.FileName()
			} else {
				// Log error and default to 404.mp4
				log.Errorf("video: failed to transcode %s", clean.Log(f.FileName))
				videoFileName = get.Config().StaticFile("video/404.mp4")
			}

			AddContentTypeHeader(c, video.ContentTypeAVC)
		} else {
			if videoCodec != "" && videoCodec != videoFileType {
				log.Debugf("video: %s is %s encoded and requires no transcoding, average bitrate %.1f MBit/s", clean.Log(f.FileName), strings.ToUpper(videoCodec), videoBitrate)
				AddContentTypeHeader(c, fmt.Sprintf("%s; codecs=\"%s\"", f.FileMime, clean.Codec(videoCodec)))
			} else {
				log.Debugf("video: %s is streamed directly, average bitrate %.1f MBit/s", clean.Log(f.FileName), videoBitrate)
				AddContentTypeHeader(c, f.FileMime)
			}
		}

		// Add HTTP cache header.
		AddVideoCacheHeader(c, conf.CdnVideo())

		// Return requested content.
		if c.Query("download") != "" {
			c.FileAttachment(videoFileName, f.DownloadName(DownloadName(c), 0))
		} else {
			c.File(videoFileName)
		}

		return
	})
}
