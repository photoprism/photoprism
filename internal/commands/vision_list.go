package commands

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize/english"
	"github.com/urfave/cli/v2"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// VisionListCommand configures the command name, flags, and action.
var VisionListCommand = &cli.Command{
	Name:   "ls",
	Usage:  "Lists the configured computer vision models",
	Flags:  report.CliFlags,
	Action: visionListAction,
}

// visionListCols are the columns of the "vision ls" table.
var visionListCols = []string{
	"Model",
	"Type",
	"Engine",
	"Provider",
	"Endpoint",
	"Format",
	"Normalize",
	"Resolution",
	"Options",
	"Schedule",
	"Status",
	"Installed",
}

// visionListSettings holds the instance settings that decide what a listed model runs with,
// since "vision.yml" names neither the execution provider nor the face embedding model.
type visionListSettings struct {
	ModelsPath string
	Provider   onnx.Provider
	FaceModel  face.ModelName
	FaceActive bool
	FaceRun    vision.RunType
}

// newVisionListSettings resolves the listing settings from the instance configuration.
func newVisionListSettings(conf *config.Config) visionListSettings {
	faceModel := conf.EffectiveFaceModel()
	active := faceModel != face.ModelNone && !conf.DisableFaces()

	// Without a model in force, the setting explains why: "none", or a model that is unavailable.
	if faceModel == face.ModelNone {
		faceModel = conf.FaceModelSetting()
	}

	return visionListSettings{
		ModelsPath: conf.ModelsPath(),
		Provider:   conf.OnnxProvider(),
		FaceModel:  faceModel,
		FaceActive: active,
		FaceRun:    conf.FaceEngineRunType(),
	}
}

// visionEndpoint renders a service endpoint for display, without the credentials that
// Service.Endpoint injects into the URL for the request itself.
func visionEndpoint(uri, method string) string {
	if uri == "" || method == "" {
		return ""
	}

	if redacted := clean.UriRedacted(clean.Uri(uri)); redacted != "" {
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
		// Show log message.
		log.Infof("found %s", english.Plural(len(vision.Config.Models), "model", "models"))

		if len(vision.Config.Models) == 0 {
			return nil
		}

		settings := newVisionListSettings(conf)
		rows := make([][]string, 0, len(vision.Config.Models))

		for _, model := range vision.Config.Models {
			if model != nil {
				rows = append(rows, visionListRow(model, settings))
			}
		}

		result, err := report.RenderFormat(rows, visionListCols, report.CliFormat(ctx))

		fmt.Printf("\n%s\n", result)

		return err
	})
}

// visionListRow renders a model as a row of the "vision ls" table. Local face entries report the
// embedding model FACE_MODEL selects, because that model runs regardless of the entry.
func visionListRow(model *vision.Model, s visionListSettings) []string {
	modelUri, modelMethod := model.Endpoint()
	remote := modelUri != "" && modelMethod != ""

	name, _, _ := model.GetModel()
	engine := model.EngineName()
	resolution := model.Resolution
	run := model.RunType()
	enabled := !model.Disabled && !model.DisabledByMode
	installed := visionInstalled(model, s.ModelsPath)

	var options string

	if model.TensorFlow != nil {
		tags := ""

		if model.TensorFlow.Tags != nil {
			tags = strings.Join(model.TensorFlow.Tags, ", ")
		}

		options = fmt.Sprintf(`{"tags":"%s"}`, tags)
	} else if o := model.GetOptions(); o != nil {
		if b, err := json.Marshal(*o); err == nil && string(b) != "{}" {
			options = string(b)
		}
	}

	// FACE_RUN schedules faces and FACE_MODEL decides whether they are embedded, while a disabled
	// entry still stops them, since face processing then has no model to run.
	if model.Type == vision.ModelTypeFace {
		run, enabled = s.FaceRun, enabled && s.FaceActive
	}

	switch {
	case remote:
	case model.Type == vision.ModelTypeLabels, model.Type == vision.ModelTypeNsfw:
		// A model named in "vision.yml" is resolved from the registry when it is loaded.
		if width := visionRegisteredWidth(model); resolution == 0 && width > 0 {
			resolution = width
		}

		if engine == vision.EngineLocal {
			engine = vision.EngineONNX
		}
	case model.Type == vision.ModelTypeFace:
		registered := face.FindEmbeddingModel(s.FaceModel)
		installed = report.NotAssigned

		// A TensorFlow entry still runs while FACE_MODEL selects FaceNet, and loads from its own
		// path when it has a custom name, so only the registered one reports installation.
		if s.FaceModel == face.ModelFaceNet && model.TensorFlow != nil {
			if registered != nil && model.Name == face.ModelFaceNet {
				installed = report.Bool(registered.Installed(s.ModelsPath), report.Yes, report.No)
			}

			break
		}

		if registered != nil {
			installed = report.Bool(registered.Installed(s.ModelsPath), report.Yes, report.No)
		}

		name, options, resolution, engine = s.FaceModel, "", 0, ""

		if registered != nil {
			engine = registered.Runtime

			if width, _ := registered.InputSize(); width > 0 {
				resolution = width
			} else if registered.Name == face.ModelFaceNet {
				resolution = vision.FacenetModel.Resolution
			}
		}
	}

	if run == vision.RunAuto {
		run = "auto"
	}

	var provider string

	if !remote && engine == vision.EngineONNX {
		provider = s.Provider.String()
	}

	var format string

	if remote {
		format = model.EndpointRequestFormat()
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

	// Normalization only runs on the response of a remote labels model, and the
	// effective mode is shown because an unset value resolves to one.
	var normalize string

	if model.Type == vision.ModelTypeLabels && remote {
		normalize = model.GetNormalize()
	}

	var resolutionText string

	if resolution > 0 {
		resolutionText = fmt.Sprintf("%d", resolution)
	}

	return []string{
		name,
		model.Type,
		engine,
		provider,
		visionEndpoint(modelUri, modelMethod),
		format,
		normalize,
		resolutionText,
		options,
		run,
		report.Bool(enabled, report.Enabled, report.Disabled),
		installed,
	}
}

// visionRegisteredWidth returns the input width of the registered label or NSFW model a local
// entry names, or zero when it names none.
func visionRegisteredWidth(model *vision.Model) int {
	switch model.Type {
	case vision.ModelTypeLabels:
		if description := classify.FindModel(classify.ModelName(model.Name)); description != nil {
			width, _ := description.ONNX.InputSize()
			return width
		}
	case vision.ModelTypeNsfw:
		if description := nsfw.FindModel(nsfw.ModelName(model.Name)); description != nil {
			width, _ := description.ONNX.InputSize()
			return width
		}
	}

	return 0
}

// visionInstalled reports artifact presence without initializing an inference session.
func visionInstalled(model *vision.Model, modelsPath string) string {
	if uri, _ := model.Endpoint(); uri != "" || model.Service.UriUnresolved() {
		return report.NotAssigned
	}
	var filename string
	switch model.Type {
	case vision.ModelTypeLabels:
		if description := classify.FindModel(classify.ModelName(model.Name)); description != nil {
			return report.Bool(description.Installed(modelsPath), report.Yes, report.No)
		}
	case vision.ModelTypeNsfw:
		if description := nsfw.FindModel(nsfw.ModelName(model.Name)); description != nil {
			return report.Bool(description.Installed(modelsPath), report.Yes, report.No)
		}
	default:
		return report.NotAssigned
	}
	filename = model.Path
	if filename == "" {
		filename = clean.TypeLowerUnderscore(model.Name)
	}
	filename = filepath.Join(modelsPath, clean.Path(filename))
	if !strings.EqualFold(filepath.Ext(filename), ".onnx") {
		file := filepath.Base(filename) + ".onnx"
		if model.ONNX != nil && model.ONNX.File != "" {
			file = model.ONNX.File
		}
		filename = filepath.Join(filename, file)
	}
	return report.Bool(fs.FileExistsNotEmpty(filename), report.Yes, report.No)
}
