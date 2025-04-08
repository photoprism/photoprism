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
	Caption    Models     `yaml:"Caption,omitempty" json:"caption,omitempty"`
	Faces      Models     `yaml:"Faces,omitempty" json:"faces,omitempty"`
	Labels     Models     `yaml:"Labels,omitempty" json:"labels,omitempty"`
	Nsfw       Models     `yaml:"Nsfw,omitempty" json:"nsfw,omitempty"`
	Thresholds Thresholds `yaml:"Thresholds" json:"thresholds"`
}

// NewOptions returns a new computer vision config with defaults.
func NewOptions() *Options {
	return &Options{
		Caption:    Models{},
		Faces:      Models{},
		Labels:     Models{NasnetModel},
		Nsfw:       Models{},
		Thresholds: Thresholds{Confidence: 10},
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
