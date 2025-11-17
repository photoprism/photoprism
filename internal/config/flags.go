package config

import (
	"fmt"

	"github.com/klauspost/cpuid/v2"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/ffmpeg/encode"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/internal/service/hub/places"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/http/scheme"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/time/tz"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Flags configures the global command-line interface (CLI) parameters.
var Flags = CliFlags{
	{
		Flag: &cli.StringFlag{
			Name:    "auth-mode",
			Aliases: []string{"a"},
			Usage:   "authentication `MODE` (public, password)",
			Value:   "password",
			EnvVars: EnvVars("AUTH_MODE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "auth-secret",
			Usage:   "secret `KEY` for signing authentication tokens",
			EnvVars: EnvVars("AUTH_SECRET"),
			Hidden:  true,
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "public",
			Aliases: []string{"p"},
			Hidden:  true,
			Usage:   "disables authentication, advanced settings, and WebDAV remote access",
			EnvVars: EnvVars("PUBLIC"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "admin-user",
			Aliases: []string{"login"},
			Usage:   "`USERNAME` of the superadmin account that is created on first startup",
			Value:   "admin",
			EnvVars: EnvVars("ADMIN_USER", "ADMIN_USERNAME"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "admin-password",
			Aliases: []string{"pw"},
			Usage:   fmt.Sprintf("initial `PASSWORD` of the superadmin account (%d-%d characters)", entity.PasswordLength, txt.ClipPassword),
			EnvVars: EnvVars("ADMIN_PASSWORD"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "password-length",
			Usage:   "minimum password `LENGTH` in characters",
			Value:   8,
			EnvVars: EnvVars("PASSWORD_LENGTH"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "oidc-uri",
			Usage:   "issuer `URI` for single sign-on via OpenID Connect, e.g. https://accounts.google.com",
			Value:   "",
			EnvVars: EnvVars("OIDC_URI"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "oidc-client",
			Usage:   "client `ID` for single sign-on via OpenID Connect",
			Value:   "",
			EnvVars: EnvVars("OIDC_CLIENT"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "oidc-secret",
			Usage:   "client `SECRET` for single sign-on via OpenID Connect",
			Value:   "",
			EnvVars: EnvVars("OIDC_SECRET"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "oidc-scopes",
			Usage:   "client authorization `SCOPES` for single sign-on via OpenID Connect",
			Value:   authn.OidcDefaultScopes,
			EnvVars: EnvVars("OIDC_SCOPES"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "oidc-provider",
			Usage:   "custom identity provider `NAME`, e.g. Google",
			Value:   "",
			EnvVars: EnvVars("OIDC_PROVIDER"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "oidc-icon",
			Usage:   "custom identity provider icon `URI`",
			Value:   "",
			EnvVars: EnvVars("OIDC_ICON"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "oidc-redirect",
			Usage:   "automatically redirects unauthenticated users to the configured identity provider",
			EnvVars: EnvVars("OIDC_REDIRECT"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "oidc-register",
			Usage:   "allows new users to create an account when they sign in with OpenID Connect",
			EnvVars: EnvVars("OIDC_REGISTER"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "oidc-username",
			Usage:   "preferred username `CLAIM` for new OpenID Connect users (preferred_username, name, nickname, email)",
			Value:   authn.OidcClaimPreferredUsername,
			EnvVars: EnvVars("OIDC_USERNAME"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "oidc-webdav",
			Usage:   "allows new OpenID Connect users to use WebDAV when they have a role that allows it",
			EnvVars: EnvVars("OIDC_WEBDAV"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-oidc",
			Usage:   "disables single sign-on via OpenID Connect, even if an identity provider has been configured",
			EnvVars: EnvVars("DISABLE_OIDC"),
		}}, {
		Flag: &cli.Int64Flag{
			Name:    "session-maxage",
			Value:   DefaultSessionMaxAge,
			Usage:   "session expiration time in `SECONDS`, doubled for accounts with 2FA (-1 to disable)",
			EnvVars: EnvVars("SESSION_MAXAGE"),
		}}, {
		Flag: &cli.Int64Flag{
			Name:    "session-timeout",
			Value:   DefaultSessionTimeout,
			Usage:   "session idle time in `SECONDS`, doubled for accounts with 2FA (-1 to disable)",
			EnvVars: EnvVars("SESSION_TIMEOUT"),
		}}, {
		Flag: &cli.Int64Flag{
			Name:    "session-cache",
			Value:   DefaultSessionCache,
			Usage:   "session cache duration in `SECONDS` (60-3600)",
			EnvVars: EnvVars("SESSION_CACHE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "log-level",
			Aliases: []string{"l"},
			Usage:   "log message verbosity `LEVEL` (trace, debug, info, warning, error)",
			Value:   "info",
			EnvVars: EnvVars("LOG_LEVEL"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "prod",
			Usage:   "disables debug mode and only logs startup warnings and errors",
			EnvVars: EnvVars("PROD"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "debug",
			Usage:   "enables debug mode for development and troubleshooting",
			EnvVars: EnvVars("DEBUG"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "trace",
			Usage:   "enables trace mode to display all debug and trace logs",
			EnvVars: EnvVars("TRACE"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:   "test",
			Hidden: true,
			Usage:  "enables test mode",
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "unsafe",
			Hidden:  true,
			Usage:   "disables safety checks",
			EnvVars: EnvVars("UNSAFE"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "demo",
			Hidden:  true,
			Usage:   "enables demo mode",
			EnvVars: EnvVars("DEMO"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "sponsor",
			Hidden:  true,
			Usage:   "your continuous support helps to pay for development and operating expenses",
			EnvVars: EnvVars("SPONSOR"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "partner-id",
			Hidden:  true,
			Usage:   "hosting partner id",
			EnvVars: EnvVars("PARTNER_ID"),
		}}, {
		Flag: &cli.PathFlag{
			Name:      "config-path",
			Aliases:   []string{"config", "c"},
			Usage:     "config storage `PATH` or options.yml filename, values in this file override CLI flags and environment variables if present",
			EnvVars:   EnvVars("CONFIG_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.StringFlag{
			Name: "defaults-yaml",
			// Alias was changed from "y" to "defaults" since "y" is a reserved alias for "yes".
			// Since our examples and end-user docs for this flag don't include any aliases, the change should be safe.
			Aliases: []string{"defaults"},
			Usage:   "loads default config values from `FILENAME` if it exists, does not override CLI flags or environment variables",
			// fs.ConfigFilePath lets existing installations keep a defaults.yml file
			// while new deployments may drop in defaults.yaml without updating the flag.
			Value:     fs.ConfigFilePath("/etc/photoprism", "defaults", fs.ExtYml),
			EnvVars:   EnvVars("DEFAULTS_YAML"),
			TakesFile: true,
		}}, {
		Flag: &cli.PathFlag{
			Name:      "originals-path",
			Aliases:   []string{"o"},
			Usage:     "storage `PATH` of your original media files (photos and videos)",
			EnvVars:   EnvVars("ORIGINALS_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.IntFlag{
			Name:    "originals-limit",
			Aliases: []string{"mb"},
			Value:   1000,
			Usage:   "maximum size of media files in `MB` (1-100000; -1 to disable)",
			EnvVars: EnvVars("ORIGINALS_LIMIT"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "resolution-limit",
			Aliases: []string{"mp"},
			Value:   DefaultResolutionLimit,
			Usage:   "maximum resolution of media files in `MEGAPIXELS` (1-900; -1 to disable)",
			EnvVars: EnvVars("RESOLUTION_LIMIT"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "users-path",
			Usage:   "relative `PATH` to create base and upload subdirectories for users",
			Value:   fs.UsersDir,
			EnvVars: EnvVars("USERS_PATH"),
		}}, {
		Flag: &cli.PathFlag{
			Name:      "storage-path",
			Aliases:   []string{"s"},
			Usage:     "writable storage `PATH` for sidecar, cache, and database files",
			EnvVars:   EnvVars("STORAGE_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.PathFlag{
			Name:      "import-path",
			Aliases:   []string{"im"},
			Usage:     "base `PATH` from which files can be imported to originals *optional*",
			EnvVars:   EnvVars("IMPORT_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.PathFlag{
			Name:      "import-dest",
			Usage:     "relative originals `PATH` in which files should be imported by default *optional*",
			EnvVars:   EnvVars("IMPORT_DEST"),
			TakesFile: true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "import-allow",
			Usage:   "restricts imports to these file types (comma-separated list of `EXTENSIONS`; leave blank to allow all)",
			EnvVars: EnvVars("IMPORT_ALLOW"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "upload-nsfw",
			Aliases: []string{"n"},
			Usage:   "allows uploads that might be offensive (detecting unsafe content requires TensorFlow)",
			EnvVars: EnvVars("UPLOAD_NSFW"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "upload-allow",
			Usage:   "restricts uploads to these file types (comma-separated list of `EXTENSIONS`; leave blank to allow all)",
			EnvVars: EnvVars("UPLOAD_ALLOW"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "upload-archives",
			Usage:   "allows upload of zip archives (will be extracted before import)",
			EnvVars: EnvVars("UPLOAD_ARCHIVES"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "upload-limit",
			Value:   1000,
			Usage:   "maximum total size of uploaded files in `MB` (1-100000; -1 to disable)",
			EnvVars: EnvVars("UPLOAD_LIMIT"),
		}}, {
		Flag: &cli.PathFlag{
			Name:      "cache-path",
			Aliases:   []string{"ca"},
			Usage:     "custom cache `PATH` for sessions and thumbnail files *optional*",
			EnvVars:   EnvVars("CACHE_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.PathFlag{
			Name:      "temp-path",
			Aliases:   []string{"tmp"},
			Usage:     "temporary file `PATH` *optional*",
			EnvVars:   EnvVars("TEMP_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.PathFlag{
			Name:      "assets-path",
			Aliases:   []string{"as"},
			Usage:     "assets `PATH` containing static resources like icons, models, and translations",
			EnvVars:   EnvVars("ASSETS_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.PathFlag{
			Name:      "theme-path",
			Usage:     "custom user interface theme `PATH` containing styles, scripts, and images",
			EnvVars:   EnvVars("THEME_PATH"),
			TakesFile: true,
			Hidden:    true,
		}}, {
		Flag: &cli.PathFlag{
			Name:      "models-path",
			Usage:     "custom model assets `PATH` where computer vision models are located",
			EnvVars:   EnvVars("MODELS_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.PathFlag{
			Name:    "sidecar-path",
			Aliases: []string{"sc"},
			Usage:   "custom relative or absolute sidecar `PATH` *optional*",
			EnvVars: EnvVars("SIDECAR_PATH"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "sidecar-yaml",
			Usage:   "creates YAML sidecar files to back up picture metadata",
			EnvVars: EnvVars("SIDECAR_YAML"),
		}, DocDefault: "true"}, {
		Flag: &cli.BoolFlag{
			Name:    "usage-info",
			Usage:   "displays storage usage information in the user interface",
			EnvVars: EnvVars("USAGE_INFO"),
		}}, {
		Flag: &cli.Uint64Flag{
			Name:    "files-quota",
			Usage:   "maximum total size of all indexed files in `GB` (0 for unlimited)",
			EnvVars: EnvVars("FILES_QUOTA"),
		}}, {
		Flag: &cli.PathFlag{
			Name:      "backup-path",
			Aliases:   []string{"ba"},
			Usage:     "custom base `PATH` for creating and restoring backups *optional*",
			EnvVars:   EnvVars("BACKUP_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "backup-schedule",
			Usage:   "backup `SCHEDULE` in cron format (e.g. \"0 12 * * *\" for daily at noon) or at a random time (daily, weekly)",
			Value:   DefaultBackupSchedule,
			EnvVars: EnvVars("BACKUP_SCHEDULE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "backup-retain",
			Usage:   "`NUMBER` of index backups to keep (-1 to keep all)",
			Value:   DefaultBackupRetain,
			EnvVars: EnvVars("BACKUP_RETAIN"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "backup-database",
			Usage:   "enables regular backups based on the configured schedule",
			EnvVars: EnvVars("BACKUP_DATABASE"),
		}, DocDefault: "true"}, {
		Flag: &cli.BoolFlag{
			Name:    "backup-albums",
			Usage:   "enables the use of YAML files for backing up album metadata",
			EnvVars: EnvVars("BACKUP_ALBUMS"),
		}, DocDefault: "true"}, {
		Flag: &cli.IntFlag{
			Name:    "index-workers",
			Aliases: []string{"workers"},
			Usage:   "maximum `NUMBER` of indexing workers, default depends on the number of physical cores",
			Value:   cpuid.CPU.PhysicalCores / 2,
			EnvVars: EnvVars("INDEX_WORKERS", "WORKERS"),
		}, DocDefault: " "}, {
		Flag: &cli.StringFlag{
			Name:    "index-schedule",
			Usage:   "indexing `SCHEDULE` in cron format (e.g. \"@every 3h\" for every 3 hours; \"\" to disable)",
			Value:   DefaultIndexSchedule,
			EnvVars: EnvVars("INDEX_SCHEDULE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "wakeup-interval",
			Aliases: []string{"i"},
			Usage:   "`TIME` between facial recognition, file sync, and metadata worker runs (1-86400s)",
			Value:   DefaultWakeupInterval.String(),
			EnvVars: EnvVars("WAKEUP_INTERVAL"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "auto-index",
			Usage:   "delay before automatically indexing files in `SECONDS` when uploading via WebDAV (-1 to disable)",
			Value:   DefaultAutoIndexDelay,
			EnvVars: EnvVars("AUTO_INDEX"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "auto-import",
			Usage:   "delay before automatically importing files in `SECONDS` when uploading via WebDAV (-1 to disable)",
			Value:   DefaultAutoImportDelay,
			EnvVars: EnvVars("AUTO_IMPORT"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "read-only",
			Aliases: []string{"r"},
			Usage:   "disables features that require write permission for the originals folder",
			EnvVars: EnvVars("READONLY"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "experimental",
			Aliases: []string{"e"},
			Usage:   "enables new features that may be incomplete or unstable",
			EnvVars: EnvVars("EXPERIMENTAL"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-frontend",
			Usage:   "disables the web user interface so that only the service API endpoints are accessible",
			EnvVars: EnvVars("DISABLE_FRONTEND"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-settings",
			Usage:   "disables the settings frontend and related API endpoints, e.g. in combination with public mode",
			EnvVars: EnvVars("DISABLE_SETTINGS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-backups",
			Usage:   "prevents database and album backups as well as YAML sidecar files from being created",
			EnvVars: EnvVars("DISABLE_BACKUPS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-restart",
			Usage:   "prevents admins from restarting the server through the user interface",
			EnvVars: EnvVars("DISABLE_RESTART"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-webdav",
			Usage:   "prevents other apps from accessing PhotoPrism as a shared network drive",
			EnvVars: EnvVars("DISABLE_WEBDAV"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-places",
			Usage:   "disables interactive world maps and reverse geocoding",
			EnvVars: EnvVars("DISABLE_PLACES"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-tensorflow",
			Usage:   "disables face recognition with TensorFlow *deprecated*",
			EnvVars: EnvVars("DISABLE_TENSORFLOW"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-faces",
			Usage:   "disables face detection and recognition (requires TensorFlow)",
			EnvVars: EnvVars("DISABLE_FACES"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-classification",
			Usage:   "disables all image classification and label generation",
			EnvVars: EnvVars("DISABLE_CLASSIFICATION"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-ffmpeg",
			Usage:   "disables video transcoding and thumbnail extraction with FFmpeg",
			EnvVars: EnvVars("DISABLE_FFMPEG"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-exiftool",
			Usage:   "disables metadata extraction with ExifTool (required for full Video, Live Photo, and XMP support)",
			EnvVars: EnvVars("DISABLE_EXIFTOOL"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-vips",
			Usage:   "disables image processing and conversion with libvips",
			EnvVars: EnvVars("DISABLE_VIPS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-sips",
			Usage:   "disables file conversion using the sips command under macOS",
			EnvVars: EnvVars("DISABLE_SIPS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-darktable",
			Usage:   "disables conversion of RAW images with Darktable",
			EnvVars: EnvVars("DISABLE_DARKTABLE"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-rawtherapee",
			Usage:   "disables conversion of RAW images with RawTherapee",
			EnvVars: EnvVars("DISABLE_RAWTHERAPEE"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-imagemagick",
			Usage:   "disables conversion of image files with ImageMagick",
			EnvVars: EnvVars("DISABLE_IMAGEMAGICK"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-heifconvert",
			Usage:   "disables conversion of HEIC images with libheif",
			EnvVars: EnvVars("DISABLE_HEIFCONVERT"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-jpegxl",
			Usage:   "disables JPEG XL file format support",
			EnvVars: EnvVars("DISABLE_JPEGXL"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-raw",
			Usage:   "disables indexing and conversion of RAW images",
			EnvVars: EnvVars("DISABLE_RAW"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "raw-presets",
			Usage:   "enables custom user presets when converting RAW images (reduces performance)",
			EnvVars: EnvVars("RAW_PRESETS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "exif-bruteforce",
			Usage:   "performs a brute-force search if no Exif headers were found",
			EnvVars: EnvVars("EXIF_BRUTEFORCE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "default-locale",
			Aliases: []string{"lang"},
			Usage:   "default user interface language `CODE`",
			Value:   i18n.Default.Locale(),
			EnvVars: EnvVars("DEFAULT_LOCALE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "default-timezone",
			Aliases: []string{"tz"},
			Usage:   "default time zone `NAME`, e.g. for scheduling backups",
			Value:   tz.Local,
			EnvVars: EnvVars("DEFAULT_TIMEZONE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "default-theme",
			Usage:   "default user interface theme `NAME`",
			EnvVars: EnvVars("DEFAULT_THEME"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "places-locale",
			Usage:   "location details language `CODE`, e.g. en, de, or local",
			Value:   places.LocalLocale,
			EnvVars: EnvVars("PLACES_LOCALE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "app-name",
			Usage:   "app `NAME` when installed as a Progressive Web App (PWA)",
			Value:   "",
			EnvVars: EnvVars("APP_NAME"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "app-mode",
			Usage:   "app display `MODE` (fullscreen, standalone, minimal-ui, browser)",
			Value:   "standalone",
			EnvVars: EnvVars("APP_MODE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "app-icon",
			Usage:   "home screen app `ICON` (logo, app, crisp, mint, bold, square)",
			EnvVars: EnvVars("APP_ICON"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "app-color",
			Usage:   "app background and splash screen `COLOR`",
			Value:   DefaultAppColor,
			EnvVars: EnvVars("APP_COLOR"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "imprint",
			Usage:   "legal information `TEXT`, displayed in the page footer",
			Value:   "",
			Hidden:  true,
			EnvVars: EnvVars("IMPRINT"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "legal-info",
			Usage:   "legal information `TEXT`, displayed in the page footer",
			Value:   "",
			EnvVars: EnvVars("LEGAL_INFO"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "imprint-url",
			Usage:   "legal information `URL`",
			Value:   "",
			Hidden:  true,
			EnvVars: EnvVars("IMPRINT_URL"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "legal-url",
			Usage:   "legal information `URL`",
			Value:   "",
			EnvVars: EnvVars("LEGAL_URL"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "wallpaper-uri",
			Usage:   "login screen background image `URI`",
			Value:   "",
			EnvVars: EnvVars("WALLPAPER_URI"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "site-url",
			Usage:   "canonical site `URL` used in generated links and to determine HTTPS/TLS (scheme://host[:port])",
			Value:   "http://localhost:2342/",
			EnvVars: EnvVars("SITE_URL"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "site-author",
			Usage:   "site `OWNER`, copyright, or artist",
			EnvVars: EnvVars("SITE_AUTHOR"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "site-title",
			Usage:   "site `TITLE`",
			Value:   "",
			EnvVars: EnvVars("SITE_TITLE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "site-caption",
			Usage:   "site `CAPTION`",
			Value:   "AI-Powered Photos App",
			EnvVars: EnvVars("SITE_CAPTION"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "site-description",
			Usage:   "site `DESCRIPTION` *optional*",
			EnvVars: EnvVars("SITE_DESCRIPTION"),
		}}, {
		Flag: &cli.StringFlag{
			Name:      "site-favicon",
			Usage:     "site favicon `FILENAME` *optional*",
			EnvVars:   EnvVars("SITE_FAVICON"),
			TakesFile: true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "site-preview",
			Usage:   "sharing preview image `URL`",
			EnvVars: EnvVars("SITE_PREVIEW"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "cdn-url",
			Usage:   "content delivery network `URL`",
			EnvVars: EnvVars("CDN_URL"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "cdn-video",
			Usage:   "streams videos over the specified CDN",
			EnvVars: EnvVars("CDN_VIDEO"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "cors-origin",
			Usage:   "origin `URL` from which browsers are allowed to perform cross-origin requests (leave blank to disable or use * to allow all)",
			EnvVars: EnvVars("CORS_ORIGIN"),
			Value:   header.DefaultAccessControlAllowOrigin,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "cors-headers",
			Usage:   "one or more `HEADERS` that browsers should see when performing a cross-origin request",
			EnvVars: EnvVars("CORS_HEADERS"),
			Value:   header.DefaultAccessControlAllowHeaders,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "cors-methods",
			Usage:   "one or more `METHODS` that may be used when performing a cross-origin request",
			EnvVars: EnvVars("CORS_METHODS"),
			Value:   header.DefaultAccessControlAllowMethods,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "cluster-domain",
			Usage:   "cluster `DOMAIN` (lowercase DNS name; 1–63 chars)",
			EnvVars: EnvVars("CLUSTER_DOMAIN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "cluster-uuid",
			Usage:   "cluster `UUID` (v4) to scope node credentials",
			EnvVars: EnvVars("CLUSTER_UUID"),
			Hidden:  true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "cluster-cidr",
			Usage:   "cluster `CIDR` (e.g., 10.0.0.0/8) for IP-based authorization",
			EnvVars: EnvVars("CLUSTER_CIDR"),
			Hidden:  true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "portal-url",
			Usage:   "base `URL` of the cluster management portal",
			Value:   DefaultPortalUrl,
			EnvVars: EnvVars("PORTAL_URL"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "join-token",
			Usage:   "secret `TOKEN` required to join a cluster; min 24 chars",
			EnvVars: EnvVars("JOIN_TOKEN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "node-name",
			Usage:   "node `NAME` (unique in cluster domain; [a-z0-9-]{1,32})",
			EnvVars: EnvVars("NODE_NAME"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "node-role",
			Usage:   fmt.Sprintf("node `ROLE` (%s or %s)", cluster.RoleApp, cluster.RoleService),
			EnvVars: EnvVars("NODE_ROLE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "node-uuid",
			Usage:   "node `UUID` (v7) that uniquely identifies this instance",
			EnvVars: EnvVars("NODE_UUID"),
			Hidden:  true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "node-client-id",
			Usage:   "node OAuth client `ID` (auto-assigned via join token)",
			EnvVars: EnvVars("NODE_CLIENT_ID"),
			Hidden:  true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "node-client-secret",
			Usage:   "node OAuth client `SECRET` (auto-assigned via join token)",
			EnvVars: EnvVars("NODE_CLIENT_SECRET"),
			Hidden:  true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "jwks-url",
			Usage:   "JWKS endpoint `URL` provided by the cluster portal for JWT verification",
			EnvVars: EnvVars("JWKS_URL"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "jwks-cache-ttl",
			Usage:   "JWKS cache lifetime in `SECONDS` (default 300, max 3600)",
			Value:   300,
			EnvVars: EnvVars("JWKS_CACHE_TTL"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "jwt-scope",
			Usage:   "allowed JWT `SCOPES` (space separated). Leave empty to accept defaults",
			Value:   DefaultJWTAllowedScopes,
			EnvVars: EnvVars("JWT_SCOPE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "jwt-leeway",
			Usage:   "JWT clock skew allowance in `SECONDS` (default 60, max 300)",
			Value:   60,
			EnvVars: EnvVars("JWT_LEEWAY"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "advertise-url",
			Usage:   "advertised `URL` for intra-cluster calls (scheme://host[:port])",
			Value:   "",
			EnvVars: EnvVars("ADVERTISE_URL"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "https-proxy",
			Usage:   "proxy server `URL` to be used for outgoing connections *optional*",
			EnvVars: EnvVars("HTTPS_PROXY"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "https-proxy-insecure",
			Usage:   "ignores invalid HTTPS certificates when using a proxy",
			EnvVars: EnvVars("HTTPS_PROXY_INSECURE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "trusted-platform",
			Usage:   "trusted client IP header `NAME`, e.g. when running behind a cloud provider load balancer",
			Value:   "",
			EnvVars: EnvVars("TRUSTED_PLATFORM"),
		}}, {
		Flag: &cli.StringSliceFlag{
			Name:    "trusted-proxy",
			Usage:   "`CIDR` ranges or IPv4/v6 addresses from which reverse proxy headers can be trusted, separated by commas",
			Value:   cli.NewStringSlice(header.CidrDockerInternal),
			EnvVars: EnvVars("TRUSTED_PROXY"),
		}}, {
		Flag: &cli.StringSliceFlag{
			Name:    "proxy-client-header",
			Usage:   "proxy client IP header `NAME`, e.g. X-Forwarded-For, X-Client-IP, X-Real-IP, or CF-Connecting-IP",
			Value:   cli.NewStringSlice(header.XForwardedFor),
			EnvVars: EnvVars("PROXY_CLIENT_HEADER"),
		}}, {
		Flag: &cli.StringSliceFlag{
			Name:    "proxy-proto-header",
			Usage:   "proxy protocol header `NAME`",
			Value:   cli.NewStringSlice(header.XForwardedProto),
			EnvVars: EnvVars("PROXY_PROTO_HEADER"),
		}}, {
		Flag: &cli.StringSliceFlag{
			Name:    "proxy-proto-https",
			Usage:   "forwarded HTTPS protocol `NAME`",
			Value:   cli.NewStringSlice(scheme.Https),
			EnvVars: EnvVars("PROXY_PROTO_HTTPS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-tls",
			Usage:   "disables HTTPS/TLS even if the site URL starts with https:// and a certificate is available",
			EnvVars: EnvVars("DISABLE_TLS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "default-tls",
			Usage:   "uses a self-signed HTTPS/TLS certificate if no other certificate is available",
			EnvVars: EnvVars("DEFAULT_TLS"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "tls-email",
			Usage:   "`EMAIL` address to enable automatic HTTPS via Let's Encrypt",
			EnvVars: EnvVars("TLS_EMAIL"),
			Hidden:  true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "tls-cert",
			Usage:   "public HTTPS certificate `FILENAME` (.crt), ignored for Unix domain sockets",
			EnvVars: EnvVars("TLS_CERT"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "tls-key",
			Usage:   "private HTTPS key `FILENAME` (.key), ignored for Unix domain sockets",
			EnvVars: EnvVars("TLS_KEY"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "http-mode",
			Aliases: []string{"mode"},
			Usage:   "Web server `MODE` (debug, release, test)",
			EnvVars: EnvVars("HTTP_MODE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "http-compression",
			Aliases: []string{"z"},
			Usage:   "Web server compression `METHOD` (gzip, none)",
			EnvVars: EnvVars("HTTP_COMPRESSION"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "http-cache-public",
			Usage:   "allows static content to be cached by a CDN or caching proxy",
			EnvVars: EnvVars("HTTP_CACHE_PUBLIC"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "http-cache-maxage",
			Value:   int(ttl.CacheDefault),
			Usage:   "time in `SECONDS` until cached content expires",
			EnvVars: EnvVars("HTTP_CACHE_MAXAGE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "http-video-maxage",
			Value:   int(ttl.CacheVideo),
			Usage:   "time in `SECONDS` until cached videos expire",
			EnvVars: EnvVars("HTTP_VIDEO_MAXAGE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "http-host",
			Aliases: []string{"ip"},
			Value:   "0.0.0.0",
			Usage:   "Web server `IP` address or Unix domain socket, e.g. unix:/var/run/photoprism.sock?force=true&mode=660",
			EnvVars: EnvVars("HTTP_HOST"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "http-port",
			Aliases: []string{"port"},
			Value:   2342,
			Usage:   "Web server port `NUMBER`, ignored for Unix domain sockets",
			EnvVars: EnvVars("HTTP_PORT"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "database-driver",
			Aliases: []string{"db"},
			Usage:   "database `DRIVER` (sqlite, mysql)",
			Value:   "sqlite",
			EnvVars: EnvVars("DATABASE_DRIVER"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "database-dsn",
			Aliases: []string{"dsn"},
			Usage:   "database connection `DSN` (sqlite file, optional for mysql)",
			EnvVars: EnvVars("DATABASE_DSN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "database-name",
			Aliases: []string{"db-name"},
			Value:   "photoprism",
			Usage:   "database schema `NAME`",
			EnvVars: EnvVars("DATABASE_NAME"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "database-server",
			Aliases: []string{"db-server"},
			Usage:   "database `HOST` incl. port, e.g. \"mariadb:3306\" (or socket path)",
			EnvVars: EnvVars("DATABASE_SERVER"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "database-user",
			Aliases: []string{"db-user"},
			Value:   "photoprism",
			Usage:   "database user `NAME`",
			EnvVars: EnvVars("DATABASE_USER"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "database-password",
			Aliases: []string{"db-pass"},
			Usage:   "database user `PASSWORD`",
			EnvVars: EnvVars("DATABASE_PASSWORD"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "database-timeout",
			Usage:   "timeout in `SECONDS` for establishing a database connection (1-60)",
			EnvVars: EnvVars("DATABASE_TIMEOUT"),
			Value:   15,
		}}, {
		Flag: &cli.IntFlag{
			Name:    "database-conns",
			Usage:   "maximum `NUMBER` of open database connections",
			EnvVars: EnvVars("DATABASE_CONNS"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "database-conns-idle",
			Usage:   "maximum `NUMBER` of idle database connections",
			EnvVars: EnvVars("DATABASE_CONNS_IDLE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "database-provision-driver",
			Usage:   "auto-provisioning `DRIVER` (auto, mysql)",
			Value:   Auto,
			EnvVars: EnvVars("DATABASE_PROVISION_DRIVER"),
			Hidden:  true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "database-provision-prefix",
			Usage:   "auto-provisioning name `PREFIX` for generated database names and users",
			Value:   cluster.DefaultDatabaseProvisionPrefix,
			EnvVars: EnvVars("DATABASE_PROVISION_PREFIX"),
			Hidden:  true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "database-provision-dsn",
			Usage:   "auto-provisioning `DSN`",
			EnvVars: EnvVars("DATABASE_PROVISION_DSN"),
			Hidden:  true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "database-provision-proxy-dsn",
			Usage:   "ProxySQL admin `DSN` (port 6032 by default) for keeping user accounts in sync",
			EnvVars: EnvVars("DATABASE_PROVISION_PROXY_DSN"),
			Hidden:  true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "ffmpeg-bin",
			Usage:   "FFmpeg `COMMAND` for video transcoding and thumbnail extraction",
			Value:   encode.FFmpegBin,
			EnvVars: EnvVars("FFMPEG_BIN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "ffmpeg-encoder",
			Aliases: []string{"vc"},
			Usage:   "FFmpeg AVC video encoder `NAME`",
			Value:   "libx264",
			EnvVars: EnvVars("FFMPEG_ENCODER"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "ffmpeg-size",
			Usage:   "encoding resolution limit in `PIXELS` (720-7680)",
			Value:   thumb.Sizes[thumb.Fit4096].Width,
			EnvVars: EnvVars("FFMPEG_SIZE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "ffmpeg-quality",
			Usage:   fmt.Sprintf("encoding `QUALITY` (%d-%d, where %d is almost lossless)", encode.WorstQuality, encode.BestQuality, encode.BestQuality),
			Value:   encode.DefaultQuality,
			EnvVars: EnvVars("FFMPEG_QUALITY"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "ffmpeg-bitrate",
			Usage:   fmt.Sprintf("bitrate `LIMIT` in Mbps for forced transcoding of non-AVC videos (%d-%d; %d to disable)", encode.MinBitrateLimit, encode.MaxBitrateLimit, encode.NoBitrateLimit),
			Value:   encode.DefaultBitrateLimit,
			EnvVars: EnvVars("FFMPEG_BITRATE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "ffmpeg-preset",
			Usage:   "FFmpeg compression `PRESET` when using an encoder that supports it, e.g. fast, medium, or slow",
			Value:   encode.PresetFast,
			EnvVars: EnvVars("FFMPEG_PRESET"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "ffmpeg-device",
			Usage:   "FFmpeg device `PATH` when using a hardware encoder that supports it as parameter",
			EnvVars: EnvVars("FFMPEG_DEVICE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "ffmpeg-map-video",
			Usage:   "transcoding video stream `MAP`",
			Value:   encode.DefaultMapVideo,
			EnvVars: EnvVars("FFMPEG_MAP_VIDEO"),
		}, DocDefault: fmt.Sprintf("`%s`", encode.DefaultMapVideo)}, {
		Flag: &cli.StringFlag{
			Name:    "ffmpeg-map-audio",
			Usage:   "transcoding audio stream `MAP`",
			Value:   encode.DefaultMapAudio,
			EnvVars: EnvVars("FFMPEG_MAP_AUDIO"),
		}, DocDefault: fmt.Sprintf("`%s`", encode.DefaultMapAudio)}, {
		Flag: &cli.StringFlag{
			Name:    "exiftool-bin",
			Usage:   "ExifTool `COMMAND` for extracting metadata",
			Value:   "exiftool",
			EnvVars: EnvVars("EXIFTOOL_BIN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "sips-bin",
			Usage:   "Sips `COMMAND` for media file conversion *macOS only*",
			Value:   "sips",
			EnvVars: EnvVars("SIPS_BIN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "sips-exclude",
			Usage:   "file `EXTENSIONS` not to be used with Sips *macOS only*",
			Value:   "avif, avifs, thm",
			EnvVars: EnvVars("SIPS_EXCLUDE", "SIPS_BLACKLIST"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "darktable-bin",
			Usage:   "Darktable CLI `COMMAND` for RAW to JPEG conversion",
			Value:   "darktable-cli",
			EnvVars: EnvVars("DARKTABLE_BIN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "darktable-exclude",
			Usage:   "file `EXTENSIONS` not to be used with Darktable",
			Value:   "thm",
			EnvVars: EnvVars("DARKTABLE_EXCLUDE", "DARKTABLE_BLACKLIST"),
		}}, {
		Flag: &cli.PathFlag{
			Name:      "darktable-cache-path",
			Usage:     "custom Darktable cache `PATH`",
			Value:     "",
			EnvVars:   EnvVars("DARKTABLE_CACHE_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.PathFlag{
			Name:      "darktable-config-path",
			Usage:     "custom Darktable config `PATH`",
			Value:     "",
			EnvVars:   EnvVars("DARKTABLE_CONFIG_PATH"),
			TakesFile: true,
		}}, {
		Flag: &cli.StringFlag{
			Name:    "rawtherapee-bin",
			Usage:   "RawTherapee CLI `COMMAND` for RAW to JPEG conversion",
			Value:   "rawtherapee-cli",
			EnvVars: EnvVars("RAWTHERAPEE_BIN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "rawtherapee-exclude",
			Usage:   "file `EXTENSIONS` not to be used with RawTherapee",
			Value:   "dng, thm",
			EnvVars: EnvVars("RAWTHERAPEE_EXCLUDE", "RAWTHERAPEE_BLACKLIST"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "imagemagick-bin",
			Usage:   "ImageMagick CLI `COMMAND` for image file conversion",
			Value:   "convert",
			EnvVars: EnvVars("IMAGEMAGICK_BIN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "imagemagick-exclude",
			Usage:   "file `EXTENSIONS` not to be used with ImageMagick",
			Value:   "heif, heic, heics, avif, avifs, jxl, thm",
			EnvVars: EnvVars("IMAGEMAGICK_EXCLUDE", "IMAGEMAGICK_BLACKLIST"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "heifconvert-bin",
			Usage:   "libheif HEIC image conversion `COMMAND`",
			Value:   "",
			EnvVars: EnvVars("HEIFCONVERT_BIN"),
		},
		DocDefault: "heif-dec"}, {
		Flag: &cli.StringFlag{
			Name:    "heifconvert-orientation",
			Usage:   "Exif `ORIENTATION` of images generated with libheif (keep, reset)",
			Value:   media.KeepOrientation,
			EnvVars: EnvVars("HEIFCONVERT_ORIENTATION"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "download-token",
			Usage:   "`DEFAULT` download URL token for originals (leave blank for a random value)",
			EnvVars: EnvVars("DOWNLOAD_TOKEN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "preview-token",
			Usage:   "`DEFAULT` thumbnail and video streaming URL token (leave blank for a random value)",
			EnvVars: EnvVars("PREVIEW_TOKEN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "thumb-library",
			Aliases: []string{"thumbs"},
			Usage:   "image processing `LIBRARY` to be used for generating thumbnails (auto, imaging, vips)",
			Value:   Auto,
			EnvVars: EnvVars("THUMB_LIBRARY"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "thumb-color",
			Usage:   "standard color `PROFILE` for thumbnails (auto, preserve, srgb, none)",
			Value:   thumb.ColorAuto,
			EnvVars: EnvVars("THUMB_COLOR"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "thumb-filter",
			Aliases: []string{"filter"},
			Usage:   "downscaling filter `NAME` (imaging best to worst: blackman, lanczos, cubic, linear, nearest)",
			Value:   thumb.ResampleAuto.String(),
			EnvVars: EnvVars("THUMB_FILTER"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "thumb-size",
			Usage:   "maximum size of pre-generated thumbnails in `PIXELS` (720-7680)",
			Value:   thumb.SizeCached,
			EnvVars: EnvVars("THUMB_SIZE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "thumb-size-uncached",
			Usage:   "maximum size of thumbnails generated on demand in `PIXELS` (720-7680)",
			Value:   thumb.SizeOnDemand,
			EnvVars: EnvVars("THUMB_SIZE_UNCACHED"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "thumb-uncached",
			Aliases: []string{"u"},
			Usage:   "generates missing thumbnails on demand (high memory and cpu usage)",
			EnvVars: EnvVars("THUMB_UNCACHED"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "jpeg-quality",
			Aliases: []string{"q"},
			Usage:   "higher values increase the image `QUALITY` and file size (25-100)",
			Value:   thumb.QualityMedium.String(),
			EnvVars: EnvVars("JPEG_QUALITY"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "jpeg-size",
			Usage:   "maximum size of generated JPEG images in `PIXELS` (720-30000)",
			Value:   7680,
			EnvVars: EnvVars("JPEG_SIZE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "png-size",
			Usage:   "maximum size of generated PNG images in `PIXELS` (720-30000)",
			Value:   7680,
			EnvVars: EnvVars("PNG_SIZE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:      "vision-yaml",
			Usage:     "computer vision model configuration `FILENAME` *optional*",
			Value:     "",
			EnvVars:   EnvVars("VISION_YAML"),
			TakesFile: true,
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "vision-api",
			Usage:   "enables the computer vision API endpoints under /api/v1/vision (requires authorization)",
			EnvVars: EnvVars("VISION_API"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "vision-uri",
			Usage:   "vision service base `URI`, e.g. https://example.com/api/v1/vision (leave blank to disable)",
			Value:   "",
			EnvVars: EnvVars("VISION_URI"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "vision-key",
			Usage:   "vision service access `TOKEN` *optional*",
			Value:   "",
			EnvVars: EnvVars("VISION_KEY"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "vision-schedule",
			Usage:   "vision worker `SCHEDULE` for background processing (e.g. \"0 12 * * *\" for daily at noon) or at a random time (daily, weekly)",
			EnvVars: EnvVars("VISION_SCHEDULE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "vision-filter",
			Usage:   "vision worker search `FILTER` applied to scheduled runs (same syntax as photoprism vision run)",
			Value:   "public:true",
			EnvVars: EnvVars("VISION_FILTER"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "detect-nsfw",
			Usage:   "flags newly added pictures as private if they might be offensive (requires TensorFlow)",
			EnvVars: EnvVars("DETECT_NSFW"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "face-engine",
			Usage:   "face detection engine `NAME` (auto, pigo, onnx)",
			Value:   face.EngineAuto,
			EnvVars: EnvVars("FACE_ENGINE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "face-engine-threads",
			Usage:   "face detection thread `COUNT` (0 uses half the available CPU cores)",
			EnvVars: EnvVars("FACE_ENGINE_THREADS"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "face-size",
			Usage:   "minimum size of faces in `PIXELS` (20-10000)",
			Value:   face.SizeThreshold,
			EnvVars: EnvVars("FACE_SIZE"),
		}}, {
		Flag: &cli.Float64Flag{
			Name:    "face-score",
			Usage:   "minimum face `QUALITY` score (1-100)",
			Value:   face.ScoreThreshold,
			EnvVars: EnvVars("FACE_SCORE"),
		}}, {
		Flag: &cli.Float64SliceFlag{
			Name:    "face-angle",
			Usage:   "face detection `ANGLE` in radians (repeatable)",
			Value:   cli.NewFloat64Slice(face.DefaultAngles...),
			EnvVars: EnvVars("FACE_ANGLE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "face-overlap",
			Usage:   "face area overlap threshold in `PERCENT` (1-100)",
			Value:   face.OverlapThreshold,
			EnvVars: EnvVars("FACE_OVERLAP"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "face-cluster-size",
			Usage:   "minimum size of automatically clustered faces in `PIXELS` (20-10000)",
			Value:   face.ClusterSizeThreshold,
			EnvVars: EnvVars("FACE_CLUSTER_SIZE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "face-cluster-score",
			Usage:   "minimum `QUALITY` score of automatically clustered faces (1-100)",
			Value:   face.ClusterScoreThreshold,
			EnvVars: EnvVars("FACE_CLUSTER_SCORE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "face-cluster-core",
			Usage:   "`NUMBER` of faces forming a cluster core (1-100)",
			Value:   face.ClusterCore,
			EnvVars: EnvVars("FACE_CLUSTER_CORE"),
		}}, {
		Flag: &cli.Float64Flag{
			Name:    "face-cluster-dist",
			Usage:   "similarity `DISTANCE` of faces forming a cluster core (0.1-1.5)",
			Value:   face.ClusterDist,
			EnvVars: EnvVars("FACE_CLUSTER_DIST"),
		}}, {
		Flag: &cli.Float64Flag{
			Name:    "face-cluster-radius",
			Usage:   "maximum cluster `RADIUS` accepted for automatic matches (0.1-1.5)",
			Value:   face.ClusterRadius,
			EnvVars: EnvVars("FACE_CLUSTER_RADIUS"),
		}}, {
		Flag: &cli.Float64Flag{
			Name:    "face-collision-dist",
			Usage:   "minimum collision discrimination `DISTANCE` (0.01-1)",
			Value:   face.CollisionDist,
			EnvVars: EnvVars("FACE_COLLISION_DIST"),
		}}, {
		Flag: &cli.Float64Flag{
			Name:    "face-epsilon-dist",
			Usage:   "collision tolerance `DELTA` appended to max match distances (0.001-0.1)",
			Value:   face.Epsilon,
			EnvVars: EnvVars("FACE_EPSILON_DIST"),
		}}, {
		Flag: &cli.Float64Flag{
			Name:    "face-match-dist",
			Usage:   "similarity `OFFSET` for matching faces with existing clusters (0.1-1.5)",
			Value:   face.MatchDist,
			EnvVars: EnvVars("FACE_MATCH_DIST"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "face-skip-children",
			Usage:   "skips automatic matching of child face embeddings",
			EnvVars: EnvVars("FACE_SKIP_CHILDREN"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "face-allow-background",
			Usage:   "allows matching of probable background embeddings",
			EnvVars: EnvVars("FACE_ALLOW_BACKGROUND"),
		}}, {
		Flag: &cli.StringFlag{
			Name:      "pid-filename",
			Usage:     "process id `FILENAME` *daemon-mode only*",
			EnvVars:   EnvVars("PID_FILENAME"),
			TakesFile: true,
		}}, {
		Flag: &cli.StringFlag{
			Name:      "log-filename",
			Usage:     "server log `FILENAME` *daemon-mode only*",
			Value:     "",
			EnvVars:   EnvVars("LOG_FILENAME"),
			TakesFile: true,
		}},
}
