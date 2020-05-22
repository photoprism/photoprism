package entity

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/ulule/deepcopier"
)

const (
	FolderRootUnknown   = ""
	FolderRootOriginals = "originals"
	FolderRootImport    = "import"
)

// Folder represents a file system directory.
type Folder struct {
	Root              string     `gorm:"type:varbinary(255);unique_index:idx_folders_root_path;" json:"Root" yaml:"Root"`
	Path              string     `gorm:"type:varbinary(1024);unique_index:idx_folders_root_path;" json:"Path" yaml:"Path"`
	FolderUUID        string     `gorm:"type:varbinary(36);primary_key;" json:"PPID,omitempty" yaml:"PPID,omitempty"`
	FolderTitle       string     `gorm:"type:varchar(255);" json:"Title" yaml:"Title,omitempty"`
	FolderDescription string     `gorm:"type:text;" json:"Description,omitempty" yaml:"Description,omitempty"`
	FolderType        string     `gorm:"type:varbinary(16);" json:"Type" yaml:"Type,omitempty"`
	FolderOrder       string     `gorm:"type:varbinary(32);" json:"Order" yaml:"Order,omitempty"`
	FolderFavorite    bool       `json:"Favorite" yaml:"Favorite"`
	FolderIgnore      bool       `json:"Ignore" yaml:"Ignore"`
	FolderHidden      bool       `json:"Hidden" yaml:"Hidden"`
	FolderWatch       bool       `json:"Watch" yaml:"Watch"`
	Links             []Link     `gorm:"foreignkey:ShareUUID;association_foreignkey:FolderUUID" json:"-" yaml:"-"`
	CreatedAt         time.Time  `json:"-" yaml:"-"`
	UpdatedAt         time.Time  `json:"-" yaml:"-"`
	ModifiedAt        *time.Time `json:"ModifiedAt,omitempty" yaml:"-"`
	DeletedAt         *time.Time `sql:"index" json:"-"`
}

// BeforeCreate creates a random UUID if needed before inserting a new row to the database.
func (m *Folder) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsPPID(m.FolderUUID, 'd') {
		return nil
	}

	return scope.SetColumn("FolderUUID", rnd.PPID('d'))
}

// NewFolder creates a new file system directory entity.
func NewFolder(root, pathName string, modTime *time.Time) Folder {
	now := time.Now().UTC()

	pathName = strings.Trim(pathName, string(os.PathSeparator))

	if pathName == "" {
		pathName = "/"
	}

	result := Folder{
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

	if s == "" || s == "/" {
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

// Saves the entity using form data and stores it in the database.
func (m *Folder) Save(f form.Folder) error {
	if err := deepcopier.Copy(m).From(f); err != nil {
		return err
	}

	return Db().Save(m).Error
}

// Returns a slice of folders in a given directory incl sub directories in recursive mode.
func Folders(root, dir string, recursive bool) ([]Folder, error) {
	dirs, err := fs.Dirs(dir, recursive)

	if err != nil {
		return []Folder{}, err
	}

	folders := make([]Folder, len(dirs))

	for i, p := range dirs {
		folders[i] = NewFolder(root, p, nil)
	}

	return folders, nil
}
