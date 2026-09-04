package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
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
			log.Warnf("vision: %s", err)
		}

		c.reportIgnoredFaceRun(visionYaml)
	}

	c.applyLabelModel()
}

// LabelModelSetting returns the configured label model without resolving auto.
func (c *Config) LabelModelSetting() classify.ModelName {
	if c == nil {
		return classify.ModelNone
	}

	return classify.ParseModelName(c.options.LabelModel)
}

// EffectiveLabelModel returns the local classifier selected for this instance.
func (c *Config) EffectiveLabelModel() classify.ModelName {
	setting := c.LabelModelSetting()
	if setting != classify.ModelAuto {
		return setting
	}

	if vision.Config != nil {
		if model := vision.Config.Model(vision.ModelTypeLabels); model != nil && !model.Default {
			return classify.NormalizeModelName(classify.ModelName(model.Name))
		}
	}

	return classify.DefaultModelName()
}

// applyLabelModel applies LABEL_MODEL to the local labels entry in vision.Config.
func (c *Config) applyLabelModel() {
	if c == nil || vision.Config == nil {
		return
	}

	current := vision.Config.Model(vision.ModelTypeLabels)
	setting := c.LabelModelSetting()

	if setting == classify.ModelNone {
		if current == nil {
			current = vision.NewLabelModel(classify.DefaultModelName())
		} else {
			current = current.Clone()
		}
		if current != nil {
			current.Disabled = true
			vision.Config.SetModel(current)
		}
		return
	}

	if setting == classify.ModelAuto && current != nil && !current.Default {
		return
	}

	selected := setting
	if selected == classify.ModelAuto {
		selected = classify.DefaultModelName()
	}

	if registered := vision.NewLabelModel(selected); registered != nil {
		if current != nil {
			registered.Run = current.Run
		}
		vision.Config.SetModel(registered)
		return
	}

	if current != nil && classify.NormalizeModelName(classify.ModelName(current.Name)) == selected {
		return
	}

	vision.Config.SetModel(&vision.Model{
		Type: vision.ModelTypeLabels,
		Name: string(selected),
		Path: string(selected),
	})
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
	if c == nil {
		return false
	}

	if t == vision.ModelTypeFace && c.DisableFaces() {
		return false
	}

	if t == vision.ModelTypeLabels && c.DisableClassification() {
		return false
	}

	if t == vision.ModelTypeNsfw && !c.DetectNSFW() {
		return false
	}

	if vision.Config == nil {
		return false
	}

	if t == vision.ModelTypeFace {
		return c.FaceEngineShouldRun(when)
	}

	return vision.Config.ShouldRun(t, when)
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
	} else if b, err := os.ReadFile(fileName); err != nil || len(b) == 0 { //nolint:gosec // path derived from config directory
		log.Warnf("config: failed to read vision key from %s (%s)", fileName, err)
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

// NasnetModelPath returns the legacy NASNet model path.
func (c *Config) NasnetModelPath() string {
	if c == nil {
		return ""
	}

	return filepath.Join(c.ModelsPath(), "nasnet")
}

// LabelModelPath returns the selected ONNX classifier path.
func (c *Config) LabelModelPath() string {
	if c == nil {
		return ""
	}

	name := c.EffectiveLabelModel()
	if name == classify.ModelNone {
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

// FacenetModelPath returns the FaceNet model path.
func (c *Config) FacenetModelPath() string {
	if c == nil {
		return ""
	}

	return filepath.Join(c.ModelsPath(), "facenet")
}

// NsfwModelPath returns the "not safe for work" TensorFlow model path.
func (c *Config) NsfwModelPath() string {
	if c == nil {
		return ""
	}

	return filepath.Join(c.ModelsPath(), "nsfw")
}

// DetectNSFW checks if NSFW photos should be detected and flagged.
func (c *Config) DetectNSFW() bool {
	if c == nil {
		return false
	}

	return c.options.DetectNSFW
}
