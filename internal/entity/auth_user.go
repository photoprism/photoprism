package entity

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// User identifier prefixes.
const (
	UserPrefix = "user"
	UserUID    = 'u'
)

// LenNameMin specifies the minimum length of the username in characters.
var LenNameMin = 3

// LenPasswordMin specifies the minimum length of the password in characters.
var LenPasswordMin = 4

// Users represents a list of users.
type Users []User

// User represents a person that may optionally log in as user.
type User struct {
	ID            int           `gorm:"primary_key" json:"-" yaml:"-"`
	UserUID       string        `gorm:"type:VARBINARY(64);column:user_uid;unique_index;" json:"UID" yaml:"UID"`
	UserUUID      string        `gorm:"type:VARBINARY(128);column:user_uuid;index;" json:"UUID,omitempty" yaml:"UUID,omitempty"`
	AuthProvider  string        `gorm:"type:VARBINARY(128);default:'';" json:"AuthProvider,omitempty" yaml:"AuthProvider,omitempty"`
	AuthID        string        `gorm:"type:VARBINARY(128);index;default:'';" json:"AuthID,omitempty" yaml:"AuthID,omitempty"`
	UserName      string        `gorm:"size:64;index;" json:"Name" yaml:"Name,omitempty"`
	DisplayName   string        `gorm:"size:200;" json:"DisplayName" yaml:"DisplayName,omitempty"`
	UserEmail     string        `gorm:"size:255;index;" json:"Email" yaml:"Email,omitempty"`
	BackupEmail   string        `gorm:"size:255;" json:"BackupEmail,omitempty" yaml:"BackupEmail,omitempty"`
	UserRole      string        `gorm:"size:64;default:'restricted';" json:"Role,omitempty" yaml:"Role,omitempty"`
	UserAttr      string        `gorm:"size:1024;" json:"Attr,omitempty" yaml:"Attr,omitempty"`
	SuperAdmin    bool          `json:"SuperAdmin,omitempty" yaml:"SuperAdmin,omitempty"`
	CanLogin      bool          `json:"CanLogin,omitempty" yaml:"CanLogin,omitempty"`
	LoginAt       *time.Time    `json:"LoginAt,omitempty" yaml:"LoginAt,omitempty"`
	CanSync       bool          `json:"CanSync,omitempty" yaml:"CanSync,omitempty"`
	CanInvite     bool          `json:"CanInvite,omitempty" yaml:"CanInvite,omitempty"`
	InviteToken   string        `gorm:"type:VARBINARY(64);index;" json:"-" yaml:"-"`
	InvitedBy     string        `gorm:"size:64;" json:"-" yaml:"-"`
	VerifyToken   string        `gorm:"type:VARBINARY(64);" json:"-" yaml:"-"`
	VerifiedAt    *time.Time    `json:"VerifiedAt,omitempty" yaml:"VerifiedAt,omitempty"`
	ConsentAt     *time.Time    `json:"ConsentAt,omitempty" yaml:"ConsentAt,omitempty"`
	BornAt        *time.Time    `sql:"index" json:"BornAt,omitempty" yaml:"BornAt,omitempty"`
	UserDetails   *UserDetails  `gorm:"PRELOAD:true;foreignkey:UserUID;association_foreignkey:UserUID;" json:"Details,omitempty" yaml:"Details,omitempty"`
	UserSettings  *UserSettings `gorm:"PRELOAD:true;foreignkey:UserUID;association_foreignkey:UserUID;" json:"Settings,omitempty" yaml:"Settings,omitempty"`
	ResetToken    string        `gorm:"type:VARBINARY(64);" json:"-" yaml:"-"`
	PreviewToken  string        `gorm:"type:VARBINARY(64);column:preview_token;" json:"-" yaml:"-"`
	DownloadToken string        `gorm:"type:VARBINARY(64);column:download_token;" json:"-" yaml:"-"`
	Thumb         string        `gorm:"type:VARBINARY(128);index;default:'';" json:"Thumb,omitempty" yaml:"Thumb,omitempty"`
	ThumbSrc      string        `gorm:"type:VARBINARY(8);default:'';" json:"ThumbSrc,omitempty" yaml:"ThumbSrc,omitempty"`
	RefID         string        `gorm:"type:VARBINARY(16);" json:"-" yaml:"-"`
	CreatedAt     time.Time     `json:"CreatedAt" yaml:"-"`
	UpdatedAt     time.Time     `json:"UpdatedAt" yaml:"-"`
	ExpiresAt     *time.Time    `sql:"index" json:"ExpiresAt,omitempty" yaml:"ExpiresAt,omitempty"`
	DeletedAt     *time.Time    `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity table name.
func (User) TableName() string {
	return "auth_users_dev"
}

// NewUser creates a new user and returns it.
func NewUser() (m *User) {
	uid := rnd.GenerateUID(UserUID)

	return &User{
		UserUID:      uid,
		UserDetails:  NewUserDetails(uid),
		UserSettings: NewUserSettings(uid),
		RefID:        rnd.RefID(UserPrefix),
	}
}

// FirstOrCreateUser returns an existing record, inserts a new record, or returns nil in case of an error.
func FirstOrCreateUser(m *User) *User {
	result := User{}

	if err := Db().Where("id = ? OR (user_uid = ? AND user_uid <> '') OR (user_name = ? AND user_name <> '')", m.ID, m.UserUID, m.UserName).First(&result).Error; err == nil {
		return &result
	} else if err = m.Create(); err != nil {
		event.AuditErr([]string{"user", "failed to create", "%s"}, err)
		return nil
	}

	return m
}

// FindUserByName finds a user by its username or returns nil if it was not found.
func FindUserByName(name string) *User {
	name = clean.Username(name)
	if name == "" {
		return nil
	}

	m := &User{}

	// Find matching record.
	if Db().First(m, "user_name = ?", name).RecordNotFound() {
		return nil
	}

	// Fetch related settings and details.
	return m.LoadRelated()
}

// FindUserByUID returns an existing user or nil if not found.
func FindUserByUID(uid string) *User {
	if uid == "" {
		return nil
	}

	m := &User{}

	// Find matching record.
	if UnscopedDb().First(m, "user_uid = ?", uid).RecordNotFound() {
		event.AuditWarn([]string{"user", "failed to find uid %s"}, clean.Log(uid))
		return nil
	}

	// Fetch related settings and details.
	return m.LoadRelated()
}

// UID returns the unique id as string.
func (m *User) UID() string {
	return m.UserUID
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
		event.AuditErr([]string{"user", "failed to change password for %s", "%s"}, clean.LogQuote(login), err)
		return false
	}

	// Change username.
	if err := m.UpdateName(login); err != nil {
		event.AuditErr([]string{"user", m.UserUID, "failed to change name to %s", "%s"}, clean.LogQuote(login), err)
	}

	return true
}

// Create new entity in the database.
func (m *User) Create() (err error) {
	err = Db().Create(m).Error

	if err == nil {
		m.SaveRelated()
	}

	return err
}

// Save entity properties.
func (m *User) Save() (err error) {
	err = Db().Save(m).Error

	if err == nil {
		m.SaveRelated()
	}

	return err
}

// Delete marks the entity as deleted.
func (m *User) Delete() error {
	if m.ID <= 1 {
		return fmt.Errorf("cannot delete system user")
	}

	return Db().Delete(m).Error
}

// Deleted checks if the user account has been deleted.
func (m *User) Deleted() bool {
	if m.DeletedAt == nil {
		return false
	}

	return !m.DeletedAt.IsZero()
}

// LoadRelated loads related settings and details.
func (m *User) LoadRelated() *User {
	m.Settings()
	m.Details()

	return m
}

// SaveRelated saves related settings and details.
func (m *User) SaveRelated() *User {
	if err := m.Settings().Save(); err != nil {
		event.AuditErr([]string{"user", m.UserUID, "failed to save settings", "%s"}, err)
	}
	if err := m.Details().Save(); err != nil {
		event.AuditErr([]string{"user", m.UserUID, "failed to save details", "%s"}, err)
	}

	return m
}

// Updates multiple properties in the database.
func (m *User) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// BeforeCreate sets a random UID if needed before inserting a new row to the database.
func (m *User) BeforeCreate(scope *gorm.Scope) error {
	if m.UserSettings != nil {
		m.UserSettings.UserUID = m.UserUID
	}

	if m.UserDetails != nil {
		m.UserDetails.UserUID = m.UserUID
	}

	if rnd.InvalidRefID(m.RefID) {
		m.RefID = rnd.RefID(UserPrefix)
		_ = scope.SetColumn("RefID", m.RefID)
	}

	if rnd.IsUnique(m.UserUID, UserUID) {
		return nil
	}

	m.UserUID = rnd.GenerateUID(UserUID)
	return scope.SetColumn("UserUID", m.UserUID)
}

// Expired checks if the user account has expired.
func (m *User) Expired() bool {
	if m.ExpiresAt == nil {
		return false
	}

	return m.ExpiresAt.Before(time.Now())
}

// Disabled checks if the user account has been deleted or has expired.
func (m *User) Disabled() bool {
	return m.Deleted() || m.Expired()
}

// LoginAllowed checks if logging in with the user account is possible.
func (m *User) LoginAllowed() bool {
	if role := m.AclRole(); m.Disabled() || !m.CanLogin || m.UserName == "" || role == acl.RoleUnauthorized {
		return false
	} else {
		return acl.Resources.Allow(acl.ResourceConfig, role, acl.AccessOwn)
	}

}

// SyncAllowed checks if file sync with the user account is possible.
func (m *User) SyncAllowed() bool {
	if role := m.AclRole(); m.Disabled() || !m.CanSync || m.UserName == "" || role == acl.RoleUnauthorized {
		return false
	} else {
		return acl.Resources.Allow(acl.ResourcePhotos, role, acl.ActionUpload)
	}
}

// String returns an identifier that can be used in logs.
func (m *User) String() string {
	if n := m.Name(); n != "" {
		return clean.Log(n)
	} else if n = m.FullName(); n != "" {
		return clean.Log(n)
	}

	return clean.Log(m.UserUID)
}

// Name returns the user's login name for authentication.
func (m *User) Name() string {
	return clean.Username(m.UserName)
}

// SetName sets the login username to the specified string.
func (m *User) SetName(login string) (err error) {
	login = clean.Username(login)

	// Empty?
	if login == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Update username and slug.
	m.UserName = login

	// Update display name.
	if m.DisplayName == "" || m.DisplayName == AdminDisplayName {
		m.DisplayName = clean.Name(login)
	}

	return nil
}

// UpdateName changes the login username and saves it to the database.
func (m *User) UpdateName(login string) (err error) {
	if err = m.SetName(login); err != nil {
		return err
	}

	// Save to database.
	return m.Updates(Values{
		"UserName":    m.UserName,
		"DisplayName": m.DisplayName,
	})
}

// Email returns the user's login email for authentication.
func (m *User) Email() string {
	return clean.Email(m.UserEmail)
}

// FullName returns the name of the user for display purposes.
func (m *User) FullName() string {
	switch {
	case m.DisplayName != "":
		return m.DisplayName
	default:
		return m.UserName
	}
}

// AclRole returns the user role for ACL permission checks.
func (m *User) AclRole() acl.Role {
	role := clean.Role(m.UserRole)

	switch {
	case m.SuperAdmin:
		return acl.RoleAdmin
	case role == "":
		return acl.RoleUnauthorized
	case m.UserName == "":
		return acl.RoleVisitor
	default:
		return acl.ValidRoles[role]
	}
}

// Settings returns the user settings and initializes them if necessary.
func (m *User) Settings() *UserSettings {
	if m.UserSettings != nil {
		m.UserSettings.UserUID = m.UserUID
		return m.UserSettings
	} else if m.UID() == "" {
		m.UserSettings = &UserSettings{}
		return m.UserSettings
	} else if err := CreateUserSettings(m); err != nil {
		m.UserSettings = NewUserSettings(m.UserUID)
	}

	return m.UserSettings
}

// Details returns user profile information and initializes it if needed.
func (m *User) Details() *UserDetails {
	if m.UserDetails != nil {
		m.UserDetails.UserUID = m.UserUID
		return m.UserDetails
	} else if m.UID() == "" {
		m.UserDetails = &UserDetails{}
		return m.UserDetails
	} else if err := CreateUserDetails(m); err != nil {
		m.UserDetails = NewUserDetails(m.UserUID)
	}

	return m.UserDetails
}

// Attr returns optional user account attributes as sanitized string.
// Example: https://learn.microsoft.com/en-us/troubleshoot/windows-server/identity/useraccountcontrol-manipulate-account-properties
func (m *User) Attr() string {
	return clean.Attr(m.UserAttr)
}

// IsRegistered checks if the user is registered e.g. has a username.
func (m *User) IsRegistered() bool {
	return m.UserName != "" && rnd.IsUID(m.UserUID, UserUID)
}

// IsAdmin checks if the user is an admin with username.
func (m *User) IsAdmin() bool {
	return m.IsRegistered() && m.AclRole() == acl.RoleAdmin
}

// IsVisitor checks if the user is a sharing link visitor.
func (m *User) IsVisitor() bool {
	return m.AclRole() == acl.RoleVisitor || m.ID == Visitor.ID
}

// IsUnknown checks if the user is unknown.
func (m *User) IsUnknown() bool {
	return !rnd.IsUID(m.UserUID, UserUID) || m.ID == UnknownUser.ID || m.UserUID == UnknownUser.UserUID
}

// DeleteSessions deletes all active user sessions except those passed as argument.
func (m *User) DeleteSessions(omit []string) (deleted int) {
	if m.UserUID == "" {
		return 0
	}

	// Find all user sessions except the session ids passed as argument.
	stmt := Db().Where("user_uid = ? AND id NOT IN (?)", m.UserUID, omit)
	sess := Sessions{}

	if err := stmt.Find(&sess).Error; err != nil {
		event.AuditErr([]string{"user", "failed to invalidate sessions", "%s"}, m.UserUID, err)
		return 0
	}

	// This will also remove the session from the cache.
	for _, s := range sess {
		if err := s.Delete(); err != nil {
			event.AuditWarn([]string{"user", "failed to invalidate session %s"}, m.UserUID, clean.Log(s.RefID))
		} else {
			deleted++
		}
	}

	// Return number of deleted sessions for logs.
	return deleted
}

// SetPassword sets a new password stored as hash.
func (m *User) SetPassword(password string) error {
	if !m.IsRegistered() {
		return fmt.Errorf("only registered users may change their password")
	}

	if len(password) < LenPasswordMin {
		return fmt.Errorf("password must have at least %d characters", LenPasswordMin)
	}

	pw := NewPassword(m.UserUID, password)

	return pw.Save()
}

// InvalidPassword returns true if the given password does not match the hash.
func (m *User) InvalidPassword(password string) bool {
	// Registered user?
	if !m.IsRegistered() {
		log.Warn("only registered users may change their password")
		return true
	}

	// Empty password?
	if password == "" {
		return true
	}

	// Fetch password.
	pw := FindPassword(m.UserUID)

	// Found?
	if pw == nil {
		return true
	}

	// Invalid?
	if pw.InvalidPassword(password) {
		return true
	}

	return false
}

// Validate checks if username, email and role are valid and returns an error otherwise.
func (m *User) Validate() (err error) {
	// Empty name?
	if m.Name() == "" {
		return errors.New("username cannot be empty")
	}

	// Name too short?
	if len(m.Name()) < LenNameMin {
		return fmt.Errorf("username must have at least %d characters", LenNameMin)
	}

	// Validate user role.
	if acl.ValidRoles[m.UserRole] == "" {
		return fmt.Errorf("role %s is invalid", clean.LogQuote(m.UserRole))
	}

	// Check if the username is unique.
	var duplicate = User{}

	if err = Db().
		Where("user_name = ? AND id <> ?", m.UserName, m.ID).
		First(&duplicate).Error; err == nil {
		return fmt.Errorf("username %s already exists", clean.LogQuote(m.UserName))
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	// Skip email check?
	if m.UserEmail == "" {
		return nil
	}

	// Parse and validate email address.
	if a, err := mail.ParseAddress(m.UserEmail); err != nil {
		return fmt.Errorf("email %s is invalid", clean.LogQuote(m.UserEmail))
	} else if email := a.Address; !strings.ContainsRune(email, '.') {
		return fmt.Errorf("email %s does not have a fully qualified domain", clean.LogQuote(m.UserEmail))
	} else {
		m.UserEmail = email
	}

	// Check if the email is unique.
	if err = Db().
		Where("user_email = ? AND id <> ?", m.UserEmail, m.ID).
		First(&duplicate).Error; err == nil {
		return fmt.Errorf("email %s already exists", clean.Log(m.UserEmail))
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	return nil
}

// SetFormValues sets the values specified in the form.
func (m *User) SetFormValues(frm form.User) *User {
	m.UserName = frm.Name()
	m.UserEmail = frm.Email()
	m.DisplayName = frm.DisplayName
	m.SuperAdmin = frm.SuperAdmin
	m.CanLogin = frm.CanLogin
	m.CanSync = frm.CanSync
	m.UserRole = frm.Role()
	m.UserAttr = frm.Attr()

	return m
}
