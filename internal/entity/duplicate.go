package entity

import (
	"fmt"

	"github.com/photoprism/photoprism/pkg/clean"
)

type Duplicates []Duplicate
type DuplicatesMap map[string]Duplicate

// Duplicate represents an exact file duplicate.
type Duplicate struct {
	FileName string `gorm:"type:VARBINARY(755);primary_key;" json:"Name" yaml:"Name"`
	FileRoot string `gorm:"type:VARBINARY(16);primary_key;default:'/';" json:"Root" yaml:"Root,omitempty"`
	FileHash string `gorm:"type:VARBINARY(128);default:'';index" json:"Hash" yaml:"Hash,omitempty"`
	FileSize int64  `json:"Size" yaml:"Size,omitempty"`
	ModTime  int64  `json:"ModTime" yaml:"-"`
}

// TableName returns the entity table name.
func (Duplicate) TableName() string {
	return "duplicates"
}

// AddDuplicate adds a duplicate.
func AddDuplicate(fileName, fileRoot, fileHash string, fileSize, modTime int64) error {
	if fileName == "" {
		return fmt.Errorf("duplicate name must not be empty")
	} else if fileHash == "" {
		return fmt.Errorf("duplicate hash must not be empty")
	} else if modTime == 0 {
		return fmt.Errorf("duplicate mod time must not be empty")
	} else if fileRoot == "" {
		return fmt.Errorf("duplicate root must not be empty")
	}

	duplicate := &Duplicate{
		FileName: fileName,
		FileRoot: fileRoot,
		FileHash: fileHash,
		FileSize: fileSize,
		ModTime:  modTime,
	}

	if err := duplicate.Create(); err == nil {
		return nil
	} else if err := duplicate.Save(); err != nil {
		return err
	}

	return nil
}

// PurgeDuplicate deletes a duplicate.
func PurgeDuplicate(fileName, fileRoot string) error {
	if fileName == "" {
		return fmt.Errorf("duplicate name must not be empty")
	} else if fileRoot == "" {
		return fmt.Errorf("duplicate root must not be empty")
	}

	if err := UnscopedDb().Delete(Duplicate{}, "file_name = ? AND file_root = ?", fileName, fileRoot).Error; err != nil {
		log.Errorf("duplicate: %s in %s (purge)", err, clean.Log(fileName))
		return err
	}

	return nil
}

// Purge deletes a duplicate.
func (m *Duplicate) Purge() error {
	return PurgeDuplicate(m.FileName, m.FileRoot)
}

// Find returns a duplicate from the database.
func (m *Duplicate) Find() error {
	return UnscopedDb().First(m, "file_name = ? AND file_root = ?", m.FileName, m.FileRoot).Error
}

// Create inserts a new row to the database.
func (m *Duplicate) Create() error {
	if m.FileName == "" {
		return fmt.Errorf("duplicate name must not be empty (create)")
	} else if m.FileHash == "" {
		return fmt.Errorf("duplicate hash must not be empty (create)")
	} else if m.ModTime == 0 {
		return fmt.Errorf("duplicate mod time must not be empty (create)")
	} else if m.FileRoot == "" {
		return fmt.Errorf("duplicate root must not be empty (create)")
	}

	return UnscopedDb().Create(m).Error
}

// Saves the duplicates in the database.
func (m *Duplicate) Save() error {
	if m.FileName == "" {
		return fmt.Errorf("duplicate name must not be empty (save)")
	} else if m.FileHash == "" {
		return fmt.Errorf("duplicate hash must not be empty (save)")
	} else if m.ModTime == 0 {
		return fmt.Errorf("duplicate mod time must not be empty (save)")
	} else if m.FileRoot == "" {
		return fmt.Errorf("duplicate root must not be empty (save)")
	}

	if err := UnscopedDb().Save(m).Error; err != nil {
		log.Errorf("duplicate: %s in %s (save)", err, clean.Log(m.FileName))
		return err
	}

	return nil
}
