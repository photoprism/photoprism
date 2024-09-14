/*
Package config provides global options, command-line flags, and user settings.

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package config

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/klauspost/cpuid/v2"
	"github.com/pbnjay/memory"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/service/hub"
	"github.com/photoprism/photoprism/internal/service/hub/places"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/checksum"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// log points to the global logger.
var log = event.Log

// Config holds database, cache and all parameters of photoprism
type Config struct {
	once     sync.Once
	cliCtx   *cli.Context
	options  *Options
	settings *customize.Settings
	db       *gorm.DB
	hub      *hub.Config
	token    string
	serial   string
	env      string
	start    bool
}

func init() {
	TotalMem = memory.TotalMemory()

	// Check available memory if not running in unsafe mode.
	if Env(EnvUnsafe) {
		// Disable features with high memory requirements?
		LowMem = TotalMem < MinMem
	}

	// Init public thumb sizes for use in client apps.
	for i := len(thumb.Names) - 1; i >= 0; i-- {
		name := thumb.Names[i]
		t := thumb.Sizes[name]

		if t.Public {
			Thumbs = append(Thumbs, ThumbSize{Size: string(name), Usage: t.Usage, Width: t.Width, Height: t.Height})
		}
	}
}

func initLogger() {
	once.Do(func() {
		log.SetFormatter(&logrus.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})

		if Env(EnvProd) {
			log.SetLevel(logrus.WarnLevel)
		} else if Env(EnvTrace) {
			log.SetLevel(logrus.TraceLevel)
		} else if Env(EnvDebug) {
			log.SetLevel(logrus.DebugLevel)
		} else {
			log.SetLevel(logrus.InfoLevel)
		}
	})
}

// NewConfig initialises a new configuration file
func NewConfig(ctx *cli.Context) *Config {
	start := false

	if ctx != nil {
		start = ctx.Command.Name == "start"
	}

	// Initialize logger.
	initLogger()

	// Initialize options from the "defaults.yml" file and CLI context.
	c := &Config{
		cliCtx:  ctx,
		options: NewOptions(ctx),
		token:   rnd.Base36(8),
		env:     os.Getenv("DOCKER_ENV"),
		start:   start,
	}

	// Override options with values from the "options.yml" file, if it exists.
	if optionsYaml := c.OptionsYaml(); fs.FileExists(optionsYaml) {
		if err := c.options.Load(optionsYaml); err != nil {
			log.Warnf("config: failed loading values from %s (%s)", clean.Log(optionsYaml), err)
		} else {
			log.Debugf("config: overriding config with values from %s", clean.Log(optionsYaml))
		}
	}

	return c
}

// Init creates directories, parses additional config files, opens a database connection and initializes dependencies.
func (c *Config) Init() error {
	start := time.Now()

	// Fail if the originals and storage path are identical.
	if c.OriginalsPath() == c.StoragePath() {
		return fmt.Errorf("config: originals and storage folder must be different directories")
	}

	// Make sure that the configured storage directories exist and are properly configured.
	if err := c.CreateDirectories(); err != nil {
		return fmt.Errorf("config: %s", err)
	}

	// Initialize the storage path with a random serial.
	if err := c.InitSerial(); err != nil {
		return fmt.Errorf("config: %s", err)
	}

	// Detect whether files are stored on a case-insensitive file system.
	if insensitive, err := c.CaseInsensitive(); err != nil {
		return err
	} else if insensitive {
		log.Infof("config: case-insensitive file system detected")
		fs.IgnoreCase()
	}

	// Detect the CPU type and available memory.
	if cpuName := cpuid.CPU.BrandName; cpuName != "" {
		log.Debugf("config: running on %s, %s memory detected", clean.Log(cpuid.CPU.BrandName), humanize.Bytes(TotalMem))
	}

	// Fail if less than 128 MB of memory were detected.
	if TotalMem < 128*MegaByte {
		return fmt.Errorf("config: %s of memory detected, %d GB required", humanize.Bytes(TotalMem), MinMem/GigaByte)
	}

	// Show warning if less than 1 GB RAM was detected.
	if LowMem {
		log.Warnf(`config: less than %d GB of memory detected, please upgrade if server becomes unstable or unresponsive`, MinMem/GigaByte)
		log.Warnf("config: tensorflow as well as indexing and conversion of RAW images have been disabled automatically")
	}

	// Show swap space disclaimer.
	if TotalMem < RecommendedMem {
		log.Infof("config: make sure your server has enough swap configured to prevent restarts when there are memory usage spikes")
	}

	// Show wake-up interval warning if face recognition is activated and the worker runs less than once an hour.
	if !c.DisableFaces() && !c.Unsafe() && c.WakeupInterval() > time.Hour {
		log.Warnf("config: the wakeup interval is %s, but must be 1h or less for face recognition to work", c.WakeupInterval().String())
	}

	// Configure HTTPS proxy for outgoing connections.
	if httpsProxy := c.HttpsProxy(); httpsProxy != "" {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: c.HttpsProxyInsecure(),
		}

		_ = os.Setenv("HTTPS_PROXY", httpsProxy)
	}

	// Load settings from the "settings.yml" config file.
	c.initSettings()

	// Connect to database.
	if err := c.connectDb(); err != nil {
		return err
	} else {
		c.RegisterDb()
	}

	// Initialize extensions.
	Ext().Init(c)

	// Initialize the thumbnail generation package.
	thumb.Init(memory.FreeMemory(), c.IndexWorkers(), c.ThumbLibrary())

	// Update package defaults.
	c.Propagate()

	// Show support information.
	if !c.Sponsor() {
		log.Info(MsgSponsor)
		log.Info(MsgSignUp)
	}

	// Show log message.
	log.Debugf("config: successfully initialized [%s]", time.Since(start))

	return nil
}

// Propagate updates config options in other packages as needed.
func (c *Config) Propagate() {
	FlushCache()
	log.SetLevel(c.LogLevel())

	// Initialize the thumbnail generation package.
	thumb.Library = c.ThumbLibrary()
	thumb.Color = c.ThumbColor()
	thumb.Filter = c.ThumbFilter()
	thumb.SizeCached = c.ThumbSizePrecached()
	thumb.SizeOnDemand = c.ThumbSizeUncached()
	thumb.JpegQualityDefault = c.JpegQuality()
	thumb.CachePublic = c.HttpCachePublic()

	// Set cache expiration defaults.
	ttl.CacheDefault = c.HttpCacheMaxAge()
	ttl.CacheVideo = c.HttpVideoMaxAge()

	// Set geocoding parameters.
	places.UserAgent = c.UserAgent()
	entity.GeoApi = c.GeoApi()

	// Set session cache duration.
	entity.SessionCacheDuration = c.SessionCacheDuration()

	// Set minimum password length.
	entity.PasswordLength = c.PasswordLength()

	// Set path for user assets.
	entity.UsersPath = c.UsersPath()

	// Set API preview and download default tokens.
	entity.PreviewToken.Set(c.PreviewToken(), entity.TokenConfig)
	entity.DownloadToken.Set(c.DownloadToken(), entity.TokenConfig)
	entity.ValidateTokens = !c.Public()

	// Set face recognition parameters.
	face.ScoreThreshold = c.FaceScore()
	face.OverlapThreshold = c.FaceOverlap()
	face.ClusterScoreThreshold = c.FaceClusterScore()
	face.ClusterSizeThreshold = c.FaceClusterSize()
	face.ClusterCore = c.FaceClusterCore()
	face.ClusterDist = c.FaceClusterDist()
	face.MatchDist = c.FaceMatchDist()

	// Set default theme and locale.
	customize.DefaultTheme = c.DefaultTheme()
	customize.DefaultLocale = c.DefaultLocale()

	c.Settings().Propagate()
	c.Hub().Propagate()
}

// Options returns the raw config options.
func (c *Config) Options() *Options {
	if c.options == nil {
		log.Warnf("config: options should not be nil - you may have found a bug")
		c.options = NewOptions(nil)
	}

	return c.options
}

// Unsafe checks if unsafe settings are allowed.
func (c *Config) Unsafe() bool {
	return c.options.Unsafe
}

// Restart checks if the application should be restarted, e.g. after an update or a config changes.
func (c *Config) Restart() bool {
	return mutex.Restart.Load()
}

// CliContext returns the cli context if set.
func (c *Config) CliContext() *cli.Context {
	if c.cliCtx == nil {
		log.Warnf("config: cli context not set - you may have found a bug")
	}

	return c.cliCtx
}

// CliGlobalString returns a global cli string flag value if set.
func (c *Config) CliGlobalString(name string) string {
	if c.cliCtx == nil {
		return ""
	}

	return c.cliCtx.GlobalString(name)
}

// readSerial reads and returns the current storage serial.
func (c *Config) readSerial() string {
	storageName := filepath.Join(c.StoragePath(), serialName)
	backupName := c.BackupPath(serialName)

	if fs.FileExists(storageName) {
		if data, err := os.ReadFile(storageName); err == nil && len(data) == 16 {
			return string(data)
		} else {
			log.Tracef("config: could not read %s (%s)", clean.Log(storageName), err)
		}
	}

	if fs.FileExists(backupName) {
		if data, err := os.ReadFile(backupName); err == nil && len(data) == 16 {
			return string(data)
		} else {
			log.Tracef("config: could not read %s (%s)", clean.Log(backupName), err)
		}
	}

	return ""
}

// InitSerial initializes storage directories with a random serial.
func (c *Config) InitSerial() (err error) {
	if c.Serial() != "" {
		return nil
	}

	c.serial = rnd.GenerateUID('z')

	storageName := filepath.Join(c.StoragePath(), serialName)
	backupName := c.BackupPath(serialName)

	if err = os.WriteFile(storageName, []byte(c.serial), fs.ModeFile); err != nil {
		return fmt.Errorf("could not create %s: %s", storageName, err)
	}

	if err = os.WriteFile(backupName, []byte(c.serial), fs.ModeFile); err != nil {
		return fmt.Errorf("could not create %s: %s", backupName, err)
	}

	return nil
}

// Serial returns the random storage serial.
func (c *Config) Serial() string {
	if c.serial == "" {
		c.serial = c.readSerial()
	}

	return c.serial
}

// SerialChecksum returns the CRC32 checksum of the storage serial.
func (c *Config) SerialChecksum() string {
	return checksum.Serial([]byte(c.Serial()))
}

// Name returns the app name.
func (c *Config) Name() string {
	if c.options.Name == "" {
		return "PhotoPrism"
	}

	return c.options.Name
}

// About returns the app about string.
func (c *Config) About() string {
	if c.options.About == "" {
		return "PhotoPrism®"
	}

	return c.options.About
}

// Edition returns the edition nane.
func (c *Config) Edition() string {
	if c.options.Edition == "" {
		return "ce"
	}

	return c.options.Edition
}

// Version returns the application version.
func (c *Config) Version() string {
	return c.options.Version
}

// VersionChecksum returns the application version checksum.
func (c *Config) VersionChecksum() uint32 {
	return checksum.Crc32([]byte(c.Version()))
}

// UserAgent returns an HTTP user agent string based on the app config and version.
func (c *Config) UserAgent() string {
	return fmt.Sprintf("%s/%s (%s)", c.Name(), c.Version(), strings.Join(append(c.Flags(), c.Serial()), "; "))
}

// Copyright returns the application copyright.
func (c *Config) Copyright() string {
	return c.options.Copyright
}

// Prod checks if production mode is enabled, hides non-essential log messages.
func (c *Config) Prod() bool {
	return c.options.Prod
}

// Debug checks if debug mode is enabled, shows non-essential log messages.
func (c *Config) Debug() bool {
	if c.Prod() {
		return false
	} else if c.Trace() {
		return true
	}

	return c.options.Debug
}

// Trace checks if trace mode is enabled, shows all log messages.
func (c *Config) Trace() bool {
	if c.Prod() {
		return false
	}

	return c.options.Trace || c.options.LogLevel == logrus.TraceLevel.String()
}

// Test checks if test mode is enabled.
func (c *Config) Test() bool {
	return c.options.Test
}

// Demo checks if demo mode is enabled.
func (c *Config) Demo() bool {
	return c.options.Demo
}

// Sponsor reports if you have chosen to support our mission.
func (c *Config) Sponsor() bool {
	if Sponsor || c.options.Sponsor {
		return true
	} else if c.hub != nil {
		Sponsor = c.Hub().Sponsor()
	}

	return Sponsor
}

// Experimental checks if experimental features should be enabled.
func (c *Config) Experimental() bool {
	return c.options.Experimental
}

// ReadOnly checks if photo directories are write protected.
func (c *Config) ReadOnly() bool {
	return c.options.ReadOnly
}

// LogLevel returns the Logrus log level.
func (c *Config) LogLevel() logrus.Level {
	// Normalize string.
	c.options.LogLevel = strings.ToLower(strings.TrimSpace(c.options.LogLevel))

	if c.Trace() {
		c.options.LogLevel = logrus.TraceLevel.String()
	} else if c.Debug() && c.options.LogLevel != logrus.TraceLevel.String() {
		c.options.LogLevel = logrus.DebugLevel.String()
	}

	if logLevel, err := logrus.ParseLevel(c.options.LogLevel); err == nil {
		return logLevel
	} else {
		return logrus.InfoLevel
	}
}

// SetLogLevel sets the Logrus log level.
func (c *Config) SetLogLevel(level logrus.Level) {
	log.SetLevel(level)
}

// Shutdown shuts down the active processes and closes the database connection.
func (c *Config) Shutdown() {
	// Send cancel signal to all workers.
	mutex.CancelAll()

	// Shutdown thumbnail library.
	thumb.Shutdown()

	// Close database connection.
	if err := c.CloseDb(); err != nil {
		log.Errorf("could not close database connection: %s", err)
	} else {
		log.Debug("closed database connection")
	}
}

// IndexWorkers returns the number of indexing workers.
func (c *Config) IndexWorkers() int {
	// Use one worker on systems with less than the recommended amount of memory.
	if TotalMem < RecommendedMem {
		return 1
	}

	// NumCPU returns the number of logical CPU cores.
	cores := runtime.NumCPU()

	// Limit to physical cores to avoid high load on HT capable CPUs.
	if cores > cpuid.CPU.PhysicalCores {
		cores = cpuid.CPU.PhysicalCores
	}

	// Limit number of workers when using SQLite3 to avoid database locking issues.
	if c.DatabaseDriver() == SQLite3 && (cores >= 8 && c.options.IndexWorkers <= 0 || c.options.IndexWorkers > 4) {
		return 4
	}

	// Return explicit value if set and not too large.
	if c.options.IndexWorkers > runtime.NumCPU() {
		return runtime.NumCPU()
	} else if c.options.IndexWorkers > 0 {
		return c.options.IndexWorkers
	}

	// Use half the available cores by default.
	if cores > 1 {
		return cores / 2
	}

	return 1
}

// IndexSchedule returns the indexing schedule in cron format, e.g. "0 */3 * * *" to start indexing every 3 hours.
func (c *Config) IndexSchedule() string {
	return Schedule(c.options.IndexSchedule)
}

// WakeupInterval returns the duration between background worker runs
// required for face recognition and index maintenance(1-86400s).
func (c *Config) WakeupInterval() time.Duration {
	if c.options.WakeupInterval <= 0 {
		if c.Unsafe() {
			// Worker can be disabled only in unsafe mode.
			return time.Duration(0)
		} else {
			// Default to 15 minutes if no interval is set.
			return DefaultWakeupInterval
		}
	}

	// Do not run more than once per minute.
	if c.options.WakeupInterval < MinWakeupInterval/time.Second {
		return MinWakeupInterval
	} else if c.options.WakeupInterval < MinWakeupInterval {
		c.options.WakeupInterval = c.options.WakeupInterval * time.Second
	}

	// Do not run less than once per day.
	if c.options.WakeupInterval > MaxWakeupInterval {
		return MaxWakeupInterval
	}

	return c.options.WakeupInterval
}

// AutoIndex returns the auto index delay duration.
func (c *Config) AutoIndex() time.Duration {
	if c.options.AutoIndex < 0 {
		return -1 * time.Second
	} else if c.options.AutoIndex == 0 || c.options.AutoIndex > 604800 {
		return DefaultAutoIndexDelay * time.Second
	}

	return time.Duration(c.options.AutoIndex) * time.Second
}

// AutoImport returns the auto import delay duration.
func (c *Config) AutoImport() time.Duration {
	if c.options.AutoImport < 0 || c.ReadOnly() {
		return -1 * time.Second
	} else if c.options.AutoImport == 0 || c.options.AutoImport > 604800 {
		return DefaultAutoImportDelay * time.Second
	}

	return time.Duration(c.options.AutoImport) * time.Second
}

// GeoApi returns the preferred geocoding api (places, or none).
func (c *Config) GeoApi() string {
	if c.options.DisablePlaces {
		return ""
	}

	return "places"
}

// OriginalsLimit returns the maximum size of originals in MB.
func (c *Config) OriginalsLimit() int {
	if c.options.OriginalsLimit <= 0 || c.options.OriginalsLimit > 100000 {
		return -1
	}

	return c.options.OriginalsLimit
}

// OriginalsByteLimit returns the maximum size of originals in bytes.
func (c *Config) OriginalsByteLimit() int64 {
	if result := c.OriginalsLimit(); result <= 0 {
		return -1
	} else {
		return int64(result) * 1024 * 1024
	}
}

// ResolutionLimit returns the maximum resolution of originals in megapixels (width x height).
func (c *Config) ResolutionLimit() int {
	result := c.options.ResolutionLimit

	// Disabling or increasing the limit is at your own risk.
	// Only sponsors receive support in case of problems.
	if result == 0 {
		return DefaultResolutionLimit
	} else if result < 0 {
		return -1
	} else if result > 900 {
		result = 900
	}

	return result
}

// RenewApiKeys renews the api credentials for maps and places.
func (c *Config) RenewApiKeys() {
	if c.hub == nil {
		return
	}

	if token := os.Getenv(EnvVar("CONNECT")); token != "" && !c.Hub().Sponsor() {
		_ = c.RenewApiKeysWithToken(token)
	} else {
		_ = c.RenewApiKeysWithToken("")
	}
}

// RenewApiKeysWithToken renews the api credentials for maps and places with an activation token.
func (c *Config) RenewApiKeysWithToken(token string) error {
	if c.hub == nil {
		return fmt.Errorf("hub is not initialized")
	}

	if err := c.hub.ReSync(token); err != nil {
		log.Debugf("config: %s, see https://docs.photoprism.app/getting-started/troubleshooting/firewall/", err)
		if token != "" {
			return i18n.Error(i18n.ErrAccountConnect)
		}
	} else if err = c.hub.Save(); err != nil {
		log.Warnf("config: failed to save api keys for maps and places (%s)", err)
		return i18n.Error(i18n.ErrSaveFailed)
	} else {
		c.hub.Propagate()
	}

	return nil
}

// initHub initializes PhotoPrism hub config.
func (c *Config) initHub() {
	if c.hub != nil {
		return
	} else if h := hub.NewConfig(c.Version(), c.HubConfigFile(), c.serial, c.env, c.UserAgent(), c.options.PartnerID); h != nil {
		c.hub = h
	}

	update := c.start

	if err := c.hub.Load(); err != nil {
		update = true
	}

	if update {
		c.RenewApiKeys()
	}

	c.hub.Propagate()

	ticker := time.NewTicker(time.Hour * 24)

	go func() {
		for {
			select {
			case <-ticker.C:
				c.RenewApiKeys()
			}
		}
	}()
}

// Hub returns the PhotoPrism hub config.
func (c *Config) Hub() *hub.Config {
	c.initHub()

	return c.hub
}
