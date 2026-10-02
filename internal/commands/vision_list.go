package commands

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"slices"
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
	"Option",
}

// visionFaceCols are the columns of the face table in text and Markdown output.
var visionFaceCols = []string{
	"Model",
	"Role",
	"Engine",
	"Provider",
	"Resolution",
	"Schedule",
	"Status",
	"Installed",
	"Option",
}

const (
	visionModelsTitle   = "VISION MODELS"
	visionFacesTitle    = "FACE DETECTION & RECOGNITION"
	visionFacesNote     = `Selected with face-detector and face-model, not in vision.yml. Run them with "photoprism vision run -m face"; see "photoprism faces status" for details.`
	visionYamlOption    = "vision-yaml"
	faceDetectorOption  = "face-detector"
	faceModelOption     = "face-model"
	faceDetectionRole   = "detection"
	faceRecognitionRole = "recognition"
)

// visionFaceRoles names the role of each face model by the option that selects it.
var visionFaceRoles = map[string]string{
	faceDetectorOption: faceDetectionRole,
	faceModelOption:    faceRecognitionRole,
}

// visionListSettings holds the instance settings that decide what a listed model runs with,
// since "vision.yml" names neither the execution provider nor the face models.
type visionListSettings struct {
	ModelsPath     string
	Provider       onnx.Provider
	FaceModel      face.ModelName
	FaceActive     bool
	FaceDetector   face.DetectorName
	DetectorActive bool
	FaceRun        vision.RunType
}

// newVisionListSettings resolves the listing settings from the instance configuration.
func newVisionListSettings(conf *config.Config) visionListSettings {
	faceModel := conf.EffectiveFaceModel()
	active := faceModel != face.ModelNone && !conf.DisableFaces()

	// Without a model in force, the setting explains why: "none", or a model that is unavailable.
	// As for the detector, "auto" with nothing to derive it from stays "none".
	if setting := conf.FaceModelSetting(); faceModel == face.ModelNone && setting != face.ModelAuto {
		faceModel = setting
	}

	detector := conf.FaceDetector()
	detectorActive := detector != face.DetectorNone && !conf.DisableFaces()

	// Likewise for the detector.
	if setting := conf.FaceDetectorSetting(); detector == face.DetectorNone && setting != face.DetectorAuto {
		detector = setting
	}

	return visionListSettings{
		ModelsPath:     conf.ModelsPath(),
		Provider:       conf.OnnxProvider(),
		FaceModel:      faceModel,
		FaceActive:     active,
		FaceDetector:   detector,
		DetectorActive: detectorActive,
		FaceRun:        conf.FaceEngineRunType(),
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

// visionListAction displays the configured computer vision models and the face models.
func visionListAction(ctx *cli.Context) error {
	return CallWithDependencies(ctx, func(conf *config.Config) error {
		var models []*vision.Model

		if vision.Config != nil {
			models = vision.Config.Models
		}

		settings := newVisionListSettings(conf)
		rows := visionListRows(models, settings)
		faceRows := visionFaceRows(models, settings)

		// Show log message.
		log.Infof("found %s", english.Plural(len(rows)+len(faceRows), "model", "models"))

		return printVisionList(report.CliFormat(ctx), rows, faceRows)
	})
}

// printVisionList prints the vision and face tables in text and Markdown output, and a single
// flat table for CSV, TSV and JSON, so that exports keep one header and one array.
func printVisionList(format report.Format, rows, faceRows [][]string) error {
	switch format {
	case report.CSV, report.TSV, report.JSON:
		result, err := report.RenderFormat(slices.Concat(rows, faceRows), visionListCols, format)
		fmt.Printf("\n%s\n", result)
		return err
	}

	fmt.Println()

	// Every row of this table is configured in "vision.yml", so its Option column would only repeat that.
	optionCol := visionListColumn("Option")
	visionRows := make([][]string, 0, len(rows))

	for _, row := range rows {
		visionRows = append(visionRows, row[:optionCol])
	}

	sections := []config.StatusSection{
		{Title: visionModelsTitle, Cols: visionListCols[:optionCol], Rows: visionRows},
		{Title: visionFacesTitle, Cols: visionFaceCols, Rows: visionFaceTable(faceRows), Note: visionFacesNote},
	}

	for _, section := range sections {
		if len(section.Rows) == 0 {
			continue
		}

		result, err := report.RenderFormat(section.Rows, section.Cols, format)

		if err != nil {
			return err
		}

		if format == report.Markdown {
			fmt.Printf("### %s\n\n", section.Title)
		} else {
			fmt.Printf("%s\n\n", section.Title)
		}

		fmt.Println(result)

		if section.Note != "" {
			fmt.Printf("%s\n\n", section.Note)
		}
	}

	return nil
}

// visionListRows renders the "vision.yml" entries, except local face entries, whose models are
// selected with face-detector and face-model and are listed by visionFaceRows instead.
func visionListRows(models []*vision.Model, s visionListSettings) [][]string {
	rows := make([][]string, 0, len(models))

	for _, model := range models {
		if model == nil || model.Type == vision.ModelTypeFace && !visionRemote(model) {
			continue
		}

		rows = append(rows, visionListRow(model, s))
	}

	return rows
}

// visionFaceRows renders the face detector and the face embedding model as rows of the flat table.
func visionFaceRows(models []*vision.Model, s visionListSettings) [][]string {
	entry := visionFaceEntry(models)
	enabled := entry != nil && !entry.Disabled && !entry.DisabledByMode

	return [][]string{
		visionDetectorRow(enabled, s),
		visionRecognitionRow(entry, s),
	}
}

// visionFaceEntry returns the face entry in force, as vision.Config.Model resolves it, or the
// last one listed when all are disabled, since its settings still describe the model.
func visionFaceEntry(models []*vision.Model) *vision.Model {
	var last *vision.Model

	for i := len(models) - 1; i >= 0; i-- {
		if m := models[i]; m == nil || m.Type != vision.ModelTypeFace {
			continue
		} else if !m.Disabled && !m.DisabledByMode {
			return m
		} else if last == nil {
			last = m
		}
	}

	return last
}

// visionDetectorRow renders the face detector face-detector selects. It runs while faces are
// enabled and a "vision.yml" face entry is in force, even when that entry uses a service.
func visionDetectorRow(entryEnabled bool, s visionListSettings) []string {
	row := make([]string, len(visionListCols))
	installed := report.NotAssigned

	if detector := face.FindDetector(s.FaceDetector); detector != nil {
		row[visionListColumn("Engine")] = vision.EngineONNX
		row[visionListColumn("Provider")] = s.Provider.String()
		installed = report.Bool(detector.Installed(s.ModelsPath), report.Yes, report.No)

		if width, _ := detector.ONNX.InputSize(); width > 0 {
			row[visionListColumn("Resolution")] = fmt.Sprintf("%d", width)
		}
	}

	row[visionListColumn("Model")] = s.FaceDetector
	row[visionListColumn("Type")] = vision.ModelTypeFace
	row[visionListColumn("Schedule")] = visionRunText(s.FaceRun)
	row[visionListColumn("Status")] = report.Bool(entryEnabled && s.DetectorActive, report.Enabled, report.Disabled)
	row[visionListColumn("Installed")] = installed
	row[visionListColumn("Option")] = faceDetectorOption

	return row
}

// visionRecognitionRow renders the face embedding model. A face entry with a service endpoint
// returns the vectors, which are recorded under the face-model name, so no local model is shown.
func visionRecognitionRow(entry *vision.Model, s visionListSettings) []string {
	if entry == nil {
		entry = vision.FacenetModel.Clone()
		entry.Disabled = true
	}

	row := visionListRow(entry, s)

	if visionRemote(entry) {
		row[visionListColumn("Model")] = s.FaceModel
		row[visionListColumn("Provider")] = ""
		row[visionListColumn("Resolution")] = ""
		row[visionListColumn("Installed")] = report.NotAssigned
	}

	row[visionListColumn("Endpoint")] = ""
	row[visionListColumn("Option")] = faceModelOption

	return row
}

// visionFaceTable projects flat face rows onto the columns of the face table.
func visionFaceTable(rows [][]string) [][]string {
	result := make([][]string, 0, len(rows))

	for _, row := range rows {
		option := row[visionListColumn("Option")]
		out := make([]string, 0, len(visionFaceCols))

		for _, col := range visionFaceCols {
			if col == "Role" {
				out = append(out, visionFaceRoles[option])
			} else {
				out = append(out, row[visionListColumn(col)])
			}
		}

		result = append(result, out)
	}

	return result
}

// visionListColumn returns the index of a column in the flat table. It panics on an unknown name,
// which only a typo in this file can produce.
func visionListColumn(col string) int {
	if i := slices.Index(visionListCols, col); i >= 0 {
		return i
	}

	panic(fmt.Sprintf("vision ls: unknown column %s", col))
}

// visionRemote reports whether a model is run by a service endpoint.
func visionRemote(model *vision.Model) bool {
	uri, method := model.Endpoint()
	return uri != "" && method != ""
}

// visionRunText renders a run type for display, naming the automatic schedule.
func visionRunText(run vision.RunType) string {
	if run == vision.RunAuto {
		return "auto"
	}

	return run
}

// visionListRow renders a model as a row of the flat table. Face entries report the embedding
// model FACE_MODEL selects, because that model runs regardless of the entry.
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
		visionRunText(run),
		report.Bool(enabled, report.Enabled, report.Disabled),
		installed,
		visionYamlOption,
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
