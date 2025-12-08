package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// VisionListCommand configures the command name, flags, and action.
var VisionListCommand = &cli.Command{
	Name:   "ls",
	Usage:  "Lists the configured computer vision models",
	Flags:  report.CliFlags,
	Action: visionListAction,
}

// visionListAction displays existing user accounts.
func visionListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		var rows [][]string

		cols := []string{
			"Model",
			"Type",
			"Engine",
			"Endpoint",
			"Format",
			"Resolution",
			"Options",
			"Schedule",
			"Status",
		}

		// Show log message.
		log.Infof("found %s", english.Plural(len(vision.Config.Models), "model", "models"))

		if n := len(vision.Config.Models); n == 0 {
			return nil
		} else {
			rows = make([][]string, n)
		}

		// Display report.
		for i, model := range vision.Config.Models {
			modelUri, modelMethod := model.Endpoint()
			tags := ""

			name, _, _ := model.GetModel()

			if model.TensorFlow != nil && model.TensorFlow.Tags != nil {
				tags = strings.Join(model.TensorFlow.Tags, ", ")
			}

			var options []byte
			if o := model.GetOptions(); o != nil {
				options, _ = json.Marshal(*o)
			}

			var format string

			if modelUri != "" && modelMethod != "" {
				if f := model.EndpointRequestFormat(); f != "" {
					format = f
				}
			}

			if responseFormat := model.GetFormat(); responseFormat != "" {
				if format != "" {
					format = fmt.Sprintf("%s:%s", format, responseFormat)
				} else {
					format = responseFormat
				}
			}

			if format == "" && model.Default {
				format = "default"
			}

			var run string

			if run = model.RunType(); run == "" {
				run = "auto"
			}

			engine := model.EngineName()

			rows[i] = []string{
				name,
				model.Type,
				engine,
				fmt.Sprintf("%s %s", modelMethod, modelUri),
				format,
				fmt.Sprintf("%d", model.Resolution),
				report.Bool(model.TensorFlow != nil, fmt.Sprintf(`{"tags":"%s"}`, tags), string(options)),
				run,
				report.Bool(model.Disabled, report.Disabled, report.Enabled),
			}
		}

		result, err := report.RenderFormat(rows, cols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}
