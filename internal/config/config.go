package config

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	gc "github.com/patrickmn/go-cache"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var log = event.Log

// Config holds database, cache and all parameters of photoprism
type Config struct {
	db     *gorm.DB
	cache  *gc.Cache
	config *Params
}

func init() {
	// initialize the Thumbnails global variable
	for name, t := range thumb.Types {
		if t.Public {
			thumbnail := Thumbnail{Name: name, Width: t.Width, Height: t.Height}
			Thumbnails = append(Thumbnails, thumbnail)
		}
	}
}

func initLogger(debug bool) {
	var once sync.Once

	once.Do(func() {
		log.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})

		if debug {
			log.SetLevel(logrus.DebugLevel)
		} else {
			log.SetLevel(logrus.InfoLevel)
		}
	})
}

// NewConfig initialises a new configuration file
func NewConfig(ctx *cli.Context) *Config {
	initLogger(ctx.GlobalBool("debug"))

	c := &Config{
		config: NewParams(ctx),
	}

	log.SetLevel(c.LogLevel())

	thumb.JpegQuality = c.ThumbQuality()
	thumb.PreRenderSize = c.ThumbSize()
	thumb.MaxRenderSize = c.ThumbLimit()
	thumb.Filter = c.ThumbFilter()

	return c
}

// Name returns the application name.
func (c *Config) Name() string {
	return c.config.Name
}

// Url returns the public server URL (default is "http://localhost:2342/").
func (c *Config) Url() string {
	if c.config.Url == "" {
		return "http://localhost:2342/"
	}

	return c.config.Url
}

// Title returns the site title (default is application name).
func (c *Config) Title() string {
	if c.config.Title == "" {
		return c.Name()
	}

	return c.config.Title
}

// Subtitle returns the site title.
func (c *Config) Subtitle() string {
	return c.config.Subtitle
}

// Description returns the site title.
func (c *Config) Description() string {
	return c.config.Description
}

// Author returns the site author / copyright.
func (c *Config) Author() string {
	return c.config.Author
}

// Twitter returns the twitter handle for sharing.
func (c *Config) Twitter() string {
	return c.config.Twitter
}

// Version returns the application version.
func (c *Config) Version() string {
	return c.config.Version
}

// Copyright returns the application copyright.
func (c *Config) Copyright() string {
	return c.config.Copyright
}

// Debug returns true if Debug mode is on.
func (c *Config) Debug() bool {
	return c.config.Debug
}

// Public returns true if app requires no authentication.
func (c *Config) Public() bool {
	return c.config.Public
}

// Experimental returns true if experimental features should be enabled.
func (c *Config) Experimental() bool {
	return c.config.Experimental
}

// ReadOnly returns true if photo directories are write protected.
func (c *Config) ReadOnly() bool {
	return c.config.ReadOnly
}

// DetectNSFW returns true if NSFW photos should be detected and flagged.
func (c *Config) DetectNSFW() bool {
	return c.config.DetectNSFW
}

// UploadNSFW returns true if NSFW photos can be uploaded.
func (c *Config) UploadNSFW() bool {
	return c.config.UploadNSFW
}

// AdminPassword returns the admin password.
func (c *Config) AdminPassword() string {
	if c.config.AdminPassword == "" {
		return "photoprism"
	}

	return c.config.AdminPassword
}

// WebDAVPassword returns the WebDAV password for remote access.
func (c *Config) WebDAVPassword() string {
	return c.config.WebDAVPassword
}

// LogLevel returns the logrus log level.
func (c *Config) LogLevel() logrus.Level {
	if c.Debug() {
		c.config.LogLevel = "debug"
	}

	if logLevel, err := logrus.ParseLevel(c.config.LogLevel); err == nil {
		return logLevel
	} else {
		return logrus.InfoLevel
	}
}

// Cache returns the in-memory cache.
func (c *Config) Cache() *gc.Cache {
	if c.cache == nil {
		c.cache = gc.New(336*time.Hour, 30*time.Minute)
	}

	return c.cache
}

// Init initialises the Database.
func (c *Config) Init(ctx context.Context) error {
	return c.connectToDatabase(ctx)
}

// Shutdown services and workers.
func (c *Config) Shutdown() {
	mutex.Worker.Cancel()

	if err := c.CloseDb(); err != nil {
		log.Errorf("could not close database connection: %s", err)
	} else {
		log.Info("closed database connection")
	}
}

// Workers returns the number of workers e.g. for indexing files.
func (c *Config) Workers() int {
	if c.config.Workers > 0 && c.config.Workers <= runtime.NumCPU() {
		return c.config.Workers
	}

	return runtime.NumCPU()
}

// ThumbQuality returns the thumbnail jpeg quality setting (25-100).
func (c *Config) ThumbQuality() int {
	if c.config.ThumbQuality > 100 {
		return 100
	}

	if c.config.ThumbQuality < 25 {
		return 25
	}

	return c.config.ThumbQuality
}

// ThumbSize returns the pre-rendered thumbnail size limit in pixels (720-3840).
func (c *Config) ThumbSize() int {
	if c.config.ThumbSize > 3840 {
		return 3840
	}

	if c.config.ThumbSize < 720 {
		return 720
	}

	return c.config.ThumbSize
}

// ThumbLimit returns the on-demand thumbnail size limit in pixels (720-3840).
func (c *Config) ThumbLimit() int {
	if c.config.ThumbLimit > 3840 {
		return 3840
	}

	if c.config.ThumbLimit < 720 {
		return 720
	}

	return c.config.ThumbLimit
}

// ThumbFilter returns the thumbnail resample filter (blackman, lanczos, cubic or linear).
func (c *Config) ThumbFilter() thumb.ResampleFilter {
	switch strings.ToLower(c.config.ThumbFilter) {
	case "blackman":
		return thumb.ResampleBlackman
	case "lanczos":
		return thumb.ResampleLanczos
	case "cubic":
		return thumb.ResampleCubic
	case "linear":
		return thumb.ResampleLinear
	default:
		return thumb.ResampleCubic
	}
}

// GeoCodingApi returns the preferred geo coding api (none, osm or places).
func (c *Config) GeoCodingApi() string {
	switch c.config.GeoCodingApi {
	case "places":
		return "places"
	case "osm":
		return "osm"
	}
	return ""
}
