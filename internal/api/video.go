package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media/video"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// GetVideo returns a video, optionally limited to a byte range for streaming.
//
//	@Summary		returns a video, optionally limited to a byte range for streaming
//	@Description	For more information see:
//	@Description	- https://docs.photoprism.app/developer-guide/api/thumbnails/#video-endpoint-uri
//	@Id				GetVideo
//	@Produce		video/mp4
//	@Tags			Files, Videos
//	@Failure		403		{object}	i18n.Response
//	@Param			hash	path		string	true	"SHA1 video file hash"
//	@Param			token	path		string	true	"user-specific security token provided with session"
//	@Param			format	path		string	true	"video format, e.g. mp4"
//	@Router			/api/v1/videos/{hash}/{token}/{format} [get]
func GetVideo(router *gin.RouterGroup) {
	router.GET("/videos/:hash/:token/:format", func(c *gin.Context) {
		fileHash := clean.Token(c.Param("hash"))

		// Check if a valid security token was provided.
		if InvalidPreviewToken(c, acl.Resources{acl.ResourcePhotos, acl.ResourceFiles}) {
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

		// If file is not a video, try to find the related video file.
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
			// Rendered through the sanitizer, so the guard holds for a value stored by any writer,
			// including rows written before the writers sanitized.
			log.Errorf("video: file has error %s", clean.Log(f.FileError))
			AbortVideo(c)
			return
		} else if f.FileHash == "" {
			log.Errorf("video: file hash missing in index")
			AbortVideo(c)
			return
		}

		// Get app config.
		conf := get.Config()
		src := newVideoSource(f)

		// If the file has a hybrid photo/video format, try to find and send the embedded video data.
		if f.MediaType == entity.MediaLive {
			info, reader, embedErr := openEmbeddedVideo(f, src.FileName)

			if embedErr != nil {
				AbortVideo(c)
				return
			}

			defer reader.Close()

			if c.Request.Header.Get("Range") == "" && info.VideoCodec == format.Codec {
				AddVideoCacheHeader(c, conf.CdnVideo())
				c.DataFromReader(http.StatusOK, info.VideoSize(), info.VideoContentType(), reader, nil)
				return
			}

			// Serve embedded videos from cache to allow streaming and transcoding.
			if src, embedErr = cacheEmbeddedVideo(conf, f, info, reader); embedErr != nil {
				AbortVideo(c)
				return
			}
		}

		videoFileName, contentType, videoErr := resolveVideo(conf, f, src, format, false)

		if errors.Is(videoErr, status.ErrInsufficientStorage) {
			Abort(c, http.StatusInsufficientStorage, i18n.ErrInsufficientStorage)
			return
		} else if videoErr != nil {
			AbortVideo(c)
			return
		}

		AddContentTypeHeader(c, contentType)

		// Add HTTP cache header.
		AddVideoCacheHeader(c, conf.CdnVideo())

		// Return requested content.
		if c.Query("download") != "" {
			c.FileAttachment(videoFileName, f.DownloadName(DownloadName(c), 0))
		} else {
			c.File(videoFileName)
		}
	})
}
