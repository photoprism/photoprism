package entity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	if err = Db().Model(m).Association("Photos").Find(&m.Photos); err != nil {
		log.Errorf("album: %s (yaml)", err)
		return out, err
	}

	return yaml.Marshal(m)
}

// SaveAsYaml writes the album metadata to a YAML backup file with the specified filename.
func (m *Album) SaveAsYaml(fileName string) error {
	switch {
	case m == nil:
		return fmt.Errorf("album entity is nil - you may have found a bug")
	case m.AlbumUID == "":
		return fmt.Errorf("album uid is empty")
	case fileName == "":
		return fmt.Errorf("yaml filename is empty")
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
	return fs.WriteFile(fileName, data, fs.ModeFile)
}

// YamlFileName returns the absolute file path for the YAML backup file.
func (m *Album) YamlFileName(backupPath string) (absolute, relative string, err error) {
	if m == nil {
		return "", "", fmt.Errorf("album entity is nil - you may have found a bug")
	} else if m.AlbumUID == "" {
		return "", "", fmt.Errorf("album uid is empty")
	}

	relative = filepath.Join(m.AlbumType, m.AlbumUID+fs.ExtYml)

	if backupPath == "" {
		return "", relative, fmt.Errorf("backup path is empty")
	}

	absolute = filepath.Join(backupPath, relative)

	return absolute, relative, err
}

// SaveBackupYaml writes the album metadata to a YAML backup file based on the specified storage paths.
func (m *Album) SaveBackupYaml(backupPath string) error {
	switch {
	case m == nil:
		return fmt.Errorf("album entity is nil - you may have found a bug")
	case m.AlbumUID == "":
		return fmt.Errorf("album uid is empty")
	case backupPath == "":
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

	filePath := filepath.Clean(fileName)
	if _, err := fs.StatFile(filePath); err != nil {
		return err
	}

	//nolint:gosec // G304: Path is normalized and validated above with fs.StatFile.
	data, err := os.ReadFile(filePath)

	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(data, m); err != nil {
		if strings.Contains(err.Error(), "gorm.DeletedAt") && strings.Count(err.Error(), "\n") == 1 {
			// try and fix the gorm.DeletedAt structure change
			deletedAt := JustDeletedAt{}
			if err = yaml.Unmarshal(data, &deletedAt); err != nil {
				log.Errorf("album: yaml: unable to reparse DeletedAt with %s", err.Error())
				return err
			} else {
				m.DeletedAt.Time = deletedAt.DeletedAt
				m.DeletedAt.Valid = true
			}

		} else {
			return err
		}
	}

	// Clip the restored path to the album_path column's byte budget so a backup
	// with a long multi-byte path cannot overflow the column on save.
	m.AlbumPath = ClipPath(m.AlbumPath)

	return nil
}
