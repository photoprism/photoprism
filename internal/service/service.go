package service

import (
	"github.com/allegro/bigcache"
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/session"
)

var log = event.Log
var conf *config.Config

var services struct {
	Cache    *bigcache.BigCache
	Classify *classify.TensorFlow
	Convert  *photoprism.Convert
	Files    *photoprism.Files
	Photos   *photoprism.Photos
	Import   *photoprism.Import
	Index    *photoprism.Index
	Moments  *photoprism.Moments
	Purge    *photoprism.Purge
	Nsfw     *nsfw.Detector
	Query    *query.Query
	Resample *photoprism.Resample
	Session  *session.Session
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
