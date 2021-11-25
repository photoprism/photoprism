package config

import (
	"errors"
	"fmt"
	"os"
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

// Options provides a struct in which application configuration is stored.
// Application code must use functions to get config options, for two reasons:
//
// 1. Some options are computed and we don't want to leak implementation details (aims at reducing refactoring overhead).
//
// 2. Paths might actually be dynamic later (if we build a multi-user version).
//
// See https://github.com/photoprism/photoprism/issues/50#issuecomment-433856358
type Options struct {
	Name                  string  `json:"-"`
	Version               string  `json:"-"`
	Copyright             string  `json:"-"`
	Debug                 bool    `yaml:"Debug" json:"Debug" flag:"debug"`
	LogLevel              string  `yaml:"LogLevel" json:"-" flag:"log-level"`
	LogFilename           string  `yaml:"LogFilename" json:"-" flag:"log-filename"`
	Test                  bool    `yaml:"-" json:"Test,omitempty" flag:"test"`
	Unsafe                bool    `yaml:"-" json:"-" flag:"unsafe"`
	Demo                  bool    `yaml:"Demo" json:"-" flag:"demo"`
	Sponsor               bool    `yaml:"-" json:"-" flag:"sponsor"`
	Public                bool    `yaml:"Public" json:"-" flag:"public"`
	AdminPassword         string  `yaml:"AdminPassword" json:"-" flag:"admin-password"`
	ReadOnly              bool    `yaml:"ReadOnly" json:"ReadOnly" flag:"read-only"`
	Experimental          bool    `yaml:"Experimental" json:"Experimental" flag:"experimental"`
	ConfigPath            string  `yaml:"ConfigPath" json:"-" flag:"config-path"`
	ConfigFile            string  `json:"-"`
	OriginalsPath         string  `yaml:"OriginalsPath" json:"-" flag:"originals-path"`
	OriginalsLimit        int64   `yaml:"OriginalsLimit" json:"OriginalsLimit" flag:"originals-limit"`
	ImportPath            string  `yaml:"ImportPath" json:"-" flag:"import-path"`
	StoragePath           string  `yaml:"StoragePath" json:"-" flag:"storage-path"`
	CachePath             string  `yaml:"CachePath" json:"-" flag:"cache-path"`
	SidecarPath           string  `yaml:"SidecarPath" json:"-" flag:"sidecar-path"`
	TempPath              string  `yaml:"TempPath" json:"-" flag:"temp-path"`
	BackupPath            string  `yaml:"BackupPath" json:"-" flag:"backup-path"`
	AssetsPath            string  `yaml:"AssetsPath" json:"-" flag:"assets-path"`
	Workers               int     `yaml:"Workers" json:"Workers" flag:"workers"`
	WakeupInterval        int     `yaml:"WakeupInterval" json:"WakeupInterval" flag:"wakeup-interval"`
	AutoIndex             int     `yaml:"AutoIndex" json:"AutoIndex" flag:"auto-index"`
	AutoImport            int     `yaml:"AutoImport" json:"AutoImport" flag:"auto-import"`
	DisableWebDAV         bool    `yaml:"DisableWebDAV" json:"DisableWebDAV" flag:"disable-webdav"`
	DisableBackups        bool    `yaml:"DisableBackups" json:"DisableBackups" flag:"disable-backups"`
	DisableSettings       bool    `yaml:"DisableSettings" json:"-" flag:"disable-settings"`
	DisablePlaces         bool    `yaml:"DisablePlaces" json:"DisablePlaces" flag:"disable-places"`
	DisableExifTool       bool    `yaml:"DisableExifTool" json:"DisableExifTool" flag:"disable-exiftool"`
	DisableFFmpeg         bool    `yaml:"DisableFFmpeg" json:"DisableFFmpeg" flag:"disable-ffmpeg"`
	DisableDarktable      bool    `yaml:"DisableDarktable" json:"DisableDarktable" flag:"disable-darktable"`
	DisableRawtherapee    bool    `yaml:"DisableRawtherapee" json:"DisableRawtherapee" flag:"disable-rawtherapee"`
	DisableSips           bool    `yaml:"DisableSips" json:"DisableSips" flag:"disable-sips"`
	DisableHeifConvert    bool    `yaml:"DisableHeifConvert" json:"DisableHeifConvert" flag:"disable-heifconvert"`
	DisableTensorFlow     bool    `yaml:"DisableTensorFlow" json:"DisableTensorFlow" flag:"disable-tensorflow"`
	DisableFaces          bool    `yaml:"DisableFaces" json:"DisableFaces" flag:"disable-faces"`
	DisableClassification bool    `yaml:"DisableClassification" json:"DisableClassification" flag:"disable-classification"`
	DetectNSFW            bool    `yaml:"DetectNSFW" json:"DetectNSFW" flag:"detect-nsfw"`
	UploadNSFW            bool    `yaml:"UploadNSFW" json:"-" flag:"upload-nsfw"`
	AppName               string  `yaml:"AppName" json:"AppName" flag:"app-name"`
	AppMode               string  `yaml:"AppMode" json:"AppMode" flag:"app-mode"`
	AppIcon               string  `yaml:"AppIcon" json:"AppIcon" flag:"app-icon"`
	CdnUrl                string  `yaml:"CdnUrl" json:"CdnUrl" flag:"cdn-url"`
	SiteUrl               string  `yaml:"SiteUrl" json:"SiteUrl" flag:"site-url"`
	SiteAuthor            string  `yaml:"SiteAuthor" json:"SiteAuthor" flag:"site-author"`
	SiteTitle             string  `yaml:"SiteTitle" json:"SiteTitle" flag:"site-title"`
	SiteCaption           string  `yaml:"SiteCaption" json:"SiteCaption" flag:"site-caption"`
	SiteDescription       string  `yaml:"SiteDescription" json:"SiteDescription" flag:"site-description"`
	SitePreview           string  `yaml:"SitePreview" json:"SitePreview" flag:"site-preview"`
	DatabaseDriver        string  `yaml:"DatabaseDriver" json:"-" flag:"database-driver"`
	DatabaseDsn           string  `yaml:"DatabaseDsn" json:"-" flag:"database-dsn"`
	DatabaseServer        string  `yaml:"DatabaseServer" json:"-" flag:"database-server"`
	DatabaseName          string  `yaml:"DatabaseName" json:"-" flag:"database-name"`
	DatabaseUser          string  `yaml:"DatabaseUser" json:"-" flag:"database-user"`
	DatabasePassword      string  `yaml:"DatabasePassword" json:"-" flag:"database-password"`
	DatabaseConns         int     `yaml:"DatabaseConns" json:"-" flag:"database-conns"`
	DatabaseConnsIdle     int     `yaml:"DatabaseConnsIdle" json:"-" flag:"database-conns-idle"`
	HttpHost              string  `yaml:"HttpHost" json:"-" flag:"http-host"`
	HttpPort              int     `yaml:"HttpPort" json:"-" flag:"http-port"`
	HttpMode              string  `yaml:"HttpMode" json:"-" flag:"http-mode"`
	HttpCompression       string  `yaml:"HttpCompression" json:"-" flag:"http-compression"`
	RawPresets            bool    `yaml:"RawPresets" json:"RawPresets" flag:"raw-presets"`
	DarktableBin          string  `yaml:"DarktableBin" json:"-" flag:"darktable-bin"`
	DarktableBlacklist    string  `yaml:"DarktableBlacklist" json:"-" flag:"darktable-blacklist"`
	RawtherapeeBin        string  `yaml:"RawtherapeeBin" json:"-" flag:"rawtherapee-bin"`
	RawtherapeeBlacklist  string  `yaml:"RawtherapeeBlacklist" json:"-" flag:"rawtherapee-blacklist"`
	SipsBin               string  `yaml:"SipsBin" json:"-" flag:"sips-bin"`
	HeifConvertBin        string  `yaml:"HeifConvertBin" json:"-" flag:"heifconvert-bin"`
	FFmpegBin             string  `yaml:"FFmpegBin" json:"-" flag:"ffmpeg-bin"`
	FFmpegEncoder         string  `yaml:"FFmpegEncoder" json:"FFmpegEncoder" flag:"ffmpeg-encoder"`
	FFmpegBitrate         int     `yaml:"FFmpegBitrate" json:"FFmpegBitrate" flag:"ffmpeg-bitrate"`
	FFmpegBuffers         int     `yaml:"FFmpegBuffers" json:"FFmpegBuffers" flag:"ffmpeg-buffers"`
	ExifToolBin           string  `yaml:"ExifToolBin" json:"-" flag:"exiftool-bin"`
	DetachServer          bool    `yaml:"DetachServer" json:"-" flag:"detach-server"`
	DownloadToken         string  `yaml:"DownloadToken" json:"-" flag:"download-token"`
	PreviewToken          string  `yaml:"PreviewToken" json:"-" flag:"preview-token"`
	ThumbFilter           string  `yaml:"ThumbFilter" json:"ThumbFilter" flag:"thumb-filter"`
	ThumbUncached         bool    `yaml:"ThumbUncached" json:"ThumbUncached" flag:"thumb-uncached"`
	ThumbSize             int     `yaml:"ThumbSize" json:"ThumbSize" flag:"thumb-size"`
	ThumbSizeUncached     int     `yaml:"ThumbSizeUncached" json:"ThumbSizeUncached" flag:"thumb-size-uncached"`
	JpegSize              int     `yaml:"JpegSize" json:"JpegSize" flag:"jpeg-size"`
	JpegQuality           int     `yaml:"JpegQuality" json:"JpegQuality" flag:"jpeg-quality"`
	FaceSize              int     `yaml:"-" json:"-" flag:"face-size"`
	FaceScore             float64 `yaml:"-" json:"-" flag:"face-score"`
	FaceOverlap           int     `yaml:"-" json:"-" flag:"face-overlap"`
	FaceClusterSize       int     `yaml:"-" json:"-" flag:"face-cluster-size"`
	FaceClusterScore      int     `yaml:"-" json:"-" flag:"face-cluster-score"`
	FaceClusterCore       int     `yaml:"-" json:"-" flag:"face-cluster-core"`
	FaceClusterDist       float64 `yaml:"-" json:"-" flag:"face-cluster-dist"`
	FaceMatchDist         float64 `yaml:"-" json:"-" flag:"face-match-dist"`
	PIDFilename           string  `yaml:"PIDFilename" json:"-" flag:"pid-filename"`
}

// NewOptions creates a new configuration entity by using two methods:
//
// 1. Load: This will initialize options from a yaml config file.
//
// 2. SetContext: Which comes after Load and overrides
//    any previous options giving an option two override file configs through the CLI.
func NewOptions(ctx *cli.Context) *Options {
	c := &Options{}

	if ctx == nil {
		return c
	}

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
func (c *Options) expandFilenames() {
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
func (c *Options) Load(fileName string) error {
	if fileName == "" {
		return nil
	}

	if !fs.FileExists(fileName) {
		return errors.New(fmt.Sprintf("config: %s not found", fileName))
	}

	yamlConfig, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(yamlConfig, c)
}

// SetContext uses options from the CLI to setup configuration overrides
// for the entity.
func (c *Options) SetContext(ctx *cli.Context) error {
	v := reflect.ValueOf(c).Elem()

	// Iterate through all config fields.
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)

		tagValue := v.Type().Field(i).Tag.Get("flag")

		// Assign value to field with "flag" tag.
		if tagValue != "" {
			switch t := fieldValue.Interface().(type) {
			case float64:
				// Only if explicitly set or current value is empty (use default).
				if ctx.IsSet(tagValue) {
					f := ctx.Float64(tagValue)
					fieldValue.SetFloat(f)
				} else if ctx.GlobalIsSet(tagValue) || fieldValue.Float() == 0 {
					f := ctx.GlobalFloat64(tagValue)
					fieldValue.SetFloat(f)
				}
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
