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
	CachePath             = ""
	ModelsPath            = ""
	DownloadUrl           = ""
	ServiceApi            = false
	ServiceUri            = ""
	ServiceKey            = ""
	ServiceTimeout        = 10 * time.Minute
	ServiceMethod         = http.MethodPost
	ServiceFileScheme     = scheme.Data
	ServiceRequestFormat  = ApiFormatVision
	ServiceResponseFormat = ApiFormatVision
	DefaultResolution     = 224
	DefaultTemperature    = 0.1
	MaxTemperature        = 2.0
	DefaultSrc            = entity.SrcImage
	DetectNSFWLabels      = false
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

	yamlConfig, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(yamlConfig, c); err != nil {
		return err
	}

	// 1. Ensure that there is at least one configuration for each model type,
	//    so that adding a copy of the default configuration to the vision.yml file
	//    is not required. We could alternatively require a model to included in
	//    the "vision.yml" file, but set the defaults if the "Default" flag is set
	//    while preserving explicit Run / Disabled overrides.
	// 2. Use the default "Thresholds" if no custom thresholds are configured.

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
		}

		if runType != RunAuto {
			c.Models[i].Run = runType
		}

		if disabled {
			c.Models[i].Disabled = disabled
		}
	}

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

// Save user settings to a file.
func (c *ConfigValues) Save(fileName string) error {
	if fileName == "" {
		return fmt.Errorf("missing config filename")
	}

	data, err := yaml.Marshal(c)

	if err != nil {
		return err
	}

	if err = os.WriteFile(fileName, data, fs.ModeConfigFile); err != nil {
		return err
	}

	return nil
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

func GetModelPath(name string) string {
	return filepath.Join(GetModelsPath(), clean.Path(clean.TypeLowerUnderscore(name)))
}

func GetNasnetModelPath() string {
	return GetModelPath(NasnetModel.Name)
}

func GetFacenetModelPath() string {
	return GetModelPath(FacenetModel.Name)
}

func GetNsfwModelPath() string {
	return GetModelPath(NsfwModel.Name)
}
