package service

import (
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/session"

	gc "github.com/patrickmn/go-cache"
)

var conf *config.Config

var services struct {
	FolderCache *gc.Cache
	CoverCache  *gc.Cache
	ThumbCache  *gc.Cache
	Classify    *classify.TensorFlow
	Convert     *photoprism.Convert
	Files       *photoprism.Files
	Photos      *photoprism.Photos
	Import      *photoprism.Import
	Index       *photoprism.Index
	Moments     *photoprism.Moments
	Faces       *photoprism.Faces
	Places      *photoprism.Places
	Purge       *photoprism.Purge
	CleanUp     *photoprism.CleanUp
	Nsfw        *nsfw.Detector
	FaceNet     *face.Net
	Query       *query.Query
	Resample    *photoprism.Resample
	Session     *session.Session
}

func SetConfig(c *config.Config) {
	if c == nil {
		panic("config is nil")
	}

	conf = c

	photoprism.SetConfig(c)
}

func Config() *config.Config {
	if conf == nil {
		panic("config is nil")
	}

	return conf
}
