package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Options hold the global configuration values without further validation or processing.
// Application code should retrieve option values via getter functions since they provide
// validation and return defaults if a value is empty.
type Options struct {
	Name                    string        `json:"-"`
	About                   string        `json:"-"`
	Edition                 string        `json:"-"`
	Version                 string        `json:"-"`
	Copyright               string        `json:"-"`
	PartnerID               string        `yaml:"-" json:"-" flag:"partner-id"`
	AuthMode                string        `yaml:"AuthMode" json:"-" flag:"auth-mode"`
	AuthSecret              string        `yaml:"AuthSecret" json:"-" flag:"auth-secret"`
	Public                  bool          `yaml:"Public" json:"-" flag:"public"`
	NoHub                   bool          `yaml:"-" json:"-" flag:"no-hub"`
	AdminUser               string        `yaml:"AdminUser" json:"-" flag:"admin-user"`
	AdminPassword           string        `yaml:"AdminPassword" json:"-" flag:"admin-password"`
	AdminScope              string        `yaml:"AdminScope" json:"-" flag:"admin-scope" tags:"pro"`
	PasswordLength          int           `yaml:"PasswordLength" json:"-" flag:"password-length"`
	PasswordResetUri        string        `yaml:"PasswordResetUri" json:"-" flag:"password-reset-uri" tags:"plus,pro"`
	RegisterUri             string        `yaml:"RegisterUri" json:"-" flag:"register-uri" tags:"pro"`
	LoginUri                string        `yaml:"-" json:"-" flag:"login-uri"`
	LoginInfo               string        `yaml:"LoginInfo" json:"-" flag:"login-info" tags:"plus,pro"`
	OIDCUri                 string        `yaml:"OIDCUri" json:"-" flag:"oidc-uri"`
	OIDCClient              string        `yaml:"OIDCClient" json:"-" flag:"oidc-client"`
	OIDCSecret              string        `yaml:"OIDCSecret" json:"-" flag:"oidc-secret"`
	OIDCScopes              string        `yaml:"OIDCScopes" json:"-" flag:"oidc-scopes"`
	OIDCProvider            string        `yaml:"OIDCProvider" json:"OIDCProvider" flag:"oidc-provider"`
	OIDCIcon                string        `yaml:"OIDCIcon" json:"OIDCIcon" flag:"oidc-icon"`
	OIDCRedirect            bool          `yaml:"OIDCRedirect" json:"OIDCRedirect" flag:"oidc-redirect"`
	OIDCRegister            bool          `yaml:"OIDCRegister" json:"OIDCRegister" flag:"oidc-register"`
	OIDCUsername            string        `yaml:"OIDCUsername" json:"-" flag:"oidc-username"`
	OIDCDomain              string        `yaml:"-" json:"-" flag:"oidc-domain" tags:"pro"`
	OIDCRole                string        `yaml:"-" json:"-" flag:"oidc-role" tags:"pro"`
	OIDCWebDAV              bool          `yaml:"OIDCWebDAV" json:"-" flag:"oidc-webdav"`
	DisableOIDC             bool          `yaml:"DisableOIDC" json:"DisableOIDC" flag:"disable-oidc"`
	SessionMaxAge           int64         `yaml:"SessionMaxAge" json:"-" flag:"session-maxage"`
	SessionTimeout          int64         `yaml:"SessionTimeout" json:"-" flag:"session-timeout"`
	SessionCache            int64         `yaml:"SessionCache" json:"-" flag:"session-cache"`
	LogLevel                string        `yaml:"LogLevel" json:"-" flag:"log-level"`
	Prod                    bool          `yaml:"Prod" json:"Prod" flag:"prod"`
	Debug                   bool          `yaml:"Debug" json:"Debug" flag:"debug"`
	Trace                   bool          `yaml:"Trace" json:"Trace" flag:"trace"`
	Test                    bool          `yaml:"-" json:"Test,omitempty" flag:"test"`
	Unsafe                  bool          `yaml:"-" json:"-" flag:"unsafe"`
	Demo                    bool          `yaml:"-" json:"-" flag:"demo"`
	Sponsor                 bool          `yaml:"-" json:"-" flag:"sponsor"`
	ConfigPath              string        `yaml:"ConfigPath" json:"-" flag:"config-path"`
	OptionsYaml             string        `json:"-" yaml:"-" flag:"-"`
	DefaultsYaml            string        `json:"-" yaml:"-" flag:"defaults-yaml"`
	OriginalsPath           string        `yaml:"OriginalsPath" json:"-" flag:"originals-path"`
	OriginalsLimit          int           `yaml:"OriginalsLimit" json:"OriginalsLimit" flag:"originals-limit"`
	ResolutionLimit         int           `yaml:"ResolutionLimit" json:"ResolutionLimit" flag:"resolution-limit"`
	UsersPath               string        `yaml:"UsersPath" json:"-" flag:"users-path"`
	StoragePath             string        `yaml:"StoragePath" json:"-" flag:"storage-path"`
	ImportPath              string        `yaml:"ImportPath" json:"-" flag:"import-path"`
	ImportDest              string        `yaml:"ImportDest" json:"-" flag:"import-dest"`
	ImportAllow             string        `yaml:"ImportAllow" json:"ImportAllow" flag:"import-allow"`
	UploadNSFW              bool          `yaml:"UploadNSFW" json:"-" flag:"upload-nsfw"`
	UploadAllow             string        `yaml:"UploadAllow" json:"-" flag:"upload-allow"`
	UploadArchives          bool          `yaml:"UploadArchives" json:"-" flag:"upload-archives"`
	UploadLimit             int           `yaml:"UploadLimit" json:"-" flag:"upload-limit"`
	CachePath               string        `yaml:"CachePath" json:"-" flag:"cache-path"`
	TempPath                string        `yaml:"TempPath" json:"-" flag:"temp-path"`
	AssetsPath              string        `yaml:"AssetsPath" json:"-" flag:"assets-path"`
	CustomAssetsPath        string        `yaml:"-" json:"-" flag:"custom-assets-path" tags:"plus,pro"`
	CustomThemePath         string        `yaml:"-" json:"-" flag:"theme-path"`
	ModelsPath              string        `yaml:"ModelsPath" json:"-" flag:"models-path"`
	SidecarPath             string        `yaml:"SidecarPath" json:"-" flag:"sidecar-path"`
	SidecarYaml             bool          `yaml:"SidecarYaml" json:"SidecarYaml" flag:"sidecar-yaml" default:"true"`
	UsageInfo               bool          `yaml:"UsageInfo" json:"UsageInfo" flag:"usage-info"`
	FilesQuota              uint64        `yaml:"FilesQuota" json:"-" flag:"files-quota"`
	UsersQuota              int           `yaml:"UsersQuota" json:"-" flag:"users-quota" tags:"pro"`
	BackupPath              string        `yaml:"BackupPath" json:"-" flag:"backup-path"`
	BackupSchedule          string        `yaml:"BackupSchedule" json:"BackupSchedule" flag:"backup-schedule"`
	BackupRetain            int           `yaml:"BackupRetain" json:"BackupRetain" flag:"backup-retain"`
	BackupDatabase          bool          `yaml:"BackupDatabase" json:"BackupDatabase" flag:"backup-database" default:"true"`
	BackupAlbums            bool          `yaml:"BackupAlbums" json:"BackupAlbums" flag:"backup-albums" default:"true"`
	IndexWorkers            int           `yaml:"IndexWorkers" json:"IndexWorkers" flag:"index-workers"`
	IndexSchedule           string        `yaml:"IndexSchedule" json:"IndexSchedule" flag:"index-schedule"`
	WakeupInterval          time.Duration `yaml:"WakeupInterval" json:"WakeupInterval" flag:"wakeup-interval"`
	AutoIndex               int           `yaml:"AutoIndex" json:"AutoIndex" flag:"auto-index"`
	AutoImport              int           `yaml:"AutoImport" json:"AutoImport" flag:"auto-import"`
	ReadOnly                bool          `yaml:"ReadOnly" json:"ReadOnly" flag:"read-only"`
	Experimental            bool          `yaml:"Experimental" json:"Experimental" flag:"experimental"`
	DisableFrontend         bool          `yaml:"DisableFrontend" json:"-" flag:"disable-frontend"`
	DisableSettings         bool          `yaml:"DisableSettings" json:"-" flag:"disable-settings"`
	DisableBackups          bool          `yaml:"DisableBackups" json:"DisableBackups" flag:"disable-backups"`
	DisableRestart          bool          `yaml:"DisableRestart" json:"-" flag:"disable-restart"`
	DisableWebDAV           bool          `yaml:"DisableWebDAV" json:"DisableWebDAV" flag:"disable-webdav"`
	DisablePlaces           bool          `yaml:"DisablePlaces" json:"DisablePlaces" flag:"disable-places"`
	DisableTensorFlow       bool          `yaml:"DisableTensorFlow" json:"DisableTensorFlow" flag:"disable-tensorflow"`
	DisableFaces            bool          `yaml:"DisableFaces" json:"DisableFaces" flag:"disable-faces"`
	DisableClassification   bool          `yaml:"DisableClassification" json:"DisableClassification" flag:"disable-classification"`
	DisableFFmpeg           bool          `yaml:"DisableFFmpeg" json:"DisableFFmpeg" flag:"disable-ffmpeg"`
	DisableExifTool         bool          `yaml:"DisableExifTool" json:"DisableExifTool" flag:"disable-exiftool"`
	DisableVips             bool          `yaml:"DisableVips" json:"DisableVips" flag:"disable-vips"`
	DisableSips             bool          `yaml:"DisableSips" json:"DisableSips" flag:"disable-sips"`
	DisableDarktable        bool          `yaml:"DisableDarktable" json:"DisableDarktable" flag:"disable-darktable"`
	DisableRawTherapee      bool          `yaml:"DisableRawTherapee" json:"DisableRawTherapee" flag:"disable-rawtherapee"`
	DisableImageMagick      bool          `yaml:"DisableImageMagick" json:"DisableImageMagick" flag:"disable-imagemagick"`
	DisableHeifConvert      bool          `yaml:"DisableHeifConvert" json:"DisableHeifConvert" flag:"disable-heifconvert"`
	DisableVectors          bool          `yaml:"DisableVectors" json:"DisableVectors" flag:"disable-vectors"`
	DisableJpegXL           bool          `yaml:"DisableJpegXL" json:"DisableJpegXL" flag:"disable-jpegxl"`
	DisableRaw              bool          `yaml:"DisableRaw" json:"DisableRaw" flag:"disable-raw"`
	RawPresets              bool          `yaml:"RawPresets" json:"RawPresets" flag:"raw-presets"`
	ExifBruteForce          bool          `yaml:"ExifBruteForce" json:"ExifBruteForce" flag:"exif-bruteforce"`
	DefaultLocale           string        `yaml:"DefaultLocale" json:"DefaultLocale" flag:"default-locale"`
	DefaultTimezone         string        `yaml:"DefaultTimezone" json:"DefaultTimezone" flag:"default-timezone"`
	DefaultTheme            string        `yaml:"DefaultTheme" json:"DefaultTheme" flag:"default-theme"`
	PlacesLocale            string        `yaml:"PlacesLocale" json:"PlacesLocale" flag:"places-locale"`
	AppName                 string        `yaml:"AppName" json:"AppName" flag:"app-name"`
	AppMode                 string        `yaml:"AppMode" json:"AppMode" flag:"app-mode"`
	AppIcon                 string        `yaml:"AppIcon" json:"AppIcon" flag:"app-icon"`
	AppColor                string        `yaml:"AppColor" json:"AppColor" flag:"app-color"`
	LegalInfo               string        `yaml:"LegalInfo" json:"LegalInfo" flag:"legal-info"`
	LegalUrl                string        `yaml:"LegalUrl" json:"LegalUrl" flag:"legal-url"`
	WallpaperUri            string        `yaml:"WallpaperUri" json:"WallpaperUri" flag:"wallpaper-uri"`
	SiteUrl                 string        `yaml:"SiteUrl" json:"SiteUrl" flag:"site-url"`
	SiteAuthor              string        `yaml:"SiteAuthor" json:"SiteAuthor" flag:"site-author"`
	SiteTitle               string        `yaml:"SiteTitle" json:"SiteTitle" flag:"site-title"`
	SiteCaption             string        `yaml:"SiteCaption" json:"SiteCaption" flag:"site-caption"`
	SiteDescription         string        `yaml:"SiteDescription" json:"SiteDescription" flag:"site-description"`
	SiteFavicon             string        `yaml:"SiteFavicon" json:"SiteFavicon" flag:"site-favicon"`
	SitePreview             string        `yaml:"SitePreview" json:"SitePreview" flag:"site-preview"`
	CdnUrl                  string        `yaml:"CdnUrl" json:"CdnUrl" flag:"cdn-url"`
	CdnVideo                bool          `yaml:"CdnVideo" json:"CdnVideo" flag:"cdn-video"`
	CORSOrigin              string        `yaml:"CORSOrigin" json:"-" flag:"cors-origin"`
	CORSHeaders             string        `yaml:"CORSHeaders" json:"-" flag:"cors-headers"`
	CORSMethods             string        `yaml:"CORSMethods" json:"-" flag:"cors-methods"`
	ClusterDomain           string        `yaml:"ClusterDomain" json:"-" flag:"cluster-domain"`
	ClusterCIDR             string        `yaml:"ClusterCIDR" json:"-" flag:"cluster-cidr"`
	ClusterUUID             string        `yaml:"ClusterUUID" json:"-" flag:"cluster-uuid"`
	PortalUrl               string        `yaml:"PortalUrl" json:"-" flag:"portal-url"`
	JoinToken               string        `yaml:"JoinToken" json:"-" flag:"join-token"`
	NodeName                string        `yaml:"NodeName" json:"-" flag:"node-name"`
	NodeUUID                string        `yaml:"NodeUUID" json:"-" flag:"node-uuid"`
	NodeRole                string        `yaml:"-" json:"-" flag:"node-role"`
	NodeClientID            string        `yaml:"NodeClientID" json:"-" flag:"node-client-id"`
	NodeClientSecret        string        `yaml:"NodeClientSecret" json:"-" flag:"node-client-secret"`
	JWKSUrl                 string        `yaml:"JWKSUrl" json:"-" flag:"jwks-url"`
	JWKSCacheTTL            int           `yaml:"JWKSCacheTTL" json:"-" flag:"jwks-cache-ttl"`
	JWTScope                string        `yaml:"JWTScope" json:"-" flag:"jwt-scope"`
	JWTLeeway               int           `yaml:"JWTLeeway" json:"-" flag:"jwt-leeway"`
	AdvertiseUrl            string        `yaml:"AdvertiseUrl" json:"-" flag:"advertise-url"`
	HttpsProxy              string        `yaml:"HttpsProxy" json:"HttpsProxy" flag:"https-proxy"`
	HttpsProxyInsecure      bool          `yaml:"HttpsProxyInsecure" json:"HttpsProxyInsecure" flag:"https-proxy-insecure"`
	TrustedPlatform         string        `yaml:"TrustedPlatform" json:"-" flag:"trusted-platform"`
	TrustedProxies          []string      `yaml:"TrustedProxies" json:"-" flag:"trusted-proxy"`
	ProxyClientHeaders      []string      `yaml:"ProxyClientHeaders" json:"-" flag:"proxy-client-header"`
	ProxyProtoHeaders       []string      `yaml:"ProxyProtoHeaders" json:"-" flag:"proxy-proto-header"`
	ProxyProtoHttps         []string      `yaml:"ProxyProtoHttps" json:"-" flag:"proxy-proto-https"`
	DisableTLS              bool          `yaml:"DisableTLS" json:"DisableTLS" flag:"disable-tls"`
	DefaultTLS              bool          `yaml:"DefaultTLS" json:"DefaultTLS" flag:"default-tls"`
	TLSEmail                string        `yaml:"TLSEmail" json:"TLSEmail" flag:"tls-email"`
	TLSCert                 string        `yaml:"TLSCert" json:"TLSCert" flag:"tls-cert"`
	TLSKey                  string        `yaml:"TLSKey" json:"TLSKey" flag:"tls-key"`
	HttpMode                string        `yaml:"HttpMode" json:"-" flag:"http-mode"`
	HttpCompression         string        `yaml:"HttpCompression" json:"-" flag:"http-compression"`
	HttpCachePublic         bool          `yaml:"HttpCachePublic" json:"HttpCachePublic" flag:"http-cache-public"`
	HttpCacheMaxAge         int           `yaml:"HttpCacheMaxAge" json:"HttpCacheMaxAge" flag:"http-cache-maxage"`
	HttpVideoMaxAge         int           `yaml:"HttpVideoMaxAge" json:"HttpVideoMaxAge" flag:"http-video-maxage"`
	HttpHost                string        `yaml:"HttpHost" json:"-" flag:"http-host"`
	HttpPort                int           `yaml:"HttpPort" json:"-" flag:"http-port"`
	HttpSocket              *url.URL      `yaml:"-" json:"-" flag:"-"`
	DatabaseDriver          string        `yaml:"DatabaseDriver" json:"-" flag:"database-driver"`
	DatabaseDSN             string        `yaml:"DatabaseDSN" json:"-" flag:"database-dsn"`
	DatabaseName            string        `yaml:"DatabaseName" json:"-" flag:"database-name"`
	DatabaseServer          string        `yaml:"DatabaseServer" json:"-" flag:"database-server"`
	DatabaseUser            string        `yaml:"DatabaseUser" json:"-" flag:"database-user"`
	DatabasePassword        string        `yaml:"DatabasePassword" json:"-" flag:"database-password"`
	DatabaseTimeout         int           `yaml:"DatabaseTimeout" json:"-" flag:"database-timeout"`
	DatabaseConns           int           `yaml:"DatabaseConns" json:"-" flag:"database-conns"`
	DatabaseConnsIdle       int           `yaml:"DatabaseConnsIdle" json:"-" flag:"database-conns-idle"`
	DatabaseProvisionDriver string        `yaml:"DatabaseProvisionDriver" json:"-" flag:"database-provision-driver"`
	DatabaseProvisionDSN    string        `yaml:"DatabaseProvisionDSN" json:"-" flag:"database-provision-dsn"`
	FFmpegBin               string        `yaml:"FFmpegBin" json:"-" flag:"ffmpeg-bin"`
	FFmpegEncoder           string        `yaml:"FFmpegEncoder" json:"FFmpegEncoder" flag:"ffmpeg-encoder"`
	FFmpegSize              int           `yaml:"FFmpegSize" json:"FFmpegSize" flag:"ffmpeg-size"`
	FFmpegQuality           int           `yaml:"FFmpegQuality" json:"FFmpegQuality" flag:"ffmpeg-quality"`
	FFmpegBitrate           int           `yaml:"FFmpegBitrate" json:"FFmpegBitrate" flag:"ffmpeg-bitrate"`
	FFmpegPreset            string        `yaml:"FFmpegPreset" json:"FFmpegPreset" flag:"ffmpeg-preset"`
	FFmpegDevice            string        `yaml:"FFmpegDevice" json:"-" flag:"ffmpeg-device"`
	FFmpegMapVideo          string        `yaml:"FFmpegMapVideo" json:"FFmpegMapVideo" flag:"ffmpeg-map-video"`
	FFmpegMapAudio          string        `yaml:"FFmpegMapAudio" json:"FFmpegMapAudio" flag:"ffmpeg-map-audio"`
	ExifToolBin             string        `yaml:"ExifToolBin" json:"-" flag:"exiftool-bin"`
	SipsBin                 string        `yaml:"SipsBin" json:"-" flag:"sips-bin"`
	SipsExclude             string        `yaml:"SipsExclude" json:"-" flag:"sips-exclude"`
	DarktableBin            string        `yaml:"DarktableBin" json:"-" flag:"darktable-bin"`
	DarktableCachePath      string        `yaml:"DarktableCachePath" json:"-" flag:"darktable-cache-path"`
	DarktableConfigPath     string        `yaml:"DarktableConfigPath" json:"-" flag:"darktable-config-path"`
	DarktableExclude        string        `yaml:"DarktableExclude" json:"-" flag:"darktable-exclude"`
	RawTherapeeBin          string        `yaml:"RawTherapeeBin" json:"-" flag:"rawtherapee-bin"`
	RawTherapeeExclude      string        `yaml:"RawTherapeeExclude" json:"-" flag:"rawtherapee-exclude"`
	ImageMagickBin          string        `yaml:"ImageMagickBin" json:"-" flag:"imagemagick-bin"`
	ImageMagickExclude      string        `yaml:"ImageMagickExclude" json:"-" flag:"imagemagick-exclude"`
	HeifConvertBin          string        `yaml:"HeifConvertBin" json:"-" flag:"heifconvert-bin"`
	HeifConvertOrientation  string        `yaml:"HeifConvertOrientation" json:"-" flag:"heifconvert-orientation"`
	RsvgConvertBin          string        `yaml:"RsvgConvertBin" json:"-" flag:"rsvgconvert-bin"`
	DownloadToken           string        `yaml:"DownloadToken" json:"-" flag:"download-token"`
	PreviewToken            string        `yaml:"PreviewToken" json:"-" flag:"preview-token"`
	ThumbLibrary            string        `yaml:"ThumbLibrary" json:"ThumbLibrary" flag:"thumb-library"`
	ThumbColor              string        `yaml:"ThumbColor" json:"ThumbColor" flag:"thumb-color"`
	ThumbFilter             string        `yaml:"ThumbFilter" json:"ThumbFilter" flag:"thumb-filter"`
	ThumbSize               int           `yaml:"ThumbSize" json:"ThumbSize" flag:"thumb-size"`
	ThumbSizeUncached       int           `yaml:"ThumbSizeUncached" json:"ThumbSizeUncached" flag:"thumb-size-uncached"`
	ThumbUncached           bool          `yaml:"ThumbUncached" json:"ThumbUncached" flag:"thumb-uncached"`
	JpegQuality             int           `yaml:"JpegQuality" json:"JpegQuality" flag:"jpeg-quality"`
	JpegSize                int           `yaml:"JpegSize" json:"JpegSize" flag:"jpeg-size"`
	PngSize                 int           `yaml:"PngSize" json:"PngSize" flag:"png-size"`
	VisionYaml              string        `yaml:"VisionYaml" json:"-" flag:"vision-yaml"`
	VisionApi               bool          `yaml:"VisionApi" json:"-" flag:"vision-api"`
	VisionUri               string        `yaml:"VisionUri" json:"-" flag:"vision-uri"`
	VisionKey               string        `yaml:"VisionKey" json:"-" flag:"vision-key"`
	VisionSchedule          string        `yaml:"VisionSchedule" json:"VisionSchedule" flag:"vision-schedule"`
	VisionFilter            string        `yaml:"VisionFilter" json:"VisionFilter" flag:"vision-filter"`
	DetectNSFW              bool          `yaml:"DetectNSFW" json:"DetectNSFW" flag:"detect-nsfw"`
	FaceEngine              string        `yaml:"FaceEngine" json:"-" flag:"face-engine"`
	FaceEngineRetry         bool          `yaml:"-" json:"-" flag:"-"`
	FaceEngineThreads       int           `yaml:"FaceEngineThreads" json:"-" flag:"face-engine-threads"`
	FaceSize                int           `yaml:"-" json:"-" flag:"face-size"`
	FaceScore               float64       `yaml:"-" json:"-" flag:"face-score"`
	FaceAngles              []float64     `yaml:"-" json:"-" flag:"face-angle"`
	FaceOverlap             int           `yaml:"-" json:"-" flag:"face-overlap"`
	FaceClusterSize         int           `yaml:"-" json:"-" flag:"face-cluster-size"`
	FaceClusterScore        int           `yaml:"-" json:"-" flag:"face-cluster-score"`
	FaceClusterCore         int           `yaml:"-" json:"-" flag:"face-cluster-core"`
	FaceClusterDist         float64       `yaml:"-" json:"-" flag:"face-cluster-dist"`
	FaceClusterRadius       float64       `yaml:"-" json:"-" flag:"face-cluster-radius"`
	FaceCollisionDist       float64       `yaml:"-" json:"-" flag:"face-collision-dist"`
	FaceEpsilonDist         float64       `yaml:"-" json:"-" flag:"face-epsilon-dist"`
	FaceMatchDist           float64       `yaml:"-" json:"-" flag:"face-match-dist"`
	FaceMatchChildren       bool          `yaml:"-" json:"-" flag:"face-match-children"`
	FaceMatchBackground     bool          `yaml:"-" json:"-" flag:"face-match-background"`
	PIDFilename             string        `yaml:"PIDFilename" json:"-" flag:"pid-filename"`
	LogFilename             string        `yaml:"LogFilename" json:"-" flag:"log-filename"`
	DetachServer            bool          `yaml:"DetachServer" json:"-" flag:"detach-server"`
	Deprecated              struct {
		DatabaseDsn string `yaml:"DatabaseDsn,omitempty" json:"-" flag:"-"`
	} `yaml:",inline,omitempty" json:"-" flag:"-"`
}

// NewOptions creates a new configuration entity by using two methods:
//
// 1. Load: This will initialize options from a yaml config file.
//
//  2. ApplyCliContext: Which comes after Load and overrides
//     any previous options giving an option two override file configs through the CLI.
func NewOptions(ctx *cli.Context) *Options {
	c := &Options{FaceEngine: face.EngineAuto}

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
	c.BackupDatabase = true
	c.BackupAlbums = true

	// Initialize options with the values from the "defaults.yml" file, if it exists.
	if defaultsYaml := ctx.String("defaults-yaml"); defaultsYaml == "" {
		log.Tracef("config: defaults file was not specified")
	} else if c.DefaultsYaml = fs.Abs(defaultsYaml); !fs.FileExists(c.DefaultsYaml) {
		log.Tracef("config: defaults file %s does not exist", clean.Log(c.DefaultsYaml))
	} else if err := c.Load(c.DefaultsYaml); err != nil {
		log.Warnf("config: failed loading defaults from %s (%s)", clean.Log(c.DefaultsYaml), err)
	}

	// Apply options specified with environment variables and command-line flags.
	if err := c.ApplyCliContext(ctx); err != nil {
		log.Error(err)
	}

	return c
}

// expandFilenames converts path in config to absolute path
func (o *Options) expandFilenames() {
	o.ConfigPath = fs.Abs(o.ConfigPath)
	o.StoragePath = fs.Abs(o.StoragePath)
	o.UsersPath = fs.Abs(o.UsersPath)
	o.BackupPath = fs.Abs(o.BackupPath)
	o.AssetsPath = fs.Abs(o.AssetsPath)
	o.CachePath = fs.Abs(o.CachePath)
	o.OriginalsPath = fs.Abs(o.OriginalsPath)
	o.ImportPath = fs.Abs(o.ImportPath)
	o.TempPath = fs.Abs(o.TempPath)
	o.PIDFilename = fs.Abs(o.PIDFilename)
	o.LogFilename = fs.Abs(o.LogFilename)
}

// Load uses a yaml config file to initiate the configuration entity.
func (o *Options) Load(fileName string) error {
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

	return yaml.Unmarshal(yamlConfig, o)
}

// ApplyCliContext uses options from the CLI to setup configuration overrides
// for the entity.
func (o *Options) ApplyCliContext(ctx *cli.Context) error {
	return ApplyCliContext(o, ctx)
}
