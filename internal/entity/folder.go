package entity

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/ulule/deepcopier"
)

// Folder represents a file system directory.
type Folder struct {
	Path              string     `gorm:"type:varbinary(255);unique_index:idx_folders_path_root;" json:"Path" yaml:"Path"`
	Root              string     `gorm:"type:varbinary(16);default:'originals';unique_index:idx_folders_path_root;" json:"Root" yaml:"Root"`
	FolderUID         string     `gorm:"type:varbinary(36);primary_key;" json:"UID,omitempty" yaml:"UID,omitempty"`
	FolderType        string     `gorm:"type:varbinary(16);" json:"Type" yaml:"Type,omitempty"`
	FolderTitle       string     `gorm:"type:varchar(255);" json:"Title" yaml:"Title,omitempty"`
	FolderCategory    string     `gorm:"type:varchar(255);index;" json:"Category" yaml:"Category,omitempty"`
	FolderDescription string     `gorm:"type:text;" json:"Description,omitempty" yaml:"Description,omitempty"`
	FolderOrder       string     `gorm:"type:varbinary(32);" json:"Order" yaml:"Order,omitempty"`
	FolderCountry     string     `gorm:"type:varbinary(2);index:idx_folders_country_year_month;default:'zz'" json:"Country" yaml:"Country,omitempty"`
	FolderYear        int        `gorm:"index:idx_folders_country_year_month;" json:"Year" yaml:"Year,omitempty"`
	FolderMonth       int        `gorm:"index:idx_folders_country_year_month;" json:"Month" yaml:"Month,omitempty"`
	FolderFavorite    bool       `json:"Favorite" yaml:"Favorite,omitempty"`
	FolderPrivate     bool       `json:"Private" yaml:"Private,omitempty"`
	FolderIgnore      bool       `json:"Ignore" yaml:"Ignore,omitempty"`
	FolderWatch       bool       `json:"Watch" yaml:"Watch,omitempty"`
	FileCount         int        `gorm:"-" json:"FileCount" yaml:"-"`
	Links             []Link     `gorm:"foreignkey:share_uid;association_foreignkey:folder_uid" json:"Links" yaml:"-"`
	CreatedAt         time.Time  `json:"-" yaml:"-"`
	UpdatedAt         time.Time  `json:"-" yaml:"-"`
	ModifiedAt        *time.Time `json:"ModifiedAt,omitempty" yaml:"-"`
	DeletedAt         *time.Time `sql:"index" json:"-"`
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Folder) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsPPID(m.FolderUID, 'd') {
		return nil
	}

	return scope.SetColumn("FolderUID", rnd.PPID('d'))
}

// NewFolder creates a new file system directory entity.
func NewFolder(root, pathName string, modTime *time.Time) Folder {
	now := time.Now().UTC()

	pathName = strings.Trim(pathName, string(os.PathSeparator))

	if pathName == RootPath {
		pathName = ""
	}

	result := Folder{
		FolderUID:   rnd.PPID('d'),
		Root:        root,
		Path:        pathName,
		FolderType:  TypeDefault,
		FolderOrder: SortOrderName,
		ModifiedAt:  modTime,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result.SetTitleFromPath()

	return result
}

// SetTitleFromPath updates the title based on the path name (e.g. when displaying it as an album).
func (m *Folder) SetTitleFromPath() {
	s := m.Path
	s = strings.TrimSpace(s)

	if s == "" || s == RootPath {
		s = m.Root
	} else {
		s = path.Base(s)
	}

	if len(m.Path) >= 6 && txt.IsUInt(s) {
		if date := txt.Time(m.Path); !date.IsZero() {
			if date.Day() > 1 {
				m.FolderTitle = date.Format("January 2, 2006")
			} else {
				m.FolderTitle = date.Format("January 2006")
			}
			return
		}
	}

	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.Title(s)

	m.FolderTitle = txt.Clip(s, txt.ClipDefault)
}

// Saves the complete entity in the database.
func (m *Folder) Create() error {
	if err := Db().Create(m).Error; err != nil {
		return err
	}

	event.Publish("count.folders", event.Data{
		"count": 1,
	})

	return nil
}

// FirstOrCreateFolder returns the first matching record or creates a new one with the given conditions.
func FirstOrCreateFolder(m *Folder) *Folder {
	result := Folder{}

	if err := Db().Where("path = ? AND root = ?", m.Path, m.Root).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Errorf("folder: %s", err)
		return nil
	}

	return m
}

// Updates selected properties in the database.
func (m *Folder) Updates(values interface{}) error {
	return Db().Model(m).Updates(values).Error
}

// SetForm updates the entity properties based on form values.
func (m *Folder) SetForm(f form.Folder) error {
	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	return nil
}
