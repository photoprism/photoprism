package vision

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/service/http/scheme"
)

var (
	CachePath             = ""
	ModelsPath            = ""
	DownloadUrl           = ""
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
	return &ConfigValues{
		Models:     DefaultModels,
		Thresholds: DefaultThresholds,
	}
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
	//    the "vision.yml" file, but set the defaults if the "Default" flag is set.
	// 2. Use the default "Thresholds" if no custom thresholds are configured.

	for i, model := range c.Models {
		if !model.Default {
			continue
		}

		switch model.Type {
		case ModelTypeLabels:
			c.Models[i] = NasnetModel
		case ModelTypeNsfw:
			c.Models[i] = NsfwModel
		case ModelTypeFace:
			c.Models[i] = FacenetModel
		case ModelTypeCaption:
			c.Models[i] = CaptionModel
		}
	}

	if c.Thresholds.Confidence <= 0 || c.Thresholds.Confidence > 100 {
		c.Thresholds.Confidence = DefaultThresholds.Confidence
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

// Model returns the first enabled model with the matching type from the configuration.
func (c *ConfigValues) Model(t ModelType) *Model {
	for _, m := range c.Models {
		if m.Type == t && !m.Disabled {
			return m
		}
	}

	return nil
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
