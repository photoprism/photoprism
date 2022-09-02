package entity

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/rnd"
)

// Tokens represents a list of auth tokens.
type Tokens []Token

// Token represents a set of authentication tokens.
// Lengths for OAuth 2.0 fields, see https://blogs.intuit.com/blog/2020/03/23/increased-lengths-for-oauth-2-0-fields/.
type Token struct {
	ID          int        `gorm:"primary_key" json:"-" yaml:"-"`
	TokenUID    string     `gorm:"type:VARBINARY(42);column:token_uid;unique_index;" json:"UID" yaml:"UID"`
	TokenName   string     `gorm:"type:VARCHAR(64);column:token_name;" json:"Name,omitempty" yaml:"Name,omitempty"`
	TokenType   string     `gorm:"type:VARCHAR(64);column:token_type;" json:"Type,omitempty" yaml:"Type,omitempty"`
	UserUID     string     `gorm:"type:VARBINARY(42);column:user_uid;index;" json:"UserUID" yaml:"UserUID"`
	Version     string     `gorm:"type:VARCHAR(32);column:version;default:'oauth20';" json:"Version,omitempty" yaml:"Version,omitempty"`
	Auth        string     `gorm:"type:VARBINARY(512);column:auth;" json:"Auth,omitempty" yaml:"Auth,omitempty"`
	Access      string     `gorm:"type:VARBINARY(4096);column:access;" json:"Access,omitempty" yaml:"Access,omitempty"`
	Refresh     string     `gorm:"type:VARBINARY(512);column:refresh;" json:"Refresh,omitempty" yaml:"Refresh,omitempty"`
	FileRoot    string     `gorm:"type:VARBINARY(16);column:file_root;" json:"FileRoot" yaml:"FileRoot,omitempty"`
	FilePath    string     `gorm:"type:VARBINARY(500);column:file_path;" json:"FilePath" yaml:"FilePath,omitempty"`
	CanIndex    bool       `gorm:"default:false;" json:"CanIndex" yaml:"CanIndex,omitempty"`
	CanImport   bool       `gorm:"default:false;" json:"CanImport" yaml:"CanImport,omitempty"`
	CanUpload   bool       `gorm:"default:false;" json:"CanUpload" yaml:"CanUpload,omitempty"`
	CanDownload bool       `gorm:"default:false;" json:"CanDownload" yaml:"CanDownload,omitempty"`
	CanSearch   bool       `gorm:"default:false;" json:"CanSearch" yaml:"CanSearch,omitempty"`
	CanShare    bool       `gorm:"default:false;" json:"CanShare" yaml:"CanShare,omitempty"`
	CanEdit     bool       `gorm:"default:false;" json:"CanEdit" yaml:"CanEdit,omitempty"`
	CanComment  bool       `gorm:"default:false;" json:"CanComment" yaml:"CanComment,omitempty"`
	CanDelete   bool       `gorm:"default:false;" json:"CanDelete" yaml:"CanDelete,CanCreate"`
	Notes       string     `gorm:"type:VARCHAR(512);column:notes;" json:"Notes,omitempty" yaml:"Notes,omitempty"`
	CreatedAt   time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt   time.Time  `json:"UpdatedAt" yaml:"-"`
	ExpiresAt   *time.Time `json:"ExpiresAt,omitempty" yaml:"ExpiresAt,omitempty"`
	DeletedAt   *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity database table name.
func (Token) TableName() string {
	return "auth_tokens"
}

// Create new entity in the database.
func (m *Token) Create() error {
	return Db().Create(m).Error
}

// Save entity properties.
func (m *Token) Save() error {
	return Db().Save(m).Error
}

// Updates multiple properties in the database.
func (m *Token) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Token) BeforeCreate(tx *gorm.DB) error {
	if rnd.ValidID(m.TokenUID, 'u') {
		return nil
	}
	m.TokenUID = rnd.GenerateUID('u')
	return nil
}
