/*
Package config provides global options, command-line flags, and user settings.

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

	This program is free software: you can redistribute it and/or modify
	it under Version 3 of the GNU Affero General Public License (the "AGPL"):
	<https://docs.photoprism.app/license/agpl>

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	The AGPL is supplemented by our Trademark and Brand Guidelines,
	which describe how our Brand Assets may be used:
	<https://www.photoprism.app/trademark/>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>
*/
package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/dustin/go-humanize"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"  // register mysql dialect
	_ "github.com/jinzhu/gorm/dialects/sqlite" // register sqlite dialect
	"github.com/klauspost/cpuid/v2"
	gc "github.com/patrickmn/go-cache"
	"github.com/pbnjay/memory"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/api/download"
	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism/dl"
	"github.com/photoprism/photoprism/internal/service/hub"
	"github.com/photoprism/photoprism/internal/service/hub/places"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/checksum"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Config aggregates CLI flags, options.yml overrides, runtime settings, and shared resources (database, caches) for the running instance.
type Config struct {
	cliCtx        *cli.Context
	options       *Options
	settings      *customize.Settings
	db            *gorm.DB
	dbVersion     string
	hub           *hub.Config
	hubCancel     context.CancelFunc
	hubLock       sync.Mutex
	faceWarned    sync.Map
	faceModel     string
	faceModelFlag string
	token         string
	serial        string
	tokenKey      []byte
	tokenKeyOnce  sync.Once
	env           string
	start         bool
	ready         atomic.Bool
	cache         *gc.Cache
}

// Values is a shorthand alias for map[string]interface{}.
type Values = map[string]any

func init() {
	TotalMem = memory.TotalMemory()

	// Check available memory if not running in unsafe mode.
	if Env(EnvUnsafe) {
		// Disable features with high memory requirements?
		LowMem = TotalMem < MinMem
	}

	// Disable entity cache if requested.
	if txt.Bool(os.Getenv(EnvVar("disable-photolabelcache"))) {
		entity.UsePhotoLabelsCache = false
	}

	initThumbs()
}

func initThumbs() {
	initThumbsMutex.Lock()
	defer initThumbsMutex.Unlock()

	maxSize := thumb.MaxSize()
	Thumbs = ThumbSizes{}

	// Init public thumb sizes for use in client apps.
	for i := len(thumb.Names) - 1; i >= 0; i-- {
		name := thumb.Names[i]
		t := thumb.Sizes[name]

		if t.Width > maxSize {
			continue
		}

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

		switch {
		case Env(EnvProd):
			SetLogLevel(logrus.WarnLevel)
		case Env(EnvTrace):
			SetLogLevel(logrus.TraceLevel)
		case Env(EnvDebug):
			SetLogLevel(logrus.DebugLevel)
		default:
			SetLogLevel(logrus.InfoLevel)
		}
	})
}

// NewConfig builds a Config from CLI context defaults and loads options.yml overrides if present.
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
		cache:   gc.New(time.Minute, 10*time.Minute),
	}

	// Keep what the environment or the command line asked for, since "options.yml" is loaded
	// last and a model it names is what an instance must keep using - see initFaceModel.
	c.faceModelFlag = c.options.FaceModel

	// Override options with values from the "options.yml" file, if it exists.
	if optionsYaml := c.OptionsYaml(); fs.FileExists(optionsYaml) {
		if err := c.options.Load(optionsYaml); err != nil {
			log.Warnf("config: failed loading values from %s (%s)", clean.Log(optionsYaml), err)
		} else if c.env == EnvDevelop {
			// Reduce the log level to minimize noise in the test logs.
			log.Tracef("config: overriding config with values from %s", clean.Log(optionsYaml))
		} else {
			log.Debugf("config: overriding config with values from %s", clean.Log(optionsYaml))
		}
	}

	return c
}

// Init creates directories, parses additional config files, opens the database connection, and initializes dependent subsystems.
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
			InsecureSkipVerify: c.HttpsProxyInsecure(), //nolint:gosec // proxy settings are user-configurable and opt-in
		}

		_ = os.Setenv("HTTPS_PROXY", httpsProxy)
	}

	// Load settings from the "settings.yml" config file.
	c.initSettings()

	// Initialize boot extensions before connecting to the database so they can
	// influence DB settings (e.g., cluster bootstrap providing MariaDB creds).
	Ext(StageBoot).Boot(c)

	// Connect to database.
	if err := c.connectDb(); err != nil {
		return err
	} else {
		c.RegisterDb()
	}

	// Initialize regular extensions.
	Ext(StageInit).Init(c)

	// Initialize thumbnail package.
	thumb.Init(memory.FreeMemory(), c.IndexWorkers(), c.ThumbLibrary())

	// Set minimum free storage space in percent.
	disk.StorageLowPct = c.StorageFree()
	DisableStorageCheck.Store(disk.StorageLowPct <= 0)

	c.LoadVisionConfig()

	// Settle which face embedding model this instance uses, which needs the database and has
	// to happen before Propagate configures the embedder from it.
	c.initFaceModel()

	// Update package defaults.
	c.Propagate()

	// Report the download token configuration.
	c.reportDownloadTokenOptions()

	// Show support information.
	if !c.Sponsor() {
		log.Info(MsgSponsor)
		log.Info(MsgSignUp)
	}

	// Show log message.
	log.Debugf("config: successfully initialized [%s]", time.Since(start))
	c.ready.Store(true)

	return nil
}

// reportDownloadTokenOptions reports the download token configuration. Both notices describe static
// option values, so Init calls this once at startup rather than Propagate, which runs again whenever an
// admin saves Advanced Settings.
func (c *Config) reportDownloadTokenOptions() {
	// A static token is an explicit opt-in whose trade-off is not visible from the URLs it produces: it
	// keeps permanent links working, but anyone holding it can download public content. Sessions are
	// unaffected, as they receive signed tokens.
	if !c.Public() && c.options.DownloadToken != "" {
		event.SystemWarn([]string{"config", "download-token", "static value configured, so it grants downloads of public content without identifying a session"})
	}

	if raw := c.options.DownloadTokenMaxAge; raw > 0 && raw < int64(ttl.DownloadTokenMinAge) {
		event.SystemWarn([]string{"config", "download-token-maxage", "%ds is below the %ds minimum and has been raised to it"}, raw, int64(ttl.DownloadTokenMinAge))
	}
}

// InitCore initializes configuration values without connecting to the database
// or running cluster bootstrap tasks.
func (c *Config) InitCore() error {
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

	// Load settings from the "settings.yml" config file.
	c.initSettings()

	// Initialize extensions that do not require database access.
	Ext(StageInit).Init(c)

	// Show log message.
	log.Debugf("config: successfully initialized [%s]", time.Since(start))
	c.ready.Store(true)

	return nil
}

// IsReady checks if the application has been successfully initialized.
func (c *Config) IsReady() bool {
	return c.ready.Load()
}

// Propagate updates config options in other packages as needed.
// It assigns package-level values without synchronization, which is safe because it runs at startup and
// otherwise only when an admin changes Advanced Settings — a rare action after which both call sites set
// mutex.Restart, as a restart is required for every change to take effect.
func (c *Config) Propagate() {
	FlushCache()
	log.SetLevel(c.LogLevel())

	// Configure thumbnail package.
	thumb.Library = c.ThumbLibrary()
	thumb.Color = c.ThumbColor()
	thumb.Filter = c.ThumbFilter()
	thumb.SizeCached = c.ThumbSizePrecached()
	thumb.SizeOnDemand = c.ThumbSizeUncached()
	thumb.JpegQualityDefault = c.JpegQuality()
	thumb.CachePublic = c.HttpCachePublic()
	thumb.SamplesPath = c.SamplesPath()
	thumb.IccProfilesPath = c.IccProfilesPath()
	initThumbs()

	// Configure FFmpeg package.
	ffmpeg.SetExclude(c.FFmpegExclude())

	// Configure video download package.
	dl.YtDlpBin = c.YtDlpBin()
	dl.FFmpegBin = c.FFmpegBin()
	dl.FFprobeBin = c.FFprobeBin()

	// Configure computer vision package.
	vision.SetCachePath(c.CachePath())
	vision.SetModelsPath(c.ModelsPath())
	vision.ServiceApi = c.VisionApi()
	vision.ServiceUri = c.VisionUri()
	vision.ServiceKey = c.VisionKey()
	vision.DownloadUrl = c.DownloadUrl()
	vision.DetectNSFWLabels = c.DetectNSFW() && c.Experimental()

	// Set allowed path in download package.
	download.AllowedPaths = []string{
		c.SidecarPath(),
		c.OriginalsPath(),
		c.ThumbCachePath(),
	}

	// Set cache expiration defaults, including the signed download token lifetime read when one is minted.
	ttl.CacheDefault = c.HttpCacheMaxAge()
	ttl.CacheVideo = c.HttpVideoMaxAge()
	ttl.DownloadToken = ttl.Duration(int(c.DownloadTokenMaxAge().Seconds()))

	// Configure signed URL tokens, which must be complete before anything mints or verifies one. A single
	// signing key covers every token kind (downloads today, previews next); the signature path is per kind.
	tokens.Download.Key = c.TokenSigningKey()
	tokens.Download.SignaturePath = c.ApiUri()

	// Configure download-token delivery: public mode delivers a placeholder, every session receives a
	// signed token, and the coarse token covers only the sessionless client configs.
	tokens.PublicMode = c.Public()
	tokens.CoarseDownload = c.DownloadToken()

	// Set geocoding parameters.
	places.UserAgent = c.UserAgent()
	places.DefaultLocale = c.PlacesLocale()
	entity.GeoApi = c.GeoApi()

	// Set session cache duration.
	entity.SessionCacheDuration = c.SessionCacheDuration()

	// Set minimum password length.
	entity.PasswordLength = c.PasswordLength()

	// Set path for user assets.
	entity.UsersPath = c.UsersPath()

	// Set the API preview default token (the download token is no longer stored per session). The
	// placeholder is never registered, so a missing signing key rejects previews instead of admitting it.
	if previewToken := c.PreviewToken(); previewToken != PreviewTokenPlaceholder {
		entity.PreviewToken.Set(entity.TokenConfig, previewToken)
	}

	entity.ValidateTokens = !c.Public()

	// Set face recognition parameters.
	face.SizeThreshold = c.FaceSize()
	face.ScoreThreshold = c.FaceScore()
	face.OverlapThreshold = c.FaceOverlap()
	face.ClusterScoreThreshold = c.FaceClusterScore()
	face.ClusterSizeThreshold = c.FaceClusterSize()
	face.ClusterCore = c.FaceClusterCore()
	// Derived rather than configured, but it still has to follow FACE_CLUSTER_CORE: leaving it at
	// the package initializer froze the clustering trigger at the shipped default, so raising the
	// core size moved the cluster definition and not the number of markers that starts a pass.
	face.SampleThreshold = c.FaceSampleThreshold()
	face.CollisionDist = c.FaceCollisionDist()
	face.Epsilon = c.FaceEpsilonDist()
	face.ClusterRadius = c.FaceClusterRadius()
	face.ClusterDist = c.FaceClusterDist()
	face.MatchDist = c.FaceMatchDist()
	if err := c.ConfigureFaceDetector(0); err != nil {
		log.Warnf("faces: %s (configure engine)", err)
	}
	if err := c.ConfigureFaceEmbedder(c.FaceModel()); err != nil {
		log.Warnf("faces: %s (configure embedding model)", err)
	}

	// Set default theme and locale.
	customize.DefaultTheme = c.DefaultTheme()
	customize.DefaultLanguage = c.DefaultLocale()
	customize.DefaultTimeZone = c.DefaultTimezone().String()

	// Propagate settings.
	c.Settings().Propagate()

	// Set default album sort orders.
	if c.settings.Albums.Order.Album != "" {
		entity.DefaultOrderAlbum = c.settings.Albums.Order.Album
	}

	if c.settings.Albums.Order.Folder != "" {
		entity.DefaultOrderFolder = c.settings.Albums.Order.Folder
	}

	if c.settings.Albums.Order.Moment != "" {
		entity.DefaultOrderMoment = c.settings.Albums.Order.Moment
	}

	if c.settings.Albums.Order.State != "" {
		entity.DefaultOrderState = c.settings.Albums.Order.State
	}

	if c.settings.Albums.Order.Month != "" {
		entity.DefaultOrderMonth = c.settings.Albums.Order.Month
	}

	c.Hub().Propagate()
}

// Options returns the raw config options.
func (c *Config) Options() *Options {
	if c.options == nil {
		log.Warnf("config: options must not be nil - you may have found a bug")
		c.options = NewOptions(nil)
	}

	return c.options
}

// SaveOptionsPatch merges a patch into options.yml, reloads in-memory options,
// and returns true when persisted values changed.
func (c *Config) SaveOptionsPatch(patch Values) (bool, error) {
	if c == nil || c.options == nil || len(patch) == 0 {
		return false, nil
	}

	fileName, values, err := c.loadOptionsYAML()
	if err != nil {
		return false, err
	}

	if !mergeOptionValues(values, patch) {
		return false, nil
	}

	if _, err = c.writeOptionsYAML(fileName, values); err != nil {
		return true, err
	}

	return true, c.applyOptionValues(patch)
}

// DeleteOptionsPatch removes the specified keys from "options.yml" and reports whether the file
// was changed. Removing a key restores the default, which writing an empty value does not: the
// loader cannot tell an option that was cleared from one that was set to nothing.
func (c *Config) DeleteOptionsPatch(keys ...string) (bool, error) {
	if c == nil || c.options == nil || len(keys) == 0 {
		return false, nil
	}

	// Nothing to remove from a file that does not exist, and reading one through loadOptionsYAML
	// would create its directory on the way. A helper that removes a setting must not leave a
	// directory tree behind as its only effect.
	if fileName := c.OptionsYaml(); fileName == "" || !fs.FileExists(fileName) {
		return false, nil
	}

	fileName, values, err := c.loadOptionsYAML()
	if err != nil {
		return false, err
	}

	changed := false

	for _, key := range keys {
		if _, ok := values[key]; ok {
			delete(values, key)
			changed = true
		}
	}

	if !changed {
		return false, nil
	}

	// Nothing is applied in return: a removed key leaves no value to read back, and the caller
	// clears its own field.
	return c.writeOptionsYAML(fileName, values)
}

// applyOptionValues applies the patched values, and only those, to the in-memory options.
//
// Reading the file back instead would apply every key it holds - including ones a flag or an
// environment variable overrode for this run, and ones another writer left there. That is how
// recording a face model came to replace a live database configuration.
func (c *Config) applyOptionValues(patch Values) error {
	if c == nil || c.options == nil || len(patch) == 0 {
		return nil
	}

	b, err := yaml.Marshal(patch)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, c.options)
}

// loadOptionsYAML loads options.yml into a writable map and returns its file path.
func (c *Config) loadOptionsYAML() (string, Values, error) {
	fileName := c.OptionsYaml()
	if fileName == "" {
		return "", nil, fmt.Errorf("invalid options.yml filename")
	}

	if err := fs.MkdirAll(filepath.Dir(fileName)); err != nil {
		return fileName, nil, err
	}

	values := Values{}

	if !fs.FileExists(fileName) {
		return fileName, values, nil
	}

	b, err := os.ReadFile(fileName) //nolint:gosec // path derived from config directory
	if err != nil || len(b) == 0 {
		return fileName, values, err
	}

	if err = yaml.Unmarshal(b, &values); err != nil {
		return fileName, nil, fmt.Errorf("failed parsing %s: %w", fileName, err)
	}

	if values == nil {
		values = Values{}
	}

	return fileName, values, nil
}

// setOptionString sets a string value in the options map.
func setOptionString(values Values, key string, value *string) {
	if values == nil || value == nil {
		return
	}

	values[key] = *value
}

// mergeOptionValues applies source values to destination and reports changes.
func mergeOptionValues(dst Values, src Values) bool {
	if dst == nil || len(src) == 0 {
		return false
	}

	changed := false

	for key, value := range src {
		if current, ok := dst[key]; ok && reflect.DeepEqual(current, value) {
			continue
		}

		dst[key] = value
		changed = true
	}

	return changed
}

// writeOptionsYAML persists merged options values. It does not touch the in-memory options,
// which the caller applies through applyOptionValues when it changed one.
func (c *Config) writeOptionsYAML(fileName string, values Values) (bool, error) {
	b, err := yaml.Marshal(values)
	if err != nil {
		return false, err
	}

	if err = os.WriteFile(fileName, b, fs.ModeConfigFile); err != nil {
		return false, err
	}

	return true, nil
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

// CliContextString returns a global cli string flag value if set.
func (c *Config) CliContextString(name string) string {
	if c.cliCtx == nil {
		return ""
	}

	return c.cliCtx.String(name)
}

// serialFiles returns the storage serial copies in lookup order, each with the mode to create it with.
// The mode is deliberately permissive: the serial must stay readable when the UID/GID of the process
// changes, which happens when an instance is reconfigured (e.g. in compose.yaml) and restarted.
func (c *Config) serialFiles() []struct {
	Name string
	Mode os.FileMode
} {
	return []struct {
		Name string
		Mode os.FileMode
	}{
		{filepath.Join(c.StoragePath(), serialName), fs.ModeFile},
		{c.BackupPath(serialName), fs.ModeFile},
	}
}

// readSerial returns the storage serial from the first copy that holds a valid one, or an empty string
// if none does, in which case InitSerial generates one.
func (c *Config) readSerial() string {
	for _, f := range c.serialFiles() {
		if serial := readSerialFile(f.Name); serial != "" {
			return serial
		}
	}

	return ""
}

// readSerialFile returns the serial stored in a single file, or an empty string if it is absent,
// unreadable, or invalid.
// Surrounding whitespace is tolerated because a stray newline would otherwise discard the serial and
// rotate the preview token derived from it.
func readSerialFile(fileName string) string {
	data, err := os.ReadFile(fileName) //nolint:gosec // path is computed from the storage and backup paths

	switch {
	case os.IsNotExist(err):
		return ""
	case err != nil:
		event.SystemWarn([]string{"config", "serial", "read %s", "%s"}, clean.Log(fileName), clean.Error(err))
		return ""
	}

	if serial := strings.TrimSpace(string(data)); rnd.IsUID(serial, serialPrefix) {
		return serial
	}

	event.SystemWarn([]string{"config", "serial", "read %s", "invalid value"}, clean.Log(fileName))

	return ""
}

// serialFileHas reports whether the file already holds the given serial, without logging.
// It guards the restore path, which must not repeat the warnings readSerialFile emits for a bad copy.
func serialFileHas(fileName, serial string) bool {
	data, err := os.ReadFile(fileName) //nolint:gosec // path is computed from the storage and backup paths
	return err == nil && strings.TrimSpace(string(data)) == serial
}

// restoreSerial rewrites the serial copies that are missing or damaged, so one surviving copy heals the
// other. Failures are reported but never fatal, since the serial is already available in memory.
func (c *Config) restoreSerial(serial string) {
	for _, f := range c.serialFiles() {
		if serialFileHas(f.Name, serial) {
			continue
		}

		if err := os.WriteFile(f.Name, []byte(serial), f.Mode); err != nil {
			event.SystemWarn([]string{"config", "serial", "restore %s", "%s"}, clean.Log(f.Name), clean.Error(err))
		} else {
			event.SystemInfo([]string{"config", "serial", "restore %s", status.Succeeded}, clean.Log(f.Name))
		}
	}
}

// InitSerial initializes the storage path with a random serial if it does not have one yet, and restores
// any copy that is missing or damaged.
// It identifies the storage across restarts, so it is kept redundantly; only a failure to store the
// authoritative copy is fatal.
func (c *Config) InitSerial() error {
	serial := c.Serial()

	if serial == "" {
		serial = rnd.GenerateUID(serialPrefix)
		storageName := filepath.Join(c.StoragePath(), serialName)

		if err := os.WriteFile(storageName, []byte(serial), fs.ModeFile); err != nil {
			return fmt.Errorf("could not create %s: %w", clean.Log(storageName), err)
		}

		// Adopt the serial only once stored, so a restart cannot silently change the preview token.
		c.serial = serial
	}

	c.restoreSerial(serial)

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

// Edition returns the edition name.
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

// Develop checks if features under development should be enabled.
func (c *Config) Develop() bool {
	return Develop || Env(EnvDevelop)
}

// Experimental checks if new features that may be incomplete or unstable should be enabled.
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

// SetLogLevel sets the application log level.
func (c *Config) SetLogLevel(level logrus.Level) {
	SetLogLevel(level)
}

// stopHubTicker stops the periodic hub renewal ticker if it is running.
func (c *Config) stopHubTicker() {
	c.hubLock.Lock()
	cancel := c.hubCancel
	c.hubCancel = nil
	c.hubLock.Unlock()

	if cancel != nil {
		cancel()
	}
}

// Shutdown shuts down the active processes and closes the database connection.
func (c *Config) Shutdown() {
	c.stopHubTicker()

	// App is no longer accepting requests.
	c.ready.Store(false)

	// Send cancel signal to all workers.
	mutex.CancelAll()

	// Shutdown thumbnail library.
	thumb.Shutdown()

	// Reported on the console-only system log, as the database backing the error log is going away.
	if err := c.CloseDb(); err != nil {
		event.SystemError([]string{"config", "database", "close", "%s"}, clean.Error(err))
	} else {
		event.SystemDebug([]string{"config", "database", "close", status.Succeeded})
	}
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
		log.Warnf("config: failed to save API keys for maps and places (%s)", err)
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

	ctx, cancel := context.WithCancel(context.Background())

	c.hubLock.Lock()
	c.hubCancel = cancel
	c.hubLock.Unlock()

	d := 23*time.Hour + time.Duration(float64(2*time.Hour)*rand.Float64()) //nolint:gosec // jitter for scheduling only, crypto not required
	ticker := time.NewTicker(d)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.RenewApiKeys()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Hub returns the PhotoPrism hub config.
func (c *Config) Hub() *hub.Config {
	c.initHub()

	return c.hub
}
