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

var albumYamlMutex = sync.Mutex{}

// Yaml returns album data as YAML string.
func (m *Album) Yaml() (out []byte, err error) {
	m.CreatedAt = m.CreatedAt.UTC().Truncate(time.Second)
	m.UpdatedAt = m.UpdatedAt.UTC().Truncate(time.Second)

	if err = Db().Model(m).Association("Photos").Find(&m.Photos).Error; err != nil {
		log.Errorf("album: %s (yaml)", err)
		return out, err
	}

	return yaml.Marshal(m)
}

// SaveAsYaml writes the album metadata to a YAML backup file with the specified filename.
func (m *Album) SaveAsYaml(fileName string) error {
	if m == nil {
		return fmt.Errorf("album entity is nil - you may have found a bug")
	} else if m.AlbumUID == "" {
		return fmt.Errorf("album uid is empty")
	} else if fileName == "" {
		return fmt.Errorf("yaml filname is empty")
	}

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

// YamlFileName returns the absolute file path for the YAML backup file.
func (m *Album) YamlFileName(backupPath string) (absolute, relative string, err error) {
	if m == nil {
		return "", "", fmt.Errorf("album entity is nil - you may have found a bug")
	} else if m.AlbumUID == "" {
		return "", "", fmt.Errorf("album uid is empty")
	}

	relative = filepath.Join(m.AlbumType, m.AlbumUID+fs.ExtYAML)

	if backupPath == "" {
		return "", relative, fmt.Errorf("backup path is empty")
	}

	absolute = filepath.Join(backupPath, relative)

	return absolute, relative, err
}

// SaveBackupYaml writes the album metadata to a YAML backup file based on the specified storage paths.
func (m *Album) SaveBackupYaml(backupPath string) error {
	if m == nil {
		return fmt.Errorf("album entity is nil - you may have found a bug")
	} else if m.AlbumUID == "" {
		return fmt.Errorf("album uid is empty")
	} else if backupPath == "" {
		return fmt.Errorf("backup path is empty")
	}

	// Get album YAML backup filename.
	fileName, relName, err := m.YamlFileName(backupPath)

	if err != nil {
		log.Warnf("album: %s (save %s)", err, clean.Log(relName))
		return err
	}

	var action string

	if fs.FileExists(fileName) {
		action = "update"
	} else {
		action = "create"
	}

	// Write album metadata to YAML backup file.
	if err = m.SaveAsYaml(fileName); err != nil {
		log.Warnf("album: %s (%s %s)", err, action, clean.Log(relName))
		return err
	} else {
		log.Debugf("album: %sd backup file %s", action, clean.Log(relName))
	}

	return nil
}

// LoadFromYaml restores the album metadata from a YAML backup file.
func (m *Album) LoadFromYaml(fileName string) error {
	if m == nil {
		return fmt.Errorf("album entity is nil - you may have found a bug")
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
