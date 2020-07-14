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
	MySQL  = "mysql"
	SQLite = "sqlite3"
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
	Name               string
	Version            string
	Copyright          string
	SiteUrl            string `yaml:"site-url" flag:"site-url"`
	SitePreview        string `yaml:"site-preview" flag:"site-preview"`
	SiteTitle          string `yaml:"site-title" flag:"site-title"`
	SiteCaption        string `yaml:"site-caption" flag:"site-caption"`
	SiteDescription    string `yaml:"site-description" flag:"site-description"`
	SiteAuthor         string `yaml:"site-author" flag:"site-author"`
	Public             bool   `yaml:"public" flag:"public"`
	Debug              bool   `yaml:"debug" flag:"debug"`
	ReadOnly           bool   `yaml:"read-only" flag:"read-only"`
	Experimental       bool   `yaml:"experimental" flag:"experimental"`
	TensorFlowOff      bool   `yaml:"tf-off" flag:"tf-off"`
	Workers            int    `yaml:"workers" flag:"workers"`
	WakeupInterval     int    `yaml:"wakeup-interval" flag:"wakeup-interval"`
	AdminPassword      string `yaml:"admin-password" flag:"admin-password"`
	LogLevel           string `yaml:"log-level" flag:"log-level"`
	AssetsPath         string `yaml:"assets-path" flag:"assets-path"`
	StoragePath        string `yaml:"storage-path" flag:"storage-path"`
	ImportPath         string `yaml:"import-path" flag:"import-path"`
	OriginalsPath      string `yaml:"originals-path" flag:"originals-path"`
	OriginalsLimit     int64  `yaml:"originals-limit" flag:"originals-limit"`
	ConfigFile         string
	SettingsPath       string `yaml:"settings-path" flag:"settings-path"`
	SettingsHidden     bool   `yaml:"settings-hidden" flag:"settings-hidden"`
	TempPath           string `yaml:"temp-path" flag:"temp-path"`
	CachePath          string `yaml:"cache-path" flag:"cache-path"`
	DatabaseDriver     string `yaml:"database-driver" flag:"database-driver"`
	DatabaseDsn        string `yaml:"database-dsn" flag:"database-dsn"`
	DatabaseConns      int    `yaml:"database-conns" flag:"database-conns"`
	DatabaseConnsIdle  int    `yaml:"database-conns-idle" flag:"database-conns-idle"`
	HttpServerHost     string `yaml:"http-host" flag:"http-host"`
	HttpServerPort     int    `yaml:"http-port" flag:"http-port"`
	HttpServerMode     string `yaml:"http-mode" flag:"http-mode"`
	HttpServerPassword string `yaml:"http-password" flag:"http-password"`
	SipsBin            string `yaml:"sips-bin" flag:"sips-bin"`
	DarktableBin       string `yaml:"darktable-bin" flag:"darktable-bin"`
	DarktableUnlock    bool   `yaml:"darktable-unlock" flag:"darktable-unlock"`
	HeifConvertBin     string `yaml:"heifconvert-bin" flag:"heifconvert-bin"`
	FFmpegBin          string `yaml:"ffmpeg-bin" flag:"ffmpeg-bin"`
	ExifToolBin        string `yaml:"exiftool-bin" flag:"exiftool-bin"`
	SidecarJson        bool   `yaml:"sidecar-json" flag:"sidecar-json"`
	SidecarYaml        bool   `yaml:"sidecar-yaml" flag:"sidecar-yaml"`
	SidecarPath        string `yaml:"sidecar-path" flag:"sidecar-path"`
	PIDFilename        string `yaml:"pid-filename" flag:"pid-filename"`
	LogFilename        string `yaml:"log-filename" flag:"log-filename"`
	DetachServer       bool   `yaml:"detach-server" flag:"detach-server"`
	DetectNSFW         bool   `yaml:"detect-nsfw" flag:"detect-nsfw"`
	UploadNSFW         bool   `yaml:"upload-nsfw" flag:"upload-nsfw"`
	GeoCodingApi       string `yaml:"geocoding-api" flag:"geocoding-api"`
	DownloadToken      string `yaml:"download-token" flag:"download-token"`
	PreviewToken       string `yaml:"preview-token" flag:"preview-token"`
	ThumbFilter        string `yaml:"thumb-filter" flag:"thumb-filter"`
	ThumbUncached      bool   `yaml:"thumb-uncached" flag:"thumb-uncached"`
	ThumbSize          int    `yaml:"thumb-size" flag:"thumb-size"`
	ThumbSizeUncached  int    `yaml:"thumb-size-uncached" flag:"thumb-size-uncached"`
	JpegSize           int    `yaml:"jpeg-size" flag:"jpeg-size"`
	JpegQuality        int    `yaml:"jpeg-quality" flag:"jpeg-quality"`
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
	c.SettingsPath = fs.Abs(c.SettingsPath)
	c.StoragePath = fs.Abs(c.StoragePath)
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
