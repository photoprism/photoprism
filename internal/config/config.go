package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/photoprism/photoprism/internal/entity"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/hub"
	"github.com/photoprism/photoprism/internal/hub/places"
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
	hub      *hub.Config
	token    string
	serial   string
}

func init() {
	// Init public thumb sizes for use in client apps.
	for i := len(thumb.DefaultTypes) - 1; i >= 0; i-- {
		size := thumb.DefaultTypes[i]
		t := thumb.Types[size]

		if t.Public {
			Thumbs = append(Thumbs, Thumb{Size: size, Use: t.Use, Width: t.Width, Height: t.Height})
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

	return c
}

// Propagate updates config values in other packages as needed.
func (c *Config) Propagate() {
	log.SetLevel(c.LogLevel())

	thumb.Size = c.ThumbSize()
	thumb.SizeUncached = c.ThumbSizeUncached()
	thumb.Filter = c.ThumbFilter()
	thumb.JpegQuality = c.JpegQuality()
	places.UserAgent = c.UserAgent()
	entity.GeoApi = c.GeoApi()

	c.Settings().Propagate()
	c.Hub().Propagate()
}

// Init creates directories, parses additional config files, opens a database connection and initializes dependencies.
func (c *Config) Init() error {
	if err := c.CreateDirectories(); err != nil {
		return err
	}

	if err := c.initStorage(); err != nil {
		return err
	}

	c.initSettings()
	c.initHub()

	c.Propagate()

	return c.connectDb()
}

// initStorage initializes storage directories with a random serial.
func (c *Config) initStorage() error {
	const serialName = "serial"

	c.serial = rnd.PPID('z')

	storageName := filepath.Join(c.StoragePath(), serialName)
	backupName := filepath.Join(c.BackupPath(), serialName)

	if data, err := ioutil.ReadFile(storageName); err == nil {
		c.serial = string(data)
	} else if data, err := ioutil.ReadFile(backupName); err == nil {
		c.serial = string(data)
	} else if err := ioutil.WriteFile(storageName, []byte(c.serial), os.ModePerm); err != nil {
		return fmt.Errorf("failed creating %s: %s", storageName, err)
	} else if err := ioutil.WriteFile(backupName, []byte(c.serial), os.ModePerm); err != nil {
		return fmt.Errorf("failed creating %s: %s", backupName, err)
	}

	return nil
}

// Name returns the application name ("PhotoPrism").
func (c *Config) Name() string {
	return c.params.Name
}

// Version returns the application version.
func (c *Config) Version() string {
	return c.params.Version
}

// UserAgent returns a HTTP user agent string based on app name & version.
func (c *Config) UserAgent() string {
	return fmt.Sprintf("%s/%s", c.Name(), c.Version())
}

// Copyright returns the application copyright.
func (c *Config) Copyright() string {
	return c.params.Copyright
}

// SiteUrl returns the public server URL (default is "http://localhost:2342/").
func (c *Config) SiteUrl() string {
	if c.params.SiteUrl == "" {
		return "http://localhost:2342/"
	}

	return c.params.SiteUrl
}

// SitePreview returns the site preview image URL for sharing.
func (c *Config) SitePreview() string {
	if c.params.SitePreview == "" {
		return c.SiteUrl() + "static/img/preview.jpg"
	}

	if !strings.HasPrefix(c.params.SitePreview, "http") {
		return c.SiteUrl() + c.params.SitePreview
	}

	return c.params.SitePreview
}

// SiteTitle returns the main site title (default is application name).
func (c *Config) SiteTitle() string {
	if c.params.SiteTitle == "" {
		return c.Name()
	}

	return c.params.SiteTitle
}

// SiteCaption returns a short site caption.
func (c *Config) SiteCaption() string {
	return c.params.SiteCaption
}

// SiteDescription returns a long site description.
func (c *Config) SiteDescription() string {
	return c.params.SiteDescription
}

// SiteAuthor returns the site author / copyright.
func (c *Config) SiteAuthor() string {
	return c.params.SiteAuthor
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

// AdminPassword returns the initial admin password.
func (c *Config) AdminPassword() string {
	return c.params.AdminPassword
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
	mutex.MetaWorker.Cancel()

	if err := c.CloseDb(); err != nil {
		log.Errorf("could not close database connection: %s", err)
	} else {
		log.Info("closed database connection")
	}
}

// Workers returns the number of workers e.g. for indexing files.
func (c *Config) Workers() int {
	numCPU := runtime.NumCPU()

	// Limit number of workers when using SQLite to avoid database locking issues.
	if c.DatabaseDriver() == SQLite && numCPU > 4 && c.params.Workers <= 0 {
		return 4
	}

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
		return 15 * time.Minute
	}

	return time.Duration(c.params.WakeupInterval) * time.Second
}

// GeoApi returns the preferred geo coding api (none, osm or places).
func (c *Config) GeoApi() string {
	switch c.params.GeoApi {
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

// UpdateHub updates backend api credentials for maps & places.
func (c *Config) UpdateHub() {
	if err := c.hub.Refresh(); err != nil {
		log.Debugf("config: %s", err)
	} else if err := c.hub.Save(); err != nil {
		log.Debugf("config: %s", err)
	} else {
		c.hub.Propagate()
	}
}

// initHub initializes PhotoPrism hub config.
func (c *Config) initHub() {
	c.hub = hub.NewConfig(c.Version(), c.HubConfigFile(), c.serial)

	if err := c.hub.Load(); err == nil {
		// Do nothing.
	} else if err := c.hub.Refresh(); err != nil {
		log.Debugf("config: %s", err)
	} else if err := c.hub.Save(); err != nil {
		log.Debugf("config: %s", err)
	}

	c.hub.Propagate()

	ticker := time.NewTicker(time.Hour * 24)

	go func() {
		for {
			select {
			case <-ticker.C:
				c.UpdateHub()
			}
		}
	}()
}

// Hub returns the PhotoPrism hub config.
func (c *Config) Hub() *hub.Config {
	if c.hub == nil {
		c.initHub()
	}

	return c.hub
}
