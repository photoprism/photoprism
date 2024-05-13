package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Options hold the global configuration values without further validation or processing.
// Application code should retrieve option values via getter functions since they provide
// validation and return defaults if a value is empty.
type Options struct {
	Name                  string        `json:"-"`
	About                 string        `json:"-"`
	Edition               string        `json:"-"`
	Version               string        `json:"-"`
	Copyright             string        `json:"-"`
	PartnerID             string        `yaml:"-" json:"-" flag:"partner-id"`
	AuthMode              string        `yaml:"AuthMode" json:"-" flag:"auth-mode"`
	Public                bool          `yaml:"Public" json:"-" flag:"public"`
	AdminUser             string        `yaml:"AdminUser" json:"-" flag:"admin-user"`
	AdminPassword         string        `yaml:"AdminPassword" json:"-" flag:"admin-password"`
	PasswordLength        int           `yaml:"PasswordLength" json:"-" flag:"password-length"`
	PasswordResetUri      string        `yaml:"PasswordResetUri" json:"-" flag:"password-reset-uri"`
	RegisterUri           string        `yaml:"RegisterUri" json:"-" flag:"register-uri"`
	LoginUri              string        `yaml:"LoginUri" json:"-" flag:"login-uri"`
	SessionMaxAge         int64         `yaml:"SessionMaxAge" json:"-" flag:"session-maxage"`
	SessionTimeout        int64         `yaml:"SessionTimeout" json:"-" flag:"session-timeout"`
	SessionCache          int64         `yaml:"SessionCache" json:"-" flag:"session-cache"`
	LogLevel              string        `yaml:"LogLevel" json:"-" flag:"log-level"`
	Prod                  bool          `yaml:"Prod" json:"Prod" flag:"prod"`
	Debug                 bool          `yaml:"Debug" json:"Debug" flag:"debug"`
	Trace                 bool          `yaml:"Trace" json:"Trace" flag:"trace"`
	Test                  bool          `yaml:"-" json:"Test,omitempty" flag:"test"`
	Unsafe                bool          `yaml:"-" json:"-" flag:"unsafe"`
	Demo                  bool          `yaml:"-" json:"-" flag:"demo"`
	Sponsor               bool          `yaml:"-" json:"-" flag:"sponsor"`
	ConfigPath            string        `yaml:"ConfigPath" json:"-" flag:"config-path"`
	DefaultsYaml          string        `json:"-" yaml:"-" flag:"defaults-yaml"`
	OriginalsPath         string        `yaml:"OriginalsPath" json:"-" flag:"originals-path"`
	OriginalsLimit        int           `yaml:"OriginalsLimit" json:"OriginalsLimit" flag:"originals-limit"`
	ResolutionLimit       int           `yaml:"ResolutionLimit" json:"ResolutionLimit" flag:"resolution-limit"`
	UsersPath             string        `yaml:"UsersPath" json:"-" flag:"users-path"`
	StoragePath           string        `yaml:"StoragePath" json:"-" flag:"storage-path"`
	SidecarPath           string        `yaml:"SidecarPath" json:"-" flag:"sidecar-path"`
	SidecarYaml           bool          `yaml:"SidecarYaml" json:"SidecarYaml" flag:"sidecar-yaml" default:"true"`
	CachePath             string        `yaml:"CachePath" json:"-" flag:"cache-path"`
	ImportPath            string        `yaml:"ImportPath" json:"-" flag:"import-path"`
	ImportDest            string        `yaml:"ImportDest" json:"-" flag:"import-dest"`
	AssetsPath            string        `yaml:"AssetsPath" json:"-" flag:"assets-path"`
	CustomAssetsPath      string        `yaml:"-" json:"-" flag:"custom-assets-path"`
	TempPath              string        `yaml:"TempPath" json:"-" flag:"temp-path"`
	BackupPath            string        `yaml:"BackupPath" json:"-" flag:"backup-path"`
	BackupSchedule        string        `yaml:"BackupSchedule" json:"BackupSchedule" flag:"backup-schedule"`
	BackupRetain          int           `yaml:"BackupRetain" json:"BackupRetain" flag:"backup-retain"`
	BackupIndex           bool          `yaml:"BackupIndex" json:"BackupIndex" flag:"backup-index" default:"true"`
	BackupAlbums          bool          `yaml:"BackupAlbums" json:"BackupAlbums" flag:"backup-albums" default:"true"`
	IndexWorkers          int           `yaml:"IndexWorkers" json:"IndexWorkers" flag:"index-workers"`
	IndexSchedule         string        `yaml:"IndexSchedule" json:"IndexSchedule" flag:"index-schedule"`
	WakeupInterval        time.Duration `yaml:"WakeupInterval" json:"WakeupInterval" flag:"wakeup-interval"`
	AutoIndex             int           `yaml:"AutoIndex" json:"AutoIndex" flag:"auto-index"`
	AutoImport            int           `yaml:"AutoImport" json:"AutoImport" flag:"auto-import"`
	ReadOnly              bool          `yaml:"ReadOnly" json:"ReadOnly" flag:"read-only"`
	Experimental          bool          `yaml:"Experimental" json:"Experimental" flag:"experimental"`
	DisableSettings       bool          `yaml:"DisableSettings" json:"-" flag:"disable-settings"`
	DisableRestart        bool          `yaml:"DisableRestart" json:"-" flag:"disable-restart"`
	DisableBackups        bool          `yaml:"DisableBackups" json:"DisableBackups" flag:"disable-backups"`
	DisableWebDAV         bool          `yaml:"DisableWebDAV" json:"DisableWebDAV" flag:"disable-webdav"`
	DisablePlaces         bool          `yaml:"DisablePlaces" json:"DisablePlaces" flag:"disable-places"`
	DisableTensorFlow     bool          `yaml:"DisableTensorFlow" json:"DisableTensorFlow" flag:"disable-tensorflow"`
	DisableFaces          bool          `yaml:"DisableFaces" json:"DisableFaces" flag:"disable-faces"`
	DisableClassification bool          `yaml:"DisableClassification" json:"DisableClassification" flag:"disable-classification"`
	DisableFFmpeg         bool          `yaml:"DisableFFmpeg" json:"DisableFFmpeg" flag:"disable-ffmpeg"`
	DisableExifTool       bool          `yaml:"DisableExifTool" json:"DisableExifTool" flag:"disable-exiftool"`
	DisableSips           bool          `yaml:"DisableSips" json:"DisableSips" flag:"disable-sips"`
	DisableDarktable      bool          `yaml:"DisableDarktable" json:"DisableDarktable" flag:"disable-darktable"`
	DisableRawTherapee    bool          `yaml:"DisableRawTherapee" json:"DisableRawTherapee" flag:"disable-rawtherapee"`
	DisableImageMagick    bool          `yaml:"DisableImageMagick" json:"DisableImageMagick" flag:"disable-imagemagick"`
	DisableHeifConvert    bool          `yaml:"DisableHeifConvert" json:"DisableHeifConvert" flag:"disable-heifconvert"`
	DisableVectors        bool          `yaml:"DisableVectors" json:"DisableVectors" flag:"disable-vectors"`
	DisableJpegXL         bool          `yaml:"DisableJpegXL" json:"DisableJpegXL" flag:"disable-jpegxl"`
	DisableRaw            bool          `yaml:"DisableRaw" json:"DisableRaw" flag:"disable-raw"`
	RawPresets            bool          `yaml:"RawPresets" json:"RawPresets" flag:"raw-presets"`
	ExifBruteForce        bool          `yaml:"ExifBruteForce" json:"ExifBruteForce" flag:"exif-bruteforce"`
	DetectNSFW            bool          `yaml:"DetectNSFW" json:"DetectNSFW" flag:"detect-nsfw"`
	UploadNSFW            bool          `yaml:"UploadNSFW" json:"-" flag:"upload-nsfw"`
	DefaultLocale         string        `yaml:"DefaultLocale" json:"DefaultLocale" flag:"default-locale"`
	DefaultTimezone       string        `yaml:"DefaultTimezone" json:"DefaultTimezone" flag:"default-timezone"`
	DefaultTheme          string        `yaml:"DefaultTheme" json:"DefaultTheme" flag:"default-theme"`
	AppName               string        `yaml:"AppName" json:"AppName" flag:"app-name"`
	AppMode               string        `yaml:"AppMode" json:"AppMode" flag:"app-mode"`
	AppIcon               string        `yaml:"AppIcon" json:"AppIcon" flag:"app-icon"`
	AppColor              string        `yaml:"AppColor" json:"AppColor" flag:"app-color"`
	LegalInfo             string        `yaml:"LegalInfo" json:"LegalInfo" flag:"legal-info"`
	LegalUrl              string        `yaml:"LegalUrl" json:"LegalUrl" flag:"legal-url"`
	WallpaperUri          string        `yaml:"WallpaperUri" json:"WallpaperUri" flag:"wallpaper-uri"`
	SiteUrl               string        `yaml:"SiteUrl" json:"SiteUrl" flag:"site-url"`
	SiteAuthor            string        `yaml:"SiteAuthor" json:"SiteAuthor" flag:"site-author"`
	SiteTitle             string        `yaml:"SiteTitle" json:"SiteTitle" flag:"site-title"`
	SiteCaption           string        `yaml:"SiteCaption" json:"SiteCaption" flag:"site-caption"`
	SiteDescription       string        `yaml:"SiteDescription" json:"SiteDescription" flag:"site-description"`
	SitePreview           string        `yaml:"SitePreview" json:"SitePreview" flag:"site-preview"`
	CdnUrl                string        `yaml:"CdnUrl" json:"CdnUrl" flag:"cdn-url"`
	CdnVideo              bool          `yaml:"CdnVideo" json:"CdnVideo" flag:"cdn-video"`
	CORSOrigin            string        `yaml:"CORSOrigin" json:"-" flag:"cors-origin"`
	CORSHeaders           string        `yaml:"CORSHeaders" json:"-" flag:"cors-headers"`
	CORSMethods           string        `yaml:"CORSMethods" json:"-" flag:"cors-methods"`
	HttpsProxy            string        `yaml:"HttpsProxy" json:"HttpsProxy" flag:"https-proxy"`
	HttpsProxyInsecure    bool          `yaml:"HttpsProxyInsecure" json:"HttpsProxyInsecure" flag:"https-proxy-insecure"`
	TrustedProxies        []string      `yaml:"TrustedProxies" json:"-" flag:"trusted-proxy"`
	ProxyProtoHeaders     []string      `yaml:"ProxyProtoHeaders" json:"-" flag:"proxy-proto-header"`
	ProxyProtoHttps       []string      `yaml:"ProxyProtoHttps" json:"-" flag:"proxy-proto-https"`
	DisableTLS            bool          `yaml:"DisableTLS" json:"DisableTLS" flag:"disable-tls"`
	DefaultTLS            bool          `yaml:"DefaultTLS" json:"DefaultTLS" flag:"default-tls"`
	TLSEmail              string        `yaml:"TLSEmail" json:"TLSEmail" flag:"tls-email"`
	TLSCert               string        `yaml:"TLSCert" json:"TLSCert" flag:"tls-cert"`
	TLSKey                string        `yaml:"TLSKey" json:"TLSKey" flag:"tls-key"`
	HttpMode              string        `yaml:"HttpMode" json:"-" flag:"http-mode"`
	HttpCompression       string        `yaml:"HttpCompression" json:"-" flag:"http-compression"`
	HttpCachePublic       bool          `yaml:"HttpCachePublic" json:"HttpCachePublic" flag:"http-cache-public"`
	HttpCacheMaxAge       int           `yaml:"HttpCacheMaxAge" json:"HttpCacheMaxAge" flag:"http-cache-maxage"`
	HttpVideoMaxAge       int           `yaml:"HttpVideoMaxAge" json:"HttpVideoMaxAge" flag:"http-video-maxage"`
	HttpHost              string        `yaml:"HttpHost" json:"-" flag:"http-host"`
	HttpPort              int           `yaml:"HttpPort" json:"-" flag:"http-port"`
	HttpSocket            string        `yaml:"-" json:"-" flag:"-"`
	DatabaseDriver        string        `yaml:"DatabaseDriver" json:"-" flag:"database-driver"`
	DatabaseDsn           string        `yaml:"DatabaseDsn" json:"-" flag:"database-dsn"`
	DatabaseName          string        `yaml:"DatabaseName" json:"-" flag:"database-name"`
	DatabaseServer        string        `yaml:"DatabaseServer" json:"-" flag:"database-server"`
	DatabaseUser          string        `yaml:"DatabaseUser" json:"-" flag:"database-user"`
	DatabasePassword      string        `yaml:"DatabasePassword" json:"-" flag:"database-password"`
	DatabaseTimeout       int           `yaml:"DatabaseTimeout" json:"-" flag:"database-timeout"`
	DatabaseConns         int           `yaml:"DatabaseConns" json:"-" flag:"database-conns"`
	DatabaseConnsIdle     int           `yaml:"DatabaseConnsIdle" json:"-" flag:"database-conns-idle"`
	SipsBin               string        `yaml:"SipsBin" json:"-" flag:"sips-bin"`
	SipsBlacklist         string        `yaml:"SipsBlacklist" json:"-" flag:"sips-blacklist"`
	FFmpegBin             string        `yaml:"FFmpegBin" json:"-" flag:"ffmpeg-bin"`
	FFmpegEncoder         string        `yaml:"FFmpegEncoder" json:"FFmpegEncoder" flag:"ffmpeg-encoder"`
	FFmpegSize            int           `yaml:"FFmpegSize" json:"FFmpegSize" flag:"ffmpeg-size"`
	FFmpegBitrate         int           `yaml:"FFmpegBitrate" json:"FFmpegBitrate" flag:"ffmpeg-bitrate"`
	FFmpegMapVideo        string        `yaml:"FFmpegMapVideo" json:"FFmpegMapVideo" flag:"ffmpeg-map-video"`
	FFmpegMapAudio        string        `yaml:"FFmpegMapAudio" json:"FFmpegMapAudio" flag:"ffmpeg-map-audio"`
	ExifToolBin           string        `yaml:"ExifToolBin" json:"-" flag:"exiftool-bin"`
	DarktableBin          string        `yaml:"DarktableBin" json:"-" flag:"darktable-bin"`
	DarktableCachePath    string        `yaml:"DarktableCachePath" json:"-" flag:"darktable-cache-path"`
	DarktableConfigPath   string        `yaml:"DarktableConfigPath" json:"-" flag:"darktable-config-path"`
	DarktableBlacklist    string        `yaml:"DarktableBlacklist" json:"-" flag:"darktable-blacklist"`
	RawTherapeeBin        string        `yaml:"RawTherapeeBin" json:"-" flag:"rawtherapee-bin"`
	RawTherapeeBlacklist  string        `yaml:"RawTherapeeBlacklist" json:"-" flag:"rawtherapee-blacklist"`
	ImageMagickBin        string        `yaml:"ImageMagickBin" json:"-" flag:"imagemagick-bin"`
	ImageMagickBlacklist  string        `yaml:"ImageMagickBlacklist" json:"-" flag:"imagemagick-blacklist"`
	HeifConvertBin        string        `yaml:"HeifConvertBin" json:"-" flag:"heifconvert-bin"`
	RsvgConvertBin        string        `yaml:"RsvgConvertBin" json:"-" flag:"rsvgconvert-bin"`
	DownloadToken         string        `yaml:"DownloadToken" json:"-" flag:"download-token"`
	PreviewToken          string        `yaml:"PreviewToken" json:"-" flag:"preview-token"`
	ThumbColor            string        `yaml:"ThumbColor" json:"ThumbColor" flag:"thumb-color"`
	ThumbFilter           string        `yaml:"ThumbFilter" json:"ThumbFilter" flag:"thumb-filter"`
	ThumbSize             int           `yaml:"ThumbSize" json:"ThumbSize" flag:"thumb-size"`
	ThumbSizeUncached     int           `yaml:"ThumbSizeUncached" json:"ThumbSizeUncached" flag:"thumb-size-uncached"`
	ThumbUncached         bool          `yaml:"ThumbUncached" json:"ThumbUncached" flag:"thumb-uncached"`
	JpegQuality           string        `yaml:"JpegQuality" json:"JpegQuality" flag:"jpeg-quality"`
	JpegSize              int           `yaml:"JpegSize" json:"JpegSize" flag:"jpeg-size"`
	PngSize               int           `yaml:"PngSize" json:"PngSize" flag:"png-size"`
	FaceSize              int           `yaml:"-" json:"-" flag:"face-size"`
	FaceScore             float64       `yaml:"-" json:"-" flag:"face-score"`
	FaceOverlap           int           `yaml:"-" json:"-" flag:"face-overlap"`
	FaceClusterSize       int           `yaml:"-" json:"-" flag:"face-cluster-size"`
	FaceClusterScore      int           `yaml:"-" json:"-" flag:"face-cluster-score"`
	FaceClusterCore       int           `yaml:"-" json:"-" flag:"face-cluster-core"`
	FaceClusterDist       float64       `yaml:"-" json:"-" flag:"face-cluster-dist"`
	FaceMatchDist         float64       `yaml:"-" json:"-" flag:"face-match-dist"`
	PIDFilename           string        `yaml:"PIDFilename" json:"-" flag:"pid-filename"`
	LogFilename           string        `yaml:"LogFilename" json:"-" flag:"log-filename"`
	DetachServer          bool          `yaml:"DetachServer" json:"-" flag:"detach-server"`
}

// NewOptions creates a new configuration entity by using two methods:
//
// 1. Load: This will initialize options from a yaml config file.
//
//  2. ApplyCliContext: Which comes after Load and overrides
//     any previous options giving an option two override file configs through the CLI.
func NewOptions(ctx *cli.Context) *Options {
	c := &Options{}

	// Has context?
	if ctx == nil {
		return c
	}

	// Set app name from metadata if possible.
	if s, ok := ctx.App.Metadata["Name"]; ok {
		c.Name = fmt.Sprintf("%s", s)
	}

	// Set app about from metadata if possible.
	if s, ok := ctx.App.Metadata["About"]; ok {
		c.About = fmt.Sprintf("%s", s)
	}

	// Set app edition from metadata if possible.
	if s, ok := ctx.App.Metadata["Edition"]; ok {
		c.Edition = fmt.Sprintf("%s", s)
	}

	// Set copyright and version information.
	c.Copyright = ctx.App.Copyright
	c.Version = ctx.App.Version

	// Enable database backups and YAML exports by default.
	c.SidecarYaml = true
	c.BackupIndex = true
	c.BackupAlbums = true

	// Load defaults from YAML file?
	if defaultsYaml := ctx.GlobalString("defaults-yaml"); defaultsYaml == "" {
		log.Tracef("config: defaults file was not specified")
	} else if c.DefaultsYaml = fs.Abs(defaultsYaml); !fs.FileExists(c.DefaultsYaml) {
		log.Tracef("config: defaults file %s does not exist", clean.Log(c.DefaultsYaml))
	} else if err := c.Load(c.DefaultsYaml); err != nil {
		log.Warnf("config: failed loading defaults from %s (%s)", clean.Log(c.DefaultsYaml), err)
	}

	if err := c.ApplyCliContext(ctx); err != nil {
		log.Error(err)
	}

	return c
}

// expandFilenames converts path in config to absolute path
func (c *Options) expandFilenames() {
	c.ConfigPath = fs.Abs(c.ConfigPath)
	c.StoragePath = fs.Abs(c.StoragePath)
	c.UsersPath = fs.Abs(c.UsersPath)
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
		return errors.New(fmt.Sprintf("%s not found", fileName))
	}

	yamlConfig, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(yamlConfig, c)
}

// ApplyCliContext uses options from the CLI to setup configuration overrides
// for the entity.
func (c *Options) ApplyCliContext(ctx *cli.Context) error {
	return ApplyCliContext(c, ctx)
}
