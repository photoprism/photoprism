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
type ClientConfig map[string]interface{}

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

	jsHash := fs.Checksum(c.HttpStaticBuildPath() + "/app.js")
	cssHash := fs.Checksum(c.HttpStaticBuildPath() + "/app.css")
	configFlags := c.Flags()

	var noPos = struct {
		PhotoUID   string    `json:"photo"`
		LocationID string    `json:"location"`
		TakenAt    time.Time `json:"utc"`
		PhotoLat   float64   `json:"lat"`
		PhotoLng   float64   `json:"lng"`
	}{}

	var count = struct {
		Photos         uint `json:"photos"`
		Videos         uint `json:"videos"`
		Hidden         uint `json:"hidden"`
		Favorites      uint `json:"favorites"`
		Private        uint `json:"private"`
		Review         uint `json:"review"`
		Stories        uint `json:"stories"`
		Albums         uint `json:"albums"`
		Folders        uint `json:"folders"`
		Files          uint `json:"files"`
		Moments        uint `json:"moments"`
		Countries      uint `json:"countries"`
		Places         uint `json:"places"`
		Labels         uint `json:"labels"`
		LabelMaxPhotos uint `json:"labelMaxPhotos"`
	}{}

	result := ClientConfig{
		"settings":        c.Settings(),
		"flags":           strings.Join(configFlags, " "),
		"name":            c.Name(),
		"url":             c.Url(),
		"title":           c.Title(),
		"subtitle":        c.Subtitle(),
		"description":     c.Description(),
		"author":          c.Author(),
		"version":         c.Version(),
		"copyright":       c.Copyright(),
		"debug":           c.Debug(),
		"readonly":        c.ReadOnly(),
		"uploadNSFW":      c.UploadNSFW(),
		"public":          c.Public(),
		"experimental":    c.Experimental(),
		"disableSettings": c.DisableSettings(),
		"albums":          []string{},
		"cameras":         []string{},
		"lenses":          []string{},
		"countries":       []string{},
		"thumbnails":      Thumbnails,
		"jsHash":          jsHash,
		"cssHash":         cssHash,
		"count":           count,
		"pos":             noPos,
		"years":           []int{},
		"colors":          colors.All.List(),
		"categories":      []string{},
		"clip":            txt.ClipDefault,
		"server":          RuntimeInfo{},
	}

	return result
}

// ClientConfig returns a loaded and set configuration entity.
func (c *Config) ClientConfig() ClientConfig {
	defer log.Debug(capture.Time(time.Now(), "config: client config created"))

	db := c.Db()

	var cameras []entity.Camera
	var lenses []entity.Lens
	var albums []entity.Album
	var countries []entity.Country

	var position struct {
		PhotoUID   string    `json:"photo"`
		LocationID string    `json:"location"`
		TakenAt    time.Time `json:"utc"`
		PhotoLat   float64   `json:"lat"`
		PhotoLng   float64   `json:"lng"`
	}

	db.Table("photos").
		Select("photo_uid, location_id, photo_lat, photo_lng, taken_at").
		Where("deleted_at IS NULL AND photo_lat != 0 AND photo_lng != 0").
		Order("taken_at DESC").
		Limit(1).Offset(0).
		Take(&position)

	var count = struct {
		Photos         uint `json:"photos"`
		Videos         uint `json:"videos"`
		Hidden         uint `json:"hidden"`
		Favorites      uint `json:"favorites"`
		Private        uint `json:"private"`
		Review         uint `json:"review"`
		Albums         uint `json:"albums"`
		Folders        uint `json:"folders"`
		Files          uint `json:"files"`
		Moments        uint `json:"moments"`
		Countries      uint `json:"countries"`
		Places         uint `json:"places"`
		Labels         uint `json:"labels"`
		LabelMaxPhotos uint `json:"labelMaxPhotos"`
	}{}

	db.Table("photos").
		Select("SUM(photo_type = 'video' AND photo_quality >= 0 AND photo_private = 0) AS videos, SUM(photo_type IN ('image','raw','live') AND photo_quality < 3 AND photo_quality >= 0 AND photo_private = 0) AS review, SUM(photo_quality = -1) AS hidden, SUM(photo_type IN ('image','raw','live') AND photo_private = 0 AND photo_quality >= 0) AS photos, SUM(photo_favorite = 1 AND photo_quality >= 0) AS favorites, SUM(photo_private = 1 AND photo_quality >= 0) AS private").
		Where("photos.id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1 AND (file_missing = 1 OR file_error <> ''))").
		Where("deleted_at IS NULL").
		Take(&count)

	db.Table("labels").
		Select("MAX(photo_count) as label_max_photos, COUNT(*) AS labels").
		Where("photo_count > 0").
		Where("deleted_at IS NULL").
		Where("(label_priority >= 0 || label_favorite = 1)").
		Take(&count)

	db.Table("albums").
		Select("SUM(album_type = '') AS albums, SUM(album_type = 'moment') AS moments, SUM(album_type = 'folder') AS folders").
		Where("deleted_at IS NULL").
		Take(&count)

	db.Table("folders").
		Select("COUNT(*) AS folders").
		Where("folder_hidden = 0").
		Where("deleted_at IS NULL").
		Take(&count)

	db.Table("files").
		Select("COUNT(*) AS files").
		Where("file_missing = 0").
		Where("deleted_at IS NULL").
		Take(&count)

	db.Table("countries").
		Select("(COUNT(*) - 1) AS countries").
		Take(&count)

	db.Table("places").
		Select("SUM(photo_count > 0) AS places").
		Where("id != 'zz'").
		Take(&count)

	db.Order("country_slug").
		Find(&countries)

	db.Where("deleted_at IS NULL").
		Limit(10000).Order("camera_slug").
		Find(&cameras)

	db.Where("deleted_at IS NULL").
		Limit(10000).Order("lens_slug").
		Find(&lenses)

	db.Where("deleted_at IS NULL AND album_favorite = 1").
		Limit(20).Order("album_name").
		Find(&albums)

	var years []int

	db.Table("photos").
		Where("photo_year > 0").
		Order("photo_year DESC").
		Pluck("DISTINCT photo_year", &years)

	type CategoryLabel struct {
		LabelUID   string `json:"UID"`
		CustomSlug string `json:"Slug"`
		LabelName  string `json:"Name"`
	}

	var categories []CategoryLabel

	db.Table("categories").
		Select("l.label_uid, l.custom_slug, l.label_name").
		Joins("JOIN labels l ON categories.category_id = l.id").
		Where("l.deleted_at IS NULL").
		Group("l.custom_slug").
		Order("l.custom_slug").
		Limit(1000).Offset(0).
		Scan(&categories)

	jsHash := fs.Checksum(c.HttpStaticBuildPath() + "/app.js")
	cssHash := fs.Checksum(c.HttpStaticBuildPath() + "/app.css")
	configFlags := c.Flags()

	result := ClientConfig{
		"flags":           strings.Join(configFlags, " "),
		"name":            c.Name(),
		"url":             c.Url(),
		"title":           c.Title(),
		"subtitle":        c.Subtitle(),
		"description":     c.Description(),
		"author":          c.Author(),
		"version":         c.Version(),
		"copyright":       c.Copyright(),
		"debug":           c.Debug(),
		"readonly":        c.ReadOnly(),
		"uploadNSFW":      c.UploadNSFW(),
		"public":          c.Public(),
		"experimental":    c.Experimental(),
		"disableSettings": c.DisableSettings(),
		"albums":          albums,
		"cameras":         cameras,
		"lenses":          lenses,
		"countries":       countries,
		"thumbnails":      Thumbnails,
		"jsHash":          jsHash,
		"cssHash":         cssHash,
		"settings":        c.Settings(),
		"count":           count,
		"pos":             position,
		"years":           years,
		"colors":          colors.All.List(),
		"categories":      categories,
		"clip":            txt.ClipDefault,
		"server":          NewRuntimeInfo(),
	}

	return result
}
