package entity

import (
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/ulule/deepcopier"

	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

var folderMutex = sync.Mutex{}

type Folders []Folder

// Folder represents a file system directory.
type Folder struct {
	Path              string     `gorm:"type:VARBINARY(1024);unique_index:idx_folders_path_root;" json:"Path" yaml:"Path"`
	Root              string     `gorm:"type:VARBINARY(16);default:'';unique_index:idx_folders_path_root;" json:"Root" yaml:"Root,omitempty"`
	FolderUID         string     `gorm:"type:VARBINARY(42);primary_key;" json:"UID,omitempty" yaml:"UID,omitempty"`
	FolderType        string     `gorm:"type:VARBINARY(16);" json:"Type" yaml:"Type,omitempty"`
	FolderTitle       string     `gorm:"type:VARCHAR(200);" json:"Title" yaml:"Title,omitempty"`
	FolderCategory    string     `gorm:"type:VARCHAR(100);index;" json:"Category" yaml:"Category,omitempty"`
	FolderDescription string     `gorm:"type:VARCHAR(2048);" json:"Description,omitempty" yaml:"Description,omitempty"`
	FolderOrder       string     `gorm:"type:VARBINARY(32);" json:"Order" yaml:"Order,omitempty"`
	FolderCountry     string     `gorm:"type:VARBINARY(2);index:idx_folders_country_year_month;default:'zz'" json:"Country" yaml:"Country,omitempty"`
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
	ModifiedAt        time.Time  `json:"ModifiedAt,omitempty" yaml:"-"`
	PublishedAt       *time.Time `sql:"index" json:"PublishedAt,omitempty" yaml:"PublishedAt,omitempty"`
	DeletedAt         *time.Time `sql:"index" json:"-"`
}

// TableName returns the entity table name.
func (Folder) TableName() string {
	return "folders"
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Folder) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUnique(m.FolderUID, 'd') {
		return nil
	}

	return scope.SetColumn("FolderUID", rnd.GenerateUID('d'))
}

// NewFolder creates a new file system directory entity.
func NewFolder(root, dir string, modTime time.Time) Folder {
	now := Now()

	dir = strings.Trim(dir, string(os.PathSeparator))

	if dir == RootPath {
		dir = ""
	}

	year := 0
	month := 0

	if !modTime.IsZero() {
		year = modTime.Year()
		month = int(modTime.Month())
	}

	result := Folder{
		FolderUID:     rnd.GenerateUID('d'),
		Root:          root,
		Path:          dir,
		FolderType:    MediaUnknown,
		FolderOrder:   sortby.Name,
		FolderCountry: UnknownCountry.ID,
		FolderYear:    year,
		FolderMonth:   month,
		ModifiedAt:    modTime.UTC(),
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

		if year := txt.Year(s); year > 0 {
			m.FolderYear = year
		}

		s = path.Base(s)
	}

	if len(m.Path) >= 6 {
		if date := txt.DateFromFilePath(m.Path); !date.IsZero() {
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
		m.FolderTitle = txt.Clip(txt.Title(s), txt.ClipLongName)
	}
}

// Slug returns a slug based on the folder title.
func (m *Folder) Slug() string {
	return txt.Slug(m.Path)
}

// RootPath returns the full folder path including root.
func (m *Folder) RootPath() string {
	return path.Join(m.Root, m.Path)
}

// Title returns the human-readable folder title.
func (m *Folder) Title() string {
	return m.FolderTitle
}

// Create inserts the entity to the index.
func (m *Folder) Create() error {
	folderMutex.Lock()
	defer folderMutex.Unlock()

	if err := Db().Create(m).Error; err != nil {
		return err
	} else if m.Root != RootOriginals || m.Path == "" {
		return nil
	}

	f := form.SearchPhotos{
		Path:   m.Path,
		Public: true,
	}

	if a := FindFolderAlbum(m.Path); a != nil {
		if a.DeletedAt != nil {
			// Ignore.
		} else if err := a.UpdateFolder(m.Path, f.Serialize()); err != nil {
			log.Errorf("folder: %s (update album)", err.Error())
		}
	} else if a := NewFolderAlbum(m.Title(), m.Path, f.Serialize()); a != nil {
		a.AlbumYear = m.FolderYear
		a.AlbumMonth = m.FolderMonth
		a.AlbumDay = m.FolderDay
		a.AlbumCountry = m.FolderCountry

		if err := a.Create(); err != nil {
			log.Errorf("folder: %s (add album)", err)
		} else {
			log.Infof("folder: added album %s (%s)", clean.Log(a.AlbumTitle), a.AlbumFilter)
		}
	}

	return nil
}

// FindFolder returns an existing folder if it exists.
func FindFolder(root, dir string) *Folder {
	dir = strings.Trim(dir, string(os.PathSeparator))

	if dir == RootPath {
		dir = ""
	}

	result := Folder{}

	if err := Db().Where("path = ? AND root = ?", dir, root).First(&result).Error; err == nil {
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
		log.Errorf("folder: %s (find or create %s)", createErr, m.Path)
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

	m.FolderTitle = txt.Clip(m.FolderTitle, txt.ClipLongName)
	m.FolderCategory = txt.Clip(m.FolderCategory, txt.ClipCategory)

	return nil
}
