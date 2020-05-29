package config

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/colors"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// ClientConfig contains HTTP client / Web UI config values
type ClientConfig struct {
	Flags           string              `json:"flags"`
	Name            string              `json:"name"`
	URL             string              `json:"url"`
	Title           string              `json:"title"`
	Subtitle        string              `json:"subtitle"`
	Description     string              `json:"description"`
	Author          string              `json:"author"`
	Version         string              `json:"version"`
	Copyright       string              `json:"copyright"`
	Debug           bool                `json:"debug"`
	ReadOnly        bool                `json:"readonly"`
	UploadNSFW      bool                `json:"uploadNSFW"`
	Public          bool                `json:"public"`
	Experimental    bool                `json:"experimental"`
	DisableSettings bool                `json:"disableSettings"`
	Albums          []entity.Album      `json:"albums"`
	Cameras         []entity.Camera     `json:"cameras"`
	Lenses          []entity.Lens       `json:"lenses"`
	Countries       []entity.Country    `json:"countries"`
	Thumbnails      []Thumbnail         `json:"thumbnails"`
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
	Labels         int `json:"labels"`
	LabelMaxPhotos int `json:"labelMaxPhotos"`
}

type CategoryLabel struct {
	LabelUID   string `json:"UID"`
	CustomSlug string `json:"Slug"`
	LabelName  string `json:"Name"`
}

type ClientPosition struct {
	PhotoUID   string    `json:"uid"`
	LocationID string    `json:"loc"`
	TakenAt    time.Time `json:"utc"`
	PhotoLat   float64   `json:"lat"`
	PhotoLng   float64   `json:"lng"`
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

	if !c.DisableSettings() {
		flags = append(flags, "settings")
	}

	return flags
}

// PublicClientConfig returns reduced config values for non-public sites.
func (c *Config) PublicClientConfig() ClientConfig {
	if c.Public() {
		return c.ClientConfig()
	}

	settings := c.Settings()

	result := ClientConfig{
		Settings:      Settings{Language: settings.Language, Theme: settings.Theme},
		Flags:         strings.Join(c.Flags(), " "),
		Name:          c.Name(),
		URL:           c.Url(),
		Title:         c.Title(),
		Subtitle:      c.Subtitle(),
		Description:   c.Description(),
		Author:        c.Author(),
		Version:       c.Version(),
		Copyright:     c.Copyright(),
		Debug:         c.Debug(),
		ReadOnly:      c.ReadOnly(),
		Public:        c.Public(),
		Experimental:  c.Experimental(),
		Thumbnails:    Thumbnails,
		Colors:        colors.All.List(),
		JSHash:        fs.Checksum(c.HttpStaticBuildPath() + "/app.js"),
		CSSHash:       fs.Checksum(c.HttpStaticBuildPath() + "/app.css"),
		Clip:          txt.ClipDefault,
		PreviewToken:  "public",
		DownloadToken: "public",
	}

	return result
}

// ClientConfig returns a loaded and set configuration entity.
func (c *Config) ClientConfig() ClientConfig {
	defer log.Debug(capture.Time(time.Now(), "config: client config created"))

	result := ClientConfig{
		Settings:        *c.Settings(),
		Flags:           strings.Join(c.Flags(), " "),
		Name:            c.Name(),
		URL:             c.Url(),
		Title:           c.Title(),
		Subtitle:        c.Subtitle(),
		Description:     c.Description(),
		Author:          c.Author(),
		Version:         c.Version(),
		Copyright:       c.Copyright(),
		Debug:           c.Debug(),
		ReadOnly:        c.ReadOnly(),
		UploadNSFW:      c.UploadNSFW(),
		DisableSettings: c.DisableSettings(),
		Public:          c.Public(),
		Experimental:    c.Experimental(),
		Colors:          colors.All.List(),
		Thumbnails:      Thumbnails,
		DownloadToken:   c.DownloadToken(),
		PreviewToken:    c.PreviewToken(),
		JSHash:          fs.Checksum(c.HttpStaticBuildPath() + "/app.js"),
		CSSHash:         fs.Checksum(c.HttpStaticBuildPath() + "/app.css"),
		Clip:            txt.ClipDefault,
		Server:          NewRuntimeInfo(),
	}

	db := c.Db()

	db.Table("photos").
		Select("photo_uid, location_id, photo_lat, photo_lng, taken_at").
		Where("deleted_at IS NULL AND photo_lat != 0 AND photo_lng != 0").
		Order("taken_at DESC").
		Limit(1).Offset(0).
		Take(&result.Pos)

	db.Table("cameras").
		Where("camera_slug <> 'zz' AND camera_slug <> ''").
		Select("COUNT(*) AS cameras").
		Take(&result.Count)

	db.Table("lenses").
		Where("lens_slug <> 'zz' AND lens_slug <> ''").
		Select("COUNT(*) AS lenses").
		Take(&result.Count)

	db.Table("photos").
		Select("SUM(photo_type = 'video' AND photo_quality >= 0 AND photo_private = 0) AS videos, SUM(photo_type IN ('image','raw','live') AND photo_quality < 3 AND photo_quality >= 0 AND photo_private = 0) AS review, SUM(photo_quality = -1) AS hidden, SUM(photo_type IN ('image','raw','live') AND photo_private = 0 AND photo_quality >= 0) AS photos, SUM(photo_favorite = 1 AND photo_quality >= 0) AS favorites, SUM(photo_private = 1 AND photo_quality >= 0) AS private").
		Where("photos.id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1 AND (file_missing = 1 OR file_error <> ''))").
		Where("deleted_at IS NULL").
		Take(&result.Count)

	db.Table("labels").
		Select("MAX(photo_count) as label_max_photos, COUNT(*) AS labels").
		Where("photo_count > 0").
		Where("deleted_at IS NULL").
		Where("(label_priority >= 0 || label_favorite = 1)").
		Take(&result.Count)

	db.Table("albums").
		Select("SUM(album_type = ?) AS albums, SUM(album_type = ?) AS moments, SUM(album_type = ?) AS months, SUM(album_type = ?) AS folders", entity.TypeAlbum, entity.TypeMoment, entity.TypeMonth, entity.TypeFolder).
		Where("deleted_at IS NULL").
		Take(&result.Count)

	db.Table("files").
		Select("COUNT(*) AS files").
		Where("file_missing = 0").
		Where("deleted_at IS NULL").
		Take(&result.Count)

	db.Table("countries").
		Select("(COUNT(*) - 1) AS countries").
		Take(&result.Count)

	db.Table("places").
		Select("SUM(photo_count > 0) AS places").
		Where("id != 'zz'").
		Take(&result.Count)

	db.Order("country_slug").
		Find(&result.Countries)

	db.Where("deleted_at IS NULL").
		Limit(10000).Order("camera_slug").
		Find(&result.Cameras)

	db.Where("deleted_at IS NULL").
		Limit(10000).Order("lens_slug").
		Find(&result.Lenses)

	db.Where("deleted_at IS NULL AND album_favorite = 1").
		Limit(20).Order("album_title").
		Find(&result.Albums)

	db.Table("photos").
		Where("photo_year > 0").
		Order("photo_year DESC").
		Pluck("DISTINCT photo_year", &result.Years)

	db.Table("categories").
		Select("l.label_uid, l.custom_slug, l.label_name").
		Joins("JOIN labels l ON categories.category_id = l.id").
		Where("l.deleted_at IS NULL").
		Group("l.custom_slug").
		Order("l.custom_slug").
		Limit(1000).Offset(0).
		Scan(&result.Categories)

	return result
}
