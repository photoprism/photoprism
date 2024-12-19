package config

import (
	"fmt"
	"time"

	"github.com/klauspost/cpuid/v2"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config/ttl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/media"
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
		Flag: &cli.BoolFlag{
			Name:    "public",
			Aliases: []string{"p"},
			Hidden:  true,
			Usage:   "disable authentication, advanced settings, and WebDAV remote access",
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
			Usage:   "automatically redirect unauthenticated users to the configured identity provider",
			EnvVars: EnvVars("OIDC_REDIRECT"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "oidc-register",
			Usage:   "allow new users to create an account when they sign in with OpenID Connect",
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
			Usage:   "allow new OpenID Connect users to use WebDAV when they have a role that allows it",
			EnvVars: EnvVars("OIDC_WEBDAV"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-oidc",
			Usage:   "disable single sign-on via OpenID Connect, even if an identity provider has been configured",
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
			Usage:   "log message verbosity `LEVEL` (trace, debug, info, warning, error, fatal, panic)",
			Value:   "info",
			EnvVars: EnvVars("LOG_LEVEL"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "prod",
			Hidden:  true,
			Usage:   "enable production mode, hide non-essential log messages",
			EnvVars: EnvVars("PROD"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "debug",
			Usage:   "enable debug mode, show non-essential log messages",
			EnvVars: EnvVars("DEBUG"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "trace",
			Usage:   "enable trace mode, show all log messages",
			EnvVars: EnvVars("TRACE"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:   "test",
			Hidden: true,
			Usage:  "enable test mode",
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "unsafe",
			Hidden:  true,
			Usage:   "disable safety checks",
			EnvVars: EnvVars("UNSAFE"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "demo",
			Hidden:  true,
			Usage:   "enable demo mode",
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
		Flag: &cli.StringFlag{
			Name:    "config-path",
			Aliases: []string{"c"},
			Usage:   "config storage `PATH`, values in options.yml override CLI flags and environment variables if present",
			EnvVars: EnvVars("CONFIG_PATH"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "defaults-yaml",
			Aliases: []string{"y"},
			Usage:   "load config defaults from `FILE` if exists, does not override CLI flags and environment variables",
			Value:   "/etc/photoprism/defaults.yml",
			EnvVars: EnvVars("DEFAULTS_YAML"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "originals-path",
			Aliases: []string{"o"},
			Usage:   "storage `PATH` of your original media files (photos and videos)",
			EnvVars: EnvVars("ORIGINALS_PATH"),
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
			Value:   "users",
			EnvVars: EnvVars("USERS_PATH"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "storage-path",
			Aliases: []string{"s"},
			Usage:   "writable storage `PATH` for sidecar, cache, and database files",
			EnvVars: EnvVars("STORAGE_PATH"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "import-path",
			Aliases: []string{"im"},
			Usage:   "base `PATH` from which files can be imported to originals *optional*",
			EnvVars: EnvVars("IMPORT_PATH"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "import-dest",
			Usage:   "relative originals `PATH` to which the files should be imported by default *optional*",
			EnvVars: EnvVars("IMPORT_DEST"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "cache-path",
			Aliases: []string{"ca"},
			Usage:   "custom cache `PATH` for sessions and thumbnail files *optional*",
			EnvVars: EnvVars("CACHE_PATH"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "temp-path",
			Aliases: []string{"tmp"},
			Usage:   "temporary file `PATH` *optional*",
			EnvVars: EnvVars("TEMP_PATH"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "assets-path",
			Aliases: []string{"as"},
			Usage:   "assets `PATH` containing static resources like icons, models, and translations",
			EnvVars: EnvVars("ASSETS_PATH"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "sidecar-path",
			Aliases: []string{"sc"},
			Usage:   "custom relative or absolute sidecar `PATH` *optional*",
			EnvVars: EnvVars("SIDECAR_PATH"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "sidecar-yaml",
			Usage:   "create YAML sidecar files to back up picture metadata",
			EnvVars: EnvVars("SIDECAR_YAML"),
		}, DocDefault: "true"}, {
		Flag: &cli.StringFlag{
			Name:    "backup-path",
			Aliases: []string{"ba"},
			Usage:   "custom base `PATH` for creating and restoring backups *optional*",
			EnvVars: EnvVars("BACKUP_PATH"),
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
			Usage:   "create regular backups based on the configured schedule",
			EnvVars: EnvVars("BACKUP_DATABASE"),
		}, DocDefault: "true"}, {
		Flag: &cli.BoolFlag{
			Name:    "backup-albums",
			Usage:   "create YAML files to back up album metadata",
			EnvVars: EnvVars("BACKUP_ALBUMS"),
		}, DocDefault: "true"}, {
		Flag: &cli.IntFlag{
			Name:    "index-workers",
			Aliases: []string{"workers"},
			Usage:   "maximum `NUMBER` of indexing workers, default depends on the number of physical cores",
			Value:   cpuid.CPU.PhysicalCores / 2,
			EnvVars: EnvVars("INDEX_WORKERS", "WORKERS"),
		}}, {
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
			Usage:   "disable features that require write permission for the originals folder",
			EnvVars: EnvVars("READONLY"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "experimental",
			Aliases: []string{"e"},
			Usage:   "enable new features currently under development",
			EnvVars: EnvVars("EXPERIMENTAL"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-settings",
			Usage:   "disable the settings user interface and server API, e.g. in combination with public mode",
			EnvVars: EnvVars("DISABLE_SETTINGS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-backups",
			Usage:   "prevent database and album backups as well as YAML sidecar files from being created",
			EnvVars: EnvVars("DISABLE_BACKUPS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-restart",
			Usage:   "prevent admins from restarting the server through the user interface",
			EnvVars: EnvVars("DISABLE_RESTART"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-webdav",
			Usage:   "prevent other apps from accessing PhotoPrism as a shared network drive",
			EnvVars: EnvVars("DISABLE_WEBDAV"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-places",
			Usage:   "disable interactive world maps and reverse geocoding",
			EnvVars: EnvVars("DISABLE_PLACES"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-tensorflow",
			Usage:   "disable features depending on TensorFlow, e.g. image classification and face recognition",
			EnvVars: EnvVars("DISABLE_TENSORFLOW"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-faces",
			Usage:   "disable face detection and recognition (requires TensorFlow)",
			EnvVars: EnvVars("DISABLE_FACES"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-classification",
			Usage:   "disable image classification (requires TensorFlow)",
			EnvVars: EnvVars("DISABLE_CLASSIFICATION"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-ffmpeg",
			Usage:   "disable video transcoding and thumbnail extraction with FFmpeg",
			EnvVars: EnvVars("DISABLE_FFMPEG"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-exiftool",
			Usage:   "disable metadata extraction with ExifTool (required for full Video, Live Photo, and XMP support)",
			EnvVars: EnvVars("DISABLE_EXIFTOOL"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-vips",
			Usage:   "disable image processing and conversion with libvips",
			EnvVars: EnvVars("DISABLE_VIPS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-sips",
			Usage:   "disable file conversion using the sips command under macOS",
			EnvVars: EnvVars("DISABLE_SIPS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-darktable",
			Usage:   "disable conversion of RAW images with Darktable",
			EnvVars: EnvVars("DISABLE_DARKTABLE"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-rawtherapee",
			Usage:   "disable conversion of RAW images with RawTherapee",
			EnvVars: EnvVars("DISABLE_RAWTHERAPEE"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-imagemagick",
			Usage:   "disable conversion of image files with ImageMagick",
			EnvVars: EnvVars("DISABLE_IMAGEMAGICK"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-heifconvert",
			Usage:   "disable conversion of HEIC images with libheif",
			EnvVars: EnvVars("DISABLE_HEIFCONVERT"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-jpegxl",
			Usage:   "disable JPEG XL file format support",
			EnvVars: EnvVars("DISABLE_JPEGXL"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-raw",
			Usage:   "disable indexing and conversion of RAW images",
			EnvVars: EnvVars("DISABLE_RAW"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "raw-presets",
			Usage:   "enables applying user presets when converting RAW images (reduces performance)",
			EnvVars: EnvVars("RAW_PRESETS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "exif-bruteforce",
			Usage:   "always perform a brute-force search if no Exif headers were found",
			EnvVars: EnvVars("EXIF_BRUTEFORCE"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "detect-nsfw",
			Usage:   "flag newly added pictures as private if they might be offensive (requires TensorFlow)",
			EnvVars: EnvVars("DETECT_NSFW"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "upload-nsfw",
			Aliases: []string{"n"},
			Usage:   "allow uploads that might be offensive (detecting unsafe content requires TensorFlow)",
			EnvVars: EnvVars("UPLOAD_NSFW"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "upload-allow",
			Usage:   "allow these file types for web uploads (comma-separated list of extensions; leave blank to allow all)",
			EnvVars: EnvVars("UPLOAD_ALLOW"),
			Hidden:  true,
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
			Value:   time.UTC.String(),
			EnvVars: EnvVars("DEFAULT_TIMEZONE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "default-theme",
			Usage:   "default user interface theme `NAME`",
			EnvVars: EnvVars("DEFAULT_THEME"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "app-name",
			Usage:   "progressive web app `NAME` when installed on a device",
			Value:   "",
			EnvVars: EnvVars("APP_NAME"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "app-mode",
			Usage:   "progressive web app `MODE` (fullscreen, standalone, minimal-ui, browser)",
			Value:   "standalone",
			EnvVars: EnvVars("APP_MODE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "app-icon",
			Usage:   "home screen `ICON` (logo, app, crisp, mint, bold, square)",
			EnvVars: EnvVars("APP_ICON"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "app-color",
			Usage:   "splash screen `COLOR` code",
			Value:   "#000000",
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
			Aliases: []string{"url"},
			Usage:   "public site `URL`",
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
			Usage:   "stream videos over the specified CDN",
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
			Name:    "https-proxy",
			Usage:   "proxy server `URL` to be used for outgoing connections *optional*",
			EnvVars: EnvVars("HTTPS_PROXY"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "https-proxy-insecure",
			Usage:   "ignore invalid HTTPS certificates when using a proxy",
			EnvVars: EnvVars("HTTPS_PROXY_INSECURE"),
		}}, {
		Flag: &cli.StringSliceFlag{
			Name:    "trusted-proxy",
			Usage:   "`CIDR` ranges or IPv4/v6 addresses from which reverse proxy headers can be trusted, separated by commas",
			Value:   cli.NewStringSlice(header.CidrDockerInternal),
			EnvVars: EnvVars("TRUSTED_PROXY"),
		}}, {
		Flag: &cli.StringSliceFlag{
			Name:    "proxy-proto-header",
			Usage:   "proxy protocol header `NAME`",
			Value:   cli.NewStringSlice(header.ForwardedProto),
			EnvVars: EnvVars("PROXY_PROTO_HEADER"),
		}}, {
		Flag: &cli.StringSliceFlag{
			Name:    "proxy-proto-https",
			Usage:   "forwarded HTTPS protocol `NAME`",
			Value:   cli.NewStringSlice(header.ProtoHttps),
			EnvVars: EnvVars("PROXY_PROTO_HTTPS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "disable-tls",
			Usage:   "disable HTTPS/TLS even if the site URL starts with https:// and a certificate is available",
			EnvVars: EnvVars("DISABLE_TLS"),
		}}, {
		Flag: &cli.BoolFlag{
			Name:    "default-tls",
			Usage:   "default to a self-signed HTTPS/TLS certificate if no other certificate is available",
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
			Usage:   "public HTTPS certificate `FILE` (.crt), ignored for Unix domain sockets",
			EnvVars: EnvVars("TLS_CERT"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "tls-key",
			Usage:   "private HTTPS key `FILE` (.key), ignored for Unix domain sockets",
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
			Usage:   "allow static content to be cached by a CDN or caching proxy",
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
			Usage:   "Web server `IP` address or Unix domain socket, e.g. unix:/var/run/photoprism.sock",
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
			Usage:   "database `HOST` incl. port e.g. \"mariadb:3306\" (or socket path)",
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
			Name:    "ffmpeg-bin",
			Usage:   "FFmpeg `COMMAND` for video transcoding and thumbnail extraction",
			Value:   ffmpeg.DefaultBin,
			EnvVars: EnvVars("FFMPEG_BIN"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "ffmpeg-encoder",
			Aliases: []string{"vc"},
			Usage:   "FFmpeg AVC encoder `NAME`",
			Value:   "libx264",
			EnvVars: EnvVars("FFMPEG_ENCODER"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "ffmpeg-size",
			Aliases: []string{"vs"},
			Usage:   "maximum video size in `PIXELS` (720-7680)",
			Value:   thumb.Sizes[thumb.Fit4096].Width,
			EnvVars: EnvVars("FFMPEG_SIZE"),
		}}, {
		Flag: &cli.IntFlag{
			Name:    "ffmpeg-bitrate",
			Aliases: []string{"vb"},
			Usage:   "maximum video `BITRATE` in Mbit/s",
			Value:   50,
			EnvVars: EnvVars("FFMPEG_BITRATE"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "ffmpeg-map-video",
			Usage:   "video `STREAMS` that should be transcoded",
			Value:   ffmpeg.MapVideoDefault,
			EnvVars: EnvVars("FFMPEG_MAP_VIDEO"),
		}, DocDefault: fmt.Sprintf("`%s`", ffmpeg.MapVideoDefault)}, {
		Flag: &cli.StringFlag{
			Name:    "ffmpeg-map-audio",
			Usage:   "audio `STREAMS` that should be transcoded",
			Value:   ffmpeg.MapAudioDefault,
			EnvVars: EnvVars("FFMPEG_MAP_AUDIO"),
		}, DocDefault: fmt.Sprintf("`%s`", ffmpeg.MapAudioDefault)}, {
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
		Flag: &cli.StringFlag{
			Name:    "darktable-cache-path",
			Usage:   "custom Darktable cache `PATH`",
			Value:   "",
			EnvVars: EnvVars("DARKTABLE_CACHE_PATH"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "darktable-config-path",
			Usage:   "custom Darktable config `PATH`",
			Value:   "",
			EnvVars: EnvVars("DARKTABLE_CONFIG_PATH"),
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
			Value:   "auto",
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
			Usage:   "generate missing thumbnails on demand (high memory and cpu usage)",
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
			Name:    "face-match-dist",
			Usage:   "similarity `OFFSET` for matching faces with existing clusters (0.1-1.5)",
			Value:   face.MatchDist,
			EnvVars: EnvVars("FACE_MATCH_DIST"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "pid-filename",
			Usage:   "process id `FILE` *daemon-mode only*",
			EnvVars: EnvVars("PID_FILENAME"),
		}}, {
		Flag: &cli.StringFlag{
			Name:    "log-filename",
			Usage:   "server log `FILE` *daemon-mode only*",
			Value:   "",
			EnvVars: EnvVars("LOG_FILENAME"),
		}},
}
