package vision

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/scheme"
)

var (
	// CachePath stores the directory used for caching downloaded vision models.
	CachePath = ""
	// ModelsPath stores the directory containing downloaded vision models.
	ModelsPath = ""
	// DownloadUrl overrides the default model download endpoint when set.
	DownloadUrl = ""
	// ServiceApi enables exposing vision APIs via the service layer when true.
	ServiceApi = false
	// ServiceUri sets the base URI for the vision service when exposed externally.
	ServiceUri = ""
	// ServiceKey provides an optional API key for the vision service.
	ServiceKey = ""
	// ServiceTimeout sets the maximum duration for service API requests.
	ServiceTimeout = 10 * time.Minute
	// ServiceMethod defines the HTTP verb used when calling the vision service.
	ServiceMethod = http.MethodPost
	// ServiceFileScheme specifies how local files are encoded when sent to the service.
	ServiceFileScheme = scheme.Data
	// ServiceRequestFormat sets the default payload format for service requests.
	ServiceRequestFormat = ApiFormatVision
	// ServiceResponseFormat sets the expected response format from the service.
	ServiceResponseFormat = ApiFormatVision
	// DefaultResolution specifies the default square resize dimension for model inputs.
	DefaultResolution = 224
	// DefaultTemperature sets the sampling temperature for compatible models.
	DefaultTemperature = 0.1
	// MaxTemperature clamps user-supplied temperatures to a safe upper bound.
	MaxTemperature = 2.0
	// DefaultSrc defines the fallback source string for generated labels.
	DefaultSrc = entity.SrcImage
	// DetectNSFWLabels toggles NSFW label detection in vision responses.
	DetectNSFWLabels = false
)

// Config reference the current configuration options.
var Config = NewConfig()

// ConfigValues represents computer vision configuration values for the supported Model types.
type ConfigValues struct {
	Models     Models     `yaml:"Models,omitempty" json:"models,omitempty"`
	Thresholds Thresholds `yaml:"Thresholds,omitempty" json:"thresholds"`
}

// NewConfig returns a new computer vision config with defaults.
func NewConfig() *ConfigValues {
	cfg := &ConfigValues{
		Models:     DefaultModels,
		Thresholds: DefaultThresholds,
	}

	for _, model := range cfg.Models {
		model.ApplyEngineDefaults()
	}

	return cfg
}

// Load user settings from file.
func (c *ConfigValues) Load(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("missing config filename")
	} else if !fs.FileExists(fileName) {
		return fmt.Errorf("%s not found", clean.Log(fileName))
	}

	yamlConfig, err := os.ReadFile(fileName) // #nosec G304 fileName is from validated config path

	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(yamlConfig, c); err != nil {
		return err
	}

	// Replace default placeholders with canonical defaults while respecting
	// explicit Run / Disabled overrides.
	c.applyDefaultModels()

	// Add missing default models so users are not required to list them in
	// vision.yml. Custom models continue to override defaults when present.
	c.ensureDefaultModels()

	for _, model := range c.Models {
		model.ApplyEngineDefaults()
	}

	if c.Thresholds.Confidence <= 0 || c.Thresholds.Confidence > 100 {
		c.Thresholds.Confidence = DefaultThresholds.Confidence
	}

	if c.Thresholds.Topicality <= 0 || c.Thresholds.Topicality > 100 {
		c.Thresholds.Topicality = DefaultThresholds.Topicality
	}

	if c.Thresholds.NSFW <= 0 || c.Thresholds.NSFW > 100 {
		c.Thresholds.NSFW = DefaultThresholds.NSFW
	}

	return nil
}

// applyDefaultModels swaps entries marked as Default with the built-in
// models while keeping user-specified Run / Disabled overrides intact.
func (c *ConfigValues) applyDefaultModels() {
	for i, model := range c.Models {
		if !model.Default {
			continue
		}

		runType := model.Run
		disabled := model.Disabled

		switch model.Type {
		case ModelTypeLabels:
			c.Models[i] = NasnetModel.Clone()
		case ModelTypeNsfw:
			c.Models[i] = NsfwModel.Clone()
		case ModelTypeFace:
			c.Models[i] = FacenetModel.Clone()
		case ModelTypeCaption:
			c.Models[i] = CaptionModel.Clone()
		default:
			continue
		}

		if runType != RunAuto {
			c.Models[i].Run = runType
		}

		if disabled {
			c.Models[i].Disabled = disabled
		}
	}
}

// ensureDefaultModels appends built-in default models for any types
// that are completely missing from the configuration. Custom models (enabled
// or disabled) block the addition for their respective types so user intent is
// preserved.
func (c *ConfigValues) ensureDefaultModels() {
	for _, defaultModel := range DefaultModels {
		if defaultModel == nil {
			continue
		}

		if c.hasModelType(defaultModel.Type) {
			continue
		}

		c.Models = append(c.Models, defaultModel.Clone())
	}
}

// hasModelType reports whether any configured model (enabled or disabled)
// matches the provided type.
func (c *ConfigValues) hasModelType(t ModelType) bool {
	for _, model := range c.Models {
		if model == nil {
			continue
		}

		if model.Type == t {
			return true
		}
	}

	return false
}

// Save user settings to a file.
func (c *ConfigValues) Save(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("missing config filename")
	}

	data, err := yaml.Marshal(c)

	if err != nil {
		return err
	}

	return os.WriteFile(fileName, data, fs.ModeConfigFile)
}

// Model returns the first enabled model with the matching type.
// It returns nil if no matching model is available or every model of that
// type is disabled, allowing callers to chain nil-safe Model methods.
func (c *ConfigValues) Model(t ModelType) *Model {
	for i := len(c.Models) - 1; i >= 0; i-- {
		m := c.Models[i]
		if m.Type == t && !m.Disabled {
			return m
		}
	}

	return nil
}

// ShouldRun reports whether the configured model for the given type is
// allowed to run in the specified context. It returns false when no
// suitable model exists or when execution is explicitly disabled.
func (c *ConfigValues) ShouldRun(t ModelType, when RunType) bool {
	m := c.Model(t)

	if m == nil {
		return false
	} else if m.Disabled {
		return false
	}

	return m.ShouldRun(when)
}

// RunType returns the normalized run type for the first enabled model matching
// the provided type. Disabled or missing models fall back to RunNever so
// callers can treat the result as authoritative scheduling information.
func (c *ConfigValues) RunType(t ModelType) RunType {
	m := c.Model(t)

	if m == nil {
		return RunNever
	} else if m.Disabled {
		return RunNever
	}

	return m.RunType()
}

// IsDefault checks whether the specified type is the built-in default model.
func (c *ConfigValues) IsDefault(t ModelType) bool {
	m := c.Model(t)

	if m == nil {
		return false
	}

	return m.IsDefault()
}

// IsCustom checks whether the specified type uses a custom model or service.
func (c *ConfigValues) IsCustom(t ModelType) bool {
	m := c.Model(t)

	if m == nil {
		return false
	}

	return !m.IsDefault()
}

// SetCachePath updates the cache path.
func SetCachePath(dir string) {
	if dir = fs.Abs(dir); dir == "" {
		return
	}

	CachePath = dir
}

// GetCachePath returns the cache path.
func GetCachePath() string {
	if CachePath != "" {
		return CachePath
	}

	CachePath = fs.Abs("../../../storage/cache")

	return CachePath
}

// SetModelsPath updates the model assets path.
func SetModelsPath(dir string) {
	if dir = fs.Abs(dir); dir == "" {
		return
	}

	ModelsPath = dir
}

// GetModelsPath returns the model assets path, or an empty string if not configured or found.
func GetModelsPath() string {
	if ModelsPath != "" {
		return ModelsPath
	}

	assetsPath := fs.Abs("../../../assets")

	if dir := filepath.Join(assetsPath, "models"); fs.PathExists(dir) {
		ModelsPath = dir
	} else if fs.PathExists(assetsPath) {
		ModelsPath = assetsPath
	}

	return ModelsPath
}

// GetModelPath returns the absolute path of a named model file in CachePath.
func GetModelPath(name string) string {
	return filepath.Join(GetModelsPath(), clean.Path(clean.TypeLowerUnderscore(name)))
}

// GetNasnetModelPath returns the absolute path of the default Nasnet model.
func GetNasnetModelPath() string {
	return GetModelPath(NasnetModel.Name)
}

// GetFacenetModelPath returns the absolute path of the default Facenet model.
func GetFacenetModelPath() string {
	return GetModelPath(FacenetModel.Name)
}

// GetNsfwModelPath returns the absolute path of the default NSFW model.
func GetNsfwModelPath() string {
	return GetModelPath(NsfwModel.Name)
}
