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
		PhotoUUID  string    `json:"photo"`
		LocationID string    `json:"location"`
		TakenAt    time.Time `json:"utc"`
		PhotoLat   float64   `json:"lat"`
		PhotoLng   float64   `json:"lng"`
	}{}

	var count = struct {
		Photos    uint `json:"photos"`
		Videos    uint `json:"videos"`
		Hidden    uint `json:"hidden"`
		Favorites uint `json:"favorites"`
		Private   uint `json:"private"`
		Stories   uint `json:"stories"`
		Labels    uint `json:"labels"`
		Albums    uint `json:"albums"`
		Countries uint `json:"countries"`
		Places    uint `json:"places"`
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

	var cameras []*entity.Camera
	var lenses []*entity.Lens
	var albums []*entity.Album

	var position struct {
		PhotoUUID  string    `json:"photo"`
		LocationID string    `json:"location"`
		TakenAt    time.Time `json:"utc"`
		PhotoLat   float64   `json:"lat"`
		PhotoLng   float64   `json:"lng"`
	}

	db.Table("photos").
		Select("photo_uuid, location_id, photo_lat, photo_lng, taken_at").
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
		Albums         uint `json:"albums"`
		Countries      uint `json:"countries"`
		Places         uint `json:"places"`
		Labels         uint `json:"labels"`
		LabelMaxPhotos uint `json:"labelMaxPhotos"`
	}{}

	db.Table("photos").
		Select("SUM(photo_video = 1 AND photo_quality >= 0) AS videos, SUM(photo_quality = -1) AS hidden, SUM(photo_quality >= 0) AS photos, SUM(photo_favorite) AS favorites, SUM(photo_private) AS private").
		Where("deleted_at IS NULL").
		Take(&count)

	db.Table("labels").
		Select("MAX(photo_count) as label_max_photos, COUNT(*) AS labels").
		Where("photo_count > 0").
		Where("deleted_at IS NULL").
		Where("(label_priority >= 0 || label_favorite = 1)").
		Take(&count)

	db.Table("albums").
		Select("COUNT(*) AS albums").
		Where("deleted_at IS NULL").
		Take(&count)

	db.Table("countries").
		Select("(COUNT(*) - 1) AS countries").
		Take(&count)

	db.Table("places").
		Select("SUM(photo_count > 0) AS places").
		Where("id != 'zz'").
		Take(&count)

	type country struct {
		ID          string `json:"code"`
		CountryName string `json:"name"`
	}

	var countries []country

	db.Model(&entity.Country{}).
		Select("id, country_name").
		Order("country_slug").
		Scan(&countries)

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
		LabelName string
		Title     string
	}

	var categories []CategoryLabel

	db.Table("categories").
		Select("l.label_name").
		Joins("JOIN labels l ON categories.category_id = l.id").
		Group("l.label_name").
		Order("l.label_name").
		Limit(1000).Offset(0).
		Scan(&categories)

	for i, l := range categories {
		categories[i].Title = strings.Title(l.LabelName)
	}

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
