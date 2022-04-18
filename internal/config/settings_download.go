package config

import "github.com/photoprism/photoprism/internal/entity"

// DownloadSettings represents content download settings.
type DownloadSettings struct {
	Name         entity.DownloadName `json:"name" yaml:"Name"`
	Disabled     bool                `json:"disabled" yaml:"Disabled"`
	Originals    bool                `json:"originals" yaml:"Originals"`
	MediaRaw     bool                `json:"mediaRaw" yaml:"MediaRaw"`
	MediaSidecar bool                `json:"mediaSidecar" yaml:"MediaSidecar"`
}

// NewDownloadSettings creates download settings with defaults.
func NewDownloadSettings() DownloadSettings {
	return DownloadSettings{
		Name:         entity.DownloadNameDefault,
		Disabled:     false,
		Originals:    true,
		MediaRaw:     false,
		MediaSidecar: false,
	}
}
