package service

import (
	"github.com/photoprism/photoprism/internal/classify"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/nsfw"
	"github.com/photoprism/photoprism/internal/photoprism"
)

var conf *config.Config

var services struct {
	Import   *photoprism.Import
	Index    *photoprism.Index
	Nsfw     *nsfw.Detector
	Convert  *photoprism.Convert
	Resample *photoprism.Resample
	Classify *classify.TensorFlow
}

func SetConfig(c *config.Config)  {
	if c == nil {
		panic("config is nil")
	}

	conf = c
}

func Config() *config.Config {
	if conf == nil {
		panic("config is nil")
	}

	return conf
}
