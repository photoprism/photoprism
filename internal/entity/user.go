package entity

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/form"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

type Users []User

// User represents a person that may optionally log in as user.
type User struct {
	ID             int        `gorm:"primary_key" json:"-" yaml:"-"`
	Address        *Address   `gorm:"association_autoupdate:false;association_autocreate:false;association_save_reference:false;PRELOAD:true;" json:"Address,omitempty" yaml:"Address,omitempty"`
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
	ExternalID     string     `gorm:"size:255;column:external_id;" json:"-" yaml:"-"`
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

// CreateDefaultUsers initializes the database with default user accounts.
func CreateDefaultUsers() {
	if user := FirstOrCreateUser(&Admin); user != nil {
		Admin = *user
	}

	if user := FirstOrCreateUser(&UnknownUser); user != nil {
		UnknownUser = *user
	}

	if user := FirstOrCreateUser(&Guest); user != nil {
		Guest = *user
	}
}

// Create new entity in the database.
func (m *User) Create() error {
	return Db().Create(m).Error
}

// Save entity properties.
func (m *User) Save() error {
	return Db().Save(m).Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *User) BeforeCreate(tx *gorm.DB) error {
	if rnd.IsUID(m.UserUID, 'u') {
		return nil
	}
	m.UserUID = rnd.PPID('u')
	return nil
}

// FirstOrCreateUser returns an existing row, inserts a new row, or nil in case of errors.
func FirstOrCreateUser(m *User) *User {
	result := User{}

	if err := Db().Preload("Address").Where("id = ? OR user_uid = ?", m.ID, m.UserUID).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Debugf("user: %s", err)
		return nil
	}

	return m
}

// CreateOrUpdateExternalUser Retrieves User by its ExternalID and updates relevant properties. If no User exists, a new one will be inserted into database.
func CreateOrUpdateExternalUser(m *User) (*User, error) {
	result := User{}

	if err := Db().Preload("Address").Where("external_id = ?", m.ExternalID).First(&result).Error; err == nil {
		mustUpdate := false
		if m.PrimaryEmail != result.PrimaryEmail {
			result.PrimaryEmail = m.PrimaryEmail
			mustUpdate = true
		}
		if m.FullName != result.FullName {
			result.FullName = m.FullName
			mustUpdate = true
		}
		if mustUpdate {
			err := result.Validate()
			if err != nil {
				log.Errorf("user: %s", err)
				return nil, err
			}
			err = result.Save()
			if err != nil {
				log.Errorf("user: %s", err)
				return nil, err
			}
		}
		log.Debugf("user: retrieved %q from db", result.UserName)
		return &result, nil
	} else {
		err := m.Validate()
		if err == ErrUsernameAlreadyExists {
			if e2 := m.AppendUsernameSuffix(); e2 != nil {
				return nil, e2
			}
		} else if err != nil {
			log.Errorf("user: %s", err)
			return nil, err
		}
		if err := m.Create(); err != nil {
			log.Debugf("user: %s", err)
			return nil, err
		}
		log.Debugf("user: created %q", result.UserName)
	}

	return m, nil
}

// FindUserByName returns an existing user or nil if not found.
func FindUserByName(userName string) *User {
	if userName == "" {
		return nil
	}

	result := User{}

	if err := Db().Preload("Address").Where("user_name = ?", userName).First(&result).Error; err == nil {
		return &result
	} else {
		log.Debugf("user %s not found", txt.Quote(userName))
		return nil
	}
}

// FindUserByUID returns an existing user or nil if not found.
func FindUserByUID(uid string) *User {
	if uid == "" {
		return nil
	}

	result := User{}

	if err := Db().Preload("Address").Where("user_uid = ?", uid).First(&result).Error; err == nil {
		return &result
	} else {
		log.Debugf("user %s not found", txt.Quote(uid))
		return nil
	}
}

// Delete marks the entity as deleted.
func (m *User) Delete() error {
	if m.ID <= 1 {
		return fmt.Errorf("can't delete system user")
	}

	return Db().Delete(m).Error
}

// Deleted tests if the entity is marked as deleted.
func (m *User) Deleted() bool {
	return m.DeletedAt != nil
}

// String returns an identifier that can be used in logs.
func (m *User) String() string {
	if m.UserName != "" {
		return m.UserName
	}

	if m.FullName != "" {
		return m.FullName
	}

	return m.UserUID
}

// Registered tests if the user is registered e.g. has a username.
func (m *User) Registered() bool {
	return m.UserName != "" && rnd.IsPPID(m.UserUID, 'u')
}

// Admin returns true if the user is an admin with user name.
func (m *User) Admin() bool {
	return m.Registered() && m.RoleAdmin
}

// Anonymous returns true if the user is unknown.
func (m *User) Anonymous() bool {
	return !rnd.IsPPID(m.UserUID, 'u') || m.ID == UnknownUser.ID || m.UserUID == UnknownUser.UserUID
}

// Guest returns true if the user is a guest.
func (m *User) Guest() bool {
	return m.RoleGuest
}

// Guest returns true if the user is a guest.
func (m *User) External() bool {
	return m.ExternalID != ""
}

// SetPassword sets a new password stored as hash.
func (m *User) SetPassword(password string) error {
	if !m.Registered() {
		return fmt.Errorf("only registered users can change their password")
	}

	if len(password) < 4 {
		return fmt.Errorf("new password for %s must be at least 4 characters", txt.Quote(m.UserName))
	}

	pw := NewPassword(m.UserUID, password)

	return pw.Save()
}

// InitPassword sets the initial user password stored as hash.
func (m *User) InitPassword(password string) {
	if !m.Registered() {
		log.Warn("only registered users can change their password")
		return
	}

	if password == "" {
		return
	}

	existing := FindPassword(m.UserUID)

	if existing != nil {
		return
	}

	pw := NewPassword(m.UserUID, password)

	if err := pw.Save(); err != nil {
		log.Error(err)
	}
}

// InvalidPassword returns true if the given password does not match the hash.
func (m *User) InvalidPassword(password string) bool {
	if !m.Registered() {
		log.Warn("only registered users can change their password")
		return true
	}

	if password == "" {
		return true
	}

	time.Sleep(time.Second * 5 * time.Duration(m.LoginAttempts))

	pw := FindPassword(m.UserUID)

	if pw == nil {
		return true
	}

	if pw.InvalidPassword(password) {
		if err := Db().Model(m).UpdateColumn("login_attempts", gorm.Expr("login_attempts + ?", 1)).Error; err != nil {
			log.Errorf("user: %s (update login attempts)", err)
		}

		return true
	}

	if err := Db().Model(m).Updates(map[string]interface{}{"login_attempts": 0, "login_at": TimeStamp()}).Error; err != nil {
		log.Errorf("user: %s (update last login)", err)
	}

	return false
}

// Role returns the user role for ACL permission checks.
func (m *User) Role() acl.Role {
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

	if m.Registered() {
		return acl.RoleMember
	}

	return acl.RoleDefault
}

var ErrUsernameAlreadyExists = errors.New("username already exists")

// Validate Makes sure username and email are unique and meet requirements. Returns error if any property is invalid
func (m *User) Validate() error {
	if m.UserName == "" {
		return errors.New("username must not be empty")
	}
	if len(m.UserName) < 4 {
		return errors.New("username must be at least 4 characters")
	}
	var err error
	var resultName = User{}
	if err = Db().Where("user_name = ? AND id <> ?", m.UserName, m.ID).First(&resultName).Error; err == nil {
		return ErrUsernameAlreadyExists
	} else if err != gorm.ErrRecordNotFound {
		return err
	}
	// stop here if no email is provided
	if m.PrimaryEmail == "" {
		return nil
	}
	// validate email address
	if a, err := mail.ParseAddress(m.PrimaryEmail); err != nil {
		return err
	} else {
		m.PrimaryEmail = a.Address // make sure email address will be used without name
	}
	var resultMail = User{}
	if err = Db().Where("primary_email = ? AND id <> ?", m.PrimaryEmail, m.ID).First(&resultMail).Error; err == nil {
		return errors.New("email already exists")
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	return nil
}

// CreateWithPassword Creates User with Password in db transaction.
func CreateWithPassword(uc form.UserCreate) error {
	u := &User{
		FullName:     uc.FullName,
		UserName:     uc.UserName,
		PrimaryEmail: uc.Email,
		RoleAdmin:    uc.Admin,
	}
	if len(uc.Password) < 4 {
		return fmt.Errorf("new password for %s must be at least 4 characters", txt.Quote(u.UserName))
	}
	err := u.Validate()
	if err != nil {
		return err
	}
	return Db().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(u).Error; err != nil {
			return err
		}
		pw := NewPassword(u.UserUID, uc.Password)
		if err := tx.Create(&pw).Error; err != nil {
			return err
		}
		log.Infof("created user %v with uid %v", txt.Quote(u.UserName), txt.Quote(u.UserUID))
		return nil
	})
}

func (m *User) AppendUsernameSuffix() error {
	oldname := m.UserName
	for i := 1; i < 100; i++ {
		m.UserName = strings.Join([]string{oldname, fmt.Sprintf("%02d", i)}, "-")
		err := m.Validate()
		if err == ErrUsernameAlreadyExists {
			continue
		}
		if err != nil {
			m.UserName = oldname
			return err
		} else {
			return nil
		}
	}
	return errors.New("username suffixes exhausted")
}
