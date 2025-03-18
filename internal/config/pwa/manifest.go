package pwa

import (
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Manifest represents a progressive web app manifest.
type Manifest struct {
	ManifestVersion     int           `json:"manifest_version"`
	ID                  string        `json:"id"`
	Name                string        `json:"name"`
	ShortName           string        `json:"short_name,omitempty"`
	Description         string        `json:"description,omitempty"`
	Categories          list.List     `json:"categories"`
	Developer           Url           `json:"developer"`
	DisplayOverride     []string      `json:"display_override"`
	Display             string        `json:"display"`
	Orientation         string        `json:"orientation"`
	DefaultLocale       string        `json:"default_locale"`
	ThemeColor          string        `json:"theme_color"`
	BackgroundColor     string        `json:"background_color"`
	Scope               string        `json:"scope"`
	StartUrl            string        `json:"start_url,omitempty"`
	Shortcuts           Urls          `json:"shortcuts"`
	Serviceworker       Serviceworker `json:"serviceworker,omitempty"`
	Permissions         list.List     `json:"permissions"`
	OptionalPermissions list.List     `json:"optional_permissions"`
	HostPermissions     []string      `json:"host_permissions"`
	Icons               Icons         `json:"icons"`
}

// NewManifest creates a new progressive web app manifest based on the config provided.
func NewManifest(c Config) (m *Manifest) {
	return &Manifest{
		ManifestVersion: 2,
		ID:              c.SiteUrl,
		Name:            c.Name,
		ShortName:       txt.Clip(c.Name, 32),
		Description:     c.Description,
		Categories:      Categories,
		Developer:       PhotoPrism,
		DisplayOverride: DisplayOverride,
		Display:         c.Mode,
		Orientation:     "any",
		DefaultLocale:   c.DefaultLocale,
		ThemeColor:      clean.Color(c.Color),
		BackgroundColor: clean.Color(c.Color),
		Scope:           c.BaseUri,
		StartUrl:        c.BaseUri + "library/",
		Shortcuts:       Shortcuts(c.BaseUri),
		Serviceworker: Serviceworker{
			Src:      "sw.js",
			Scope:    c.BaseUri,
			UseCache: true,
		},
		Permissions:         Permissions,
		OptionalPermissions: OptionalPermissions,
		HostPermissions:     HostPermissions(c.SiteUrl, c.CdnUrl),
		Icons:               NewIcons(c.StaticUri, c.Icon),
	}
}
