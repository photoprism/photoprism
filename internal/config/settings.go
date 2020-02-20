package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/photoprism/photoprism/pkg/fs"
	"gopkg.in/yaml.v2"
)

// Settings contains Web UI settings
type Settings struct {
	Theme    string `json:"theme" yaml:"theme" flag:"theme"`
	Language string `json:"language" yaml:"language" flag:"language"`
}

// NewSettings returns a empty Settings
func NewSettings() *Settings {
	return &Settings{}
}

// SetValuesFromFile uses a yaml config file to initiate the configuration entity.
func (s *Settings) SetValuesFromFile(fileName string) error {
	if !fs.FileExists(fileName) {
		return fmt.Errorf("settings file not found: \"%s\"", fileName)
	}

	yamlConfig, err := ioutil.ReadFile(fileName)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(yamlConfig, s)
}

// WriteValuesToFile uses a yaml config file to initiate the configuration entity.
func (s *Settings) WriteValuesToFile(fileName string) error {
	if !fs.FileExists(fileName) {
		return fmt.Errorf("settings file not found: \"%s\"", fileName)
	}

	data, err := yaml.Marshal(s)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, data, os.ModePerm)
}

// Settings returns the current user settings.
func (c *Config) Settings() *Settings {
	s := NewSettings()
	p := c.SettingsFile()

	if err := s.SetValuesFromFile(p); err != nil {
		log.Error(err)
	}

	return s
}
