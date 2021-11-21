package config

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/query"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/colors"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// ClientConfig represents HTTP client / Web UI config options.
type ClientConfig struct {
	Mode            string              `json:"mode"`
	Name            string              `json:"name"`
	Version         string              `json:"version"`
	Copyright       string              `json:"copyright"`
	Flags           string              `json:"flags"`
	BaseUri         string              `json:"baseUri"`
	StaticUri       string              `json:"staticUri"`
	ApiUri          string              `json:"apiUri"`
	ContentUri      string              `json:"contentUri"`
	SiteUrl         string              `json:"siteUrl"`
	SiteAuthor      string              `json:"siteAuthor"`
	SiteTitle       string              `json:"siteTitle"`
	SiteCaption     string              `json:"siteCaption"`
	SiteDescription string              `json:"siteDescription"`
	SitePreview     string              `json:"sitePreview"`
	AppName         string              `json:"appName"`
	AppMode         string              `json:"appMode"`
	AppIcon         string              `json:"appIcon"`
	Debug           bool                `json:"debug"`
	Test            bool                `json:"test"`
	Demo            bool                `json:"demo"`
	Sponsor         bool                `json:"sponsor"`
	ReadOnly        bool                `json:"readonly"`
	UploadNSFW      bool                `json:"uploadNSFW"`
	Public          bool                `json:"public"`
	Experimental    bool                `json:"experimental"`
	AlbumCategories []string            `json:"albumCategories"`
	Albums          entity.Albums       `json:"albums"`
	Cameras         entity.Cameras      `json:"cameras"`
	Lenses          entity.Lenses       `json:"lenses"`
	Countries       entity.Countries    `json:"countries"`
	People          entity.People       `json:"people"`
	Thumbs          ThumbSizes          `json:"thumbs"`
	Status          string              `json:"status"`
	MapKey          string              `json:"mapKey"`
	DownloadToken   string              `json:"downloadToken"`
	PreviewToken    string              `json:"previewToken"`
	JSHash          string              `json:"jsHash"`
	CSSHash         string              `json:"cssHash"`
	ManifestHash    string              `json:"manifestHash"`
	Settings        Settings            `json:"settings"`
	Disable         ClientDisable       `json:"disable"`
	Count           ClientCounts        `json:"count"`
	Pos             ClientPosition      `json:"pos"`
	Years           Years               `json:"years"`
	Colors          []map[string]string `json:"colors"`
	Categories      CategoryLabels      `json:"categories"`
	Clip            int                 `json:"clip"`
	Server          RuntimeInfo         `json:"server"`
}

// Years represents a list of years.
type Years []int

// ClientDisable represents disabled client features a user can't turn back on.
type ClientDisable struct {
	Backups        bool `json:"backups"`
	WebDAV         bool `json:"webdav"`
	Settings       bool `json:"settings"`
	Places         bool `json:"places"`
	ExifTool       bool `json:"exiftool"`
	FFmpeg         bool `json:"ffmpeg"`
	Darktable      bool `json:"darktable"`
	Rawtherapee    bool `json:"rawtherapee"`
	Sips           bool `json:"sips"`
	HeifConvert    bool `json:"heifconvert"`
	TensorFlow     bool `json:"tensorflow"`
	Faces          bool `json:"faces"`
	Classification bool `json:"classification"`
}

// ClientCounts represents photo, video and album counts for the client UI.
type ClientCounts struct {
	All            int `json:"all"`
	Photos         int `json:"photos"`
	Videos         int `json:"videos"`
	Cameras        int `json:"cameras"`
	Lenses         int `json:"lenses"`
	Countries      int `json:"countries"`
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
	People         int `json:"people"`
	Places         int `json:"places"`
	States         int `json:"states"`
	Labels         int `json:"labels"`
	LabelMaxPhotos int `json:"labelMaxPhotos"`
}

type CategoryLabels []CategoryLabel

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

	if c.Sponsor() {
		flags = append(flags, "sponsor")
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

	if !c.Settings().UI.Scrollbar {
		flags = append(flags, "hide-scrollbar")
	}

	return flags
}

// PublicConfig returns public client config options with as little information as possible.
func (c *Config) PublicConfig() ClientConfig {
	if c.Public() {
		return c.UserConfig()
	}

	settings := c.Settings()

	result := ClientConfig{
		Settings: Settings{
			UI:       settings.UI,
			Maps:     settings.Maps,
			Features: settings.Features,
			Share:    settings.Share,
		},
		Disable: ClientDisable{
			Backups:        true,
			WebDAV:         true,
			Settings:       c.DisableSettings(),
			Places:         c.DisablePlaces(),
			ExifTool:       true,
			FFmpeg:         true,
			Darktable:      true,
			Rawtherapee:    true,
			Sips:           true,
			HeifConvert:    true,
			TensorFlow:     true,
			Faces:          true,
			Classification: true,
		},
		Flags:           strings.Join(c.Flags(), " "),
		Mode:            "public",
		Name:            c.Name(),
		BaseUri:         c.BaseUri(""),
		StaticUri:       c.StaticUri(),
		ApiUri:          c.ApiUri(),
		ContentUri:      c.ContentUri(),
		SiteUrl:         c.SiteUrl(),
		SiteAuthor:      c.SiteAuthor(),
		SiteTitle:       c.SiteTitle(),
		SiteCaption:     c.SiteCaption(),
		SiteDescription: c.SiteDescription(),
		SitePreview:     c.SitePreview(),
		AppName:         c.AppName(),
		AppMode:         c.AppMode(),
		AppIcon:         c.AppIcon(),
		Version:         c.Version(),
		Copyright:       c.Copyright(),
		Debug:           c.Debug(),
		Test:            c.Test(),
		Demo:            c.Demo(),
		Sponsor:         c.Sponsor(),
		ReadOnly:        c.ReadOnly(),
		Public:          c.Public(),
		Experimental:    c.Experimental(),
		Status:          "",
		MapKey:          "",
		Thumbs:          Thumbs,
		Colors:          colors.All.List(),
		JSHash:          fs.Checksum(c.BuildPath() + "/app.js"),
		CSSHash:         fs.Checksum(c.BuildPath() + "/app.css"),
		ManifestHash:    fs.Checksum(c.TemplatesPath() + "/manifest.json"),
		Clip:            txt.ClipDefault,
		PreviewToken:    "public",
		DownloadToken:   "public",
	}

	return result
}

// GuestConfig returns client config options for the sharing with guests.
func (c *Config) GuestConfig() ClientConfig {
	settings := c.Settings()

	result := ClientConfig{
		Settings: Settings{
			UI:       settings.UI,
			Maps:     settings.Maps,
			Features: settings.Features,
			Share:    settings.Share,
		},
		Disable: ClientDisable{
			Backups:        true,
			WebDAV:         c.DisableWebDAV(),
			Settings:       c.DisableSettings(),
			Places:         c.DisablePlaces(),
			ExifTool:       true,
			FFmpeg:         true,
			Darktable:      true,
			Rawtherapee:    true,
			Sips:           true,
			HeifConvert:    true,
			TensorFlow:     true,
			Faces:          true,
			Classification: true,
		},
		Flags:           strings.Join(c.Flags(), " "),
		Mode:            "guest",
		Name:            c.Name(),
		BaseUri:         c.BaseUri(""),
		StaticUri:       c.StaticUri(),
		ApiUri:          c.ApiUri(),
		ContentUri:      c.ContentUri(),
		SiteUrl:         c.SiteUrl(),
		SiteAuthor:      c.SiteAuthor(),
		SiteTitle:       c.SiteTitle(),
		SiteCaption:     c.SiteCaption(),
		SiteDescription: c.SiteDescription(),
		SitePreview:     c.SitePreview(),
		AppName:         c.AppName(),
		AppMode:         c.AppMode(),
		AppIcon:         c.AppIcon(),
		Version:         c.Version(),
		Copyright:       c.Copyright(),
		Debug:           c.Debug(),
		Test:            c.Test(),
		Demo:            c.Demo(),
		Sponsor:         c.Sponsor(),
		ReadOnly:        true,
		UploadNSFW:      c.UploadNSFW(),
		Public:          true,
		Experimental:    false,
		Colors:          colors.All.List(),
		Thumbs:          Thumbs,
		Status:          c.Hub().Status,
		MapKey:          c.Hub().MapKey(),
		DownloadToken:   c.DownloadToken(),
		PreviewToken:    c.PreviewToken(),
		JSHash:          fs.Checksum(c.BuildPath() + "/share.js"),
		CSSHash:         fs.Checksum(c.BuildPath() + "/share.css"),
		ManifestHash:    fs.Checksum(c.TemplatesPath() + "/manifest.json"),
		Clip:            txt.ClipDefault,
	}

	return result
}

// UserConfig returns client configuration options for registered users.
func (c *Config) UserConfig() ClientConfig {
	result := ClientConfig{
		Settings: *c.Settings(),
		Disable: ClientDisable{
			Backups:        c.DisableBackups(),
			WebDAV:         c.DisableWebDAV(),
			Settings:       c.DisableSettings(),
			Places:         c.DisablePlaces(),
			ExifTool:       c.DisableExifTool(),
			FFmpeg:         c.DisableFFmpeg(),
			Darktable:      c.DisableDarktable(),
			Rawtherapee:    c.DisableRawtherapee(),
			Sips:           c.DisableSips(),
			HeifConvert:    c.DisableHeifConvert(),
			TensorFlow:     c.DisableTensorFlow(),
			Faces:          c.DisableFaces(),
			Classification: c.DisableClassification(),
		},
		Flags:           strings.Join(c.Flags(), " "),
		Mode:            "user",
		Name:            c.Name(),
		BaseUri:         c.BaseUri(""),
		StaticUri:       c.StaticUri(),
		ApiUri:          c.ApiUri(),
		ContentUri:      c.ContentUri(),
		SiteUrl:         c.SiteUrl(),
		SiteAuthor:      c.SiteAuthor(),
		SiteTitle:       c.SiteTitle(),
		SiteCaption:     c.SiteCaption(),
		SiteDescription: c.SiteDescription(),
		SitePreview:     c.SitePreview(),
		AppName:         c.AppName(),
		AppMode:         c.AppMode(),
		AppIcon:         c.AppIcon(),
		Version:         c.Version(),
		Copyright:       c.Copyright(),
		Debug:           c.Debug(),
		Test:            c.Test(),
		Demo:            c.Demo(),
		Sponsor:         c.Sponsor(),
		ReadOnly:        c.ReadOnly(),
		UploadNSFW:      c.UploadNSFW(),
		Public:          c.Public(),
		Experimental:    c.Experimental(),
		Colors:          colors.All.List(),
		Thumbs:          Thumbs,
		Status:          c.Hub().Status,
		MapKey:          c.Hub().MapKey(),
		DownloadToken:   c.DownloadToken(),
		PreviewToken:    c.PreviewToken(),
		JSHash:          fs.Checksum(c.BuildPath() + "/app.js"),
		CSSHash:         fs.Checksum(c.BuildPath() + "/app.css"),
		ManifestHash:    fs.Checksum(c.TemplatesPath() + "/manifest.json"),
		Clip:            txt.ClipDefault,
		Server:          NewRuntimeInfo(),
	}

	c.Db().
		Table("photos").
		Select("photo_uid, cell_id, photo_lat, photo_lng, taken_at").
		Where("deleted_at IS NULL AND photo_lat <> 0 AND photo_lng <> 0").
		Order("taken_at DESC").
		Limit(1).Offset(0).
		Take(&result.Pos)

	c.Db().
		Table("cameras").
		Where("camera_slug <> 'zz' AND camera_slug <> ''").
		Select("COUNT(*) AS cameras").
		Take(&result.Count)

	c.Db().
		Table("lenses").
		Where("lens_slug <> 'zz' AND lens_slug <> ''").
		Select("COUNT(*) AS lenses").
		Take(&result.Count)

	c.Db().
		Table("photos").
		Select("SUM(photo_type = 'video' AND photo_quality >= 0 AND photo_private = 0) AS videos, SUM(photo_type IN ('image','raw','live') AND photo_quality < 3 AND photo_quality >= 0 AND photo_private = 0) AS review, SUM(photo_quality = -1) AS hidden, SUM(photo_type IN ('image','raw','live') AND photo_private = 0 AND photo_quality >= 0) AS photos, SUM(photo_favorite = 1 AND photo_private = 0 AND photo_quality >= 0) AS favorites, SUM(photo_private = 1 AND photo_quality >= 0) AS private").
		Where("photos.id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1 AND (file_missing = 1 OR file_error <> ''))").
		Where("deleted_at IS NULL").
		Take(&result.Count)

	result.Count.All = result.Count.Photos + result.Count.Videos

	c.Db().
		Table("labels").
		Select("MAX(photo_count) AS label_max_photos, COUNT(*) AS labels").
		Where("photo_count > 0").
		Where("deleted_at IS NULL").
		Where("(label_priority >= 0 OR label_favorite = 1)").
		Take(&result.Count)

	c.Db().
		Table("albums").
		Select("SUM(album_type = ?) AS albums, SUM(album_type = ?) AS moments, SUM(album_type = ?) AS months, SUM(album_type = ?) AS states, SUM(album_type = ?) AS folders", entity.AlbumDefault, entity.AlbumMoment, entity.AlbumMonth, entity.AlbumState, entity.AlbumFolder).
		Where("deleted_at IS NULL AND (albums.album_type <> 'folder' OR albums.album_path IN (SELECT photos.photo_path FROM photos WHERE photos.deleted_at IS NULL))").
		Take(&result.Count)

	c.Db().
		Table("files").
		Select("COUNT(*) AS files").
		Where("file_missing = 0").
		Where("deleted_at IS NULL").
		Take(&result.Count)

	c.Db().
		Table("countries").
		Select("(COUNT(*) - 1) AS countries").
		Take(&result.Count)

	c.Db().
		Table("places").
		Select("SUM(photo_count > 0) AS places").
		Where("id <> 'zz'").
		Take(&result.Count)

	c.Db().
		Order("country_slug").
		Find(&result.Countries)

	// People are subjects with type person.
	result.Count.People, _ = query.PeopleCount()
	result.People, _ = query.People()

	c.Db().
		Where("id IN (SELECT photos.camera_id FROM photos WHERE photos.photo_quality >= 0 OR photos.deleted_at IS NULL)").
		Where("deleted_at IS NULL").
		Limit(10000).Order("camera_slug").
		Find(&result.Cameras)

	c.Db().
		Where("deleted_at IS NULL").
		Limit(10000).Order("lens_slug").
		Find(&result.Lenses)

	c.Db().
		Where("deleted_at IS NULL AND album_favorite = 1").
		Limit(20).Order("album_title").
		Find(&result.Albums)

	c.Db().
		Table("photos").
		Where("photo_year > 0 AND (photos.photo_quality >= 0 OR photos.deleted_at IS NULL)").
		Order("photo_year DESC").
		Pluck("DISTINCT photo_year", &result.Years)

	c.Db().
		Table("categories").
		Select("l.label_uid, l.custom_slug, l.label_name").
		Joins("JOIN labels l ON categories.category_id = l.id").
		Where("l.deleted_at IS NULL").
		Group("l.custom_slug").
		Order("l.custom_slug").
		Limit(1000).Offset(0).
		Scan(&result.Categories)

	c.Db().
		Table("albums").
		Select("album_category").
		Where("deleted_at IS NULL AND album_category <> ''").
		Group("album_category").
		Order("album_category").
		Limit(1000).Offset(0).
		Pluck("album_category", &result.AlbumCategories)

	return result
}
