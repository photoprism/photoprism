package entity

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/ulule/deepcopier"
)

type Folders []Folder

// Folder represents a file system directory.
type Folder struct {
	Path              string     `gorm:"type:varbinary(255);unique_index:idx_folders_path_root;" json:"Path" yaml:"Path"`
	Root              string     `gorm:"type:varbinary(16);default:'';unique_index:idx_folders_path_root;" json:"Root" yaml:"Root,omitempty"`
	FolderUID         string     `gorm:"type:varbinary(42);primary_key;" json:"UID,omitempty" yaml:"UID,omitempty"`
	FolderType        string     `gorm:"type:varbinary(16);" json:"Type" yaml:"Type,omitempty"`
	FolderTitle       string     `gorm:"type:varchar(255);" json:"Title" yaml:"Title,omitempty"`
	FolderCategory    string     `gorm:"type:varchar(255);index;" json:"Category" yaml:"Category,omitempty"`
	FolderDescription string     `gorm:"type:text;" json:"Description,omitempty" yaml:"Description,omitempty"`
	FolderOrder       string     `gorm:"type:varbinary(32);" json:"Order" yaml:"Order,omitempty"`
	FolderCountry     string     `gorm:"type:varbinary(2);index:idx_folders_country_year_month;default:'zz'" json:"Country" yaml:"Country,omitempty"`
	FolderYear        int        `gorm:"index:idx_folders_country_year_month;" json:"Year" yaml:"Year,omitempty"`
	FolderMonth       int        `gorm:"index:idx_folders_country_year_month;" json:"Month" yaml:"Month,omitempty"`
	FolderDay         int        `json:"Day" yaml:"Day,omitempty"`
	FolderFavorite    bool       `json:"Favorite" yaml:"Favorite,omitempty"`
	FolderPrivate     bool       `json:"Private" yaml:"Private,omitempty"`
	FolderIgnore      bool       `json:"Ignore" yaml:"Ignore,omitempty"`
	FolderWatch       bool       `json:"Watch" yaml:"Watch,omitempty"`
	FileCount         int        `gorm:"-" json:"FileCount" yaml:"-"`
	CreatedAt         time.Time  `json:"-" yaml:"-"`
	UpdatedAt         time.Time  `json:"-" yaml:"-"`
	ModifiedAt        *time.Time `json:"ModifiedAt,omitempty" yaml:"-"`
	DeletedAt         *time.Time `sql:"index" json:"-"`
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Folder) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.FolderUID, 'd') {
		return nil
	}

	return scope.SetColumn("FolderUID", rnd.PPID('d'))
}

// NewFolder creates a new file system directory entity.
func NewFolder(root, pathName string, modTime *time.Time) Folder {
	now := Timestamp()

	pathName = strings.Trim(pathName, string(os.PathSeparator))

	if pathName == RootPath {
		pathName = ""
	}

	result := Folder{
		FolderUID:     rnd.PPID('d'),
		Root:          root,
		Path:          pathName,
		FolderType:    TypeDefault,
		FolderOrder:   SortOrderName,
		FolderCountry: UnknownCountry.ID,
		FolderYear:    0,
		FolderMonth:   0,
		ModifiedAt:    modTime,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	result.SetValuesFromPath()

	return result
}

// SetValuesFromPath updates the title and other values based on the path name.
func (m *Folder) SetValuesFromPath() {
	s := m.Path
	s = strings.TrimSpace(s)

	if s == "" || s == RootPath {
		if m.Root == RootOriginals {
			m.FolderTitle = "Originals"

			return
		} else {
			s = m.Root
		}
	} else {
		m.FolderCountry = txt.CountryCode(s)
		m.FolderYear = txt.Year(s)
		s = path.Base(s)
	}

	if len(m.Path) >= 6 {
		if date := txt.Time(m.Path); !date.IsZero() {
			if txt.IsUInt(s) || txt.IsTime(s) {
				if date.Day() > 1 {
					m.FolderTitle = date.Format("January 2, 2006")
				} else {
					m.FolderTitle = date.Format("January 2006")
				}
			}

			m.FolderYear = date.Year()
			m.FolderMonth = int(date.Month())
			m.FolderDay = date.Day()
		}
	}

	if m.FolderTitle == "" {
		s = strings.ReplaceAll(s, "_", " ")
		s = strings.ReplaceAll(s, "-", " ")
		s = strings.Title(s)

		m.FolderTitle = txt.Clip(s, txt.ClipDefault)
	}
}

// Slug returns a slug based on the folder title.
func (m *Folder) Slug() string {
	return slug.Make(m.FolderTitle)
}

// Title returns a human readable folder title.
func (m *Folder) Title() string {
	return m.FolderTitle
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

// FindFolder returns an existing row if exists.
func FindFolder(root, pathName string) *Folder {
	pathName = strings.Trim(pathName, string(os.PathSeparator))

	result := Folder{}

	if err := Db().Where("path = ? AND root = ?", pathName, root).First(&result).Error; err == nil {
		return &result
	}

	return nil
}

// FirstOrCreateFolder returns the existing row, inserts a new row or nil in case of errors.
func FirstOrCreateFolder(m *Folder) *Folder {
	result := Folder{}

	if err := Db().Where("path = ? AND root = ?", m.Path, m.Root).First(&result).Error; err == nil {
		return &result
	} else if createErr := m.Create(); createErr == nil {
		return m
	} else if err := Db().Where("path = ? AND root = ?", m.Path, m.Root).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("folder: %s (first or create %s)", createErr, m.Path)
	}

	return nil
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
