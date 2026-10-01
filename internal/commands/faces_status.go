package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/txt/report"
)

// FacesStatusCommand reports the face detection and recognition configuration together with the
// state a database connection reveals: the model the library holds, whether embeddings are paused,
// and which threshold is holding automatic clustering back. It is named for the status rather than
// the configuration, because `photoprism config` reports the same options and cannot state any of it.
var FacesStatusCommand = &cli.Command{
	Name:    "status",
	Aliases: []string{"config"},
	Usage:   "Reports the current status and configuration details",
	Flags:   report.CliFlags,
	Action:  facesStatusAction,
}

// facesStatusAction prints the current status and configuration details.
func facesStatusAction(ctx *cli.Context) error {
	conf, err := InitCoreConfig(ctx, true)

	if err != nil {
		log.Debug(err)
	}

	// Detecting the model asks the library which one produced its vectors, so without a
	// connection this report could not name the model that is in force. Connecting is
	// idempotent and fails over to the old behavior.
	conf.RegisterDb()

	// A report states what is configured and must not change it, so this resolves the model
	// for display without writing the result to "options.yml".
	conf.ResolveFaceModel()

	format, formatErr := report.CliFormatStrict(ctx)

	if formatErr != nil {
		return formatErr
	}

	status := conf.FaceStatus()
	sections := conf.FaceReportSections()

	return printStatusReport(format, status, sections)
}
