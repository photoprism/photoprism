package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/nsfw"
	"github.com/photoprism/photoprism/internal/ai/onnx"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
)

// VisionYaml returns the path to the computer-vision configuration file,
// preferring an explicit override and otherwise letting fs.ConfigFilePath pick
// the right `.yml`/`.yaml` variant in the config directory.
func (c *Config) VisionYaml() string {
	if c == nil {
		return ""
	}

	if c.options.VisionYaml != "" {
		return fs.Abs(c.options.VisionYaml)
	} else {
		return fs.ConfigFilePath(c.ConfigPath(), "vision", fs.ExtYml)
	}
}

// LoadVisionConfig applies the optional "vision.yml", which schedules the label, NSFW and
// caption models. Faces are configured through FACE_* options only, so a face entry there is
// read and reported as ignored rather than obeyed.
func (c *Config) LoadVisionConfig() {
	if c == nil || vision.Config == nil {
		return
	}

	visionYaml := c.VisionYaml()

	if fs.FileExistsNotEmpty(visionYaml) {
		if err := vision.Config.Load(visionYaml); err != nil {
			log.Warnf("vision: %s", clean.Error(err))
		}

		c.reportIgnoredFaceRun(visionYaml)
	}

	c.applyLabelModel()
	c.applyNSFWModel()
	c.reportVisionModes()
	c.reportUnscreenedUploads()
}

// PropagateVision applies the settings the vision package reads, such as the global service URI
// that routes models without their own endpoint. Propagate calls it, and a command that reports
// on the core config calls it directly, since that config is never propagated.
func (c *Config) PropagateVision() {
	vision.SetCachePath(c.CachePath())
	vision.SetModelsPath(c.ModelsPath())
	vision.SetOnnxProvider(c.OnnxProvider())
	vision.ServiceApi = c.VisionApi()
	vision.ServiceUri = c.VisionUri()
	vision.ServiceKey = c.VisionKey()
	vision.DownloadUrl = c.DownloadUrl()
	vision.DetectNSFWLabels = c.DetectNSFWLabels()
}

// NSFWModelSetting returns the dedicated, disabled, or labels detection mode.
func (c *Config) NSFWModelSetting() nsfw.ModelName {
	if c == nil {
		return nsfw.ModelNone
	}
	switch strings.ToLower(strings.TrimSpace(c.options.NsfwModel)) {
	case "none":
		return nsfw.ModelNone
	case "labels":
		return nsfw.ModelName("labels")
	default:
		return nsfw.ModelAuto
	}
}

// EffectiveNSFWModel returns the local detector selected for this instance.
func (c *Config) EffectiveNSFWModel() nsfw.ModelName {
	if c.NSFWModelSetting() != nsfw.ModelAuto {
		return nsfw.ModelNone
	}
	if vision.Config != nil {
		if model := configuredVisionModel(vision.Config, vision.ModelTypeNsfw); model != nil {
			if model.Disabled || model.DisabledByMode {
				return nsfw.ModelNone
			}
			if !model.Default {
				// A model served by an endpoint is reported by the name it is requested with.
				if visionModelRemote(model) {
					name, _, _ := model.GetModel()
					return nsfw.ModelName(name)
				}

				return nsfw.NormalizeModelName(nsfw.ModelName(model.Name))
			}
		}
	}
	return c.installedNSFWModel()
}

// installedNSFWModel returns the first installed detector in automatic preference order.
func (c *Config) installedNSFWModel() nsfw.ModelName {
	if c == nil {
		return nsfw.ModelNone
	}

	modelsPath := c.ModelsPath()
	for _, candidate := range nsfw.AutoModelPreference {
		if nsfw.FindModel(candidate).Installed(modelsPath) {
			return candidate
		}
	}

	return nsfw.ModelNone
}

// applyNSFWModel applies NSFW_MODEL to the local detector entry in vision.Config.
func (c *Config) applyNSFWModel() {
	if c == nil || vision.Config == nil {
		return
	}
	current := configuredVisionModel(vision.Config, vision.ModelTypeNsfw)
	if current == nil {
		current = vision.NewNsfwModel(nsfw.DefaultModelName())
		vision.Config.SetModel(current)
	}
	for _, model := range vision.Config.Models {
		if model != nil && model.Type == vision.ModelTypeNsfw {
			model.DisabledByMode = c.NSFWModelSetting() != nsfw.ModelAuto
		}
	}
	if current.DisabledByMode || !current.Default {
		return
	}
	if selected := c.installedNSFWModel(); selected != nsfw.ModelNone {
		registered := vision.NewNsfwModel(selected)
		registered.Default, registered.Run, registered.Disabled = true, current.Run, current.Disabled
		vision.Config.SetModel(registered)
	}
}

// reportUnscreenedUploads warns when upload screening has no configured detector.
func (c *Config) reportUnscreenedUploads() {
	if c == nil || c.UploadNSFW() {
		return
	}
	if c.NSFWModelSetting() == nsfw.ModelName("labels") {
		event.SystemWarn([]string{"config", "uploads are not screened in nsfw-model labels mode"})
		return
	}
	if vision.Config == nil || vision.Config.Model(vision.ModelTypeNsfw) == nil {
		event.SystemWarn([]string{"config", "uploads are screened for offensive content, but no nsfw model is configured"})
		return
	}
	model := vision.Config.Model(vision.ModelTypeNsfw)
	if uri, _ := model.Endpoint(); uri != "" {
		return
	}
	if description := nsfw.FindModel(nsfw.ModelName(model.Name)); description != nil && !description.Installed(c.ModelsPath()) {
		event.SystemWarn([]string{"config", "uploads cannot be screened because nsfw model %s is not installed; run download-models.sh %s and restart PhotoPrism"}, model.Name, model.Name)
	}
}

// LabelModelSetting returns the labels mode, including deprecated disablement, which applies
// unless labels-model is set to a supported mode.
func (c *Config) LabelModelSetting() classify.ModelName {
	if c == nil {
		return classify.ModelNone
	}

	switch setting := strings.ToLower(strings.TrimSpace(c.options.LabelsModel)); setting {
	case "none":
		return classify.ModelNone
	case "auto":
		if c.options.DisableClassification {
			if _, reported := c.warnedOnce.LoadOrStore("disable-classification-ignored", true); !reported {
				log.Infof("config: disable-classification is ignored, because labels-model %s is configured", setting)
			}
		}
		return classify.ModelAuto
	default:
		if c.options.DisableClassification {
			return classify.ModelNone
		}
		return classify.ModelAuto
	}
}

// EffectiveLabelModel returns the local classifier selected for this instance.
func (c *Config) EffectiveLabelModel() classify.ModelName {
	setting := c.LabelModelSetting()
	if vision.Config != nil {
		if model := configuredVisionModel(vision.Config, vision.ModelTypeLabels); model != nil && (model.Disabled || model.DisabledByMode) {
			return classify.ModelNone
		}
	}

	if setting != classify.ModelAuto {
		return setting
	}

	if vision.Config != nil {
		if model := configuredVisionModel(vision.Config, vision.ModelTypeLabels); model != nil {
			if !model.Default {
				// A model served by an endpoint is reported by the name it is requested with.
				if visionModelRemote(model) {
					name, _, _ := model.GetModel()
					return classify.ModelName(name)
				}

				return classify.NormalizeModelName(classify.ModelName(model.Name))
			}
		}
	}

	return c.installedLabelModel()
}

// installedLabelModel returns the first installed classifier in automatic preference order.
func (c *Config) installedLabelModel() classify.ModelName {
	if c == nil {
		return classify.ModelNone
	}

	modelsPath := c.ModelsPath()
	for _, candidate := range classify.AutoModelPreference {
		if classify.FindModel(candidate).Installed(modelsPath) {
			return candidate
		}
	}

	return classify.ModelNone
}

// applyLabelModel applies LABELS_MODEL to the local labels entry in vision.Config.
func (c *Config) applyLabelModel() {
	if c == nil || vision.Config == nil {
		return
	}
	current := configuredVisionModel(vision.Config, vision.ModelTypeLabels)
	if current == nil {
		current = vision.NewLabelModel(classify.DefaultModelName())
		vision.Config.SetModel(current)
	}
	for _, model := range vision.Config.Models {
		if model != nil && model.Type == vision.ModelTypeLabels {
			model.DisabledByMode = c.LabelModelSetting() == classify.ModelNone
		}
	}
	if current.DisabledByMode || !current.Default {
		return
	}
	if selected := c.installedLabelModel(); selected != classify.ModelNone {
		registered := vision.NewLabelModel(selected)
		registered.Default, registered.Run, registered.Disabled = true, current.Run, current.Disabled
		vision.Config.SetModel(registered)
	}
}

// visionModelRemote reports whether a model is served by an endpoint, so it has no local artifact.
func visionModelRemote(model *vision.Model) bool {
	if model == nil {
		return false
	}

	uri, _ := model.Endpoint()
	return uri != "" || model.Service.UriUnresolved()
}

// visionModelOf returns the enabled model of a type, or nil when there is none.
func visionModelOf(modelType vision.ModelType) *vision.Model {
	if vision.Config == nil {
		return nil
	}

	return vision.Config.Model(modelType)
}

// labelsModelReturnsNSFW reports whether the enabled labels model can return NSFW fields, which
// only Ollama and OpenAI services do.
func labelsModelReturnsNSFW() bool {
	model := visionModelOf(vision.ModelTypeLabels)

	if model == nil {
		return false
	} else if uri, method := model.Endpoint(); uri == "" || method == "" {
		return false
	}

	switch model.EndpointRequestFormat() {
	case vision.ApiFormatOpenAI, vision.ApiFormatOllama:
		return true
	default:
		return false
	}
}

// configuredVisionModel returns the latest configured model of a type, including disabled models.
func configuredVisionModel(config *vision.ConfigValues, modelType vision.ModelType) *vision.Model {
	if config == nil {
		return nil
	}

	for i := len(config.Models) - 1; i >= 0; i-- {
		model := config.Models[i]
		if model != nil && model.Type == modelType {
			return model
		}
	}

	return nil
}

// reportIgnoredFaceRun reports a face schedule left in "vision.yml", which no longer decides
// anything. Two ways to set one thing raise a question nobody can answer from the outside -
// which wins, and where to change it - so the file is read and ignored rather than obeyed.
func (c *Config) reportIgnoredFaceRun(visionYaml string) {
	m := vision.Config.Model(vision.ModelTypeFace)

	if m == nil || vision.ParseRunType(m.Run) == vision.RunAuto {
		return
	}

	// Warned rather than noted: this used to be the documented way to turn face detection off,
	// so an operator who set "never" has it running again after an upgrade.
	c.warnFaceConfig("face-run-ignored", "config: face run type %s in %s is ignored, set FACE_RUN instead",
		clean.Log(m.Run), clean.Log(visionYaml))
}

// VisionSchedule returns the cron schedule configured for the vision worker, or "" if disabled.
func (c *Config) VisionSchedule() string {
	if c == nil {
		return ""
	}

	return Schedule(c.options.VisionSchedule)
}

// VisionFilter returns the search filter to use for scheduled vision runs.
func (c *Config) VisionFilter() string {
	if c == nil {
		return ""
	}

	return strings.TrimSpace(c.options.VisionFilter)
}

// VisionModelShouldRun reports whether the configured vision model of the
// specified type should execute in a given scheduling context. Face detection
// delegates to FaceEngineShouldRun so detection and embedding stay aligned.
func (c *Config) VisionModelShouldRun(t vision.ModelType, when vision.RunType) bool {
	return c.VisionModelSkipReason(t, when) == ""
}

// VisionModelSkipReason returns why the vision model of the specified type does not run in a given
// scheduling context, or "" when it runs, so a caller can report what keeps a requested model idle.
func (c *Config) VisionModelSkipReason(t vision.ModelType, when vision.RunType) string {
	switch {
	case c == nil:
		return "the configuration is missing"
	case t == vision.ModelTypeFace && c.DisableFaces():
		return "faces are disabled"
	case t == vision.ModelTypeLabels && c.DisableClassification():
		return "image classification is disabled"
	case t == vision.ModelTypeNsfw && !c.DetectNSFW():
		return "detect-nsfw is off"
	case t == vision.ModelTypeNsfw && c.NSFWModelSetting() != nsfw.ModelAuto:
		return fmt.Sprintf("nsfw-model is %s", c.NSFWModelSetting())
	case vision.Config == nil:
		return "the vision configuration is missing"
	case t == vision.ModelTypeFace:
		if c.FaceEngineShouldRun(when) {
			return ""
		} else if c.FaceEngine() == face.EngineNone {
			return "no face detector is in force"
		} else if reason := faceEmbeddingsUnavailable(); reason != "" {
			return reason
		}

		return fmt.Sprintf("face-run is %s", vision.ReportRunType(c.FaceEngineRunType()))
	case vision.Config.ShouldRun(t, when):
		return ""
	}

	if model := vision.Config.Model(t); model != nil {
		return fmt.Sprintf("its run type is %s", vision.ReportRunType(model.RunType()))
	}

	return fmt.Sprintf("no enabled %s model is configured", clean.Log(t))
}

// VisionApi checks whether the Computer Vision API endpoints should be enabled.
func (c *Config) VisionApi() bool {
	if c == nil {
		return false
	}

	return c.options.VisionApi && !c.options.Demo
}

// VisionUri returns the remote computer vision service URI, e.g. https://example.com/api/v1/vision.
func (c *Config) VisionUri() string {
	if c == nil {
		return ""
	}

	return clean.Uri(c.options.VisionUri)
}

// VisionKey returns the remote computer vision service access token, if any.
func (c *Config) VisionKey() string {
	if c == nil {
		return ""
	}

	// Try to read access token from file if c.options.VisionKey is not set.
	if c.options.VisionKey != "" {
		return clean.Password(c.options.VisionKey)
	} else if fileName := FlagFilePath("VISION_KEY"); fileName == "" {
		// No access token set, this is not an error.
		return ""
	} else if b, err := os.ReadFile(fileName); err != nil { //nolint:gosec // path derived from config directory
		event.SystemWarn([]string{"config", "vision key", "read %s", "%s"}, clean.Log(fileName), clean.ErrorFull(err))
		return ""
	} else if len(b) == 0 {
		// FlagFilePath resolves a name only while the file is not empty, so this reports a
		// file truncated between that check and the read.
		event.SystemWarn([]string{"config", "vision key", "read %s", "file is empty"}, clean.Log(fileName))
		return ""
	} else {
		return clean.Password(string(b))
	}
}

// ModelsPath returns the path where the machine learning models are located.
func (c *Config) ModelsPath() string {
	if c == nil {
		return ""
	}

	if c.options.ModelsPath != "" {
		return fs.Abs(c.options.ModelsPath)
	}

	if dir := filepath.Join(c.AssetsPath(), fs.ModelsDir); fs.PathExists(dir) {
		c.options.ModelsPath = dir
		return c.options.ModelsPath
	}

	c.options.ModelsPath = fs.FindDir(fs.ModelsPaths)

	return c.options.ModelsPath
}

// LabelModelPath returns the selected ONNX classifier path.
func (c *Config) LabelModelPath() string {
	if c == nil {
		return ""
	}

	name := c.EffectiveLabelModel()
	if name == classify.ModelNone || visionModelRemote(visionModelOf(vision.ModelTypeLabels)) {
		return ""
	}

	if model := classify.FindModel(name); model != nil {
		return model.ONNX.FilePath(filepath.Join(c.ModelsPath(), string(model.Name)))
	}

	if vision.Config == nil {
		return ""
	}

	model := vision.Config.Model(vision.ModelTypeLabels)
	if model == nil {
		return ""
	}

	path := model.Path
	if path == "" {
		path = clean.TypeLowerUnderscore(model.Name)
	}
	path = filepath.Join(c.ModelsPath(), clean.Path(path))
	if strings.EqualFold(filepath.Ext(path), ".onnx") {
		return path
	}

	fileName := filepath.Base(path) + ".onnx"
	if model.ONNX != nil && model.ONNX.File != "" {
		fileName = model.ONNX.File
	}

	return filepath.Join(path, fileName)
}

// LabelModelRuntime returns the engine used by the configured labels model.
func (c *Config) LabelModelRuntime() string {
	if c == nil || c.EffectiveLabelModel() == classify.ModelNone {
		return "none"
	}

	if vision.Config != nil {
		if model := vision.Config.Model(vision.ModelTypeLabels); model != nil {
			if runtime := model.EngineName(); runtime != vision.EngineLocal {
				return runtime
			}

			return vision.EngineONNX
		}
	}

	return vision.EngineONNX
}

// NsfwModelPath returns the selected ONNX detector path.
func (c *Config) NsfwModelPath() string {
	if c == nil {
		return ""
	}

	name := c.EffectiveNSFWModel()
	if name == nsfw.ModelNone || visionModelRemote(visionModelOf(vision.ModelTypeNsfw)) {
		return ""
	}
	if model := nsfw.FindModel(name); model != nil {
		return model.ONNX.FilePath(filepath.Join(c.ModelsPath(), string(model.Name)))
	}
	if vision.Config == nil {
		return ""
	}
	model := vision.Config.Model(vision.ModelTypeNsfw)
	if model == nil {
		return ""
	}
	path := model.Path
	if path == "" {
		path = clean.TypeLowerUnderscore(model.Name)
	}
	path = filepath.Join(c.ModelsPath(), clean.Path(path))
	if strings.EqualFold(filepath.Ext(path), ".onnx") {
		return path
	}
	fileName := filepath.Base(path) + ".onnx"
	if model.ONNX != nil && model.ONNX.File != "" {
		fileName = model.ONNX.File
	}
	return filepath.Join(path, fileName)
}

// NsfwModelRuntime returns the engine used by the configured NSFW model.
func (c *Config) NsfwModelRuntime() string {
	if c == nil || c.EffectiveNSFWModel() == nsfw.ModelNone {
		return "none"
	}
	if vision.Config != nil {
		if model := vision.Config.Model(vision.ModelTypeNsfw); model != nil {
			if runtime := model.EngineName(); runtime != vision.EngineLocal {
				return runtime
			}
			return vision.EngineONNX
		}
	}
	return vision.EngineONNX
}

// OnnxProvider returns the execution provider that ONNX inference sessions should use.
//
// An unrecognized value resolves to the default rather than stopping inference, and is reported
// once because the getter is called per loaded model and from the config report.
func (c *Config) OnnxProvider() onnx.Provider {
	if c == nil {
		return onnx.DefaultProvider
	}

	provider, ok := onnx.ParseProvider(c.options.OnnxProvider)

	if !ok {
		c.warnVisionConfig("onnx-provider", "config: unsupported onnx provider %s, using %s",
			clean.Log(c.options.OnnxProvider), provider)
	}

	return provider
}

// warnVisionConfig reports a computer-vision configuration problem once, because the getters
// are called from Propagate and from the config report rather than a single time per start.
func (c *Config) warnVisionConfig(key, format string, args ...any) {
	if _, warned := c.warnedOnce.LoadOrStore(key, true); !warned {
		log.Warnf(format, args...)
	}
}

// DetectNSFW checks if NSFW photos should be detected and flagged.
func (c *Config) DetectNSFW() bool {
	if c == nil {
		return false
	}

	return c.options.DetectNSFW
}

// visionKeyWarnings returns a warning for each character problem of the Vision API key. Incoming requests
// compare the key as written; outgoing requests send it with environment variables expanded.
func visionKeyWarnings(key string, incoming, outgoing bool) (warnings []string) {
	if key == "" || !incoming && !outgoing {
		return nil
	}

	stripped := incoming && header.ID(key) != key

	if sent := strings.TrimSpace(os.ExpandEnv(key)); !stripped && outgoing && header.ID(sent) != sent {
		stripped = true
	}

	if stripped {
		warnings = append(warnings, "vision key contains characters that are removed from access tokens, so it cannot authenticate with a PhotoPrism Vision API")
	}

	if incoming && strings.Contains(key, "$") {
		warnings = append(warnings, "vision key contains $ and is compared as written, without expanding environment variables")
	}

	return warnings
}

// warnVisionKey writes the warnings visionKeyWarnings returns for the configured Vision API key to
// the system log once, as they describe a secret.
func (c *Config) warnVisionKey() {
	for _, w := range visionKeyWarnings(c.VisionKey(), c.VisionApi(), c.VisionUri() != "") {
		if _, warned := c.warnedOnce.LoadOrStore("vision-key: "+w, true); !warned {
			event.SystemWarn([]string{"config", "%s"}, w)
		}
	}
}

// DetectNSFWLabels reports whether label responses may flag offensive content.
func (c *Config) DetectNSFWLabels() bool {
	return c != nil && c.DetectNSFW() && c.NSFWModelSetting() == nsfw.ModelName("labels")
}

// reportVisionModes reports invalid modes and unavailable label detection at startup.
func (c *Config) reportVisionModes() {
	for _, mode := range []struct {
		name, value string
		labels      bool
	}{
		{"labels-model", c.options.LabelsModel, false}, {"nsfw-model", c.options.NsfwModel, true},
	} {
		value := strings.ToLower(strings.TrimSpace(mode.value))
		if value != "" && value != "auto" && value != "none" && (!mode.labels || value != "labels") {
			resolved := string(c.NSFWModelSetting())
			if !mode.labels {
				resolved = string(c.LabelModelSetting())
			}
			event.SystemWarn([]string{"config", "unsupported %s mode %s, using %s; choose the model in vision.yml"}, mode.name, clean.Log(mode.value), resolved)
		}
	}
	if model := vision.Config.Model(vision.ModelTypeLabels); model != nil {
		if uri, _ := model.Endpoint(); uri == "" {
			if description := classify.FindModel(classify.ModelName(model.Name)); description != nil && !description.Installed(c.ModelsPath()) {
				event.SystemWarn([]string{"config", "label model %s is not installed; run download-models.sh %s and restart PhotoPrism"}, model.Name, model.Name)
			}
		}
	}
	if !c.DetectNSFWLabels() || labelsModelReturnsNSFW() {
		return
	}
	event.SystemWarn([]string{"config", "no nsfw detection takes place in labels mode because the labels model cannot return nsfw fields"})
}
