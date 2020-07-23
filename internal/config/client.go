package config

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/colors"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// ClientConfig contains HTTP client / Web UI config values
type ClientConfig struct {
	Name            string              `json:"name"`
	Version         string              `json:"version"`
	Copyright       string              `json:"copyright"`
	Flags           string              `json:"flags"`
	SiteUrl         string              `json:"siteUrl"`
	SitePreview     string              `json:"sitePreview"`
	SiteTitle       string              `json:"siteTitle"`
	SiteCaption     string              `json:"siteCaption"`
	SiteDescription string              `json:"siteDescription"`
	SiteAuthor      string              `json:"siteAuthor"`
	Debug           bool                `json:"debug"`
	ReadOnly        bool                `json:"readonly"`
	UploadNSFW      bool                `json:"uploadNSFW"`
	Public          bool                `json:"public"`
	Experimental    bool                `json:"experimental"`
	DisableSettings bool                `json:"disableSettings"`
	AlbumCategories []string            `json:"albumCategories"`
	Albums          []entity.Album      `json:"albums"`
	Cameras         []entity.Camera     `json:"cameras"`
	Lenses          []entity.Lens       `json:"lenses"`
	Countries       []entity.Country    `json:"countries"`
	Thumbs          []Thumb             `json:"thumbs"`
	DownloadToken   string              `json:"downloadToken"`
	PreviewToken    string              `json:"previewToken"`
	JSHash          string              `json:"jsHash"`
	CSSHash         string              `json:"cssHash"`
	Settings        Settings            `json:"settings"`
	Count           ClientCounts        `json:"count"`
	Pos             ClientPosition      `json:"pos"`
	Years           []int               `json:"years"`
	Colors          []map[string]string `json:"colors"`
	Categories      []CategoryLabel     `json:"categories"`
	Clip            int                 `json:"clip"`
	Server          RuntimeInfo         `json:"server"`
}

type ClientCounts struct {
	Cameras        int `json:"cameras"`
	Lenses         int `json:"lenses"`
	Countries      int `json:"countries"`
	Photos         int `json:"photos"`
	Videos         int `json:"videos"`
	Hidden         int `json:"hidden"`
	Favorites      int `json:"favorites"`
	Private        int `json:"private"`
	Review         int `json:"review"`
	Stories        int `json:"stories"`
	Albums         int `json:"albums"`
	Moments        int `json:"moments"`
	Months         int `json:"months"`
	Folders        int `json:"folders"`
	Files          int `json:"files"`
	Places         int `json:"places"`
	States         int `json:"states"`
	Labels         int `json:"labels"`
	LabelMaxPhotos int `json:"labelMaxPhotos"`
}

type CategoryLabel struct {
	LabelUID   string `json:"UID"`
	CustomSlug string `json:"Slug"`
	LabelName  string `json:"Name"`
}

type ClientPosition struct {
	PhotoUID string    `json:"uid"`
	CellID   string    `json:"cid"`
	TakenAt  time.Time `json:"utc"`
	PhotoLat float64   `json:"lat"`
	PhotoLng float64   `json:"lng"`
}

// Flags returns config flags as string slice.
func (c *Config) Flags() (flags []string) {
	if c.Public() {
		flags = append(flags, "public")
	}

	if c.Debug() {
		flags = append(flags, "debug")
	}

	if c.Experimental() {
		flags = append(flags, "experimental")
	}

	if c.ReadOnly() {
		flags = append(flags, "readonly")
	}

	if !c.SettingsHidden() {
		flags = append(flags, "settings")
	}

	return flags
}

// PublicConfig returns public client config values with as little information as possible.
func (c *Config) PublicConfig() ClientConfig {
	if c.Public() {
		return c.UserConfig()
	}

	settings := c.Settings()

	result := ClientConfig{
		Settings: Settings{
			Language: settings.Language,
			Theme:    settings.Theme,
			Maps:     settings.Maps,
			Features: settings.Features,
		},
		Flags:           strings.Join(c.Flags(), " "),
		Name:            c.Name(),
		SiteUrl:         c.SiteUrl(),
		SitePreview:     c.SitePreview(),
		SiteTitle:       c.SiteTitle(),
		SiteCaption:     c.SiteCaption(),
		SiteDescription: c.SiteDescription(),
		SiteAuthor:      c.SiteAuthor(),
		Version:         c.Version(),
		Copyright:       c.Copyright(),
		Debug:           c.Debug(),
		ReadOnly:        c.ReadOnly(),
		Public:          c.Public(),
		Experimental:    c.Experimental(),
		Thumbs:          Thumbs,
		Colors:          colors.All.List(),
		JSHash:          fs.Checksum(c.BuildPath() + "/app.js"),
		CSSHash:         fs.Checksum(c.BuildPath() + "/app.css"),
		Clip:            txt.ClipDefault,
		PreviewToken:    "public",
		DownloadToken:   "public",
	}

	return result
}

// GuestConfig returns client config values for the sharing with guests.
func (c *Config) GuestConfig() ClientConfig {
	settings := c.Settings()

	result := ClientConfig{
		Settings: Settings{
			Language: settings.Language,
			Theme:    settings.Theme,
			Maps:     settings.Maps,
			Features: settings.Features,
		},
		Flags:           "readonly public shared",
		Name:            c.Name(),
		SiteUrl:         c.SiteUrl(),
		SitePreview:     c.SitePreview(),
		SiteTitle:       c.SiteTitle(),
		SiteCaption:     c.SiteCaption(),
		SiteDescription: c.SiteDescription(),
		SiteAuthor:      c.SiteAuthor(),
		Version:         c.Version(),
		Copyright:       c.Copyright(),
		Debug:           c.Debug(),
		ReadOnly:        true,
		UploadNSFW:      c.UploadNSFW(),
		DisableSettings: true,
		Public:          true,
		Experimental:    false,
		Colors:          colors.All.List(),
		Thumbs:          Thumbs,
		DownloadToken:   c.DownloadToken(),
		PreviewToken:    c.PreviewToken(),
		JSHash:          fs.Checksum(c.BuildPath() + "/share.js"),
		CSSHash:         fs.Checksum(c.BuildPath() + "/share.css"),
		Clip:            txt.ClipDefault,
	}

	return result
}

// UserConfig returns client configuration values for registered users.
func (c *Config) UserConfig() ClientConfig {
	result := ClientConfig{
		Settings:        *c.Settings(),
		Flags:           strings.Join(c.Flags(), " "),
		Name:            c.Name(),
		SiteUrl:         c.SiteUrl(),
		SitePreview:     c.SitePreview(),
		SiteTitle:       c.SiteTitle(),
		SiteCaption:     c.SiteCaption(),
		SiteDescription: c.SiteDescription(),
		SiteAuthor:      c.SiteAuthor(),
		Version:         c.Version(),
		Copyright:       c.Copyright(),
		Debug:           c.Debug(),
		ReadOnly:        c.ReadOnly(),
		UploadNSFW:      c.UploadNSFW(),
		DisableSettings: c.SettingsHidden(),
		Public:          c.Public(),
		Experimental:    c.Experimental(),
		Colors:          colors.All.List(),
		Thumbs:          Thumbs,
		DownloadToken:   c.DownloadToken(),
		PreviewToken:    c.PreviewToken(),
		JSHash:          fs.Checksum(c.BuildPath() + "/app.js"),
		CSSHash:         fs.Checksum(c.BuildPath() + "/app.css"),
		Clip:            txt.ClipDefault,
		Server:          NewRuntimeInfo(),
	}

	c.Db().Table("photos").
		Select("photo_uid, cell_id, photo_lat, photo_lng, taken_at").
		Where("deleted_at IS NULL AND photo_lat != 0 AND photo_lng != 0").
		Order("taken_at DESC").
		Limit(1).Offset(0).
		Take(&result.Pos)

	c.Db().Table("cameras").
		Where("camera_slug <> 'zz' AND camera_slug <> ''").
		Select("COUNT(*) AS cameras").
		Take(&result.Count)

	c.Db().Table("lenses").
		Where("lens_slug <> 'zz' AND lens_slug <> ''").
		Select("COUNT(*) AS lenses").
		Take(&result.Count)

	c.Db().Table("photos").
		Select("SUM(photo_type = 'video' AND photo_quality >= 0 AND photo_private = 0) AS videos, SUM(photo_type IN ('image','raw','live') AND photo_quality < 3 AND photo_quality >= 0 AND photo_private = 0) AS review, SUM(photo_quality = -1) AS hidden, SUM(photo_type IN ('image','raw','live') AND photo_private = 0 AND photo_quality >= 0) AS photos, SUM(photo_favorite = 1 AND photo_private = 0 AND photo_quality >= 0) AS favorites, SUM(photo_private = 1 AND photo_quality >= 0) AS private").
		Where("photos.id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1 AND (file_missing = 1 OR file_error <> ''))").
		Where("deleted_at IS NULL").
		Take(&result.Count)

	c.Db().Table("labels").
		Select("MAX(photo_count) as label_max_photos, COUNT(*) AS labels").
		Where("photo_count > 0").
		Where("deleted_at IS NULL").
		Where("(label_priority >= 0 OR label_favorite = 1)").
		Take(&result.Count)

	c.Db().Table("albums").
		Select("SUM(album_type = ?) AS albums, SUM(album_type = ?) AS moments, SUM(album_type = ?) AS months, SUM(album_type = ?) AS states, SUM(album_type = ?) AS folders", entity.AlbumDefault, entity.AlbumMoment, entity.AlbumMonth, entity.AlbumState, entity.AlbumFolder).
		Where("deleted_at IS NULL").
		Take(&result.Count)

	c.Db().Table("files").
		Select("COUNT(*) AS files").
		Where("file_missing = 0").
		Where("deleted_at IS NULL").
		Take(&result.Count)

	c.Db().Table("countries").
		Select("(COUNT(*) - 1) AS countries").
		Take(&result.Count)

	c.Db().Table("places").
		Select("SUM(photo_count > 0) AS places").
		Where("id != 'zz'").
		Take(&result.Count)

	c.Db().Order("country_slug").
		Find(&result.Countries)

	c.Db().Where("deleted_at IS NULL").
		Limit(10000).Order("camera_slug").
		Find(&result.Cameras)

	c.Db().Where("deleted_at IS NULL").
		Limit(10000).Order("lens_slug").
		Find(&result.Lenses)

	c.Db().Where("deleted_at IS NULL AND album_favorite = 1").
		Limit(20).Order("album_title").
		Find(&result.Albums)

	c.Db().Table("photos").
		Where("photo_year > 0").
		Order("photo_year DESC").
		Pluck("DISTINCT photo_year", &result.Years)

	c.Db().Table("categories").
		Select("l.label_uid, l.custom_slug, l.label_name").
		Joins("JOIN labels l ON categories.category_id = l.id").
		Where("l.deleted_at IS NULL").
		Group("l.custom_slug").
		Order("l.custom_slug").
		Limit(1000).Offset(0).
		Scan(&result.Categories)

	c.Db().Table("albums").
		Select("album_category").
		Where("deleted_at IS NULL AND album_category <> ''").
		Group("album_category").
		Order("album_category").
		Limit(1000).Offset(0).
		Pluck("album_category", &result.AlbumCategories)

	return result
}
