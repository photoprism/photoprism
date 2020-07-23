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
	ID            int        `gorm:"primary_key" json:"ID" yaml:"-"`
	PersonUID     string     `gorm:"type:varbinary(42);unique_index;" json:"UID" yaml:"UID"`
	UserName      string     `gorm:"type:varchar(32);" json:"UserName" yaml:"UserName,omitempty"`
	FirstName     string     `gorm:"type:varchar(32);" json:"FirstName" yaml:"FirstName,omitempty"`
	LastName      string     `gorm:"type:varchar(32);" json:"LastName" yaml:"LastName,omitempty"`
	DisplayName   string     `gorm:"type:varchar(64);" json:"DisplayName" yaml:"DisplayName,omitempty"`
	UserEmail     string     `gorm:"type:varchar(255);" json:"Email" yaml:"Email,omitempty"`
	UserInfo      string     `gorm:"type:text;" json:"Info" yaml:"Info,omitempty"`
	UserPath      string     `json:"UserPath" yaml:"UserPath,omitempty"`
	UserActive    bool       `json:"Active" yaml:"Active,omitempty"`
	UserConfirmed bool       `json:"Confirmed" yaml:"Confirmed,omitempty"`
	RoleAdmin     bool       `json:"Admin" yaml:"Admin,omitempty"`
	RoleGuest     bool       `json:"Guest" yaml:"Guest,omitempty"`
	RoleChild     bool       `json:"Child" yaml:"Child,omitempty"`
	RoleFamily    bool       `json:"Family" yaml:"Family,omitempty"`
	RoleFriend    bool       `json:"Friend" yaml:"Friend,omitempty"`
	IsArtist      bool       `json:"Artist" yaml:"Artist,omitempty"`
	IsSubject     bool       `json:"Subject" yaml:"Subject,omitempty"`
	CanEdit       bool       `json:"CanEdit" yaml:"CanEdit,omitempty"`
	CanComment    bool       `json:"CanComment" yaml:"CanComment,omitempty"`
	CanUpload     bool       `json:"CanUpload" yaml:"CanUpload,omitempty"`
	CanDownload   bool       `json:"CanDownload" yaml:"CanDownload,omitempty"`
	WebDAV        bool       `gorm:"column:webdav" json:"WebDAV" yaml:"WebDAV,omitempty"`
	ApiToken      string     `json:"ApiToken" yaml:"ApiToken,omitempty"`
	BirthYear     int        `json:"BirthYear" yaml:"BirthYear,omitempty"`
	BirthMonth    int        `json:"BirthMonth" yaml:"BirthMonth,omitempty"`
	BirthDay      int        `json:"BirthDay" yaml:"BirthDay,omitempty"`
	LoginAttempts int        `json:"-" yaml:"-,omitempty"`
	LoginAt       *time.Time `json:"-" yaml:"-"`
	CreatedAt     time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt     time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt     *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// Default admin user.
var Admin = Person{
	ID:            1,
	UserName:      "admin",
	DisplayName:   "Admin",
	RoleAdmin:     true,
	UserActive:    true,
	UserConfirmed: true,
}

// Anonymous, public user without own account.
var UnknownPerson = Person{
	ID:            -1,
	PersonUID:     "u000000000000001",
	UserName:      "",
	DisplayName:   "Anonymous",
	RoleAdmin:     false,
	RoleGuest:     false,
	UserActive:    false,
	UserConfirmed: false,
}

// Guest user without own account for link sharing.
var Guest = Person{
	ID:            -2,
	PersonUID:     "u000000000000002",
	UserName:      "",
	DisplayName:   "Guest",
	RoleAdmin:     false,
	RoleGuest:     true,
	UserActive:    false,
	UserConfirmed: false,
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

	if err := Db().Where("id = ? OR person_uid = ?", m.ID, m.PersonUID).First(&result).Error; err == nil {
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

	if err := Db().Where("user_name = ?", userName).First(&result).Error; err == nil {
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

	if err := Db().Where("person_uid = ?", uid).First(&result).Error; err == nil {
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
