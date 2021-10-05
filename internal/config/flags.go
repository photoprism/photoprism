package config

import (
	"github.com/klauspost/cpuid/v2"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/face"
)

// GlobalFlags describes global command-line parameters and flags.
var GlobalFlags = []cli.Flag{
	cli.BoolFlag{
		Name:   "debug",
		Usage:  "enables the debug mode, shows additional log messages",
		EnvVar: "PHOTOPRISM_DEBUG",
	},
	cli.BoolFlag{
		Name:   "test",
		Hidden: true,
		Usage:  "enables the test mode",
	},
	cli.BoolFlag{
		Name:   "demo",
		Hidden: true,
		Usage:  "enables the demo mode",
		EnvVar: "PHOTOPRISM_DEMO",
	},
	cli.BoolFlag{
		Name:   "sponsor",
		Hidden: true,
		Usage:  "your continuous support helps to pay for development and operating expenses",
		EnvVar: "PHOTOPRISM_SPONSOR",
	},
	cli.BoolFlag{
		Name:   "public, p",
		Usage:  "disables password authentication",
		EnvVar: "PHOTOPRISM_PUBLIC",
	},
	cli.BoolFlag{
		Name:   "read-only, r",
		Usage:  "disables import, upload, and delete",
		EnvVar: "PHOTOPRISM_READONLY",
	},
	cli.BoolFlag{
		Name:   "experimental, e",
		Usage:  "enables experimental features",
		EnvVar: "PHOTOPRISM_EXPERIMENTAL",
	},
	cli.StringFlag{
		Name:   "admin-password",
		Usage:  "sets the initial admin `PASSWORD`, min 4 characters",
		EnvVar: "PHOTOPRISM_ADMIN_PASSWORD",
	},
	cli.StringFlag{
		Name:   "config-file, c",
		Usage:  "loads config options from `FILENAME`",
		EnvVar: "PHOTOPRISM_CONFIG_FILE",
	},
	cli.StringFlag{
		Name:   "config-path",
		Usage:  "sets the config `PATH` for storing settings and other config files",
		EnvVar: "PHOTOPRISM_CONFIG_PATH",
	},
	cli.StringFlag{
		Name:   "originals-path",
		Usage:  "sets the originals `PATH` containing your photo and video files",
		EnvVar: "PHOTOPRISM_ORIGINALS_PATH",
	},
	cli.IntFlag{
		Name:   "originals-limit",
		Value:  1000,
		Usage:  "limits the file size in `MB`",
		EnvVar: "PHOTOPRISM_ORIGINALS_LIMIT",
	},
	cli.StringFlag{
		Name:   "import-path",
		Usage:  "sets an optional import `PATH` from which files can be added to originals",
		EnvVar: "PHOTOPRISM_IMPORT_PATH",
	},
	cli.StringFlag{
		Name:   "storage-path",
		Usage:  "sets the storage base `PATH` for cache, database, and sidecar files",
		EnvVar: "PHOTOPRISM_STORAGE_PATH",
	},
	cli.StringFlag{
		Name:   "sidecar-path",
		Usage:  "sets a custom relative or absolute sidecar `PATH`",
		EnvVar: "PHOTOPRISM_SIDECAR_PATH",
	},
	cli.StringFlag{
		Name:   "cache-path",
		Usage:  "sets a custom cache `PATH` for storing sessions and thumbnails",
		EnvVar: "PHOTOPRISM_CACHE_PATH",
	},
	cli.StringFlag{
		Name:   "temp-path",
		Usage:  "sets a custom temp `PATH` for storing temporary files",
		EnvVar: "PHOTOPRISM_TEMP_PATH",
	},
	cli.StringFlag{
		Name:   "backup-path",
		Usage:  "sets a custom backup `PATH`",
		EnvVar: "PHOTOPRISM_BACKUP_PATH",
	},
	cli.StringFlag{
		Name:   "assets-path",
		Usage:  "sets the assets `PATH` containing static resources like icons and templates",
		EnvVar: "PHOTOPRISM_ASSETS_PATH",
	},
	cli.IntFlag{
		Name:   "workers, w",
		Usage:  "limits the `NUMBER` of indexing workers",
		EnvVar: "PHOTOPRISM_WORKERS",
		Value:  cpuid.CPU.PhysicalCores / 2,
	},
	cli.IntFlag{
		Name:   "wakeup-interval",
		Usage:  "adjusts the background worker wakeup interval in `SECONDS`",
		Value:  DefaultWakeupInterval,
		EnvVar: "PHOTOPRISM_WAKEUP_INTERVAL",
	},
	cli.IntFlag{
		Name:   "auto-index",
		Usage:  "adjusts the WebDAV auto indexing safety delay in `SECONDS` (-1 to disable)",
		Value:  DefaultAutoIndexDelay,
		EnvVar: "PHOTOPRISM_AUTO_INDEX",
	},
	cli.IntFlag{
		Name:   "auto-import",
		Usage:  "adjusts the WebDAV auto import safety delay in `SECONDS` (-1 to disable)",
		Value:  DefaultAutoImportDelay,
		EnvVar: "PHOTOPRISM_AUTO_IMPORT",
	},
	cli.BoolFlag{
		Name:   "disable-backups",
		Usage:  "disables creating YAML metadata backup sidecar files",
		EnvVar: "PHOTOPRISM_DISABLE_BACKUPS",
	},
	cli.BoolFlag{
		Name:   "disable-webdav",
		Usage:  "disables the built-in WebDAV server",
		EnvVar: "PHOTOPRISM_DISABLE_WEBDAV",
	},
	cli.BoolFlag{
		Name:   "disable-settings",
		Usage:  "disables the settings UI and API",
		EnvVar: "PHOTOPRISM_DISABLE_SETTINGS",
	},
	cli.BoolFlag{
		Name:   "disable-places",
		Usage:  "disables reverse geocoding and maps",
		EnvVar: "PHOTOPRISM_DISABLE_PLACES",
	},
	cli.BoolFlag{
		Name:   "disable-exiftool",
		Usage:  "disables metadata extraction with ExifTool",
		EnvVar: "PHOTOPRISM_DISABLE_EXIFTOOL",
	},
	cli.BoolFlag{
		Name:   "disable-ffmpeg",
		Usage:  "disables video transcoding and thumbnail generation with FFmpeg",
		EnvVar: "PHOTOPRISM_DISABLE_FFMPEG",
	},
	cli.BoolFlag{
		Name:   "disable-darktable",
		Usage:  "disables RAW file conversion with Darktable",
		EnvVar: "PHOTOPRISM_DISABLE_DARKTABLE",
	},
	cli.BoolFlag{
		Name:   "disable-rawtherapee",
		Usage:  "disables RAW file conversion with RawTherapee",
		EnvVar: "PHOTOPRISM_DISABLE_RAWTHERAPEE",
	},
	cli.BoolFlag{
		Name:   "disable-sips",
		Usage:  "disables RAW file conversion with Sips on macOS",
		EnvVar: "PHOTOPRISM_DISABLE_SIPS",
	},
	cli.BoolFlag{
		Name:   "disable-heifconvert",
		Usage:  "disables HEIC/HEIF file conversion",
		EnvVar: "PHOTOPRISM_DISABLE_HEIFCONVERT",
	},
	cli.BoolFlag{
		Name:   "disable-tensorflow",
		Usage:  "disables all features depending on TensorFlow",
		EnvVar: "PHOTOPRISM_DISABLE_TENSORFLOW",
	},
	cli.BoolFlag{
		Name:   "disable-faces",
		Usage:  "disables facial recognition",
		EnvVar: "PHOTOPRISM_DISABLE_FACES",
	},
	cli.BoolFlag{
		Name:   "disable-classification",
		Usage:  "disables image classification",
		EnvVar: "PHOTOPRISM_DISABLE_CLASSIFICATION",
	},
	cli.BoolFlag{
		Name:   "detect-nsfw",
		Usage:  "flags photos as private that may be offensive (requires TensorFlow)",
		EnvVar: "PHOTOPRISM_DETECT_NSFW",
	},
	cli.BoolFlag{
		Name:   "upload-nsfw",
		Usage:  "allows uploads that may be offensive",
		EnvVar: "PHOTOPRISM_UPLOAD_NSFW",
	},
	cli.StringFlag{
		Name:   "log-level, l",
		Usage:  "adjusts the log level to trace, debug, info, warning, error, fatal or panic",
		Value:  "info",
		EnvVar: "PHOTOPRISM_LOG_LEVEL",
	},
	cli.StringFlag{
		Name:   "log-filename",
		Usage:  "sets the server log `FILENAME`",
		EnvVar: "PHOTOPRISM_LOG_FILENAME",
		Value:  "",
	},
	cli.StringFlag{
		Name:   "pid-filename",
		Usage:  "sets the server process id `FILENAME`",
		EnvVar: "PHOTOPRISM_PID_FILENAME",
	},
	cli.StringFlag{
		Name:   "cdn-url",
		Usage:  "sets an optional content delivery network `URL` (optional)",
		EnvVar: "PHOTOPRISM_CDN_URL",
	},
	cli.StringFlag{
		Name:   "site-url",
		Usage:  "sets the public site `URL`",
		Value:  "http://localhost:2342/",
		EnvVar: "PHOTOPRISM_SITE_URL",
	},
	cli.StringFlag{
		Name:   "site-preview",
		Usage:  "sets the public preview image `URL`",
		EnvVar: "PHOTOPRISM_SITE_PREVIEW",
	},
	cli.StringFlag{
		Name:   "site-title",
		Usage:  "sets the site title",
		Value:  "PhotoPrism",
		EnvVar: "PHOTOPRISM_SITE_TITLE",
	},
	cli.StringFlag{
		Name:   "site-author",
		Usage:  "sets the artist name or copyright",
		EnvVar: "PHOTOPRISM_SITE_AUTHOR",
	},
	cli.StringFlag{
		Name:   "site-caption",
		Usage:  "sets the site caption",
		Value:  "Browse Your Life",
		EnvVar: "PHOTOPRISM_SITE_CAPTION",
	},
	cli.StringFlag{
		Name:   "site-description",
		Usage:  "sets an optional site description",
		EnvVar: "PHOTOPRISM_SITE_DESCRIPTION",
	},
	cli.IntFlag{
		Name:   "http-port",
		Value:  2342,
		Usage:  "sets the http server port `NUMBER`",
		EnvVar: "PHOTOPRISM_HTTP_PORT",
	},
	cli.StringFlag{
		Name:   "http-host",
		Usage:  "sets the http server `IP` address",
		EnvVar: "PHOTOPRISM_HTTP_HOST",
	},
	cli.StringFlag{
		Name:   "http-mode, m",
		Usage:  "selects the http server `MODE` (debug, release, or test)",
		EnvVar: "PHOTOPRISM_HTTP_MODE",
	},
	cli.StringFlag{
		Name:   "http-compression, z",
		Usage:  "improves transfer speed and bandwidth utilization (none or gzip)",
		EnvVar: "PHOTOPRISM_HTTP_COMPRESSION",
	},
	cli.StringFlag{
		Name:   "database-driver",
		Usage:  "selects the database `DRIVER` (sqlite or mysql)",
		Value:  "sqlite",
		EnvVar: "PHOTOPRISM_DATABASE_DRIVER",
	},
	cli.StringFlag{
		Name:   "database-dsn",
		Usage:  "sets the sqlite file name, specifying a `DSN` is optional other drivers",
		EnvVar: "PHOTOPRISM_DATABASE_DSN",
	},
	cli.StringFlag{
		Name:   "database-server",
		Usage:  "sets the database server `HOST` and port e.g. mysql:3306",
		EnvVar: "PHOTOPRISM_DATABASE_SERVER",
	},
	cli.StringFlag{
		Name:   "database-name",
		Value:  "photoprism",
		Usage:  "sets the database schema `NAME`",
		EnvVar: "PHOTOPRISM_DATABASE_NAME",
	},
	cli.StringFlag{
		Name:   "database-user",
		Value:  "photoprism",
		Usage:  "sets the database user `NAME`",
		EnvVar: "PHOTOPRISM_DATABASE_USER",
	},
	cli.StringFlag{
		Name:   "database-password",
		Usage:  "sets the database user `PASSWORD`",
		EnvVar: "PHOTOPRISM_DATABASE_PASSWORD",
	},
	cli.IntFlag{
		Name:   "database-conns",
		Usage:  "limits the `NUMBER` of open database connections",
		EnvVar: "PHOTOPRISM_DATABASE_CONNS",
	},
	cli.IntFlag{
		Name:   "database-conns-idle",
		Usage:  "limits the `NUMBER` of idle database connections",
		EnvVar: "PHOTOPRISM_DATABASE_CONNS_IDLE",
	},
	cli.BoolFlag{
		Name:   "raw-presets",
		Usage:  "enables RAW file converter presets (may reduce performance)",
		EnvVar: "PHOTOPRISM_RAW_PRESETS",
	},
	cli.StringFlag{
		Name:   "darktable-bin",
		Usage:  "sets the Darktable CLI `COMMAND` for RAW file conversion",
		Value:  "darktable-cli",
		EnvVar: "PHOTOPRISM_DARKTABLE_BIN",
	},
	cli.StringFlag{
		Name:   "darktable-blacklist",
		Usage:  "disables using Darktable for specific file `EXTENSIONS`",
		Value:  "raf,cr3,dng",
		EnvVar: "PHOTOPRISM_DARKTABLE_BLACKLIST",
	},
	cli.StringFlag{
		Name:   "rawtherapee-bin",
		Usage:  "sets the RawTherapee CLI `COMMAND` for RAW file conversion",
		Value:  "rawtherapee-cli",
		EnvVar: "PHOTOPRISM_RAWTHERAPEE_BIN",
	},
	cli.StringFlag{
		Name:   "rawtherapee-blacklist",
		Usage:  "disables using RawTherapee for specific file `EXTENSIONS`",
		Value:  "",
		EnvVar: "PHOTOPRISM_RAWTHERAPEE_BLACKLIST",
	},
	cli.StringFlag{
		Name:   "sips-bin",
		Usage:  "sets the Sips `COMMAND` for RAW file conversion on macOS",
		Value:  "sips",
		EnvVar: "PHOTOPRISM_SIPS_BIN",
	},
	cli.StringFlag{
		Name:   "heifconvert-bin",
		Usage:  "sets the HEIC/HEIF image convert `COMMAND`",
		Value:  "heif-convert",
		EnvVar: "PHOTOPRISM_HEIFCONVERT_BIN",
	},
	cli.StringFlag{
		Name:   "ffmpeg-bin",
		Usage:  "sets the FFmpeg `COMMAND` for video transcoding and cover images",
		Value:  "ffmpeg",
		EnvVar: "PHOTOPRISM_FFMPEG_BIN",
	},
	cli.StringFlag{
		Name:   "ffmpeg-encoder",
		Usage:  "sets the FFmpeg AVC encoder `NAME`",
		Value:  "libx264",
		EnvVar: "PHOTOPRISM_FFMPEG_ENCODER",
	},
	cli.IntFlag{
		Name:   "ffmpeg-bitrate",
		Usage:  "limits the FFmpeg encoding `BITRATE` (Mbit/s)",
		Value:  50,
		EnvVar: "PHOTOPRISM_FFMPEG_BITRATE",
	},
	cli.IntFlag{
		Name:   "ffmpeg-buffers",
		Usage:  "adjusts the number of FFmpeg capture buffers",
		Value:  32,
		EnvVar: "PHOTOPRISM_FFMPEG_BUFFERS",
	},
	cli.StringFlag{
		Name:   "exiftool-bin",
		Usage:  "sets the ExifTool `COMMAND` for extracting metadata",
		Value:  "exiftool",
		EnvVar: "PHOTOPRISM_EXIFTOOL_BIN",
	},
	cli.StringFlag{
		Name:   "download-token",
		Usage:  "sets a custom download url `TOKEN`",
		EnvVar: "PHOTOPRISM_DOWNLOAD_TOKEN",
	},
	cli.StringFlag{
		Name:   "preview-token",
		Usage:  "sets a custom thumbnail and video streaming url `TOKEN`",
		EnvVar: "PHOTOPRISM_PREVIEW_TOKEN",
	},
	cli.StringFlag{
		Name:   "thumb-filter",
		Usage:  "selects the thumbnail downscaling `FILTER` (best to worst: blackman, lanczos, cubic, linear)",
		Value:  "lanczos",
		EnvVar: "PHOTOPRISM_THUMB_FILTER",
	},
	cli.IntFlag{
		Name:   "thumb-size, s",
		Usage:  "limits the pre-cached thumbnail size in `PIXELS` (720-7680)",
		Value:  2048,
		EnvVar: "PHOTOPRISM_THUMB_SIZE",
	},
	cli.BoolFlag{
		Name:   "thumb-uncached, u",
		Usage:  "enables on-demand thumbnail generation (high memory and cpu usage)",
		EnvVar: "PHOTOPRISM_THUMB_UNCACHED",
	},
	cli.IntFlag{
		Name:   "thumb-size-uncached, x",
		Usage:  "limits the on-demand thumbnail size in `PIXELS` (720-7680)",
		Value:  7680,
		EnvVar: "PHOTOPRISM_THUMB_SIZE_UNCACHED",
	},
	cli.IntFlag{
		Name:   "jpeg-size",
		Usage:  "limits the size of generated JPEG files in `PIXELS` (720-30000)",
		Value:  7680,
		EnvVar: "PHOTOPRISM_JPEG_SIZE",
	},
	cli.IntFlag{
		Name:   "jpeg-quality, q",
		Usage:  "adjusts the `QUALITY` of generated JPEG files, a higher value reduces compression (25-100)",
		Value:  92,
		EnvVar: "PHOTOPRISM_JPEG_QUALITY",
	},
	cli.IntFlag{
		Name:   "face-size",
		Usage:  "sets the min face size in `PIXELS`",
		Value:  face.SizeThreshold,
		EnvVar: "PHOTOPRISM_FACE_SIZE",
	},
	cli.Float64Flag{
		Name:   "face-score",
		Usage:  "adjusts the quality `THRESHOLD` for faces",
		Value:  face.ScoreThreshold,
		EnvVar: "PHOTOPRISM_FACE_SCORE",
	},
	cli.IntFlag{
		Name:   "face-overlap",
		Usage:  "adjusts the face area overlap threshold in `PERCENT`",
		Value:  face.OverlapThreshold,
		EnvVar: "PHOTOPRISM_FACE_OVERLAP",
	},
	cli.IntFlag{
		Name:   "face-cluster-size",
		Usage:  "sets the min size of faces forming a cluster in `PIXELS`",
		Value:  face.ClusterSizeThreshold,
		EnvVar: "PHOTOPRISM_FACE_CLUSTER_SIZE",
	},
	cli.IntFlag{
		Name:   "face-cluster-score",
		Usage:  "adjusts the quality `THRESHOLD` for faces forming a cluster",
		Value:  face.ClusterScoreThreshold,
		EnvVar: "PHOTOPRISM_FACE_CLUSTER_SCORE",
	},
	cli.IntFlag{
		Name:   "face-cluster-core",
		Usage:  "determines the `NUMBER` of faces forming a cluster core",
		Value:  face.ClusterCore,
		EnvVar: "PHOTOPRISM_FACE_CLUSTER_CORE",
	},
	cli.Float64Flag{
		Name:   "face-cluster-dist",
		Usage:  "determines how close faces forming a cluster core must be to each other",
		Value:  face.ClusterDist,
		EnvVar: "PHOTOPRISM_FACE_CLUSTER_DIST",
	},
	cli.Float64Flag{
		Name:   "face-match-dist",
		Usage:  "adjusts the similarity offset for matching faces with existing clusters",
		Value:  face.MatchDist,
		EnvVar: "PHOTOPRISM_FACE_MATCH_DIST",
	},
}
