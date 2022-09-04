package entity

import (
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

var UsernameLength = 3
var PasswordLength = 4

// Users represents a list of users.
type Users []User

// User represents a person that may optionally log in as user.
type User struct {
	ID             int        `gorm:"primary_key" json:"-" yaml:"-"`
	UserUID        string     `gorm:"type:VARBINARY(42);unique_index;" json:"UID" yaml:"UID"`
	UserSlug       string     `gorm:"type:VARBINARY(160);unique_index;" json:"Slug" yaml:"Slug,omitempty"`
	Username       string     `gorm:"size:64;index;" json:"Username" yaml:"Username,omitempty"`
	Email          string     `gorm:"size:255;index;" json:"Email" yaml:"Email,omitempty"`
	UserRole       string     `gorm:"size:32;" json:"Role" yaml:"Role,omitempty"`
	SuperAdmin     bool       `gorm:"default:false;" json:"SuperAdmin,omitempty" yaml:"SuperAdmin,omitempty"`
	CanLogin       bool       `gorm:"default:false;" json:"CanLogin,omitempty" yaml:"CanLogin,omitempty"`
	CanInvite      bool       `gorm:"default:false;" json:"CanInvite,omitempty" yaml:"CanInvite,omitempty"`
	ShareUID       string     `gorm:"type:VARBINARY(42);index;" json:"ShareUID" yaml:"ShareUID,omitempty"`
	AuthUID        string     `gorm:"type:VARBINARY(512);column:auth_uid;" json:"-" yaml:"-"`
	AuthSrc        string     `gorm:"type:VARBINARY(64);column:auth_src;" json:"-" yaml:"-"`
	WebDAV         string     `gorm:"size:16;column:webdav;" json:"WebDAV,omitempty" yaml:"WebDAV,omitempty"`
	AvatarURL      string     `gorm:"type:VARBINARY(255);column:avatar_url;" json:"AvatarURL" yaml:"AvatarURL,omitempty"`
	AvatarSrc      string     `gorm:"type:VARBINARY(64);column:avatar_src;" json:"AvatarSrc" yaml:"AvatarSrc,omitempty"`
	UserCountry    string     `gorm:"type:VARBINARY(2);" json:"Country" yaml:"Country,omitempty"`
	UserLocale     string     `gorm:"type:VARBINARY(64);" json:"Locale" yaml:"Locale,omitempty"`
	TimeZone       string     `gorm:"type:VARBINARY(64);default:'';" json:"TimeZone" yaml:"TimeZone,omitempty"`
	PlaceID        string     `gorm:"type:VARBINARY(42);index;default:'zz'" json:"PlaceID,omitempty" yaml:"-"`
	PlaceSrc       string     `gorm:"type:VARBINARY(8);" json:"PlaceSrc,omitempty" yaml:"PlaceSrc,omitempty"`
	CellID         string     `gorm:"type:VARBINARY(42);index;default:'zz'" json:"CellID" yaml:"CellID"`
	SubjUID        string     `gorm:"type:VARBINARY(42);index;" json:"SubjUID" yaml:"SubjUID,omitempty"`
	UserBio        string     `gorm:"size:255;" json:"Bio,omitempty" yaml:"Bio,omitempty"`
	UserStatus     string     `gorm:"size:32;" json:"Status,omitempty" yaml:"Status,omitempty"`
	UserURL        string     `gorm:"size:255;column:user_url" json:"URL,omitempty" yaml:"URL,omitempty"`
	UserPhone      string     `gorm:"size:32;" json:"Phone,omitempty" yaml:"Phone,omitempty"`
	FullName       string     `gorm:"size:128;" json:"FullName" yaml:"FullName,omitempty"`
	DisplayName    string     `gorm:"size:64;" json:"DisplayName" yaml:"DisplayName,omitempty"`
	UserAlias      string     `gorm:"size:64;" json:"Alias" yaml:"Alias,omitempty"`
	ArtistName     string     `gorm:"size:64;" json:"ArtistName,omitempty" yaml:"ArtistName,omitempty"`
	UserArtist     bool       `gorm:"default:false;" json:"Artist,omitempty" yaml:"Artist,omitempty"`
	UserFavorite   bool       `gorm:"default:false;" json:"Favorite" yaml:"Favorite,omitempty"`
	UserHidden     bool       `gorm:"default:false;" json:"Hidden" yaml:"Hidden,omitempty"`
	UserPrivate    bool       `gorm:"default:false;" json:"Private" yaml:"Private,omitempty"`
	UserExcluded   bool       `gorm:"default:false;" json:"Excluded" yaml:"Excluded,omitempty"`
	CompanyName    string     `gorm:"size:128;" json:"CompanyName,omitempty" yaml:"CompanyName,omitempty"`
	DepartmentName string     `gorm:"size:128;" json:"DepartmentName,omitempty" yaml:"DepartmentName,omitempty"`
	JobTitle       string     `gorm:"size:64;" json:"JobTitle,omitempty" yaml:"JobTitle,omitempty"`
	BusinessURL    string     `gorm:"size:255" json:"BusinessURL,omitempty" yaml:"BusinessURL,omitempty"`
	BusinessPhone  string     `gorm:"size:32;" json:"BusinessPhone,omitempty" yaml:"BusinessPhone,omitempty"`
	BusinessEmail  string     `gorm:"size:255;" json:"BusinessEmail,omitempty" yaml:"BusinessEmail,omitempty"`
	BackupEmail    string     `gorm:"size:255;" json:"BackupEmail,omitempty" yaml:"BackupEmail,omitempty"`
	BirthYear      int        `gorm:"default:-1;" json:"BirthYear" yaml:"BirthYear,omitempty"`
	BirthMonth     int        `gorm:"default:-1;" json:"BirthMonth" yaml:"BirthMonth,omitempty"`
	BirthDay       int        `gorm:"default:-1;" json:"BirthDay" yaml:"BirthDay,omitempty"`
	FileRoot       string     `gorm:"type:VARBINARY(16);column:file_root;" json:"FileRoot,omitempty" yaml:"FileRoot,omitempty"`
	FilePath       string     `gorm:"type:VARBINARY(500);column:file_path;" json:"FilePath,omitempty" yaml:"FilePath,omitempty"`
	InviteToken    string     `gorm:"type:VARBINARY(32);" json:"-" yaml:"-"`
	InvitedBy      string     `gorm:"type:VARBINARY(32);" json:"-" yaml:"-"`
	DownloadToken  string     `gorm:"column:download_token;type:VARBINARY(128);" json:"-" yaml:"-"`
	PreviewToken   string     `gorm:"column:preview_token;type:VARBINARY(128);" json:"-" yaml:"-"`
	ResetToken     string     `gorm:"type:VARBINARY(64);" json:"-" yaml:"-"`
	ConfirmToken   string     `gorm:"type:VARBINARY(64);" json:"-" yaml:"-"`
	ConfirmedAt    *time.Time `json:"ConfirmedAt,omitempty" yaml:"ConfirmedAt,omitempty"`
	TermsAccepted  *time.Time `json:"TermsAccepted,omitempty" yaml:"TermsAccepted,omitempty"`
	LoginAttempts  int        `json:"-" yaml:"-"`
	LoginAt        *time.Time `json:"-" yaml:"-"`
	CreatedAt      time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt      time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt      *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity database table name.
func (User) TableName() string {
	return "auth_users"
}

// InitAccount sets the name and password of the initial admin account.
func (m *User) InitAccount(login, password string) (updated bool) {
	if !m.IsRegistered() {
		log.Warn("only registered users can change their password")
		return false
	}

	// Password must not be empty.
	if password == "" {
		return false
	}

	existing := FindPassword(m.UserUID)

	if existing != nil {
		return false
	}

	pw := NewPassword(m.UserUID, password)

	// Save password.
	if err := pw.Save(); err != nil {
		log.Error(err)
		return false
	}

	// Change username.
	if err := m.UpdateName(login); err != nil {
		log.Debugf("auth: cannot change username of %s to %s (%s)", clean.Log(m.UserUID), clean.LogQuote(login), err.Error())
	}

	return true
}

// Create new entity in the database.
func (m *User) Create() error {
	return Db().Create(m).Error
}

// Save entity properties.
func (m *User) Save() error {
	return Db().Save(m).Error
}

// Updates multiple properties in the database.
func (m *User) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *User) BeforeCreate(scope *gorm.Scope) error {
	m.UserSlug = m.GenerateSlug()

	if err := scope.SetColumn("UserSlug", m.UserSlug); err != nil {
		return err
	}

	if rnd.ValidID(m.UserUID, 'u') {
		return nil
	}

	m.UserUID = rnd.GenerateUID('u')

	return scope.SetColumn("UserUID", m.UserUID)
}

// GenerateSlug returns an updated slug.
func (m *User) GenerateSlug() string {
	if l := clean.Login(m.Username); l != "" {
		return txt.Slug(l)
	} else if m.UserSlug == "" {
		return rnd.GenerateToken(8)
	}

	return m.UserSlug
}

// FirstOrCreateUser returns an existing row, inserts a new row, or nil in case of errors.
func FirstOrCreateUser(m *User) *User {
	result := User{}

	m.UserSlug = m.GenerateSlug()

	if err := Db().Where("id = ? OR (user_uid = ? AND user_uid <> '') OR (user_slug = ? AND user_slug <> '') OR (username = ? AND username <> '')", m.ID, m.UserUID, m.UserSlug, m.Username).First(&result).Error; err == nil {
		return &result
	} else if err := m.Create(); err != nil {
		log.Debugf("user: %s", err)
		return nil
	}

	return m
}

// FindUserByLogin returns an existing user or nil if not found.
func FindUserByLogin(s string) *User {
	if s == "" {
		return nil
	}

	result := User{}

	// Find by Login.
	if name := clean.Login(s); name == "" {
		return nil
	} else if err := Db().Where("username = ?", name).First(&result).Error; err == nil {
		return &result
	} else {
		log.Debugf("username %s not found", clean.LogQuote(name))
	}

	// Find by Email.
	if email := clean.Email(s); email == "" {
		return nil
	} else if err := Db().Where("email = ?", email).First(&result).Error; err == nil {
		return &result
	} else {
		log.Debugf("email %s not found", clean.LogQuote(email))
	}

	return nil
}

// FindUserByUID returns an existing user or nil if not found.
func FindUserByUID(uid string) *User {
	if uid == "" {
		return nil
	}

	result := User{}

	if err := Db().Where("user_uid = ?", uid).First(&result).Error; err == nil {
		return &result
	} else {
		log.Debugf("user uid %s not found", clean.LogQuote(uid))
		return nil
	}
}

// Delete marks the entity as deleted.
func (m *User) Delete() error {
	if m.ID <= 1 {
		return fmt.Errorf("cannot delete system user")
	}

	return Db().Delete(m).Error
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
	if n := m.UserName(); n != "" {
		return clean.Log(n)
	} else if n = m.RealName(); n != "" {
		return clean.Log(n)
	} else if m.UserSlug != "" {
		return clean.Log(m.UserSlug)
	}

	return clean.Log(m.UserUID)
}

// UserName returns the login username.
func (m *User) UserName() string {
	return clean.Login(m.Username)
}

// UserEmail returns the login email address.
func (m *User) UserEmail() string {
	switch {
	case m.Email != "":
		return m.Email
	case m.BackupEmail != "":
		return m.BackupEmail
	case m.BusinessEmail != "":
		return m.BusinessEmail
	default:
		return ""
	}
}

// RealName returns the user' real name if known.
func (m *User) RealName() string {
	switch {
	case m.FullName != "":
		return m.FullName
	case m.DisplayName != "":
		return m.DisplayName
	case m.ArtistName != "":
		return m.ArtistName
	default:
		return ""
	}
}

// SetUsername sets the login username to the specified string.
func (m *User) SetUsername(login string) (err error) {
	login = clean.Login(login)

	// Empty?
	if login == "" {
		return fmt.Errorf("new username is empty")
	}

	// Update username and slug.
	m.Username = login
	m.UserSlug = m.GenerateSlug()

	// Update display name.
	if m.DisplayName == "" || m.DisplayName == AdminDisplayName {
		m.DisplayName = clean.Name(login)
	}

	return nil
}

// UpdateName changes the login username and saves it to the database.
func (m *User) UpdateName(login string) (err error) {
	if err = m.SetUsername(login); err != nil {
		return err
	}

	// Save to database.
	return m.Updates(Values{
		"UserSlug":    m.UserSlug,
		"Username":    m.Username,
		"DisplayName": m.DisplayName,
	})
}

// AclRole returns the user role for ACL permission checks.
func (m *User) AclRole() acl.Role {
	role := clean.Role(m.UserRole)

	switch {
	case m.SuperAdmin:
		return acl.RoleAdmin
	case role == "":
		return acl.RoleDefault
	case acl.RoleAdmin.Equal(role):
		return acl.RoleAdmin
	case acl.RoleEditor.Equal(role):
		return acl.RoleEditor
	case acl.RoleViewer.Equal(role):
		return acl.RoleViewer
	case acl.RoleGuest.Equal(role):
		return acl.RoleGuest
	default:
		return acl.Role(role)
	}
}

// IsRegistered checks if the user is registered e.g. has a username.
func (m *User) IsRegistered() bool {
	return m.UserName() != "" && rnd.EntityUID(m.UserUID, 'u')
}

// IsAdmin checks if the user is an admin with username.
func (m *User) IsAdmin() bool {
	return m.IsRegistered() && m.AclRole() == acl.RoleAdmin
}

// IsEditor checks if the user is an editor with username.
func (m *User) IsEditor() bool {
	return m.IsRegistered() && m.AclRole() == acl.RoleEditor
}

// IsViewer checks if the user is a viewer with username.
func (m *User) IsViewer() bool {
	return m.IsRegistered() && m.AclRole() == acl.RoleViewer
}

// IsGuest checks if the user is a guest.
func (m *User) IsGuest() bool {
	return m.AclRole() == acl.RoleGuest
}

// IsAnonymous checks if the user is unknown.
func (m *User) IsAnonymous() bool {
	return !rnd.EntityUID(m.UserUID, 'u') || m.ID == UnknownUser.ID || m.UserUID == UnknownUser.UserUID
}

// SetPassword sets a new password stored as hash.
func (m *User) SetPassword(password string) error {
	if !m.IsRegistered() {
		return fmt.Errorf("only registered users can change their password")
	}

	if len(password) < PasswordLength {
		return fmt.Errorf("password must have at least %d characters", PasswordLength)
	}

	pw := NewPassword(m.UserUID, password)

	return pw.Save()
}

// InvalidPassword returns true if the given password does not match the hash.
func (m *User) InvalidPassword(password string) bool {
	if !m.IsRegistered() {
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

// Validate Makes sure username and email are unique and meet requirements. Returns error if any property is invalid
func (m *User) Validate() error {
	if m.UserName() == "" {
		return errors.New("username must not be empty")
	}

	if len(m.UserName()) < UsernameLength {
		return fmt.Errorf("username must have at least %d characters", UsernameLength)
	}

	var err error
	var resultName = User{}

	if err = Db().Where("username = ? AND id <> ?", m.Username, m.ID).First(&resultName).Error; err == nil {
		return errors.New("username already exists")
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	// stop here if no email is provided
	if m.Email == "" {
		return nil
	}

	// validate email address.
	if a, err := mail.ParseAddress(m.Email); err != nil {
		return err
	} else {
		m.Email = a.Address // make sure email will be used without name.
	}

	var resultMail = User{}

	if err = Db().Where("email = ? AND id <> ?", m.Email, m.ID).First(&resultMail).Error; err == nil {
		return errors.New("email already exists")
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	return nil
}

// CreateWithPassword Creates User with Password in db transaction.
func CreateWithPassword(uc form.UserCreate) error {
	u := &User{
		Username:   uc.Username,
		Email:      uc.Email,
		SuperAdmin: true,
	}

	if len(uc.Password) < PasswordLength {
		return fmt.Errorf("password must have at least %d characters", PasswordLength)
	}

	if err := u.Validate(); err != nil {
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
		log.Infof("added user %s with uid %s", clean.Log(u.UserName()), clean.Log(u.UserUID))
		return nil
	})
}
