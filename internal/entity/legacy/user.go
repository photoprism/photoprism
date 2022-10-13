package legacy

import (
	"time"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

type Users []User

// User represents a person that may optionally log in as user.
type User struct {
	ID             int        `gorm:"primary_key" json:"-" yaml:"-"`
	AddressID      int        `gorm:"default:1" json:"-" yaml:"-"`
	UserUID        string     `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	MotherUID      string     `gorm:"type:VARBINARY(42);" json:"MotherUID" yaml:"MotherUID,omitempty"`
	FatherUID      string     `gorm:"type:VARBINARY(42);" json:"FatherUID" yaml:"FatherUID,omitempty"`
	GlobalUID      string     `gorm:"type:VARBINARY(42);index;" json:"GlobalUID" yaml:"GlobalUID,omitempty"`
	FullName       string     `gorm:"size:128;" json:"FullName" yaml:"FullName,omitempty"`
	NickName       string     `gorm:"size:64;" json:"NickName" yaml:"NickName,omitempty"`
	MaidenName     string     `gorm:"size:64;" json:"MaidenName" yaml:"MaidenName,omitempty"`
	ArtistName     string     `gorm:"size:64;" json:"ArtistName" yaml:"ArtistName,omitempty"`
	UserName       string     `gorm:"size:64;" json:"UserName" yaml:"UserName,omitempty"`
	UserStatus     string     `gorm:"size:32;" json:"UserStatus" yaml:"UserStatus,omitempty"`
	UserDisabled   bool       `json:"UserDisabled" yaml:"UserDisabled,omitempty"`
	UserSettings   string     `gorm:"type:LONGTEXT;" json:"-" yaml:"-"`
	PrimaryEmail   string     `gorm:"size:255;index;" json:"PrimaryEmail" yaml:"PrimaryEmail,omitempty"`
	EmailConfirmed bool       `json:"EmailConfirmed" yaml:"EmailConfirmed,omitempty"`
	BackupEmail    string     `gorm:"size:255;" json:"BackupEmail" yaml:"BackupEmail,omitempty"`
	PersonURL      string     `gorm:"type:VARBINARY(255);" json:"PersonURL" yaml:"PersonURL,omitempty"`
	PersonPhone    string     `gorm:"size:32;" json:"PersonPhone" yaml:"PersonPhone,omitempty"`
	PersonStatus   string     `gorm:"size:32;" json:"PersonStatus" yaml:"PersonStatus,omitempty"`
	PersonAvatar   string     `gorm:"type:VARBINARY(255);" json:"PersonAvatar" yaml:"PersonAvatar,omitempty"`
	PersonLocation string     `gorm:"size:128;" json:"PersonLocation" yaml:"PersonLocation,omitempty"`
	PersonBio      string     `gorm:"type:TEXT;" json:"PersonBio" yaml:"PersonBio,omitempty"`
	PersonAccounts string     `gorm:"type:LONGTEXT;" json:"-" yaml:"-"`
	BusinessURL    string     `gorm:"type:VARBINARY(255);" json:"BusinessURL" yaml:"BusinessURL,omitempty"`
	BusinessPhone  string     `gorm:"size:32;" json:"BusinessPhone" yaml:"BusinessPhone,omitempty"`
	BusinessEmail  string     `gorm:"size:255;" json:"BusinessEmail" yaml:"BusinessEmail,omitempty"`
	CompanyName    string     `gorm:"size:128;" json:"CompanyName" yaml:"CompanyName,omitempty"`
	DepartmentName string     `gorm:"size:128;" json:"DepartmentName" yaml:"DepartmentName,omitempty"`
	JobTitle       string     `gorm:"size:64;" json:"JobTitle" yaml:"JobTitle,omitempty"`
	BirthYear      int        `json:"BirthYear" yaml:"BirthYear,omitempty"`
	BirthMonth     int        `json:"BirthMonth" yaml:"BirthMonth,omitempty"`
	BirthDay       int        `json:"BirthDay" yaml:"BirthDay,omitempty"`
	TermsAccepted  bool       `json:"TermsAccepted" yaml:"TermsAccepted,omitempty"`
	IsArtist       bool       `json:"IsArtist" yaml:"IsArtist,omitempty"`
	IsSubject      bool       `json:"IsSubject" yaml:"IsSubject,omitempty"`
	RoleAdmin      bool       `json:"RoleAdmin" yaml:"RoleAdmin,omitempty"`
	RoleGuest      bool       `json:"RoleGuest" yaml:"RoleGuest,omitempty"`
	RoleChild      bool       `json:"RoleChild" yaml:"RoleChild,omitempty"`
	RoleFamily     bool       `json:"RoleFamily" yaml:"RoleFamily,omitempty"`
	RoleFriend     bool       `json:"RoleFriend" yaml:"RoleFriend,omitempty"`
	WebDAV         bool       `gorm:"column:webdav" json:"WebDAV" yaml:"WebDAV,omitempty"`
	StoragePath    string     `gorm:"column:storage_path;type:VARBINARY(500);" json:"StoragePath" yaml:"StoragePath,omitempty"`
	CanInvite      bool       `json:"CanInvite" yaml:"CanInvite,omitempty"`
	InviteToken    string     `gorm:"type:VARBINARY(32);" json:"-" yaml:"-"`
	InvitedBy      string     `gorm:"type:VARBINARY(32);" json:"-" yaml:"-"`
	ConfirmToken   string     `gorm:"type:VARBINARY(64);" json:"-" yaml:"-"`
	ResetToken     string     `gorm:"type:VARBINARY(64);" json:"-" yaml:"-"`
	ApiToken       string     `gorm:"column:api_token;type:VARBINARY(128);" json:"-" yaml:"-"`
	ApiSecret      string     `gorm:"column:api_secret;type:VARBINARY(128);" json:"-" yaml:"-"`
	LoginAttempts  int        `json:"-" yaml:"-"`
	LoginAt        *time.Time `json:"-" yaml:"-"`
	CreatedAt      time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt      time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt      *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity database table name.
func (User) TableName() string {
	return "users"
}

// Admin is the default admin user.
var Admin = User{
	ID:           1,
	AddressID:    1,
	UserName:     "admin",
	FullName:     "Admin",
	RoleAdmin:    true,
	UserDisabled: false,
}

// UnknownUser is an anonymous, public user without own account.
var UnknownUser = User{
	ID:           -1,
	AddressID:    1,
	UserUID:      "u000000000000001",
	UserName:     "",
	FullName:     "Anonymous",
	RoleAdmin:    false,
	RoleGuest:    false,
	UserDisabled: true,
}

// Guest is a user without own account e.g. for link sharing.
var Guest = User{
	ID:           -2,
	AddressID:    1,
	UserUID:      "u000000000000002",
	UserName:     "",
	FullName:     "Guest",
	RoleAdmin:    false,
	RoleGuest:    true,
	UserDisabled: true,
}

// Deleted tests if the entity is marked as deleted.
func (m *User) Deleted() bool {
	if m.DeletedAt == nil {
		return false
	}

	return !m.DeletedAt.IsZero()
}

// String returns an identifier that can be used in logs.
func (m *User) String() string {
	if n := m.Username(); n != "" {
		return clean.Log(n)
	}

	if m.FullName != "" {
		return clean.Log(m.FullName)
	}

	return clean.Log(m.UserUID)
}

// Username returns the normalized username.
func (m *User) Username() string {
	return clean.Username(m.UserName)
}

// Registered tests if the user is registered e.g. has a username.
func (m *User) Registered() bool {
	return m.Username() != "" && rnd.IsUID(m.UserUID, 'u')
}

// Admin returns true if the user is an admin with user name.
func (m *User) Admin() bool {
	return m.Registered() && m.RoleAdmin
}

// Anonymous returns true if the user is unknown.
func (m *User) Anonymous() bool {
	return !rnd.IsUID(m.UserUID, 'u') || m.ID == UnknownUser.ID || m.UserUID == UnknownUser.UserUID
}

// Guest returns true if the user is a guest.
func (m *User) Guest() bool {
	return m.RoleGuest
}
