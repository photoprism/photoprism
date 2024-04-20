package commands

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/report"
)

// ShowConfigOptionsCommand configures the command name, flags, and action.
var ShowConfigOptionsCommand = cli.Command{
	Name:   "config-options",
	Usage:  "Displays supported environment variables and CLI flags",
	Flags:  report.CliFlags,
	Action: showConfigOptionsAction,
}

var faceFlagsInfo = `!!! info ""
    To [recognize faces](https://docs.photoprism.app/user-guide/organize/people/), PhotoPrism first extracts crops from your images using a [library](https://github.com/esimov/pigo) based on [pixel intensity comparisons](https://dl.photoprism.app/pdf/20140820-Pixel_Intensity_Comparisons.pdf). These are then fed into TensorFlow to compute [512-dimensional vectors](https://dl.photoprism.app/pdf/20150101-FaceNet.pdf) for characterization. In the final step, the [DBSCAN algorithm](https://en.wikipedia.org/wiki/DBSCAN) attempts to cluster these so-called face embeddings, so they can be matched to persons with just a few clicks. A reasonable range for the similarity distance between face embeddings is between 0.60 and 0.70, with a higher value being more aggressive and leading to larger clusters with more false positives. To cluster a smaller number of faces, you can reduce the core to 3 or 2 similar faces.

We recommend that only advanced users change these parameters:`

// showConfigOptionsAction displays supported environment variables and CLI flags.
func showConfigOptionsAction(ctx *cli.Context) error {
	conf := config.NewConfig(ctx)
	conf.SetLogLevel(logrus.FatalLevel)

	rows, cols := config.Flags.Report()

	// CSV Export?
	if ctx.Bool("csv") || ctx.Bool("tsv") {
		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Println(result)

		return err
	}

	type Section struct {
		Start string
		Title string
		Info  string
	}

	s := []Section{
		{Start: "PHOTOPRISM_ADMIN_PASSWORD", Title: "Authentication"},
		{Start: "PHOTOPRISM_LOG_LEVEL", Title: "Logging"},
		{Start: "PHOTOPRISM_CONFIG_PATH", Title: "Storage"},
		{Start: "PHOTOPRISM_WORKERS", Title: "Index Workers"},
		{Start: "PHOTOPRISM_READONLY", Title: "Feature Flags"},
		{Start: "PHOTOPRISM_DEFAULT_LOCALE", Title: "Customization"},
		{Start: "PHOTOPRISM_SITE_URL", Title: "Site Information"},
		{Start: "PHOTOPRISM_HTTPS_PROXY", Title: "Proxy Servers"},
		{Start: "PHOTOPRISM_DISABLE_TLS", Title: "Web Server"},
		{Start: "PHOTOPRISM_DATABASE_DRIVER", Title: "Database Connection"},
		{Start: "PHOTOPRISM_SIPS_BIN", Title: "File Converters"},
		{Start: "PHOTOPRISM_DOWNLOAD_TOKEN", Title: "Security Tokens"},
		{Start: "PHOTOPRISM_THUMB_COLOR", Title: "Image Quality"},
		{Start: "PHOTOPRISM_FACE_SIZE", Title: "Face Recognition",
			Info: faceFlagsInfo},
		{Start: "PHOTOPRISM_PID_FILENAME", Title: "Daemon Mode",
			Info: "If you start the server as a *daemon* in the background, you can additionally specify a filename for the log and the process ID:"},
	}

	j := 0

	for i, sec := range s {
		fmt.Printf("### %s ###\n\n", sec.Title)
		if sec.Info != "" && ctx.Bool("md") {
			fmt.Printf("%s\n\n", sec.Info)
		}

		secRows := make([][]string, 0, len(rows))

		for {
			row := rows[j]

			if len(row) < 1 {
				continue
			}

			if i < len(s)-1 {
				if s[i+1].Start == row[0] {
					break
				}
			}

			secRows = append(secRows, row)
			j++

			if j >= len(rows) {
				break
			}
		}

		result, err := report.RenderFormat(secRows, cols, report.CliFormat(ctx))

		if err != nil {
			return err
		}

		fmt.Println(result)

		if j >= len(rows) {
			break
		}
	}

	return nil
}
