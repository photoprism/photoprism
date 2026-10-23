package api

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/media/video"
)

// errVideoUnavailable is returned when no playable file can be sent for a video request.
var errVideoUnavailable = errors.New("video unavailable")

// videoSource is the file a video response is read from.
type videoSource struct {
	FileName    string
	FileType    fs.Type
	ContentType string
	Bitrate     float64
}

// newVideoSource returns the indexed file f as a video source.
func newVideoSource(f *entity.File) videoSource {
	return videoSource{
		FileName:    photoprism.FileName(f.FileRoot, f.FileName),
		FileType:    f.Type(),
		ContentType: f.ContentType(),
		Bitrate:     f.Bitrate(),
	}
}

// openEmbeddedVideo probes the video embedded in a live photo and returns a reader for it.
func openEmbeddedVideo(f *entity.File, fileName string) (video.Info, *video.Reader, error) {
	info, err := video.ProbeFile(fileName)

	if info.VideoOffset < 0 || !info.Compatible || err != nil {
		logErr("video", err)
		log.Warnf("video: no embedded media found in %s", clean.Log(f.FileName))
		return info, nil, errVideoUnavailable
	}

	reader, err := video.NewReader(fileName, info.VideoOffset)

	if err != nil {
		log.Errorf("video: failed to read media embedded in %s (%s)", clean.Log(f.FileName), clean.Error(err))
		return info, nil, errVideoUnavailable
	}

	return info, reader, nil
}

// cacheEmbeddedVideo caches the embedded video read from reader and returns it as the source.
func cacheEmbeddedVideo(conf *config.Config, f *entity.File, info video.Info, reader *video.Reader) (videoSource, error) {
	cacheName, err := fs.CacheFileFromReader(filepath.Join(conf.MediaFileCachePath(f.FileHash), f.FileHash+info.VideoFileExt()), reader)

	if err != nil {
		log.Errorf("video: failed to cache %s embedded in %s (%s)", f.Type().ToUpper(), clean.Log(f.FileName), clean.Error(err))
		return videoSource{}, errVideoUnavailable
	}

	log.Debugf("video: streaming %s encoded %s in %s from cache", strings.ToUpper(info.VideoCodec), info.VideoFileType().ToUpper(), clean.Log(f.FileName))

	return videoSource{
		FileName:    cacheName,
		FileType:    info.VideoFileType(),
		ContentType: info.VideoContentType(),
		Bitrate:     info.VideoBitrate(),
	}, nil
}

// resolveVideo returns the name and content type of the file to send for src in the requested format,
// transcoding it to AVC if needed. A download is never transcoded for its bitrate alone, and Insta360
// footage is only downloaded from an existing AVC.
func resolveVideo(conf *config.Config, f *entity.File, src videoSource, format video.Type, download bool) (fileName, contentType string, err error) {
	mediaFile, mediaErr := photoprism.NewMediaFile(src.FileName)

	if mediaErr != nil {
		// Set missing flag so that the file doesn't show up in search results anymore.
		logErr("video", f.Update("FileMissing", true))
		log.Errorf("video: file %s is missing", clean.Log(f.FileName))
		return "", "", errVideoUnavailable
	}

	// Footage that can be dewarped is never converted inline in an HTTP request: index/import workers
	// create its AVC, and an older LRV derivative is a safe fallback. Other footage plays as it is.
	playable, ready := photoprism.Insta360PlaybackFile(mediaFile)

	if !ready {
		log.Warnf("video: equirectangular derivative for %s is not ready", clean.Log(f.FileName))
		return "", "", errVideoUnavailable
	} else if download && insta360Capture(f) && playable == mediaFile {
		if avcName := get.Convert().FindAvc(mediaFile); avcName == "" {
			log.Debugf("video: %s has no AVC derivative", clean.Log(f.FileName))
			return "", "", errVideoUnavailable
		} else if playable, mediaErr = photoprism.NewMediaFile(avcName); mediaErr != nil {
			log.Errorf("video: failed to open AVC derivative of %s (%s)", clean.Log(f.FileName), clean.Error(mediaErr))
			return "", "", errVideoUnavailable
		}
	}

	if playable != mediaFile {
		mediaFile = playable
		src.FileName = playable.FileName()
		src.FileType = playable.FileType()
		src.ContentType = playable.ContentType()
		src.Bitrate = playable.VideoInfo().VideoBitrate()
	}

	// Verify video format support and compatibility, and check the bitrate against the configured limit.
	supported := video.Compatible(src.ContentType, format.ContentType)
	bitrateExceeded := !download && conf.FFmpegEnabled() && conf.FFmpegBitrateExceeded(src.Bitrate)

	switch {
	case supported && !bitrateExceeded:
		if src.FileType != f.Type() {
			contentType = mediaFile.ContentType()
		} else {
			contentType = f.ContentType()
		}

		log.Debugf("video: %s has content type %s (average bitrate %.1f Mbps)", clean.Log(f.FileName), clean.Log(contentType), src.Bitrate)

		return src.FileName, contentType, nil
	case download && insta360Capture(f):
		log.Warnf("video: AVC derivative of %s has content type %s", clean.Log(f.FileName), clean.Log(src.ContentType))
		return "", "", errVideoUnavailable
	case supported:
		log.Debugf("video: %s has an average bitrate of %.1f Mbps, which exceeds the %d Mbps limit", clean.Log(f.FileName), src.Bitrate, conf.FFmpegBitrate())
	default:
		log.Debugf("video: %s has content type %s and cannot be streamed directly (average bitrate %.1f Mbps)", clean.Log(f.FileName), clean.Log(src.ContentType), src.Bitrate)
	}

	avcFile, avcErr := get.Convert().ToAvc(mediaFile, conf.FFmpegEncoder(), false, false)

	switch {
	case avcFile != nil && avcErr == nil:
		return avcFile.FileName(), header.ContentTypeMp4AvcMain, nil
	case errors.Is(avcErr, status.ErrInsufficientStorage):
		log.Warnf("video: insufficient storage to transcode %s", clean.Log(f.FileName))
		return "", "", status.ErrInsufficientStorage
	case avcErr != nil:
		log.Debugf("video: %s", clean.Error(avcErr))
	}

	log.Errorf("video: failed to transcode %s", clean.Log(f.FileName))

	return "", "", errVideoUnavailable
}

// insta360Capture reports whether f is an Insta360 lens or proxy video.
func insta360Capture(f *entity.File) bool {
	switch f.Type() {
	case fs.VideoInsv, fs.VideoLrv:
		return true
	default:
		return false
	}
}
