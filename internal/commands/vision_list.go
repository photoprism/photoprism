package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// VisionListCommand configures the command name, flags, and action.
var VisionListCommand = &cli.Command{
	Name:   "ls",
	Usage:  "Lists the configured computer vision models",
	Flags:  report.CliFlags,
	Action: visionListAction,
}

// visionEndpoint renders a service endpoint for display, without the credentials that
// Service.Endpoint injects into the URL for the request itself.
func visionEndpoint(uri, method string) string {
	if uri == "" || method == "" {
		return ""
	}

	if redacted := clean.UriRedacted(uri); redacted != "" {
		uri = redacted
	} else {
		// An unparsable URI is shown as a placeholder: it may still carry credentials.
		uri = "?"
	}

	return fmt.Sprintf("%s %s", method, uri)
}

// visionListAction displays the configured computer vision models.
func visionListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		var rows [][]string

		cols := []string{
			"Model",
			"Type",
			"Engine",
			"Endpoint",
			"Format",
			"Normalize",
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

			// Normalization only runs on the response of a remote labels model, and the
			// effective mode is shown because an unset value resolves to one.
			var normalize string

			if model.Type == vision.ModelTypeLabels && modelUri != "" && modelMethod != "" {
				normalize = model.GetNormalize()
			}

			rows[i] = []string{
				name,
				model.Type,
				engine,
				visionEndpoint(modelUri, modelMethod),
				format,
				normalize,
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
