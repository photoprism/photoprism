package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/photoprism/photoprism/internal/util"
	"gopkg.in/yaml.v2"
)

type Settings struct {
	Theme    string `json:"theme" yaml:"theme" flag:"theme"`
	Language string `json:"language" yaml:"language" flag:"language"`
}

func NewSettings() *Settings {
	return &Settings{}
}

// SetValuesFromFile uses a yaml config file to initiate the configuration entity.
func (s *Settings) SetValuesFromFile(fileName string) error {
	if !util.Exists(fileName) {
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
	if !util.Exists(fileName) {
		return fmt.Errorf("settings file not found: \"%s\"", fileName)
	}

	data, err := yaml.Marshal(s)

	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, data, os.ModePerm)
}

