package entity

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/pkg/fs"
)

var albumYamlMutex = sync.Mutex{}

// Yaml returns album data as YAML string.
func (m *Album) Yaml() (out []byte, err error) {
	m.CreatedAt = m.CreatedAt.UTC().Truncate(time.Second)
	m.UpdatedAt = m.UpdatedAt.UTC().Truncate(time.Second)

	if err := Db().Model(m).Association("Photos").Find(&m.Photos).Error; err != nil {
		log.Errorf("album: %s (yaml)", err)
		return out, err
	}

	return yaml.Marshal(m)
}

// SaveAsYaml saves album data as YAML file.
func (m *Album) SaveAsYaml(fileName string) error {
	data, err := m.Yaml()

	if err != nil {
		return err
	}

	// Make sure directory exists.
	if err = fs.MkdirAll(filepath.Dir(fileName)); err != nil {
		return err
	}

	albumYamlMutex.Lock()
	defer albumYamlMutex.Unlock()

	// Write YAML data to file.
	if err = fs.WriteFile(fileName, data); err != nil {
		return err
	}

	return nil
}

// LoadFromYaml photo data from a YAML file.
func (m *Album) LoadFromYaml(fileName string) error {
	data, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, m); err != nil {
		return err
	}

	return nil
}

// YamlFileName returns the YAML file name.
func (m *Album) YamlFileName(albumsPath string) string {
	return filepath.Join(albumsPath, m.AlbumType, m.AlbumUID+fs.ExtYAML)
}
