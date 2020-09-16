package entity

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

type People []Person

// Person represents a real person that can also be a user if a password is set.
type Person struct {
	ID              int        `gorm:"primary_key" json:"-" yaml:"-"`
	Address         *Address   `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false;PRELOAD:true;" json:"Address,omitempty" yaml:"Address,omitempty"`
	AddressID       int        `gorm:"default:1" json:"-" yaml:"-"`
	PersonUID       string     `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	ParentUID       string     `gorm:"type:VARBINARY(42);" json:"ParentUID" yaml:"ParentUID,omitempty"`
	GlobalUID       string     `gorm:"type:VARBINARY(42);index;" json:"GlobalUID" yaml:"GlobalUID,omitempty"`
	DisplayName     string     `gorm:"size:128;" json:"DisplayName" yaml:"DisplayName,omitempty"`
	DisplayLocation string     `gorm:"size:128;" json:"DisplayLocation" yaml:"DisplayLocation,omitempty"`
	DisplayBio      string     `gorm:"type:TEXT;" json:"DisplayBio" yaml:"DisplayBio,omitempty"`
	NamePrefix      string     `gorm:"size:64;" json:"NamePrefix" yaml:"NamePrefix,omitempty"`
	GivenName       string     `gorm:"size:128;" json:"GivenName" yaml:"GivenName,omitempty"`
	FamilyName      string     `gorm:"size:128;" json:"FamilyName" yaml:"FamilyName,omitempty"`
	NameSuffix      string     `gorm:"size:64;" json:"NameSuffix" yaml:"NameSuffix,omitempty"`
	PrimaryEmail    string     `gorm:"size:255;index;" json:"PrimaryEmail" yaml:"PrimaryEmail,omitempty"`
	BackupEmail     string     `gorm:"size:255;" json:"BackupEmail" yaml:"BackupEmail,omitempty"`
	PersonURL       string     `gorm:"type:VARBINARY(255);" json:"PersonURL" yaml:"PersonURL,omitempty"`
	PersonPhone     string     `gorm:"size:32;" json:"PersonPhone" yaml:"PersonPhone,omitempty"`
	PersonStatus    string     `gorm:"type:VARBINARY(32);" json:"PersonStatus" yaml:"PersonStatus,omitempty"`
	PersonAvatar    string     `gorm:"type:VARBINARY(255);" json:"PersonAvatar" yaml:"PersonAvatar,omitempty"`
	PersonAccounts  string     `gorm:"type:LONGTEXT;" json:"-" yaml:"-"`
	CompanyURL      string     `gorm:"type:VARBINARY(255);" json:"CompanyURL" yaml:"CompanyURL,omitempty"`
	CompanyPhone    string     `gorm:"size:32;" json:"CompanyPhone" yaml:"CompanyPhone,omitempty"`
	CompanyName     string     `gorm:"size:128;" json:"CompanyName" yaml:"CompanyName,omitempty"`
	DepartmentName  string     `gorm:"size:128;" json:"DepartmentName" yaml:"DepartmentName,omitempty"`
	JobTitle        string     `gorm:"size:64;" json:"JobTitle" yaml:"JobTitle,omitempty"`
	BirthYear       int        `json:"BirthYear" yaml:"BirthYear,omitempty"`
	BirthMonth      int        `json:"BirthMonth" yaml:"BirthMonth,omitempty"`
	BirthDay        int        `json:"BirthDay" yaml:"BirthDay,omitempty"`
	UserName        string     `gorm:"size:64;" json:"UserName" yaml:"UserName,omitempty"`
	UserSettings    string     `gorm:"type:LONGTEXT;" json:"-" yaml:"-"`
	IsActive        bool       `json:"IsActive" yaml:"IsActive,omitempty"`
	IsConfirmed     bool       `json:"IsConfirmed" yaml:"IsConfirmed,omitempty"`
	IsArtist        bool       `json:"IsArtist" yaml:"IsArtist,omitempty"`
	IsSubject       bool       `json:"IsSubject" yaml:"IsSubject,omitempty"`
	RoleAdmin       bool       `json:"RoleAdmin" yaml:"RoleAdmin,omitempty"`
	RoleGuest       bool       `json:"RoleGuest" yaml:"RoleGuest,omitempty"`
	RoleChild       bool       `json:"RoleChild" yaml:"RoleChild,omitempty"`
	RoleFamily      bool       `json:"RoleFamily" yaml:"RoleFamily,omitempty"`
	RoleFriend      bool       `json:"RoleFriend" yaml:"RoleFriend,omitempty"`
	WebDAV          bool       `gorm:"column:webdav" json:"WebDAV" yaml:"WebDAV,omitempty"`
	StoragePath     string     `gorm:"column:storage_path;type:VARBINARY(255);" json:"StoragePath" yaml:"StoragePath,omitempty"`
	ConfirmToken    string     `gorm:"type:VARBINARY(128);" json:"-" yaml:"-"`
	ResetToken      string     `gorm:"type:VARBINARY(128);" json:"-" yaml:"-"`
	ApiToken        string     `gorm:"column:api_token;type:VARBINARY(128);" json:"-" yaml:"-"`
	ApiSecret       string     `gorm:"column:api_secret;type:VARBINARY(128);" json:"-" yaml:"-"`
	LoginAttempts   int        `json:"-" yaml:"-"`
	LoginAt         *time.Time `json:"-" yaml:"-"`
	EulaSigned      *time.Time `json:"EulaSigned" yaml:"EulaSigned,omitempty"`
	CreatedAt       time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt       time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt       *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName the database table name.
func (Person) TableName() string {
	return "people"
}

// Default admin user.
var Admin = Person{
	ID:          1,
	AddressID:   1,
	UserName:    "admin",
	DisplayName: "Admin",
	RoleAdmin:   true,
	IsActive:    true,
}

// Anonymous, public user without own account.
var UnknownPerson = Person{
	ID:          -1,
	AddressID:   1,
	PersonUID:   "u000000000000001",
	UserName:    "",
	DisplayName: "Anonymous",
	RoleAdmin:   false,
	RoleGuest:   false,
	IsActive:    false,
}

// Guest user without own account for link sharing.
var Guest = Person{
	ID:          -2,
	AddressID:   1,
	PersonUID:   "u000000000000002",
	UserName:    "",
	DisplayName: "Guest",
	RoleAdmin:   false,
	RoleGuest:   true,
	IsActive:    false,
}

// CreateDefaultUsers initializes the database with default user accounts.
func CreateDefaultUsers() {
	if user := FirstOrCreatePerson(&Admin); user != nil {
		Admin = *user
	}

	if user := FirstOrCreatePerson(&UnknownPerson); user != nil {
		UnknownPerson = *user
	}

	if user := FirstOrCreatePerson(&Guest); user != nil {
		Guest = *user
	}
}

// Create inserts a new row to the database.
func (m *Person) Create() error {
	return Db().Create(m).Error
}

// Saves the new row to the database.
func (m *Person) Save() error {
	return Db().Save(m).Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Person) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.PersonUID, 'u') {
		return nil
	}

	return scope.SetColumn("PersonUID", rnd.PPID('u'))
}

// FirstOrCreatePerson returns an existing row, inserts a new row or nil in case of errors.
func FirstOrCreatePerson(m *Person) *Person {
	result := Person{}

	if err := Db().Preload("Address").Where("id = ? OR person_uid = ?", m.ID, m.PersonUID).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Debugf("person: %s", err)
		return nil
	}

	return m
}

// FindPersonByUserName returns an existing user or nil if not found.
func FindPersonByUserName(userName string) *Person {
	if userName == "" {
		return nil
	}

	result := Person{}

	if err := Db().Preload("Address").Where("user_name = ?", userName).First(&result).Error; err == nil {
		return &result
	} else {
		log.Debugf("user %s not found", txt.Quote(userName))
		return nil
	}
}

// FindPersonByUID returns an existing user or nil if not found.
func FindPersonByUID(uid string) *Person {
	if uid == "" {
		return nil
	}

	result := Person{}

	if err := Db().Preload("Address").Where("person_uid = ?", uid).First(&result).Error; err == nil {
		return &result
	} else {
		log.Debugf("user %s not found", txt.Quote(uid))
		return nil
	}
}

// String returns an identifier that can be used in logs.
func (m *Person) String() string {
	if m.UserName != "" {
		return m.UserName
	}

	if m.DisplayName != "" {
		return m.DisplayName
	}

	return m.PersonUID
}

// User returns true if the person has a user name.
func (m *Person) Registered() bool {
	return m.UserName != "" && rnd.IsPPID(m.PersonUID, 'u')
}

// Admin returns true if the person is an admin with user name.
func (m *Person) Admin() bool {
	return m.Registered() && m.RoleAdmin
}

// Anonymous returns true if the person is unknown.
func (m *Person) Anonymous() bool {
	return !rnd.IsPPID(m.PersonUID, 'u') || m.ID == UnknownPerson.ID || m.PersonUID == UnknownPerson.PersonUID
}

// Guest returns true if the person is a guest.
func (m *Person) Guest() bool {
	return m.RoleGuest
}

// SetPassword sets a new password stored as hash.
func (m *Person) SetPassword(password string) error {
	if !m.Registered() {
		return fmt.Errorf("only registered users can change their password")
	}

	if len(password) < 6 {
		return fmt.Errorf("new password for %s must be at least 6 characters", txt.Quote(m.UserName))
	}

	pw := NewPassword(m.PersonUID, password)

	return pw.Save()
}

// InitPassword sets the initial user password stored as hash.
func (m *Person) InitPassword(password string) {
	if !m.Registered() {
		log.Warn("only registered users can change their password")
		return
	}

	if password == "" {
		return
	}

	existing := FindPassword(m.PersonUID)

	if existing != nil {
		return
	}

	pw := NewPassword(m.PersonUID, password)

	if err := pw.Save(); err != nil {
		log.Error(err)
	}
}

// InvalidPassword returns true if the given password does not match the hash.
func (m *Person) InvalidPassword(password string) bool {
	if !m.Registered() {
		log.Warn("only registered users can change their password")
		return true
	}

	if password == "" {
		return true
	}

	pw := FindPassword(m.PersonUID)

	if pw == nil {
		return true
	}

	return pw.InvalidPassword(password)
}

// Role returns the user role for ACL permission checks.
func (m *Person) Role() acl.Role {
	if m.RoleAdmin {
		return acl.RoleAdmin
	}

	if m.RoleChild {
		return acl.RoleChild
	}

	if m.RoleFamily {
		return acl.RoleFamily
	}

	if m.RoleFriend {
		return acl.RoleFriend
	}

	if m.RoleGuest {
		return acl.RoleGuest
	}

	return acl.RoleDefault
}
