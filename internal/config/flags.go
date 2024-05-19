package config

import (
	"fmt"
	"time"

	"github.com/klauspost/cpuid/v2"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/face"
	"github.com/photoprism/photoprism/internal/ffmpeg"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/ttl"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Flags configures the global command-line interface (CLI) parameters.
var Flags = CliFlags{
	{
		Flag: cli.StringFlag{
			Name:   "auth-mode, a",
			Usage:  "authentication `MODE` (public, password)",
			Value:  "password",
			EnvVar: EnvVar("AUTH_MODE"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "public, p",
			Hidden: true,
			Usage:  "disable authentication, advanced settings, and WebDAV remote access",
			EnvVar: EnvVar("PUBLIC"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "admin-user, login",
			Usage:  "`USERNAME` of the superadmin account that is created on first startup",
			Value:  "admin",
			EnvVar: EnvVar("ADMIN_USER"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "admin-password, pw",
			Usage:  fmt.Sprintf("initial `PASSWORD` of the superadmin account (%d-%d characters)", entity.PasswordLength, txt.ClipPassword),
			EnvVar: EnvVar("ADMIN_PASSWORD"),
		}}, {
		Flag: cli.Int64Flag{
			Name:   "session-maxage",
			Value:  DefaultSessionMaxAge,
			Usage:  "session expiration time in `SECONDS`, doubled for accounts with 2FA (-1 to disable)",
			EnvVar: EnvVar("SESSION_MAXAGE"),
		}}, {
		Flag: cli.Int64Flag{
			Name:   "session-timeout",
			Value:  DefaultSessionTimeout,
			Usage:  "session idle time in `SECONDS`, doubled for accounts with 2FA (-1 to disable)",
			EnvVar: EnvVar("SESSION_TIMEOUT"),
		}}, {
		Flag: cli.Int64Flag{
			Name:   "session-cache",
			Value:  DefaultSessionCache,
			Usage:  "session cache duration in `SECONDS` (60-3600)",
			EnvVar: EnvVar("SESSION_CACHE"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "log-level, l",
			Usage:  "log message verbosity `LEVEL` (trace, debug, info, warning, error, fatal, panic)",
			Value:  "info",
			EnvVar: EnvVar("LOG_LEVEL"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "prod",
			Hidden: true,
			Usage:  "enable production mode, hide non-essential log messages",
			EnvVar: EnvVar("PROD"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "debug",
			Usage:  "enable debug mode, show non-essential log messages",
			EnvVar: EnvVar("DEBUG"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "trace",
			Usage:  "enable trace mode, show all log messages",
			EnvVar: EnvVar("TRACE"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "test",
			Hidden: true,
			Usage:  "enable test mode",
		}}, {
		Flag: cli.BoolFlag{
			Name:   "unsafe",
			Hidden: true,
			Usage:  "disable safety checks",
			EnvVar: EnvVar("UNSAFE"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "demo",
			Hidden: true,
			Usage:  "enable demo mode",
			EnvVar: EnvVar("DEMO"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "sponsor",
			Hidden: true,
			Usage:  "your continuous support helps to pay for development and operating expenses",
			EnvVar: EnvVar("SPONSOR"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "partner-id",
			Hidden: true,
			Usage:  "hosting partner id",
			EnvVar: EnvVar("PARTNER_ID"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "config-path, c",
			Usage:  "config storage `PATH`, values in options.yml override CLI flags and environment variables if present",
			EnvVar: EnvVar("CONFIG_PATH"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "defaults-yaml, y",
			Usage:  "load config defaults from `FILE` if exists, does not override CLI flags and environment variables",
			Value:  "/etc/photoprism/defaults.yml",
			EnvVar: EnvVar("DEFAULTS_YAML"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "originals-path, o",
			Usage:  "storage `PATH` of your original media files (photos and videos)",
			EnvVar: EnvVar("ORIGINALS_PATH"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "originals-limit, mb",
			Value:  1000,
			Usage:  "maximum size of media files in `MB` (1-100000; -1 to disable)",
			EnvVar: EnvVar("ORIGINALS_LIMIT"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "resolution-limit, mp",
			Value:  DefaultResolutionLimit,
			Usage:  "maximum resolution of media files in `MEGAPIXELS` (1-900; -1 to disable)",
			EnvVar: EnvVar("RESOLUTION_LIMIT"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "users-path",
			Usage:  "relative `PATH` to create base and upload subdirectories for users",
			Value:  "users",
			EnvVar: EnvVar("USERS_PATH"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "storage-path, s",
			Usage:  "writable storage `PATH` for sidecar, cache, and database files",
			EnvVar: EnvVar("STORAGE_PATH"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "sidecar-path, sc",
			Usage:  "custom relative or absolute sidecar `PATH` *optional*",
			EnvVar: EnvVar("SIDECAR_PATH"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "sidecar-yaml",
			Usage:  "create YAML sidecar files to back up index metadata",
			EnvVar: EnvVar("SIDECAR_YAML"),
		}, DocDefault: "true"}, {
		Flag: cli.StringFlag{
			Name:   "cache-path, ca",
			Usage:  "custom cache `PATH` for sessions and thumbnail files *optional*",
			EnvVar: EnvVar("CACHE_PATH"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "import-path, im",
			Usage:  "base `PATH` from which files can be imported to originals *optional*",
			EnvVar: EnvVar("IMPORT_PATH"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "import-dest",
			Usage:  "relative originals `PATH` to which the files should be imported by default *optional*",
			EnvVar: EnvVar("IMPORT_DEST"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "assets-path, as",
			Usage:  "assets `PATH` containing static resources like icons, models, and translations",
			EnvVar: EnvVar("ASSETS_PATH"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "temp-path, tmp",
			Usage:  "temporary file `PATH` *optional*",
			EnvVar: EnvVar("TEMP_PATH"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "backup-path, ba",
			Usage:  "custom base `PATH` for creating and restoring backups *optional*",
			EnvVar: EnvVar("BACKUP_PATH"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "backup-schedule",
			Usage:  "backup `SCHEDULE` in cron format (e.g. \"0 12 * * *\" for daily at noon) or at a random time (daily, weekly)",
			Value:  DefaultBackupSchedule,
			EnvVar: EnvVar("BACKUP_SCHEDULE"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "backup-retain",
			Usage:  "`NUMBER` of index backups to keep (-1 to keep all)",
			Value:  DefaultBackupRetain,
			EnvVar: EnvVar("BACKUP_RETAIN"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "backup-database",
			Usage:  "create index backups based on the configured schedule",
			EnvVar: EnvVar("BACKUP_DATABASE"),
		}, DocDefault: "true"}, {
		Flag: cli.BoolFlag{
			Name:   "backup-albums",
			Usage:  "create YAML backup files for album metadata",
			EnvVar: EnvVar("BACKUP_ALBUMS"),
		}, DocDefault: "true"}, {
		Flag: cli.IntFlag{
			Name:   "index-workers, workers",
			Usage:  "maximum `NUMBER` of indexing workers, default depends on the number of physical cores",
			Value:  cpuid.CPU.PhysicalCores / 2,
			EnvVar: EnvVar("INDEX_WORKERS") + ", " + EnvVar("WORKERS"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "index-schedule",
			Usage:  "regular indexing `SCHEDULE` in cron format (e.g. \"@every 3h\" for every 3 hours; \"\" to disable)",
			Value:  DefaultIndexSchedule,
			EnvVar: EnvVar("INDEX_SCHEDULE"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "wakeup-interval, i",
			Usage:  "`TIME` between facial recognition, file sync, and metadata worker runs (1-86400s)",
			Value:  DefaultWakeupInterval.String(),
			EnvVar: EnvVar("WAKEUP_INTERVAL"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "auto-index",
			Usage:  "delay before automatically indexing files in `SECONDS` (-1 to disable) when uploading files via WebDAV",
			Value:  DefaultAutoIndexDelay,
			EnvVar: EnvVar("AUTO_INDEX"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "auto-import",
			Usage:  "delay before automatically importing files in `SECONDS` (-1 to disable) when uploading files via WebDAV",
			Value:  DefaultAutoImportDelay,
			EnvVar: EnvVar("AUTO_IMPORT"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "read-only, r",
			Usage:  "disable import, upload, delete, and all other operations that require write permissions",
			EnvVar: EnvVar("READONLY"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "experimental, e",
			Usage:  "enable experimental features",
			EnvVar: EnvVar("EXPERIMENTAL"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-settings",
			Usage:  "disable settings UI and API",
			EnvVar: EnvVar("DISABLE_SETTINGS"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-restart",
			Usage:  "disable restarting the server from the user interface",
			EnvVar: EnvVar("DISABLE_RESTART"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-backups",
			Usage:  "disable creating and updating YAML sidecar files (does not affect index and album backups)",
			EnvVar: EnvVar("DISABLE_BACKUPS"),
			Hidden: true,
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-webdav",
			Usage:  "disable built-in WebDAV server",
			EnvVar: EnvVar("DISABLE_WEBDAV"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-places",
			Usage:  "disable reverse geocoding and maps",
			EnvVar: EnvVar("DISABLE_PLACES"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-tensorflow",
			Usage:  "disable all features depending on TensorFlow",
			EnvVar: EnvVar("DISABLE_TENSORFLOW"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-faces",
			Usage:  "disable face detection and recognition (requires TensorFlow)",
			EnvVar: EnvVar("DISABLE_FACES"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-classification",
			Usage:  "disable image classification (requires TensorFlow)",
			EnvVar: EnvVar("DISABLE_CLASSIFICATION"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-sips",
			Usage:  "disable conversion of media files with Sips *macOS only*",
			EnvVar: EnvVar("DISABLE_SIPS"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-ffmpeg",
			Usage:  "disable video transcoding and thumbnail extraction with FFmpeg",
			EnvVar: EnvVar("DISABLE_FFMPEG"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-exiftool",
			Usage:  "disable creating JSON metadata sidecar files with ExifTool",
			EnvVar: EnvVar("DISABLE_EXIFTOOL"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-darktable",
			Usage:  "disable conversion of RAW images with Darktable",
			EnvVar: EnvVar("DISABLE_DARKTABLE"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-rawtherapee",
			Usage:  "disable conversion of RAW images with RawTherapee",
			EnvVar: EnvVar("DISABLE_RAWTHERAPEE"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-imagemagick",
			Usage:  "disable conversion of image files with ImageMagick",
			EnvVar: EnvVar("DISABLE_IMAGEMAGICK"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-heifconvert",
			Usage:  "disable conversion of HEIC images with libheif",
			EnvVar: EnvVar("DISABLE_HEIFCONVERT"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-jpegxl",
			Usage:  "disable JPEG XL file format support",
			EnvVar: EnvVar("DISABLE_JPEGXL"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-raw",
			Usage:  "disable indexing and conversion of RAW images",
			EnvVar: EnvVar("DISABLE_RAW"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "raw-presets",
			Usage:  "enables applying user presets when converting RAW images (reduces performance)",
			EnvVar: EnvVar("RAW_PRESETS"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "exif-bruteforce",
			Usage:  "always perform a brute-force search if no Exif headers were found",
			EnvVar: EnvVar("EXIF_BRUTEFORCE"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "detect-nsfw",
			Usage:  "automatically flag pictures as private that MAY be offensive (requires TensorFlow)",
			EnvVar: EnvVar("DETECT_NSFW"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "upload-nsfw, n",
			Usage:  "allow uploads that MAY be offensive (no effect without TensorFlow)",
			EnvVar: EnvVar("UPLOAD_NSFW"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "default-locale, lang",
			Usage:  "default user interface language `CODE`",
			Value:  i18n.Default.Locale(),
			EnvVar: EnvVar("DEFAULT_LOCALE"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "default-timezone, tz",
			Usage:  "default time zone `NAME`, e.g. for scheduling backups",
			Value:  time.UTC.String(),
			EnvVar: EnvVar("DEFAULT_TIMEZONE"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "default-theme",
			Usage:  "default user interface theme `NAME`",
			EnvVar: EnvVar("DEFAULT_THEME"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "app-name",
			Usage:  "progressive web app `NAME` when installed on a device",
			Value:  "",
			EnvVar: EnvVar("APP_NAME"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "app-mode",
			Usage:  "progressive web app `MODE` (fullscreen, standalone, minimal-ui, browser)",
			Value:  "standalone",
			EnvVar: EnvVar("APP_MODE"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "app-icon",
			Usage:  "home screen `ICON` (logo, app, crisp, mint, bold, square)",
			EnvVar: EnvVar("APP_ICON"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "app-color",
			Usage:  "splash screen `COLOR` code",
			Value:  "#000000",
			EnvVar: EnvVar("APP_COLOR"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "imprint",
			Usage:  "legal information `TEXT`, displayed in the page footer",
			Value:  "",
			Hidden: true,
			EnvVar: EnvVar("IMPRINT"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "legal-info",
			Usage:  "legal information `TEXT`, displayed in the page footer",
			Value:  "",
			EnvVar: EnvVar("LEGAL_INFO"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "imprint-url",
			Usage:  "legal information `URL`",
			Value:  "",
			Hidden: true,
			EnvVar: EnvVar("IMPRINT_URL"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "legal-url",
			Usage:  "legal information `URL`",
			Value:  "",
			EnvVar: EnvVar("LEGAL_URL"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "wallpaper-uri",
			Usage:  "login screen background image `URI`",
			Value:  "",
			EnvVar: EnvVar("WALLPAPER_URI"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "site-url, url",
			Usage:  "public site `URL`",
			Value:  "http://localhost:2342/",
			EnvVar: EnvVar("SITE_URL"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "site-author",
			Usage:  "site `OWNER`, copyright, or artist",
			EnvVar: EnvVar("SITE_AUTHOR"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "site-title",
			Usage:  "site `TITLE`",
			Value:  "",
			EnvVar: EnvVar("SITE_TITLE"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "site-caption",
			Usage:  "site `CAPTION`",
			Value:  "AI-Powered Photos App",
			EnvVar: EnvVar("SITE_CAPTION"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "site-description",
			Usage:  "site `DESCRIPTION` *optional*",
			EnvVar: EnvVar("SITE_DESCRIPTION"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "site-preview",
			Usage:  "sharing preview image `URL`",
			EnvVar: EnvVar("SITE_PREVIEW"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "cdn-url",
			Usage:  "content delivery network `URL`",
			EnvVar: EnvVar("CDN_URL"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "cdn-video",
			Usage:  "stream videos over the specified CDN",
			EnvVar: EnvVar("CDN_VIDEO"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "cors-origin",
			Usage:  "origin `URL` from which browsers are allowed to perform cross-origin requests (leave empty to disable or use * to allow all)",
			EnvVar: EnvVar("CORS_ORIGIN"),
			Value:  header.DefaultAccessControlAllowOrigin,
		}}, {
		Flag: cli.StringFlag{
			Name:   "cors-headers",
			Usage:  "one or more `HEADERS` that browsers should see when performing a cross-origin request",
			EnvVar: EnvVar("CORS_HEADERS"),
			Value:  header.DefaultAccessControlAllowHeaders,
		}}, {
		Flag: cli.StringFlag{
			Name:   "cors-methods",
			Usage:  "one or more `METHODS` that may be used when performing a cross-origin request",
			EnvVar: EnvVar("CORS_METHODS"),
			Value:  header.DefaultAccessControlAllowMethods,
		}}, {
		Flag: cli.StringFlag{
			Name:   "https-proxy",
			Usage:  "proxy server `URL` to be used for outgoing connections *optional*",
			EnvVar: EnvVar("HTTPS_PROXY"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "https-proxy-insecure",
			Usage:  "ignore invalid HTTPS certificates when using a proxy",
			EnvVar: EnvVar("HTTPS_PROXY_INSECURE"),
		}}, {
		Flag: cli.StringSliceFlag{
			Name:   "trusted-proxy",
			Usage:  "`CIDR` ranges or IPv4/v6 addresses from which reverse proxy headers can be trusted, separated by commas",
			Value:  &cli.StringSlice{header.CidrDockerInternal},
			EnvVar: EnvVar("TRUSTED_PROXY"),
		}}, {
		Flag: cli.StringSliceFlag{
			Name:   "proxy-proto-header",
			Usage:  "proxy protocol header `NAME`",
			Value:  &cli.StringSlice{header.ForwardedProto},
			EnvVar: EnvVar("PROXY_PROTO_HEADER"),
		}}, {
		Flag: cli.StringSliceFlag{
			Name:   "proxy-proto-https",
			Usage:  "forwarded HTTPS protocol `NAME`",
			Value:  &cli.StringSlice{header.ProtoHttps},
			EnvVar: EnvVar("PROXY_PROTO_HTTPS"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "disable-tls",
			Usage:  "disable HTTPS/TLS even if the site URL starts with https:// and a certificate is available",
			EnvVar: EnvVar("DISABLE_TLS"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "default-tls",
			Usage:  "default to a self-signed HTTPS/TLS certificate if no other certificate is available",
			EnvVar: EnvVar("DEFAULT_TLS"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "tls-email",
			Usage:  "`EMAIL` address to enable automatic HTTPS via Let's Encrypt",
			EnvVar: EnvVar("TLS_EMAIL"),
			Hidden: true,
		}}, {
		Flag: cli.StringFlag{
			Name:   "tls-cert",
			Usage:  "public HTTPS certificate `FILE` (.crt), ignored for Unix domain sockets",
			EnvVar: EnvVar("TLS_CERT"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "tls-key",
			Usage:  "private HTTPS key `FILE` (.key), ignored for Unix domain sockets",
			EnvVar: EnvVar("TLS_KEY"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "http-mode, mode",
			Usage:  "Web server `MODE` (debug, release, test)",
			EnvVar: EnvVar("HTTP_MODE"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "http-compression, z",
			Usage:  "Web server compression `METHOD` (gzip, none)",
			EnvVar: EnvVar("HTTP_COMPRESSION"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "http-cache-public",
			Usage:  "allow static content to be cached by a CDN or caching proxy",
			EnvVar: EnvVar("HTTP_CACHE_PUBLIC"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "http-cache-maxage",
			Value:  int(ttl.CacheDefault),
			Usage:  "time in `SECONDS` until cached content expires",
			EnvVar: EnvVar("HTTP_CACHE_MAXAGE"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "http-video-maxage",
			Value:  int(ttl.CacheVideo),
			Usage:  "time in `SECONDS` until cached videos expire",
			EnvVar: EnvVar("HTTP_VIDEO_MAXAGE"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "http-host, ip",
			Value:  "0.0.0.0",
			Usage:  "Web server `IP` address or Unix domain socket, e.g. unix:/var/run/photoprism.sock",
			EnvVar: EnvVar("HTTP_HOST"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "http-port, port",
			Value:  2342,
			Usage:  "Web server port `NUMBER`, ignored for Unix domain sockets",
			EnvVar: EnvVar("HTTP_PORT"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "database-driver, db",
			Usage:  "database `DRIVER` (sqlite, mysql)",
			Value:  "sqlite",
			EnvVar: EnvVar("DATABASE_DRIVER"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "database-dsn, dsn",
			Usage:  "database connection `DSN` (sqlite file, optional for mysql)",
			EnvVar: EnvVar("DATABASE_DSN"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "database-name, db-name",
			Value:  "photoprism",
			Usage:  "database schema `NAME`",
			EnvVar: EnvVar("DATABASE_NAME"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "database-server, db-server",
			Usage:  "database `HOST` incl. port e.g. \"mariadb:3306\" (or socket path)",
			EnvVar: EnvVar("DATABASE_SERVER"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "database-user, db-user",
			Value:  "photoprism",
			Usage:  "database user `NAME`",
			EnvVar: EnvVar("DATABASE_USER"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "database-password, db-pass",
			Usage:  "database user `PASSWORD`",
			EnvVar: EnvVar("DATABASE_PASSWORD"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "database-timeout",
			Usage:  "timeout in `SECONDS` for establishing a database connection (1-60)",
			EnvVar: EnvVar("DATABASE_TIMEOUT"),
			Value:  15,
		}}, {
		Flag: cli.IntFlag{
			Name:   "database-conns",
			Usage:  "maximum `NUMBER` of open database connections",
			EnvVar: EnvVar("DATABASE_CONNS"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "database-conns-idle",
			Usage:  "maximum `NUMBER` of idle database connections",
			EnvVar: EnvVar("DATABASE_CONNS_IDLE"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "sips-bin",
			Usage:  "Sips `COMMAND` for media file conversion *macOS only*",
			Value:  "sips",
			EnvVar: EnvVar("SIPS_BIN"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "sips-blacklist",
			Usage:  "do not use Sips to convert files with these `EXTENSIONS` *macOS only*",
			Value:  "avif,avifs,thm",
			EnvVar: EnvVar("SIPS_BLACKLIST"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "ffmpeg-bin",
			Usage:  "FFmpeg `COMMAND` for video transcoding and thumbnail extraction",
			Value:  ffmpeg.DefaultBin,
			EnvVar: EnvVar("FFMPEG_BIN"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "ffmpeg-encoder, vc",
			Usage:  "FFmpeg AVC encoder `NAME`",
			Value:  "libx264",
			EnvVar: EnvVar("FFMPEG_ENCODER"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "ffmpeg-size, vs",
			Usage:  "maximum video size in `PIXELS` (720-7680)",
			Value:  thumb.Sizes[thumb.Fit4096].Width,
			EnvVar: EnvVar("FFMPEG_SIZE"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "ffmpeg-bitrate, vb",
			Usage:  "maximum video `BITRATE` in Mbit/s",
			Value:  50,
			EnvVar: EnvVar("FFMPEG_BITRATE"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "ffmpeg-map-video",
			Usage:  "video `STREAMS` that should be transcoded",
			Value:  ffmpeg.MapVideoDefault,
			EnvVar: EnvVar("FFMPEG_MAP_VIDEO"),
		}, DocDefault: fmt.Sprintf("`%s`", ffmpeg.MapVideoDefault)}, {
		Flag: cli.StringFlag{
			Name:   "ffmpeg-map-audio",
			Usage:  "audio `STREAMS` that should be transcoded",
			Value:  ffmpeg.MapAudioDefault,
			EnvVar: EnvVar("FFMPEG_MAP_AUDIO"),
		}, DocDefault: fmt.Sprintf("`%s`", ffmpeg.MapAudioDefault)}, {
		Flag: cli.StringFlag{
			Name:   "exiftool-bin",
			Usage:  "ExifTool `COMMAND` for extracting metadata",
			Value:  "exiftool",
			EnvVar: EnvVar("EXIFTOOL_BIN"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "darktable-bin",
			Usage:  "Darktable CLI `COMMAND` for RAW to JPEG conversion",
			Value:  "darktable-cli",
			EnvVar: EnvVar("DARKTABLE_BIN"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "darktable-blacklist",
			Usage:  "do not use Darktable to convert files with these `EXTENSIONS`",
			Value:  "thm",
			EnvVar: EnvVar("DARKTABLE_BLACKLIST"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "darktable-cache-path",
			Usage:  "custom Darktable cache `PATH`",
			Value:  "",
			EnvVar: EnvVar("DARKTABLE_CACHE_PATH"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "darktable-config-path",
			Usage:  "custom Darktable config `PATH`",
			Value:  "",
			EnvVar: EnvVar("DARKTABLE_CONFIG_PATH"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "rawtherapee-bin",
			Usage:  "RawTherapee CLI `COMMAND` for RAW to JPEG conversion",
			Value:  "rawtherapee-cli",
			EnvVar: EnvVar("RAWTHERAPEE_BIN"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "rawtherapee-blacklist",
			Usage:  "do not use RawTherapee to convert files with these `EXTENSIONS`",
			Value:  "dng,thm",
			EnvVar: EnvVar("RAWTHERAPEE_BLACKLIST"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "imagemagick-bin",
			Usage:  "ImageMagick CLI `COMMAND` for image file conversion",
			Value:  "convert",
			EnvVar: EnvVar("IMAGEMAGICK_BIN"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "imagemagick-blacklist",
			Usage:  "do not use ImageMagick to convert files with these `EXTENSIONS`",
			Value:  "heif,heic,heics,avif,avifs,jxl,thm",
			EnvVar: EnvVar("IMAGEMAGICK_BLACKLIST"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "heifconvert-bin",
			Usage:  "libheif HEIC image conversion `COMMAND`",
			Value:  "heif-convert",
			EnvVar: EnvVar("HEIFCONVERT_BIN"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "download-token",
			Usage:  "`DEFAULT` download URL token for originals (leave empty for a random value)",
			EnvVar: EnvVar("DOWNLOAD_TOKEN"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "preview-token",
			Usage:  "`DEFAULT` thumbnail and video streaming URL token (leave empty for a random value)",
			EnvVar: EnvVar("PREVIEW_TOKEN"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "thumb-library, thumbs",
			Usage:  "image processing `LIBRARY` to be used for generating thumbnails (auto, imaging, vips)",
			Value:  "auto",
			EnvVar: EnvVar("THUMB_LIBRARY"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "thumb-color",
			Usage:  "standard color `PROFILE` for thumbnails (auto, preserve, srgb, none)",
			Value:  thumb.ColorAuto,
			EnvVar: EnvVar("THUMB_COLOR"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "thumb-filter, filter",
			Usage:  "downscaling filter `NAME` (imaging best to worst: blackman, lanczos, cubic, linear, nearest)",
			Value:  thumb.ResampleAuto.String(),
			EnvVar: EnvVar("THUMB_FILTER"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "thumb-size",
			Usage:  "maximum size of pre-generated thumbnails in `PIXELS` (720-7680)",
			Value:  thumb.SizeCached,
			EnvVar: EnvVar("THUMB_SIZE"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "thumb-size-uncached",
			Usage:  "maximum size of thumbnails generated on demand in `PIXELS` (720-7680)",
			Value:  thumb.SizeOnDemand,
			EnvVar: EnvVar("THUMB_SIZE_UNCACHED"),
		}}, {
		Flag: cli.BoolFlag{
			Name:   "thumb-uncached, u",
			Usage:  "generate missing thumbnails on demand (high memory and cpu usage)",
			EnvVar: EnvVar("THUMB_UNCACHED"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "jpeg-quality, q",
			Usage:  "higher values increase the image `QUALITY` and file size (25-100)",
			Value:  thumb.JpegQualityDefault.String(),
			EnvVar: EnvVar("JPEG_QUALITY"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "jpeg-size",
			Usage:  "maximum size of generated JPEG images in `PIXELS` (720-30000)",
			Value:  7680,
			EnvVar: EnvVar("JPEG_SIZE"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "png-size",
			Usage:  "maximum size of generated PNG images in `PIXELS` (720-30000)",
			Value:  7680,
			EnvVar: EnvVar("PNG_SIZE"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "face-size",
			Usage:  "minimum size of faces in `PIXELS` (20-10000)",
			Value:  face.SizeThreshold,
			EnvVar: EnvVar("FACE_SIZE"),
		}}, {
		Flag: cli.Float64Flag{
			Name:   "face-score",
			Usage:  "minimum face `QUALITY` score (1-100)",
			Value:  face.ScoreThreshold,
			EnvVar: EnvVar("FACE_SCORE"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "face-overlap",
			Usage:  "face area overlap threshold in `PERCENT` (1-100)",
			Value:  face.OverlapThreshold,
			EnvVar: EnvVar("FACE_OVERLAP"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "face-cluster-size",
			Usage:  "minimum size of automatically clustered faces in `PIXELS` (20-10000)",
			Value:  face.ClusterSizeThreshold,
			EnvVar: EnvVar("FACE_CLUSTER_SIZE"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "face-cluster-score",
			Usage:  "minimum `QUALITY` score of automatically clustered faces (1-100)",
			Value:  face.ClusterScoreThreshold,
			EnvVar: EnvVar("FACE_CLUSTER_SCORE"),
		}}, {
		Flag: cli.IntFlag{
			Name:   "face-cluster-core",
			Usage:  "`NUMBER` of faces forming a cluster core (1-100)",
			Value:  face.ClusterCore,
			EnvVar: EnvVar("FACE_CLUSTER_CORE"),
		}}, {
		Flag: cli.Float64Flag{
			Name:   "face-cluster-dist",
			Usage:  "similarity `DISTANCE` of faces forming a cluster core (0.1-1.5)",
			Value:  face.ClusterDist,
			EnvVar: EnvVar("FACE_CLUSTER_DIST"),
		}}, {
		Flag: cli.Float64Flag{
			Name:   "face-match-dist",
			Usage:  "similarity `OFFSET` for matching faces with existing clusters (0.1-1.5)",
			Value:  face.MatchDist,
			EnvVar: EnvVar("FACE_MATCH_DIST"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "pid-filename",
			Usage:  "process id `FILE` *daemon-mode only*",
			EnvVar: EnvVar("PID_FILENAME"),
		}}, {
		Flag: cli.StringFlag{
			Name:   "log-filename",
			Usage:  "server log `FILE` *daemon-mode only*",
			Value:  "",
			EnvVar: EnvVar("LOG_FILENAME"),
		}},
}
