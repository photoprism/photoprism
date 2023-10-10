package pwa

import (
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Manifest represents a progressive web app manifest.
type Manifest struct {
	Name            string        `json:"name"`
	ShortName       string        `json:"short_name,omitempty"`
	Description     string        `json:"description,omitempty"`
	Categories      list.List     `json:"categories"`
	Display         string        `json:"display"`
	Orientation     string        `json:"orientation"`
	ThemeColor      string        `json:"theme_color"`
	BackgroundColor string        `json:"background_color"`
	Scope           string        `json:"scope"`
	StartUrl        string        `json:"start_url,omitempty"`
	Serviceworker   Serviceworker `json:"serviceworker,omitempty"`
	Permissions     list.List     `json:"permissions"`
	Icons           Icons         `json:"icons"`
}

// NewManifest creates a new progressive web app manifest based on the config provided.
func NewManifest(c Config) (m *Manifest) {
	return &Manifest{
		Name:            c.Name,
		ShortName:       txt.Clip(c.Name, 32),
		Description:     c.Description,
		Categories:      Categories,
		Display:         c.Mode,
		Orientation:     "any",
		ThemeColor:      clean.Color(c.Color),
		BackgroundColor: clean.Color(c.Color),
		Scope:           c.BaseUri,
		StartUrl:        c.BaseUri + "library/",
		Serviceworker: Serviceworker{
			Src:      "sw.js",
			Scope:    c.BaseUri,
			UseCache: true,
		},
		Permissions: Permissions,
		Icons:       NewIcons(c.StaticUri, c.Icon),
	}
}
