package entity

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Yaml returns photo data as YAML string.
func (m *Photo) Yaml() ([]byte, error) {
	out, err := yaml.Marshal(m)

	if err != nil {
		return []byte{}, err
	}

	return out, err
}

// SaveAsYaml saves photo data as YAML file.
func (m *Photo) SaveAsYaml(fileName string) error {
	data, err := m.Yaml()

	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(fileName, data, os.ModePerm); err != nil {
		return err
	}

	return nil
}

// LoadFromYaml photo data from a YAML file.
func (m *Photo) LoadFromYaml(fileName string) error {
	data, err := ioutil.ReadFile(fileName)

	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, m); err != nil {
		return err
	}

	return nil
}
