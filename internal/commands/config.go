package commands

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/urfave/cli"
)

// ConfigCommand is used to register the display config cli command
var ConfigCommand = cli.Command{
	Name:   "config",
	Usage:  "Displays global configuration values",
	Action: configAction,
}

// configAction prints current configuration
func configAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)

	fmt.Printf("%-25s VALUE\n", "NAME")

	// Description
	fmt.Printf("%-25s %s\n", "name", conf.Name())
	fmt.Printf("%-25s %s\n", "url", conf.Url())
	fmt.Printf("%-25s %s\n", "title", conf.Title())
	fmt.Printf("%-25s %s\n", "subtitle", conf.Subtitle())
	fmt.Printf("%-25s %s\n", "description", conf.Description())
	fmt.Printf("%-25s %s\n", "author", conf.Author())
	fmt.Printf("%-25s %s\n", "version", conf.Version())
	fmt.Printf("%-25s %s\n", "copyright", conf.Copyright())

	// Flags
	fmt.Printf("%-25s %t\n", "debug", conf.Debug())
	fmt.Printf("%-25s %t\n", "read-only", conf.ReadOnly())
	fmt.Printf("%-25s %t\n", "public", conf.Public())
	fmt.Printf("%-25s %t\n", "experimental", conf.Experimental())
	fmt.Printf("%-25s %t\n", "disable-settings", conf.DisableSettings())

	// TensorFlow
	fmt.Printf("%-25s %t\n", "disable-tf", conf.DisableTensorFlow())
	fmt.Printf("%-25s %s\n", "tf-version", conf.TensorFlowVersion())
	fmt.Printf("%-25s %s\n", "tf-model-path", conf.TensorFlowModelPath())
	fmt.Printf("%-25s %t\n", "detect-nsfw", conf.DetectNSFW())
	fmt.Printf("%-25s %t\n", "upload-nsfw", conf.UploadNSFW())

	// Passwords
	fmt.Printf("%-25s %s\n", "admin-password", conf.AdminPassword())
	fmt.Printf("%-25s %s\n", "webdav-password", conf.WebDAVPassword())

	// Background workers and logging
	fmt.Printf("%-25s %d\n", "workers", conf.Workers())
	fmt.Printf("%-25s %d\n", "wakeup-interval", conf.WakeupInterval()/time.Second)
	fmt.Printf("%-25s %s\n", "log-level", conf.LogLevel())

	// Path and file names
	fmt.Printf("%-25s %s\n", "log-filename", conf.LogFilename())
	fmt.Printf("%-25s %s\n", "pid-filename", conf.PIDFilename())
	fmt.Printf("%-25s %s\n", "config-file", conf.ConfigFile())
	fmt.Printf("%-25s %s\n", "config-path", conf.ConfigPath())
	fmt.Printf("%-25s %s\n", "assets-path", conf.AssetsPath())
	fmt.Printf("%-25s %s\n", "originals-path", conf.OriginalsPath())
	fmt.Printf("%-25s %s\n", "import-path", conf.ImportPath())
	fmt.Printf("%-25s %s\n", "temp-path", conf.TempPath())
	fmt.Printf("%-25s %s\n", "cache-path", conf.CachePath())
	fmt.Printf("%-25s %s\n", "resources-path", conf.ResourcesPath())

	// HTTP server config
	fmt.Printf("%-25s %s\n", "favicons-path", conf.HttpFaviconsPath())
	fmt.Printf("%-25s %s\n", "static-path", conf.HttpStaticPath())
	fmt.Printf("%-25s %s\n", "static-build-path", conf.HttpStaticBuildPath())
	fmt.Printf("%-25s %s\n", "templates-path", conf.HttpTemplatesPath())
	fmt.Printf("%-25s %s\n", "http-template", conf.HttpDefaultTemplate())
	fmt.Printf("%-25s %s\n", "http-host", conf.HttpServerHost())
	fmt.Printf("%-25s %d\n", "http-port", conf.HttpServerPort())
	fmt.Printf("%-25s %s\n", "http-mode", conf.HttpServerMode())

	// Built-in TiDB server config
	fmt.Printf("%-25s %s\n", "tidb-host", conf.TidbServerHost())
	fmt.Printf("%-25s %d\n", "tidb-port", conf.TidbServerPort())
	fmt.Printf("%-25s %s\n", "tidb-password", conf.TidbServerPassword())
	fmt.Printf("%-25s %s\n", "tidb-path", conf.TidbServerPath())

	// Database config
	fmt.Printf("%-25s %s\n", "database-driver", conf.DatabaseDriver())
	fmt.Printf("%-25s %s\n", "database-dsn", conf.DatabaseDsn())

	// External binaries
	fmt.Printf("%-25s %s\n", "sips-bin", conf.SipsBin())
	fmt.Printf("%-25s %s\n", "darktable-bin", conf.DarktableBin())
	fmt.Printf("%-25s %s\n", "heifconvert-bin", conf.HeifConvertBin())
	fmt.Printf("%-25s %s\n", "ffmpeg-bin", conf.FFmpegBin())
	fmt.Printf("%-25s %s\n", "exiftool-bin", conf.ExifToolBin())
	fmt.Printf("%-25s %t\n", "sidecar-json", conf.SidecarJson())

	// Places / Geocoding API
	fmt.Printf("%-25s %s\n", "geocoding-api", conf.GeoCodingApi())

	// Thumbnails
	fmt.Printf("%-25s %d\n", "jpeg-quality", conf.JpegQuality())
	fmt.Printf("%-25s %s\n", "thumb-filter", conf.ThumbFilter())
	fmt.Printf("%-25s %t\n", "thumb-uncached", conf.ThumbUncached())
	fmt.Printf("%-25s %d\n", "thumb-size", conf.ThumbSize())
	fmt.Printf("%-25s %d\n", "thumb-limit", conf.ThumbLimit())
	fmt.Printf("%-25s %s\n", "thumb-path", conf.ThumbPath())

	return nil
}
