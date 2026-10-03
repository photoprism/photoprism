package commands

import (
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/pkg/txt/report"
)

// VisionStatusCommand reports which models the computer vision features use, when they run, and
// the options that decide it, which "vision ls" cannot state for modes and detection settings.
var VisionStatusCommand = &cli.Command{
	Name:    "status",
	Aliases: []string{"config"},
	Usage:   "Reports the current status and configuration details",
	Flags:   report.CliFlags,
	Action:  visionStatusAction,
}

// visionStatusAction prints the current status and configuration details.
func visionStatusAction(ctx *cli.Context) error {
	conf, err := InitCoreConfig(ctx, true)

	if err != nil {
		log.Debug(err)
	}

	// The face status line names the model the library holds, which needs a connection, and
	// resolving it for display does not write the result to "options.yml".
	conf.RegisterDb()
	conf.ResolveFaceModel()

	// The core config neither loads "vision.yml", which selects the models this report names, nor
	// applies the settings that route models to a service, so both happen here as in Init.
	conf.LoadVisionConfig()
	conf.PropagateVision()

	format, formatErr := report.CliFormatStrict(ctx)

	if formatErr != nil {
		return formatErr
	}

	return printStatusReport(format, conf.VisionStatus(), conf.VisionReportSections())
}
