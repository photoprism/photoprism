package config

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var log = event.Log
var once sync.Once

// Config holds database, cache and all parameters of photoprism
type Config struct {
	once     sync.Once
	db       *gorm.DB
	params   *Params
	settings *Settings
	token    string
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
		params: NewParams(ctx),
		token:  rnd.Token(8),
	}

	c.initSettings()

	return c
}

// Propagate updates config values in other packages as needed.
func (c *Config) Propagate() {
	log.SetLevel(c.LogLevel())

	thumb.Size = c.ThumbSize()
	thumb.Limit = c.ThumbLimit()
	thumb.Filter = c.ThumbFilter()
	thumb.JpegQuality = c.JpegQuality()

	c.Settings().Propagate()
}

// Init initialises the database connection and dependencies.
func (c *Config) Init(ctx context.Context) error {
	c.Propagate()
	return c.connectToDatabase(ctx)
}

// Name returns the application name.
func (c *Config) Name() string {
	return c.params.Name
}

// Url returns the public server URL (default is "http://localhost:2342/").
func (c *Config) Url() string {
	if c.params.Url == "" {
		return "http://localhost:2342/"
	}

	return c.params.Url
}

// Title returns the site title (default is application name).
func (c *Config) Title() string {
	if c.params.Title == "" {
		return c.Name()
	}

	return c.params.Title
}

// Subtitle returns the site title.
func (c *Config) Subtitle() string {
	return c.params.Subtitle
}

// Description returns the site title.
func (c *Config) Description() string {
	return c.params.Description
}

// Author returns the site author / copyright.
func (c *Config) Author() string {
	return c.params.Author
}

// Version returns the application version.
func (c *Config) Version() string {
	return c.params.Version
}

// Copyright returns the application copyright.
func (c *Config) Copyright() string {
	return c.params.Copyright
}

// Debug returns true if Debug mode is on.
func (c *Config) Debug() bool {
	return c.params.Debug
}

// Public returns true if app requires no authentication.
func (c *Config) Public() bool {
	return c.params.Public
}

// Experimental returns true if experimental features should be enabled.
func (c *Config) Experimental() bool {
	return c.params.Experimental
}

// ReadOnly returns true if photo directories are write protected.
func (c *Config) ReadOnly() bool {
	return c.params.ReadOnly
}

// DetectNSFW returns true if NSFW photos should be detected and flagged.
func (c *Config) DetectNSFW() bool {
	return c.params.DetectNSFW
}

// UploadNSFW returns true if NSFW photos can be uploaded.
func (c *Config) UploadNSFW() bool {
	return c.params.UploadNSFW
}

// AdminPassword returns the admin password.
func (c *Config) AdminPassword() string {
	if c.params.AdminPassword == "" {
		return "photoprism"
	}

	return c.params.AdminPassword
}

// WebDAVPassword returns the WebDAV password for remote access.
func (c *Config) WebDAVPassword() string {
	return c.params.WebDAVPassword
}

// LogLevel returns the logrus log level.
func (c *Config) LogLevel() logrus.Level {
	if c.Debug() {
		c.params.LogLevel = "debug"
	}

	if logLevel, err := logrus.ParseLevel(c.params.LogLevel); err == nil {
		return logLevel
	} else {
		return logrus.InfoLevel
	}
}

// Shutdown services and workers.
func (c *Config) Shutdown() {
	mutex.MainWorker.Cancel()
	mutex.ShareWorker.Cancel()
	mutex.SyncWorker.Cancel()
	mutex.PrismWorker.Cancel()

	if err := c.CloseDb(); err != nil {
		log.Errorf("could not close database connection: %s", err)
	} else {
		log.Info("closed database connection")
	}
}

// Workers returns the number of workers e.g. for indexing files.
func (c *Config) Workers() int {
	numCPU := runtime.NumCPU()

	if c.params.Workers > 0 && c.params.Workers <= numCPU {
		return c.params.Workers
	}

	if numCPU > 1 {
		return numCPU - 1
	}

	return 1
}

// WakeupInterval returns the background worker wakeup interval.
func (c *Config) WakeupInterval() time.Duration {
	if c.params.WakeupInterval <= 0 {
		return 5 * time.Minute
	}

	return time.Duration(c.params.WakeupInterval) * time.Second
}

// GeoCodingApi returns the preferred geo coding api (none, osm or places).
func (c *Config) GeoCodingApi() string {
	switch c.params.GeoCodingApi {
	case "places":
		return "places"
	case "osm":
		return "osm"
	}
	return ""
}

// OriginalsLimit returns the file size limit for originals.
func (c *Config) OriginalsLimit() int64 {
	if c.params.OriginalsLimit <= 0 || c.params.OriginalsLimit > 100000 {
		return -1
	}

	// Megabyte.
	return c.params.OriginalsLimit * 1024 * 1024
}
