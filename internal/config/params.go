package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

// Database drivers (sql dialects).
const (
	MySQL    = "mysql"
	MariaDB  = "mariadb"
	SQLite   = "sqlite3"
	Postgres = "postgres" // TODO: Requires GORM 2.0 for generic column data types
)

// Params provides a struct in which application configuration is stored.
// Application code must use functions to get config values, for two reasons:
//
// 1. Some values are computed and we don't want to leak implementation details (aims at reducing refactoring overhead).
//
// 2. Paths might actually be dynamic later (if we build a multi-user version).
//
// See https://github.com/photoprism/photoprism/issues/50#issuecomment-433856358
type Params struct {
	Name              string
	Version           string
	Copyright         string
	Debug             bool   `yaml:"Debug" flag:"debug"`
	Public            bool   `yaml:"Public" flag:"public"`
	ReadOnly          bool   `yaml:"ReadOnly" flag:"read-only"`
	Experimental      bool   `yaml:"Experimental" flag:"experimental"`
	ConfigPath        string `yaml:"ConfigPath" flag:"config-path"`
	ConfigFile        string
	AdminPassword     string `yaml:"AdminPassword" flag:"admin-password"`
	OriginalsPath     string `yaml:"OriginalsPath" flag:"originals-path"`
	OriginalsLimit    int64  `yaml:"OriginalsLimit" flag:"originals-limit"`
	ImportPath        string `yaml:"ImportPath" flag:"import-path"`
	StoragePath       string `yaml:"StoragePath" flag:"storage-path"`
	SidecarPath       string `yaml:"SidecarPath" flag:"sidecar-path"`
	TempPath          string `yaml:"TempPath" flag:"temp-path"`
	BackupPath        string `yaml:"BackupPath" flag:"backup-path"`
	AssetsPath        string `yaml:"AssetsPath" flag:"assets-path"`
	CachePath         string `yaml:"CachePath" flag:"cache-path"`
	Workers           int    `yaml:"Workers" flag:"workers"`
	WakeupInterval    int    `yaml:"WakeupInterval" flag:"wakeup-interval"`
	DisableBackups    bool   `yaml:"DisableBackups" flag:"disable-backups"`
	DisableSettings   bool   `yaml:"DisableSettings" flag:"disable-settings"`
	DisablePlaces     bool   `yaml:"DisablePlaces" flag:"disable-places"`
	DisableExifTool   bool   `yaml:"DisableExifTool" flag:"disable-exiftool"`
	DisableTensorFlow bool   `yaml:"DisableTensorFlow" flag:"disable-tensorflow"`
	DetectNSFW        bool   `yaml:"DetectNSFW" flag:"detect-nsfw"`
	UploadNSFW        bool   `yaml:"UploadNSFW" flag:"upload-nsfw"`
	LogLevel          string `yaml:"LogLevel" flag:"log-level"`
	LogFilename       string `yaml:"LogFilename" flag:"log-filename"`
	PIDFilename       string `yaml:"PIDFilename" flag:"pid-filename"`
	SiteUrl           string `yaml:"SiteUrl" flag:"site-url"`
	SitePreview       string `yaml:"SitePreview" flag:"site-preview"`
	SiteTitle         string `yaml:"SiteTitle" flag:"site-title"`
	SiteCaption       string `yaml:"SiteCaption" flag:"site-caption"`
	SiteDescription   string `yaml:"SiteDescription" flag:"site-description"`
	SiteAuthor        string `yaml:"SiteAuthor" flag:"site-author"`
	DatabaseDriver    string `yaml:"DatabaseDriver" flag:"database-driver"`
	DatabaseDsn       string `yaml:"DatabaseDsn" flag:"database-dsn"`
	DatabaseServer    string `yaml:"DatabaseServer" flag:"database-server"`
	DatabaseName      string `yaml:"DatabaseName" flag:"database-name"`
	DatabaseUser      string `yaml:"DatabaseUser" flag:"database-user"`
	DatabasePassword  string `yaml:"DatabasePassword" flag:"database-password"`
	DatabaseConns     int    `yaml:"DatabaseConns" flag:"database-conns"`
	DatabaseConnsIdle int    `yaml:"DatabaseConnsIdle" flag:"database-conns-idle"`
	HttpHost          string `yaml:"HttpHost" flag:"http-host"`
	HttpPort          int    `yaml:"HttpPort" flag:"http-port"`
	HttpMode          string `yaml:"HttpMode" flag:"http-mode"`
	SipsBin           string `yaml:"SipsBin" flag:"sips-bin"`
	RawtherapeeBin    string `yaml:"RawtherapeeBin" flag:"rawtherapee-bin"`
	DarktableBin      string `yaml:"DarktableBin" flag:"darktable-bin"`
	DarktablePresets  bool   `yaml:"DarktablePresets" flag:"darktable-presets"`
	HeifConvertBin    string `yaml:"HeifConvertBin" flag:"heifconvert-bin"`
	FFmpegBin         string `yaml:"FFmpegBin" flag:"ffmpeg-bin"`
	ExifToolBin       string `yaml:"ExifToolBin" flag:"exiftool-bin"`
	BackupYaml        bool   `yaml:"BackupYaml" flag:"backup-yaml"`
	DetachServer      bool   `yaml:"DetachServer" flag:"detach-server"`
	DownloadToken     string `yaml:"DownloadToken" flag:"download-token"`
	PreviewToken      string `yaml:"PreviewToken" flag:"preview-token"`
	ThumbFilter       string `yaml:"ThumbFilter" flag:"thumb-filter"`
	ThumbUncached     bool   `yaml:"ThumbUncached" flag:"thumb-uncached"`
	ThumbSize         int    `yaml:"ThumbSize" flag:"thumb-size"`
	ThumbSizeUncached int    `yaml:"ThumbSizeUncached" flag:"thumb-size-uncached"`
	JpegSize          int    `yaml:"JpegSize" flag:"jpeg-size"`
	JpegQuality       int    `yaml:"JpegQuality" flag:"jpeg-quality"`
}

// NewParams creates a new configuration entity by using two methods:
//
// 1. Load: This will initialize values from a yaml config file.
//
// 2. SetContext: Which comes after Load and overrides
//    any previous values giving an option two override file configs through the CLI.
func NewParams(ctx *cli.Context) *Params {
	c := &Params{}

	c.Name = ctx.App.Name
	c.Copyright = ctx.App.Copyright
	c.Version = ctx.App.Version
	c.ConfigFile = fs.Abs(ctx.GlobalString("config-file"))

	if err := c.Load(c.ConfigFile); err != nil {
		log.Debug(err)
	}

	if err := c.SetContext(ctx); err != nil {
		log.Error(err)
	}

	return c
}

// expandFilenames converts path in config to absolute path
func (c *Params) expandFilenames() {
	c.ConfigPath = fs.Abs(c.ConfigPath)
	c.StoragePath = fs.Abs(c.StoragePath)
	c.BackupPath = fs.Abs(c.BackupPath)
	c.AssetsPath = fs.Abs(c.AssetsPath)
	c.CachePath = fs.Abs(c.CachePath)
	c.OriginalsPath = fs.Abs(c.OriginalsPath)
	c.ImportPath = fs.Abs(c.ImportPath)
	c.TempPath = fs.Abs(c.TempPath)
	c.PIDFilename = fs.Abs(c.PIDFilename)
	c.LogFilename = fs.Abs(c.LogFilename)
}

// Load uses a yaml config file to initiate the configuration entity.
func (c *Params) Load(fileName string) error {
	if fileName == "" {
		return nil
	}

	if !fs.FileExists(fileName) {
		return errors.New(fmt.Sprintf("config: %s not found", fileName))
	}

	yamlConfig, err := ioutil.ReadFile(fileName)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(yamlConfig, c)
}

// SetContext uses values from the CLI to setup configuration overrides
// for the entity.
func (c *Params) SetContext(ctx *cli.Context) error {
	v := reflect.ValueOf(c).Elem()

	// Iterate through all config fields.
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		tagValue := v.Type().Field(i).Tag.Get("flag")

		// Automatically assign values to fields with "flag" tag.
		if tagValue != "" {
			switch t := fieldValue.Interface().(type) {
			case int, int64:
				// Only if explicitly set or current value is empty (use default).
				if ctx.IsSet(tagValue) {
					f := ctx.Int64(tagValue)
					fieldValue.SetInt(f)
				} else if ctx.GlobalIsSet(tagValue) || fieldValue.Int() == 0 {
					f := ctx.GlobalInt64(tagValue)
					fieldValue.SetInt(f)
				}
			case uint, uint64:
				// Only if explicitly set or current value is empty (use default).
				if ctx.IsSet(tagValue) {
					f := ctx.Uint64(tagValue)
					fieldValue.SetUint(f)
				} else if ctx.GlobalIsSet(tagValue) || fieldValue.Uint() == 0 {
					f := ctx.GlobalUint64(tagValue)
					fieldValue.SetUint(f)
				}
			case string:
				// Only if explicitly set or current value is empty (use default)
				if ctx.IsSet(tagValue) {
					f := ctx.String(tagValue)
					fieldValue.SetString(f)
				} else if ctx.GlobalIsSet(tagValue) || fieldValue.String() == "" {
					f := ctx.GlobalString(tagValue)
					fieldValue.SetString(f)
				}
			case bool:
				if ctx.IsSet(tagValue) {
					f := ctx.Bool(tagValue)
					fieldValue.SetBool(f)
				} else if ctx.GlobalIsSet(tagValue) {
					f := ctx.GlobalBool(tagValue)
					fieldValue.SetBool(f)
				}
			default:
				log.Warnf("can't assign value of type %s from cli flag %s", t, tagValue)
			}
		}
	}

	return nil
}
