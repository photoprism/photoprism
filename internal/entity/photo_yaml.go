package entity

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

var photoYamlMutex = sync.Mutex{}

// Yaml returns photo data as YAML string.
func (m *Photo) Yaml() ([]byte, error) {
	// Load details if not done yet.
	m.GetDetails()

	m.CreatedAt = m.CreatedAt.UTC().Truncate(time.Second)
	m.UpdatedAt = m.UpdatedAt.UTC().Truncate(time.Second)

	out, err := yaml.Marshal(m)

	if err != nil {
		return []byte{}, err
	}

	return out, err
}

// SaveAsYaml writes the photo metadata to a YAML sidecar file with the specified filename.
func (m *Photo) SaveAsYaml(fileName string) error {
	if m == nil {
		return fmt.Errorf("photo entity is nil - you may have found a bug")
	} else if fileName == "" {
		return fmt.Errorf("yaml filename is empty")
	} else if m.PhotoUID == "" {
		return fmt.Errorf("photo uid is empty")
	}

	data, err := m.Yaml()

	if err != nil {
		return err
	}

	// Make sure directory exists.
	if err = fs.MkdirAll(filepath.Dir(fileName)); err != nil {
		return err
	}

	photoYamlMutex.Lock()
	defer photoYamlMutex.Unlock()

	// Write YAML data to file.
	if err = fs.WriteFile(fileName, data); err != nil {
		return err
	}

	return nil
}

// YamlFileName returns both the absolute file path and the relative name for the YAML sidecar file, e.g. for logging.
func (m *Photo) YamlFileName(originalsPath, sidecarPath string) (absolute, relative string, err error) {
	absolute, err = fs.FileName(filepath.Join(originalsPath, m.PhotoPath, m.PhotoName), sidecarPath, originalsPath, fs.ExtYAML)
	relative = filepath.Join(m.PhotoPath, m.PhotoName) + fs.ExtYAML

	return absolute, relative, err
}

// SaveSidecarYaml writes the photo metadata to a YAML sidecar file based on the specified storage paths.
func (m *Photo) SaveSidecarYaml(originalsPath, sidecarPath string) error {
	if m == nil {
		return fmt.Errorf("photo entity is nil - you may have found a bug")
	} else if m.PhotoName == "" {
		return fmt.Errorf("photo name is empty")
	} else if m.PhotoUID == "" {
		return fmt.Errorf("photo uid is empty")
	}

	// Write metadata to YAML sidecar file.
	if fileName, relName, err := m.YamlFileName(originalsPath, sidecarPath); err != nil {
		log.Warnf("photo: %s (save %s)", err, clean.Log(relName))
		return err
	} else if err = m.SaveAsYaml(fileName); err != nil {
		log.Warnf("photo: %s (save %s)", err, clean.Log(relName))
		return err
	} else {
		log.Infof("photo: saved sidecar file %s", clean.Log(relName))
	}

	return nil
}

// LoadFromYaml restores the photo metadata from a YAML sidecar file.
func (m *Photo) LoadFromYaml(fileName string) error {
	if m == nil {
		return fmt.Errorf("photo entity is nil - you may have found a bug")
	} else if fileName == "" {
		return fmt.Errorf("yaml filename is empty")
	}

	data, err := os.ReadFile(fileName)

	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(data, m); err != nil {
		return err
	}

	return nil
}
