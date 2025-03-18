package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media/http/header"
	"github.com/photoprism/photoprism/pkg/media/video"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// GetVideo returns a video, optionally limited to a byte range for streaming.
//
//	@Summary		returns a video, optionally limited to a byte range for streaming
//	@Description	Fore more information see:
//	@Description	- https://docs.photoprism.app/developer-guide/api/thumbnails/#video-endpoint-uri
//	@Id				GetVideo
//	@Produce		video/mp4
//	@Tags			Files, Videos
//	@Failure		403		{object}	i18n.Response
//	@Param			thumb	path		string	true	"SHA1 video file hash"
//	@Param			token	path		string	true	"user-specific security token provided with session"
//	@Param			format	path		string	true	"video format, e.g. mp4"
//	@Router			/api/v1/videos/{hash}/{token}/{format} [get]
func GetVideo(router *gin.RouterGroup) {
	router.GET("/videos/:hash/:token/:format", func(c *gin.Context) {
		fileHash := clean.Token(c.Param("hash"))

		// Check if a valid security token was provided.
		if InvalidPreviewToken(c) {
			c.Data(http.StatusForbidden, "image/svg+xml", videoIconSvg)
			return
		}

		// Check if a valid file hash was provided.
		if !rnd.IsSHA(fileHash) {
			log.Debugf("video: invalid file hash %s", clean.Log(fileHash))
			AbortVideo(c)
			return
		}

		// Check if a supported video format was provided.
		formatName := clean.Token(c.Param("format"))
		format, ok := video.Types[formatName]

		if !ok {
			log.Errorf("video: invalid format %s", clean.Log(formatName))
			c.Data(http.StatusBadRequest, "image/svg+xml", videoIconSvg)
			return
		}

		// Find media file by SHA hash.
		f, err := query.FileByHash(fileHash)

		if err != nil {
			log.Errorf("video: requested file not found (%s)", err)
			AbortVideo(c)
			return
		}

		// If file is not a video, try to find the realted video file.
		if !f.FileVideo {
			f, err = query.VideoByPhotoUID(f.PhotoUID)

			if err != nil {
				log.Errorf("video: no playable file found (%s)", err)
				AbortVideo(c)
				return
			}
		}

		// Return a broken video if the file could not be found.
		if f.FileError != "" {
			log.Errorf("video: file has error %s", f.FileError)
			AbortVideo(c)
			return
		} else if f.FileHash == "" {
			log.Errorf("video: file hash missing in index")
			AbortVideo(c)
			return
		}

		// Get app config.
		conf := get.Config()

		// Get video bitrate, codec, and file type.
		videoFileType := f.Type()
		videoCodec := f.FileCodec
		videoBitrate := f.Bitrate()
		videoFileName := photoprism.FileName(f.FileRoot, f.FileName)
		videoContentType := f.ContentType()

		// If the file has a hybrid photo/video format, try to find and send the embedded video data.
		if f.MediaType == entity.MediaLive {
			if info, videoErr := video.ProbeFile(videoFileName); info.VideoOffset < 0 || !info.Compatible || videoErr != nil {
				logErr("video", videoErr)
				log.Warnf("video: no embedded media found in %s", clean.Log(f.FileName))
				AbortVideo(c)
				return
			} else if reader, readErr := video.NewReader(videoFileName, info.VideoOffset); readErr != nil {
				log.Errorf("video: failed to read media embedded in %s (%s)", clean.Log(f.FileName), readErr)
				AbortVideo(c)
				return
			} else if c.Request.Header.Get("Range") == "" && info.VideoCodec == format.Codec {
				defer reader.Close()
				AddVideoCacheHeader(c, conf.CdnVideo())
				c.DataFromReader(http.StatusOK, info.VideoSize(), info.VideoContentType(), reader, nil)
				return
			} else if cacheName, cacheErr := fs.CacheFileFromReader(filepath.Join(conf.MediaFileCachePath(f.FileHash), f.FileHash+info.VideoFileExt()), reader); cacheErr != nil {
				log.Errorf("video: failed to cache %s embedded in %s (%s)", videoFileType.ToUpper(), clean.Log(f.FileName), cacheErr)
				AbortVideo(c)
				return
			} else {
				// Serve embedded videos from cache to allow streaming and transcoding.
				videoBitrate = info.VideoBitrate()
				videoCodec = info.VideoCodec
				videoFileType = info.VideoFileType()
				videoFileName = cacheName
				videoContentType = info.VideoContentType()
				log.Debugf("video: streaming %s encoded %s in %s from cache", strings.ToUpper(videoCodec), videoFileType.ToUpper(), clean.Log(f.FileName))
			}
		}

		// Verify video format support and compatibility.
		supported := video.Compatible(videoContentType, format.ContentType)

		// Check video bitrate against the configured limit.
		transcode := !supported || conf.FFmpegEnabled() && conf.FFmpegBitrateExceeded(videoBitrate)

		if mediaFile, mediaErr := photoprism.NewMediaFile(videoFileName); mediaErr != nil {
			// Set missing flag so that the file doesn't show up in search results anymore.
			logErr("video", f.Update("FileMissing", true))

			// Log error and default to 404.mp4
			log.Errorf("video: file %s is missing", clean.Log(f.FileName))
			AbortVideo(c)
			return
		} else if transcode {
			log.Debugf(
				"video: %s has content type %s and cannot be streamed directly, average bitrate %.1f MBit/s",
				clean.Log(f.FileName),
				clean.Log(videoContentType),
				videoBitrate,
			)

			conv := get.Convert()

			if avcFile, avcErr := conv.ToAvc(mediaFile, get.Config().FFmpegEncoder(), false, false); avcFile != nil && avcErr == nil {
				videoFileName = avcFile.FileName()
				AddContentTypeHeader(c, header.ContentTypeMp4AvcMain)
			} else {
				// Log error and default to 404.mp4
				log.Errorf("video: failed to transcode %s", clean.Log(f.FileName))
				AbortVideo(c)
				return
			}
		} else {
			var contentType string

			if videoFileType != f.Type() {
				contentType = mediaFile.ContentType()
			} else {
				contentType = f.ContentType()
			}

			log.Debugf(
				"video: %s has content type %s, average bitrate %.1f MBit/s",
				clean.Log(f.FileName),
				clean.Log(contentType),
				videoBitrate,
			)

			AddContentTypeHeader(c, contentType)
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
