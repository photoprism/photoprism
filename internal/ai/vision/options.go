package vision

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Config reference the current configuration options.
var Config = NewOptions()

// Options represents a computer vision configuration for the supported Model types.
type Options struct {
	Models     Models     `yaml:"Models,omitempty" json:"models,omitempty"`
	Thresholds Thresholds `yaml:"Thresholds" json:"thresholds"`
}

// NewOptions returns a new computer vision config with defaults.
func NewOptions() *Options {
	return &Options{
		Models:     DefaultModels,
		Thresholds: DefaultThresholds,
	}
}

// Load user settings from file.
func (c *Options) Load(fileName string) error {
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

	return nil
}

// Save user settings to a file.
func (c *Options) Save(fileName string) error {
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
func (c *Options) Model(t ModelType) *Model {
	for _, m := range c.Models {
		if m.Type == t && !m.Disabled {
			return m
		}
	}

	return nil
}
