package fs

// Common file names and patterns used across packages.
const (
	PPIgnoreAll       = "*"
	PPIgnoreFilename  = ".ppignore"
	PPStorageFilename = ".ppstorage"
	PPHiddenPathname  = ".photoprism"
)

// Common directory names used across packages (sorted by name).
const (
	AlbumsDir       = "albums"
	DownloadDir     = "download"
	BackupDir       = "backup"
	BuildDir        = "build"
	CacheDir        = "cache"
	CertificatesDir = "certificates"
	NodeDir         = "node"
	PortalDir       = "portal"
	SecretsDir      = "secrets"
	CmdDir          = "cmd"
	ConfigDir       = "config"
	IconsDir        = "icons"
	ImgDir          = "img"
	KeysDir         = "keys"
	LocalesDir      = "locales"
	MediaDir        = "media"
	ModelsDir       = "models"
	ProfilesDir     = "profiles"
	SamplesDir      = "samples"
	SettingsDir     = "settings"
	SidecarDir      = "sidecar"
	StaticDir       = "static"
	WebDir          = "web"
	StorageDir      = "storage"
	TemplatesDir    = "templates"
	TestdataDir     = "testdata"
	ThemeDir        = "theme"
	ThumbnailsDir   = "thumbnails"
	UploadDir       = "upload"
	UsersDir        = "users"
	ZipDir          = "zip"
)

// Common file names used across packages (sorted by name).
const (
	AppJsFile            = "app.js"
	AssetsJsonFile       = "assets.json"
	ManifestJsonFile     = "manifest.json"
	SwJsFile             = "sw.js"
	SwScopeCleanupJsFile = "sw-scope-cleanup.js"
	VersionTxtFile       = "version.txt"
	JoinTokenFile        = "join_token"
	ClientSecretFile     = "client_secret"
)

// EnvFileName is the standard environment-file basename.
const EnvFileName = ".env"

// SerialFile is the storage-identity filename.
const SerialFile = "serial"

// SigningKeyFile is the instance signing-key filename.
const SigningKeyFile = "signing.key"

// IgnoreFilePattern matches hidden ignore-configuration basenames.
const IgnoreFilePattern = ".*ignore"

// Common configuration basenames retain extension selection through ConfigFilePath.
const (
	ConfigDefaultsName = "defaults"
	ConfigHubName      = "hub"
	ConfigOptionsName  = "options"
	ConfigSettingsName = "settings"
)
