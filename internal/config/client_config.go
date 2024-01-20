package config

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/colors"
	"github.com/photoprism/photoprism/pkg/env"
	"github.com/photoprism/photoprism/pkg/txt"
)

type ClientType string

const (
	ClientPublic ClientType = "public"
	ClientShare  ClientType = "share"
	ClientUser   ClientType = "user"
)

// ClientConfig represents HTTP client / Web UI config options.
type ClientConfig struct {
	Mode             string              `json:"mode"`
	Name             string              `json:"name"`
	About            string              `json:"about"`
	Edition          string              `json:"edition"`
	Version          string              `json:"version"`
	Copyright        string              `json:"copyright"`
	Flags            string              `json:"flags"`
	BaseUri          string              `json:"baseUri"`
	StaticUri        string              `json:"staticUri"`
	CssUri           string              `json:"cssUri"`
	JsUri            string              `json:"jsUri"`
	ManifestUri      string              `json:"manifestUri"`
	ApiUri           string              `json:"apiUri"`
	ContentUri       string              `json:"contentUri"`
	VideoUri         string              `json:"videoUri"`
	WallpaperUri     string              `json:"wallpaperUri"`
	SiteUrl          string              `json:"siteUrl"`
	SiteDomain       string              `json:"siteDomain"`
	SiteAuthor       string              `json:"siteAuthor"`
	SiteTitle        string              `json:"siteTitle"`
	SiteCaption      string              `json:"siteCaption"`
	SiteDescription  string              `json:"siteDescription"`
	SitePreview      string              `json:"sitePreview"`
	LegalInfo        string              `json:"legalInfo"`
	LegalUrl         string              `json:"legalUrl"`
	AppName          string              `json:"appName"`
	AppMode          string              `json:"appMode"`
	AppIcon          string              `json:"appIcon"`
	AppColor         string              `json:"appColor"`
	Restart          bool                `json:"restart"`
	Debug            bool                `json:"debug"`
	Trace            bool                `json:"trace"`
	Test             bool                `json:"test"`
	Demo             bool                `json:"demo"`
	Sponsor          bool                `json:"sponsor"`
	ReadOnly         bool                `json:"readonly"`
	UploadNSFW       bool                `json:"uploadNSFW"`
	Public           bool                `json:"public"`
	AuthMode         string              `json:"authMode"`
	UsersPath        string              `json:"usersPath"`
	LoginUri         string              `json:"loginUri"`
	RegisterUri      string              `json:"registerUri"`
	PasswordLength   int                 `json:"passwordLength"`
	PasswordResetUri string              `json:"passwordResetUri"`
	Experimental     bool                `json:"experimental"`
	AlbumCategories  []string            `json:"albumCategories"`
	Albums           entity.Albums       `json:"albums"`
	Cameras          entity.Cameras      `json:"cameras"`
	Lenses           entity.Lenses       `json:"lenses"`
	Countries        entity.Countries    `json:"countries"`
	People           entity.People       `json:"people"`
	Thumbs           ThumbSizes          `json:"thumbs"`
	Tier             int                 `json:"tier"`
	Membership       string              `json:"membership"`
	Customer         string              `json:"customer"`
	MapKey           string              `json:"mapKey"`
	DownloadToken    string              `json:"downloadToken,omitempty"`
	PreviewToken     string              `json:"previewToken,omitempty"`
	Disable          ClientDisable       `json:"disable"`
	Count            ClientCounts        `json:"count"`
	Pos              ClientPosition      `json:"pos"`
	Years            Years               `json:"years"`
	Colors           []map[string]string `json:"colors"`
	Categories       CategoryLabels      `json:"categories"`
	Clip             int                 `json:"clip"`
	Server           env.Resources       `json:"server"`
	Settings         *customize.Settings `json:"settings,omitempty"`
	ACL              acl.Grants          `json:"acl,omitempty"`
	Ext              Values              `json:"ext"`
}

// ApplyACL updates the client config values based on the ACL and Role provided.
func (c ClientConfig) ApplyACL(a acl.ACL, r acl.Role) ClientConfig {
	if c.Settings != nil {
		c.Settings = c.Settings.ApplyACL(a, r)
	}

	c.ACL = a.Grants(r)

	return c
}

// Years represents a list of years.
type Years []int

// ClientDisable represents disabled client features a user cannot turn back on.
type ClientDisable struct {
	WebDAV         bool `json:"webdav"`
	Settings       bool `json:"settings"`
	Places         bool `json:"places"`
	Backups        bool `json:"backups"`
	TensorFlow     bool `json:"tensorflow"`
	Faces          bool `json:"faces"`
	Classification bool `json:"classification"`
	Sips           bool `json:"sips"`
	FFmpeg         bool `json:"ffmpeg"`
	ExifTool       bool `json:"exiftool"`
	Darktable      bool `json:"darktable"`
	RawTherapee    bool `json:"rawtherapee"`
	ImageMagick    bool `json:"imagemagick"`
	HeifConvert    bool `json:"heifconvert"`
	Vectors        bool `json:"vectors"`
	JpegXL         bool `json:"jpegxl"`
	Raw            bool `json:"raw"`
}

// ClientCounts represents photo, video and album counts for the client UI.
type ClientCounts struct {
	All            int `json:"all"`
	Photos         int `json:"photos"`
	Live           int `json:"live"`
	Videos         int `json:"videos"`
	Cameras        int `json:"cameras"`
	Lenses         int `json:"lenses"`
	Countries      int `json:"countries"`
	Hidden         int `json:"hidden"`
	Archived       int `json:"archived"`
	Favorites      int `json:"favorites"`
	Review         int `json:"review"`
	Stories        int `json:"stories"`
	Private        int `json:"private"`
	Albums         int `json:"albums"`
	PrivateAlbums  int `json:"private_albums"`
	Moments        int `json:"moments"`
	PrivateMoments int `json:"private_moments"`
	Months         int `json:"months"`
	PrivateMonths  int `json:"private_months"`
	States         int `json:"states"`
	PrivateStates  int `json:"private_states"`
	Folders        int `json:"folders"`
	PrivateFolders int `json:"private_folders"`
	Files          int `json:"files"`
	People         int `json:"people"`
	Places         int `json:"places"`
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

	if c.Test() {
		flags = append(flags, "test")
	}

	if c.Demo() {
		flags = append(flags, "demo")
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

// ClientPublic returns config values for use by the JavaScript UI and other clients.
func (c *Config) ClientPublic() ClientConfig {
	if c.Public() {
		return c.ClientUser(true).ApplyACL(acl.Resources, acl.RoleAdmin)
	}

	a := c.ClientAssets()

	cfg := ClientConfig{
		Settings: c.PublicSettings(),
		ACL:      acl.Resources.Grants(acl.RoleNone),
		Disable: ClientDisable{
			WebDAV:         true,
			Settings:       c.DisableSettings(),
			Places:         c.DisablePlaces(),
			Backups:        true,
			TensorFlow:     true,
			Faces:          true,
			Classification: true,
			Sips:           true,
			FFmpeg:         true,
			ExifTool:       true,
			Darktable:      true,
			RawTherapee:    true,
			ImageMagick:    true,
			HeifConvert:    true,
			Vectors:        c.DisableVectors(),
			JpegXL:         true,
			Raw:            true,
		},
		Flags:            strings.Join(c.Flags(), " "),
		Mode:             string(ClientPublic),
		Name:             c.Name(),
		About:            c.About(),
		Edition:          c.Edition(),
		BaseUri:          c.BaseUri(""),
		StaticUri:        c.StaticUri(),
		CssUri:           a.AppCssUri(),
		JsUri:            a.AppJsUri(),
		ApiUri:           c.ApiUri(),
		ContentUri:       c.ContentUri(),
		VideoUri:         c.VideoUri(),
		SiteUrl:          c.SiteUrl(),
		SiteDomain:       c.SiteDomain(),
		SiteAuthor:       c.SiteAuthor(),
		SiteTitle:        c.SiteTitle(),
		SiteCaption:      c.SiteCaption(),
		SiteDescription:  c.SiteDescription(),
		SitePreview:      c.SitePreview(),
		LegalInfo:        c.LegalInfo(),
		LegalUrl:         c.LegalUrl(),
		AppName:          c.AppName(),
		AppMode:          c.AppMode(),
		AppIcon:          c.AppIcon(),
		AppColor:         c.AppColor(),
		WallpaperUri:     c.WallpaperUri(),
		Version:          c.Version(),
		Copyright:        c.Copyright(),
		Restart:          c.Restart(),
		Debug:            c.Debug(),
		Trace:            c.Trace(),
		Test:             c.Test(),
		Demo:             c.Demo(),
		Sponsor:          c.Sponsor(),
		ReadOnly:         c.ReadOnly(),
		Public:           c.Public(),
		AuthMode:         c.AuthMode(),
		UsersPath:        c.UsersPath(),
		LoginUri:         c.LoginUri(),
		RegisterUri:      c.RegisterUri(),
		PasswordResetUri: c.PasswordResetUri(),
		Experimental:     c.Experimental(),
		Albums:           entity.Albums{},
		Cameras:          entity.Cameras{},
		Lenses:           entity.Lenses{},
		Countries:        entity.Countries{},
		People:           entity.People{},
		Tier:             c.Hub().Tier(),
		Membership:       c.Hub().Membership(),
		Customer:         "",
		MapKey:           "",
		Thumbs:           Thumbs,
		Colors:           colors.All.List(),
		ManifestUri:      c.ClientManifestUri(),
		Clip:             txt.ClipDefault,
		PreviewToken:     entity.TokenPublic,
		DownloadToken:    entity.TokenPublic,
		Ext:              ClientExt(c, ClientPublic),
	}

	return cfg
}

// ClientShare returns reduced client config values for share link visitors.
func (c *Config) ClientShare() ClientConfig {
	a := c.ClientAssets()

	cfg := ClientConfig{
		Settings: c.ShareSettings(),
		ACL:      acl.Resources.Grants(acl.RoleVisitor),
		Disable: ClientDisable{
			WebDAV:         c.DisableWebDAV(),
			Settings:       c.DisableSettings(),
			Places:         c.DisablePlaces(),
			Backups:        true,
			TensorFlow:     true,
			Faces:          c.DisableFaces(),
			Classification: c.DisableClassification(),
			Sips:           true,
			FFmpeg:         true,
			ExifTool:       true,
			Darktable:      true,
			RawTherapee:    true,
			ImageMagick:    true,
			HeifConvert:    true,
			Vectors:        c.DisableVectors(),
			JpegXL:         c.DisableJpegXL(),
			Raw:            c.DisableRaw(),
		},
		Flags:            strings.Join(c.Flags(), " "),
		Mode:             string(ClientShare),
		Name:             c.Name(),
		About:            c.About(),
		Edition:          c.Edition(),
		BaseUri:          c.BaseUri(""),
		StaticUri:        c.StaticUri(),
		CssUri:           a.AppCssUri(),
		JsUri:            a.ShareJsUri(),
		ApiUri:           c.ApiUri(),
		ContentUri:       c.ContentUri(),
		VideoUri:         c.VideoUri(),
		SiteUrl:          c.SiteUrl(),
		SiteDomain:       c.SiteDomain(),
		SiteAuthor:       c.SiteAuthor(),
		SiteTitle:        c.SiteTitle(),
		SiteCaption:      c.SiteCaption(),
		SiteDescription:  c.SiteDescription(),
		SitePreview:      c.SitePreview(),
		LegalInfo:        c.LegalInfo(),
		LegalUrl:         c.LegalUrl(),
		AppName:          c.AppName(),
		AppMode:          c.AppMode(),
		AppIcon:          c.AppIcon(),
		AppColor:         c.AppColor(),
		WallpaperUri:     c.WallpaperUri(),
		Version:          c.Version(),
		Copyright:        c.Copyright(),
		Restart:          c.Restart(),
		Debug:            c.Debug(),
		Trace:            c.Trace(),
		Test:             c.Test(),
		Demo:             c.Demo(),
		Sponsor:          c.Sponsor(),
		ReadOnly:         c.ReadOnly(),
		UploadNSFW:       c.UploadNSFW(),
		Public:           c.Public(),
		AuthMode:         c.AuthMode(),
		UsersPath:        "",
		LoginUri:         c.LoginUri(),
		RegisterUri:      c.RegisterUri(),
		PasswordResetUri: c.PasswordResetUri(),
		Experimental:     c.Experimental(),
		Albums:           entity.Albums{},
		Cameras:          entity.Cameras{},
		Lenses:           entity.Lenses{},
		Countries:        entity.Countries{},
		People:           entity.People{},
		Colors:           colors.All.List(),
		Thumbs:           Thumbs,
		Tier:             c.Hub().Tier(),
		Membership:       c.Hub().Membership(),
		Customer:         c.Hub().Customer(),
		MapKey:           c.Hub().MapKey(),
		DownloadToken:    c.DownloadToken(),
		PreviewToken:     c.PreviewToken(),
		ManifestUri:      c.ClientManifestUri(),
		Clip:             txt.ClipDefault,
		Ext:              ClientExt(c, ClientShare),
	}

	return cfg
}

// ClientUser returns complete client config values for users with full access.
func (c *Config) ClientUser(withSettings bool) ClientConfig {
	a := c.ClientAssets()

	var s *customize.Settings

	if withSettings {
		s = c.Settings()
	}

	cfg := ClientConfig{
		Settings: s,
		Disable: ClientDisable{
			WebDAV:         c.DisableWebDAV(),
			Settings:       c.DisableSettings(),
			Places:         c.DisablePlaces(),
			Backups:        c.DisableBackups(),
			TensorFlow:     c.DisableTensorFlow(),
			Faces:          c.DisableFaces(),
			Classification: c.DisableClassification(),
			Sips:           c.DisableSips(),
			FFmpeg:         c.DisableFFmpeg(),
			ExifTool:       c.DisableExifTool(),
			Darktable:      c.DisableDarktable(),
			RawTherapee:    c.DisableRawTherapee(),
			ImageMagick:    c.DisableImageMagick(),
			HeifConvert:    c.DisableHeifConvert(),
			Vectors:        c.DisableVectors(),
			JpegXL:         c.DisableJpegXL(),
			Raw:            c.DisableRaw(),
		},
		Flags:            strings.Join(c.Flags(), " "),
		Mode:             string(ClientUser),
		Name:             c.Name(),
		About:            c.About(),
		Edition:          c.Edition(),
		BaseUri:          c.BaseUri(""),
		StaticUri:        c.StaticUri(),
		CssUri:           a.AppCssUri(),
		JsUri:            a.AppJsUri(),
		ApiUri:           c.ApiUri(),
		ContentUri:       c.ContentUri(),
		VideoUri:         c.VideoUri(),
		SiteUrl:          c.SiteUrl(),
		SiteDomain:       c.SiteDomain(),
		SiteAuthor:       c.SiteAuthor(),
		SiteTitle:        c.SiteTitle(),
		SiteCaption:      c.SiteCaption(),
		SiteDescription:  c.SiteDescription(),
		SitePreview:      c.SitePreview(),
		LegalInfo:        c.LegalInfo(),
		LegalUrl:         c.LegalUrl(),
		AppName:          c.AppName(),
		AppMode:          c.AppMode(),
		AppIcon:          c.AppIcon(),
		AppColor:         c.AppColor(),
		WallpaperUri:     c.WallpaperUri(),
		Version:          c.Version(),
		Copyright:        c.Copyright(),
		Restart:          c.Restart(),
		Debug:            c.Debug(),
		Trace:            c.Trace(),
		Test:             c.Test(),
		Demo:             c.Demo(),
		Sponsor:          c.Sponsor(),
		ReadOnly:         c.ReadOnly(),
		UploadNSFW:       c.UploadNSFW(),
		Public:           c.Public(),
		AuthMode:         c.AuthMode(),
		UsersPath:        c.UsersPath(),
		LoginUri:         c.LoginUri(),
		RegisterUri:      c.RegisterUri(),
		PasswordLength:   c.PasswordLength(),
		PasswordResetUri: c.PasswordResetUri(),
		Experimental:     c.Experimental(),
		Albums:           entity.Albums{},
		Cameras:          entity.Cameras{},
		Lenses:           entity.Lenses{},
		Countries:        entity.Countries{},
		People:           entity.People{},
		Colors:           colors.All.List(),
		Thumbs:           Thumbs,
		Tier:             c.Hub().Tier(),
		Membership:       c.Hub().Membership(),
		Customer:         c.Hub().Customer(),
		MapKey:           c.Hub().MapKey(),
		DownloadToken:    c.DownloadToken(),
		PreviewToken:     c.PreviewToken(),
		ManifestUri:      c.ClientManifestUri(),
		Clip:             txt.ClipDefault,
		Server:           env.Info(),
		Ext:              ClientExt(c, ClientUser),
	}

	// Query start time.
	start := time.Now()

	hidePrivate := c.Settings().Features.Private

	c.Db().
		Table("photos").
		Select("photo_uid, cell_id, photo_lat, photo_lng, taken_at").
		Where("deleted_at IS NULL AND photo_lat <> 0 AND photo_lng <> 0").
		Order("taken_at DESC").
		Limit(1).Offset(0).
		Take(&cfg.Pos)

	c.Db().
		Table("cameras").
		Where("camera_slug <> 'zz' AND camera_slug <> ''").
		Select("COUNT(*) AS cameras").
		Take(&cfg.Count)

	c.Db().
		Table("lenses").
		Where("lens_slug <> 'zz' AND lens_slug <> ''").
		Select("COUNT(*) AS lenses").
		Take(&cfg.Count)

	if hidePrivate {
		c.Db().
			Table("photos").
			Select("SUM(photo_type = 'video' AND photo_quality > -1 AND photo_private = 0) AS videos, " +
				"SUM(photo_type = 'live' AND photo_quality > -1 AND photo_private = 0) AS live, " +
				"SUM(photo_quality = -1) AS hidden, " +
				"SUM(photo_type NOT IN ('live', 'video') AND photo_quality > -1 AND photo_private = 0) AS photos, " +
				"SUM(photo_quality BETWEEN 0 AND 2) AS review, " +
				"SUM(photo_favorite = 1 AND photo_private = 0 AND photo_quality > -1) AS favorites, " +
				"SUM(photo_private = 1 AND photo_quality > -1) AS private").
			Where("photos.id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1 AND (file_missing = 1 OR file_error <> ''))").
			Where("deleted_at IS NULL").
			Take(&cfg.Count)
	} else {
		c.Db().
			Table("photos").
			Select("SUM(photo_type = 'video' AND photo_quality > -1) AS videos, " +
				"SUM(photo_type = 'live' AND photo_quality > -1) AS live, " +
				"SUM(photo_quality = -1) AS hidden, " +
				"SUM(photo_type NOT IN ('live', 'video') AND photo_quality > -1) AS photos, " +
				"SUM(photo_quality BETWEEN 0 AND 2) AS review, " +
				"SUM(photo_favorite = 1 AND photo_quality > -1) AS favorites, " +
				"0 AS private").
			Where("photos.id NOT IN (SELECT photo_id FROM files WHERE file_primary = 1 AND (file_missing = 1 OR file_error <> ''))").
			Where("deleted_at IS NULL").
			Take(&cfg.Count)
	}

	// Get number of archived pictures.
	if c.Settings().Features.Archive {
		c.Db().
			Table("photos").
			Select("SUM(photo_quality > -1) AS archived").
			Where("deleted_at IS NOT NULL").
			Take(&cfg.Count)
	}

	// Calculate total count.
	cfg.Count.All = cfg.Count.Photos + cfg.Count.Live + cfg.Count.Videos

	// Exclude pictures in review from total count.
	if c.Settings().Features.Review {
		cfg.Count.All = cfg.Count.All - cfg.Count.Review
	}

	c.Db().
		Table("labels").
		Select("MAX(photo_count) AS label_max_photos, COUNT(*) AS labels").
		Where("photo_count > 0").
		Where("deleted_at IS NULL").
		Where("(label_priority >= 0 OR label_favorite = 1)").
		Take(&cfg.Count)

	if hidePrivate {
		c.Db().
			Table("albums").
			Select("SUM(album_type = ?) AS albums, "+
				"SUM(album_type = ?) AS moments, "+
				"SUM(album_type = ?) AS months, "+
				"SUM(album_type = ?) AS states, "+
				"SUM(album_type = ?) AS folders, "+
				"SUM(album_type = ? AND album_private = 1) AS private_albums, "+
				"SUM(album_type = ? AND album_private = 1) AS private_moments, "+
				"SUM(album_type = ? AND album_private = 1) AS private_months, "+
				"SUM(album_type = ? AND album_private = 1) AS private_states, "+
				"SUM(album_type = ? AND album_private = 1) AS private_folders",
				entity.AlbumManual, entity.AlbumMoment, entity.AlbumMonth, entity.AlbumState, entity.AlbumFolder,
				entity.AlbumManual, entity.AlbumMoment, entity.AlbumMonth, entity.AlbumState, entity.AlbumFolder).
			Where("deleted_at IS NULL AND (albums.album_type <> 'folder' OR albums.album_path IN (SELECT photos.photo_path FROM photos WHERE photos.photo_private = 0 AND photos.deleted_at IS NULL))").
			Take(&cfg.Count)
	} else {
		c.Db().
			Table("albums").
			Select("SUM(album_type = ?) AS albums, "+
				"SUM(album_type = ?) AS moments, "+
				"SUM(album_type = ?) AS months, "+
				"SUM(album_type = ?) AS states, "+
				"SUM(album_type = ?) AS folders",
				entity.AlbumManual, entity.AlbumMoment, entity.AlbumMonth, entity.AlbumState, entity.AlbumFolder).
			Where("deleted_at IS NULL AND (albums.album_type <> 'folder' OR albums.album_path IN (SELECT photos.photo_path FROM photos WHERE photos.deleted_at IS NULL))").
			Take(&cfg.Count)
	}

	c.Db().
		Table("files").
		Select("COUNT(*) AS files").
		Where("file_missing = 0 AND file_root = ? AND deleted_at IS NULL", entity.RootOriginals).
		Take(&cfg.Count)

	c.Db().
		Table("countries").
		Select("(COUNT(*) - 1) AS countries").
		Take(&cfg.Count)

	c.Db().
		Table("places").
		Select("SUM(photo_count > 0) AS places").
		Where("id <> 'zz'").
		Take(&cfg.Count)

	c.Db().
		Order("country_slug").
		Find(&cfg.Countries)

	// People are subjects with type person.
	cfg.Count.People, _ = query.PeopleCount()
	cfg.People, _ = query.People()

	c.Db().
		Where("id IN (SELECT photos.camera_id FROM photos WHERE photos.photo_quality > -1 OR photos.deleted_at IS NULL)").
		Where("deleted_at IS NULL").
		Limit(10000).Order("camera_slug").
		Find(&cfg.Cameras)

	c.Db().
		Where("deleted_at IS NULL").
		Limit(10000).Order("lens_slug").
		Find(&cfg.Lenses)

	c.Db().
		Where("deleted_at IS NULL AND album_favorite = 1").
		Limit(20).Order("album_title").
		Find(&cfg.Albums)

	c.Db().
		Table("photos").
		Where("photo_year > 0 AND (photos.photo_quality > -1 OR photos.deleted_at IS NULL)").
		Order("photo_year DESC").
		Pluck("DISTINCT photo_year", &cfg.Years)

	c.Db().
		Table("categories").
		Select("l.label_uid, l.custom_slug, l.label_name").
		Joins("JOIN labels l ON categories.category_id = l.id").
		Where("l.deleted_at IS NULL").
		Group("l.custom_slug, l.label_uid, l.label_name").
		Order("l.custom_slug").
		Limit(10000).Offset(0).
		Scan(&cfg.Categories)

	c.Db().
		Table("albums").
		Select("album_category").
		Where("deleted_at IS NULL AND album_category <> ''").
		Group("album_category").
		Order("album_category").
		Limit(10000).Offset(0).
		Pluck("album_category", &cfg.AlbumCategories)

	// Trace log for performance measurement.
	log.Tracef("config: updated counts [%s]", time.Since(start))

	return cfg
}

// ClientRole provides the client config values for the specified user role.
func (c *Config) ClientRole(role acl.Role) ClientConfig {
	return c.ClientUser(true).ApplyACL(acl.Resources, role)
}

// ClientSession provides the client config values for the specified session.
func (c *Config) ClientSession(sess *entity.Session) (cfg ClientConfig) {
	if sess.NoUser() && sess.IsClient() {
		cfg = c.ClientUser(false).ApplyACL(acl.Resources, sess.ClientRole())
		cfg.Settings = c.SessionSettings(sess)
	} else if sess.User().IsVisitor() {
		cfg = c.ClientShare()
	} else if sess.User().IsRegistered() {
		cfg = c.ClientUser(false).ApplyACL(acl.Resources, sess.UserRole())
		cfg.Settings = c.SessionSettings(sess)
	} else {
		cfg = c.ClientPublic()
	}

	if c.Public() {
		cfg.PreviewToken = entity.TokenPublic
		cfg.DownloadToken = entity.TokenPublic
	} else if sess.PreviewToken != "" || sess.DownloadToken != "" {
		cfg.PreviewToken = sess.PreviewToken
		cfg.DownloadToken = sess.DownloadToken
	}

	return cfg
}
