package entity

import (
	"errors"
	"fmt"
	"net/mail"
	"path"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/ulule/deepcopier"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// User identifier prefixes.
const (
	UserUID               = byte('u')
	UserPrefix            = "user"
	OwnerUnknown          = ""
	UsernameLengthDefault = 1
	PasswordLengthDefault = 8
)

// UsernameLength specifies the minimum length of the username in characters.
var UsernameLength = UsernameLengthDefault

// PasswordLength specifies the minimum length of a password in characters (runes, not bytes).
var PasswordLength = PasswordLengthDefault

// UsersPath is the relative path for user assets.
var UsersPath = "users"

// Users represents a list of users.
type Users []User

// User represents a person that may optionally log in as user.
type User struct {
	ID            int           `gorm:"primary_key" json:"ID" yaml:"-"`
	UUID          string        `gorm:"type:VARBINARY(64);column:user_uuid;index;" json:"UUID,omitempty" yaml:"UUID,omitempty"`
	UserUID       string        `gorm:"type:VARBINARY(42);column:user_uid;unique_index;" json:"UID" yaml:"UID"`
	AuthProvider  string        `gorm:"type:VARBINARY(128);default:'';" json:"AuthProvider" yaml:"AuthProvider,omitempty"`
	AuthMethod    string        `gorm:"type:VARBINARY(128);default:'';" json:"AuthMethod" yaml:"AuthMethod,omitempty"`
	AuthIssuer    string        `gorm:"type:VARBINARY(255);default:'';" json:"AuthIssuer,omitempty" yaml:"AuthIssuer,omitempty"`
	AuthID        string        `gorm:"type:VARBINARY(255);index;default:'';" json:"AuthID" yaml:"AuthID,omitempty"`
	UserName      string        `gorm:"size:200;index;" json:"Name" yaml:"Name,omitempty"`
	DisplayName   string        `gorm:"size:200;" json:"DisplayName" yaml:"DisplayName,omitempty"`
	UserEmail     string        `gorm:"size:255;index;" json:"Email" yaml:"Email,omitempty"`
	BackupEmail   string        `gorm:"size:255;" json:"BackupEmail,omitempty" yaml:"BackupEmail,omitempty"`
	UserRole      string        `gorm:"size:64;default:'';" json:"Role" yaml:"Role,omitempty"`
	UserAttr      string        `gorm:"size:1024;" json:"Attr" yaml:"Attr,omitempty"`
	SuperAdmin    bool          `json:"SuperAdmin" yaml:"SuperAdmin,omitempty"`
	CanLogin      bool          `json:"CanLogin" yaml:"CanLogin,omitempty"`
	LoginAt       *time.Time    `json:"LoginAt" yaml:"LoginAt,omitempty"`
	ExpiresAt     *time.Time    `sql:"index" json:"ExpiresAt,omitempty" yaml:"ExpiresAt,omitempty"`
	WebDAV        bool          `gorm:"column:webdav;" json:"WebDAV" yaml:"WebDAV,omitempty"`
	BasePath      string        `gorm:"type:VARBINARY(1024);" json:"BasePath" yaml:"BasePath,omitempty"`
	UploadPath    string        `gorm:"type:VARBINARY(1024);" json:"UploadPath" yaml:"UploadPath,omitempty"`
	CanInvite     bool          `json:"CanInvite" yaml:"CanInvite,omitempty"`
	InviteToken   string        `gorm:"type:VARBINARY(64);index;" json:"-" yaml:"-"`
	InvitedBy     string        `gorm:"size:64;" json:"-" yaml:"-"`
	VerifyToken   string        `gorm:"type:VARBINARY(64);" json:"-" yaml:"-"`
	VerifiedAt    *time.Time    `json:"VerifiedAt,omitempty" yaml:"VerifiedAt,omitempty"`
	ConsentAt     *time.Time    `json:"ConsentAt,omitempty" yaml:"ConsentAt,omitempty"`
	BornAt        *time.Time    `sql:"index" json:"BornAt,omitempty" yaml:"BornAt,omitempty"`
	UserDetails   *UserDetails  `gorm:"PRELOAD:true;foreignkey:UserUID;association_foreignkey:UserUID;" json:"Details,omitempty" yaml:"Details,omitempty"`
	UserSettings  *UserSettings `gorm:"PRELOAD:true;foreignkey:UserUID;association_foreignkey:UserUID;" json:"Settings,omitempty" yaml:"Settings,omitempty"`
	UserShares    UserShares    `gorm:"-" json:"Shares,omitempty" yaml:"Shares,omitempty"`
	ResetToken    string        `gorm:"type:VARBINARY(64);" json:"-" yaml:"-"`
	PreviewToken  string        `gorm:"type:VARBINARY(64);column:preview_token;" json:"-" yaml:"-"`
	DownloadToken string        `gorm:"type:VARBINARY(64);column:download_token;" json:"-" yaml:"-"`
	Thumb         string        `gorm:"type:VARBINARY(128);index;default:'';" json:"Thumb" yaml:"Thumb,omitempty"`
	ThumbSrc      string        `gorm:"type:VARBINARY(8);default:'';" json:"ThumbSrc" yaml:"ThumbSrc,omitempty"`
	RefID         string        `gorm:"type:VARBINARY(16);" json:"-" yaml:"-"`
	CreatedAt     time.Time     `json:"CreatedAt" yaml:"-"`
	UpdatedAt     time.Time     `json:"UpdatedAt" yaml:"-"`
	DeletedAt     *time.Time    `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity table name.
func (User) TableName() string {
	return "auth_users"
}

// NewUser creates a new user entity with defaults.
func NewUser() (m *User) {
	uid := rnd.GenerateUID(UserUID)

	return &User{
		UserUID:       uid,
		UserDetails:   NewUserDetails(uid),
		UserSettings:  NewUserSettings(uid),
		PreviewToken:  GenerateToken(),
		DownloadToken: GenerateToken(),
		RefID:         rnd.RefID(UserPrefix),
	}
}

// OidcUser creates a new OIDC user entity.
func OidcUser(userInfo *oidc.UserInfo, issuer, userName string) User {
	authId := clean.Auth(userInfo.Subject)

	if authId == "" {
		return User{}
	}

	return User{
		UserName:     userName,
		DisplayName:  userInfo.Name,
		UserEmail:    clean.Email(userInfo.Email),
		AuthProvider: authn.ProviderOIDC.String(),
		AuthIssuer:   clean.Uri(issuer),
		AuthID:       authId,
	}
}

// LdapUser creates a new LDAP user entity.
func LdapUser(username, dn string) User {
	return User{
		UserName:     clean.Username(username),
		AuthID:       dn,
		AuthProvider: authn.ProviderLDAP.String(),
	}
}

// FindUser returns the matching user or nil if it was not found.
func FindUser(find User) *User {
	m := &User{}

	// Build query.
	stmt := UnscopedDb()
	if find.ID != 0 && find.UserName != "" {
		stmt = stmt.Where("id = ? OR user_name = ?", find.ID, find.UserName)
	} else if find.ID != 0 {
		stmt = stmt.Where("id = ?", find.ID)
	} else if find.HasUID() {
		stmt = stmt.Where("user_uid = ?", find.UserUID)
	} else if authn.ProviderOIDC.Equal(find.AuthProvider) && find.AuthID != "" {
		if find.AuthIssuer == "" {
			stmt = stmt.Where("auth_provider = ? AND auth_id = ?", find.AuthProvider, find.AuthID)
		} else {
			stmt = stmt.Where("auth_provider = ? AND (auth_issuer = '' OR auth_issuer = ?) AND auth_id = ?", find.AuthProvider, find.AuthIssuer, find.AuthID)
		}
	} else if find.AuthProvider != "" && find.AuthID != "" && find.UserName != "" {
		stmt = stmt.Where("auth_provider = ? AND auth_id = ? OR user_name = ?", find.AuthProvider, find.AuthID, find.UserName)
	} else if find.UserName != "" {
		stmt = stmt.Where("user_name = ?", find.UserName)
	} else if find.UserEmail != "" {
		stmt = stmt.Where("user_email = ?", find.UserEmail)
	} else if find.AuthProvider != "" && find.AuthID != "" {
		stmt = stmt.Where("auth_provider = ? AND auth_id = ?", find.AuthProvider, find.AuthID)
	} else {
		return nil
	}

	// Find matching record.
	if err := stmt.First(m).Error; err != nil {
		return nil
	}

	// Fetch related records.
	return m.LoadRelated()
}

// FirstOrCreateUser returns an existing record, inserts a new record, or returns nil in case of an error.
func FirstOrCreateUser(m *User) *User {
	if m == nil {
		return nil
	}

	if found := FindUser(*m); found != nil {
		return found
	} else if err := m.Create(); err != nil {
		event.AuditErr([]string{"user", "failed to create", "%s"}, err)
		return nil
	} else {
		return m
	}
}

// FindUserByName returns the matching user or nil if it was not found.
func FindUserByName(userName string) *User {
	userName = clean.Username(userName)

	if userName == "" {
		return nil
	}

	return FindUser(User{UserName: userName})
}

// FindLocalUser returns the matching local user or nil if it was not found.
func FindLocalUser(userName string) *User {
	name := clean.Username(userName)

	if name == "" {
		return nil
	}

	m := &User{}
	providers := authn.LocalProviders

	// Build query.
	if err := UnscopedDb().
		Where("user_name = ? AND auth_provider IN (?)", name, providers).
		First(m).Error; err != nil {
		return nil
	}

	// Return with related records.
	return m.LoadRelated()
}

// FindUserByUID returns the matching user or nil if it was not found.
func FindUserByUID(uid string) *User {
	if rnd.InvalidUID(uid, UserUID) {
		return nil
	}

	return FindUser(User{UserUID: uid})
}

// GetUID returns the unique id as string.
func (m *User) GetUID() string {
	if m == nil {
		return ""
	}

	return m.UserUID
}

// HasUID tests if the user has a valid uid.
func (m *User) HasUID() bool {
	return rnd.IsUID(m.UserUID, UserUID)
}

// InvalidUID tests if the user does not have a valid uid.
func (m *User) InvalidUID() bool {
	return !m.HasUID()
}

// SameUID checks if the given uid matches the own uid.
func (m *User) SameUID(uid string) bool {
	if m == nil {
		return false
	} else if m.UserUID == "" || rnd.InvalidUID(uid, UserUID) {
		return false
	}

	return m.UserUID == uid
}

// InitAccount sets the name and password of the initial admin account.
func (m *User) InitAccount(initName, initPasswd string) (updated bool) {
	// User must exist and the password must not be empty.
	initPasswd = strings.TrimSpace(initPasswd)
	if rnd.InvalidUID(m.UserUID, UserUID) || initPasswd == "" {
		return false
	} else if !m.CanLogIn() {
		log.Warnf("users: %s account is not allowed to log in", m.String())
	}

	// Abort if user has a password.
	existingPasswd := FindPassword(m.UserUID)

	if existingPasswd != nil {
		return false
	}

	// Set initial password.
	initialPasswd := NewPassword(m.UserUID, initPasswd, true)

	// Save password.
	if err := initialPasswd.Save(); err != nil {
		event.AuditErr([]string{"user %s", "failed to change password", "%s"}, m.RefID, err)
		return false
	}

	// Change username if needed.
	if initName != "" && initName != m.UserName {
		if err := m.UpdateUsername(initName); err != nil {
			event.AuditErr([]string{"user %s", "failed to change username to %s", "%s"}, m.RefID, clean.Log(initName), err)
		}
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

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *User) Save() (err error) {
	m.GenerateTokens(false)

	err = UnscopedDb().Save(m).Error

	if err == nil {
		m.SaveRelated()
	}

	return err
}

// Delete marks the entity as deleted.
func (m *User) Delete() (err error) {
	if m.ID <= 1 {
		return fmt.Errorf("cannot delete system user")
	} else if m.UserUID == "" {
		return fmt.Errorf("uid is required to delete user")
	}

	if err = UnscopedDb().Delete(Session{}, "user_uid = ?", m.UserUID).Error; err != nil {
		event.AuditErr([]string{"user %s", "delete", "failed to remove sessions", "%s"}, m.RefID, err)
	}

	err = Db().Delete(m).Error

	FlushSessionCache()

	return err
}

// IsDeleted checks if the user account has been deleted.
func (m *User) IsDeleted() bool {
	if m.DeletedAt == nil {
		return false
	}

	return !m.DeletedAt.IsZero()
}

// LoadRelated loads related settings and details.
func (m *User) LoadRelated() *User {
	m.Details()
	m.Settings()
	m.RefreshShares()

	return m
}

// SaveRelated saves related settings and details.
func (m *User) SaveRelated() *User {
	if err := m.Settings().Save(); err != nil {
		event.AuditErr([]string{"user %s", "failed to save settings", "%s"}, m.RefID, err)
	}
	if err := m.Details().Save(); err != nil {
		event.AuditErr([]string{"user %s", "failed to save details", "%s"}, m.RefID, err)
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

	m.GenerateTokens(false)

	if rnd.InvalidRefID(m.RefID) {
		m.RefID = rnd.RefID(UserPrefix)
		Log("user", "set ref id", scope.SetColumn("RefID", m.RefID))
	}

	if rnd.IsUnique(m.UserUID, UserUID) {
		return nil
	}

	m.UserUID = rnd.GenerateUID(UserUID)
	return scope.SetColumn("UserUID", m.UserUID)
}

// IsExpired checks if the user account has expired.
func (m *User) IsExpired() bool {
	if m.ExpiresAt == nil {
		return false
	}

	return m.ExpiresAt.Before(time.Now())
}

// IsDisabled checks if the user account has been deleted or has expired.
func (m *User) IsDisabled() bool {
	if m == nil {
		return true
	}

	return m.IsDeleted() || m.IsExpired() && !m.SuperAdmin
}

// UpdateLoginTime updates the login timestamp and returns it if successful.
func (m *User) UpdateLoginTime() *time.Time {
	if m == nil {
		return nil
	} else if m.IsDeleted() {
		return nil
	}

	timeStamp := TimeStamp()

	if err := Db().Model(m).UpdateColumn("LoginAt", timeStamp).Error; err != nil {
		return nil
	}

	m.LoginAt = timeStamp

	return timeStamp
}

// CanLogIn checks if the user is allowed to log in and use the web UI.
func (m *User) CanLogIn() bool {
	if m == nil {
		return false
	} else if m.IsDeleted() || m.HasProvider(authn.ProviderNone) {
		return false
	} else if !m.CanLogin && !m.SuperAdmin || m.ID <= 0 || m.UserName == "" {
		return false
	} else if m.IsDisabled() || m.IsUnknown() || !m.IsRegistered() {
		return false
	} else {
		return acl.Rules.Allow(acl.ResourceConfig, m.AclRole(), acl.AccessOwn)
	}
}

// CanUseWebDAV checks whether the user is allowed to use WebDAV to synchronize files.
func (m *User) CanUseWebDAV() bool {
	if m == nil {
		// Abort check if user is nil for any reason.
		return false
	} else if !m.WebDAV || m.ID <= 0 || m.IsDisabled() || m.IsUnknown() || !m.IsRegistered() || m.HasProvider(authn.ProviderNone) {
		// Deny WebDAV access if WebDAV is disabled, the user does not have a
		// regular, registered account, or the account has been deactivated.
		return false
	} else {
		// Check if the ACL allows downloading files via WebDAV based on the user role.
		return acl.Rules.Allow(acl.ResourceWebDAV, m.AclRole(), acl.ActionDownload)
	}
}

// CanUpload checks if the user is allowed to upload files.
func (m *User) CanUpload() bool {
	if m == nil {
		// Abort check if user is nil for any reason.
		return false
	} else if m.IsDisabled() || m.HasProvider(authn.ProviderNone) || m.IsUnknown() {
		// Deny uploading if the user is unknown or the account has been deactivated.
		return false
	} else {
		// Check if the ACL allows uploading photos based on the user role.
		return acl.Rules.Allow(acl.ResourcePhotos, m.AclRole(), acl.ActionUpload)
	}
}

// DefaultBasePath returns the default base path of the user based on the user name.
func (m *User) DefaultBasePath() string {
	if s := m.Handle(); s == "" {
		return ""
	} else {
		return path.Join(UsersPath, s)
	}
}

// GetBasePath returns the user's relative base path.
func (m *User) GetBasePath() string {
	if m.BasePath == "" && m.HasRole("contributor") {
		m.BasePath = m.DefaultBasePath()
	}

	return m.BasePath
}

// SetBasePath changes the user's relative base path.
func (m *User) SetBasePath(dir string) *User {
	if list.Contains(list.List{"", ".", "./", "/", "\\"}, dir) {
		m.BasePath = ""
	} else if dir == "~" && m.UserName != "" {
		m.BasePath = m.DefaultBasePath()
	} else {
		m.BasePath = clean.UserPath(dir)
	}

	return m
}

// GetUploadPath returns the user's relative upload path.
func (m *User) GetUploadPath() string {
	basePath := m.GetBasePath()

	if list.Contains(list.List{"", ".", "./"}, m.UploadPath) {
		return basePath
	} else if basePath != "" && strings.HasPrefix(m.UploadPath, basePath+"/") {
		return m.UploadPath
	} else if basePath == "" && m.UploadPath == "~" && m.UserName != "" {
		return m.DefaultBasePath()
	}

	return path.Join(basePath, m.UploadPath)
}

// SetUploadPath changes the user's relative upload path.
func (m *User) SetUploadPath(dir string) *User {
	basePath := m.GetBasePath()

	if list.Contains(list.List{"", ".", "./", "/", "\\"}, dir) {
		m.UploadPath = ""
	} else if basePath == "" && dir == "~" && m.UserName != "" {
		m.UploadPath = m.DefaultBasePath()
	} else {
		m.UploadPath = clean.UserPath(dir)
	}

	return m
}

// String returns an identifier that can be used in logs.
func (m *User) String() string {
	if m == nil {
		return "User<nil>"
	}

	if n := m.Username(); n != "" {
		return clean.LogQuote(n)
	} else if n = m.FullName(); n != "" {
		return clean.LogQuote(n)
	}

	return clean.Log(m.UserUID)
}

// Provider returns the authentication provider name.
func (m *User) Provider() authn.ProviderType {
	if m.AuthProvider != "" {
		return authn.ProviderType(m.AuthProvider)
	} else if m.ID == Visitor.ID {
		return authn.ProviderLink
	} else if m.ID == 1 {
		return authn.ProviderLocal
	} else if m.UserName != "" && m.ID > 0 {
		return authn.ProviderDefault
	}

	return authn.ProviderNone
}

// HasProvider checks if the user has the given auth provider.
func (m *User) HasProvider(t authn.ProviderType) bool {
	return m.Provider().Is(t)
}

// SetProvider set the authentication provider.
func (m *User) SetProvider(t authn.ProviderType) *User {
	if m == nil {
		return nil
	} else if t == "" {
		return m
	}

	m.AuthProvider = t.String()

	if !m.Provider().SupportsPasscodeAuthentication() && m.Method().Is(authn.Method2FA) {
		m.AuthMethod = ""
	}

	return m
}

// Method returns the authentication method.
func (m *User) Method() authn.MethodType {
	return authn.Method(m.AuthMethod)
}

// SetMethod sets a custom authentication method.
func (m *User) SetMethod(method authn.MethodType) *User {
	if m == nil {
		return nil
	} else if method == "" {
		return m
	}

	// It must not be possible to activate 2FA if the authentication provider does not support passcodes.
	if !m.Provider().SupportsPasscodeAuthentication() && method.Is(authn.Method2FA) {
		return m
	}

	m.AuthMethod = method.String()

	return m
}

// SetAuthID sets a custom authentication identifier.
func (m *User) SetAuthID(id, issuer string) *User {
	// Update auth id if not empty.
	if authId := clean.Auth(id); authId == "" {
		return m
	} else {
		m.AuthID = authId
		m.AuthIssuer = clean.Uri(issuer)
	}

	// Make sure other users do not use the same identifier.
	if m.HasUID() && m.AuthProvider != "" {
		if err := UnscopedDb().Model(&User{}).
			Where("user_uid <> ? AND auth_provider = ? AND auth_id = ? AND super_admin = 0", m.UserUID, m.AuthProvider, m.AuthID).
			Updates(map[string]interface{}{"auth_id": "", "auth_provider": authn.ProviderNone}).Error; err != nil {
			event.AuditErr([]string{"user %s", "failed to resolve auth id conflicts", "%s"}, m.RefID, err)
		}
	}

	return m
}

// UpdateAuthID updated the custom authentication identifier.
func (m *User) UpdateAuthID(id, issuer string) error {
	if m.InvalidUID() {
		return errors.New("invalid user uid")
	}

	// Update auth id and issuer record.
	return m.SetAuthID(id, issuer).Updates(Map{
		"AuthID":     m.AuthID,
		"AuthIssuer": m.AuthIssuer,
	})
}

// AuthInfo returns information about the authentication type.
func (m *User) AuthInfo() string {
	provider := m.Provider()
	method := m.Method()

	if method.IsDefault() {
		return provider.Pretty()
	}

	return fmt.Sprintf("%s (%s)", provider.Pretty(), method.Pretty())
}

// Passcode returns the related auth key by type, if one exists.
func (m *User) Passcode(t authn.KeyType) *Passcode {
	if m == nil {
		return nil
	} else if m.UserUID == "" {
		return nil
	}

	return FindPasscode(Passcode{UID: m.GetUID(), KeyType: t.String()})
}

// Username returns the user's login name as sanitized string.
func (m *User) Username() string {
	return clean.Username(m.UserName)
}

// SetUsername sets the login username to the specified string.
func (m *User) SetUsername(login string) (err error) {
	if m.ID < 0 {
		return fmt.Errorf("system users cannot be modified")
	}

	login = clean.Username(login)

	// Empty?
	if login == "" {
		return fmt.Errorf("username is empty")
	} else if m.UserName == login {
		return nil
	} else if m.UserName != "" && m.ID != 1 {
		return fmt.Errorf("username cannot be changed")
	}

	// Update username and slug.
	m.UserName = login

	// Update display name.
	if m.DisplayName == "" || m.DisplayName == AdminDisplayName && m.ID == 1 {
		m.DisplayName = m.FullName()
	}

	return nil
}

// UpdateUsername changes the login name in the database.
func (m *User) UpdateUsername(login string) (err error) {
	// Check if the name already exists or has not changed.
	if m.UserName == login || m.ID <= 0 {
		return nil
	} else if user := FindUserByName(login); user != nil {
		return fmt.Errorf("user %s already exists", clean.LogQuote(login))
	}

	// Set new username.
	if err = m.SetUsername(login); err != nil {
		return err
	}

	// Save to database.
	return m.Updates(Map{
		"UserName":    m.UserName,
		"DisplayName": m.DisplayName,
	})
}

// Email returns the user's login email for authentication.
func (m *User) Email() string {
	return clean.Email(m.UserEmail)
}

// Handle returns the user's login handle.
func (m *User) Handle() string {
	return clean.Handle(m.UserName)
}

// FullName returns the name of the user for display purposes.
func (m *User) FullName() string {
	if m.DisplayName != "" {
		return m.DisplayName
	}

	if n := m.Details().DisplayName(); n != "" {
		return n
	}

	return clean.NameCapitalized(strings.ReplaceAll(m.Handle(), ".", " "))
}

// SetRole sets the user role specified as string.
func (m *User) SetRole(role string) *User {
	role = clean.Role(role)

	switch role {
	case "", "0", "false", "nil", "null", "nan":
		m.UserRole = acl.RoleNone.String()
	default:
		m.UserRole = acl.UserRoles[role].String()
	}

	return m
}

// HasRole checks the user role specified as string.
func (m *User) HasRole(role acl.Role) bool {
	return m.AclRole() == role
}

// AclRole returns the user role for ACL permission checks.
func (m *User) AclRole() acl.Role {
	role := clean.Role(m.UserRole)

	switch {
	case m.SuperAdmin:
		return acl.RoleAdmin
	case role == "":
		return acl.RoleNone
	case m.UserName == "":
		return acl.RoleVisitor
	default:
		return acl.UserRoles[role]
	}
}

// Settings returns the user settings and initializes them if necessary.
func (m *User) Settings() *UserSettings {
	if m.UserSettings != nil {
		m.UserSettings.UserUID = m.UserUID
		return m.UserSettings
	} else if m.GetUID() == "" {
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
	} else if m.GetUID() == "" {
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

// IsRegistered checks if this user has a registered account with a valid ID, username, and role.
func (m *User) IsRegistered() bool {
	if m == nil {
		return false
	}

	// Registered users must have an ID, a UID, a username and a known role, except visitor.
	return m.ID > 0 && m.UserName != "" && m.HasUID() && !m.IsVisitor()
}

// NotRegistered checks if the user is not registered with an own account.
func (m *User) NotRegistered() bool {
	return !m.IsRegistered()
}

// Equal returns true if the user specified matches.
func (m *User) Equal(u *User) bool {
	if m == nil || u == nil {
		return false
	}

	return m.UserUID == u.UserUID
}

// IsAdmin checks if the user is an admin with username.
func (m *User) IsAdmin() bool {
	if m == nil {
		return false
	}

	return m.IsSuperAdmin() || m.IsRegistered() && m.AclRole() == acl.RoleAdmin
}

// IsSuperAdmin checks if the user is a super admin.
func (m *User) IsSuperAdmin() bool {
	if m == nil {
		return false
	}

	return m.SuperAdmin
}

// IsVisitor checks if the user is a sharing link visitor.
func (m *User) IsVisitor() bool {
	return m.AclRole() == acl.RoleVisitor || m.ID == Visitor.ID
}

// HasSharedAccessOnly checks if the user as only access to shared resources.
func (m *User) HasSharedAccessOnly(resource acl.Resource) bool {
	if acl.Rules.Deny(resource, m.AclRole(), acl.AccessShared) {
		return false
	}

	return acl.Rules.DenyAll(resource, m.AclRole(), acl.Permissions{acl.AccessAll, acl.AccessLibrary})
}

// IsUnknown checks if the user is unknown.
func (m *User) IsUnknown() bool {
	if m == nil {
		return true
	}

	return m.InvalidUID() || m.ID == UnknownUser.ID || m.UserUID == UnknownUser.UserUID || m.HasRole(acl.RoleNone)
}

// DeleteSessions deletes all active user sessions except those passed as argument.
func (m *User) DeleteSessions(omit []string) (deleted int) {
	if m.UserUID == "" {
		return 0
	}

	// Compose update statement.
	stmt := Db()

	// Find all user sessions except the session ids passed as argument.
	if len(omit) == 0 {
		stmt = stmt.Where("user_uid = ?", m.UserUID)
	} else {
		stmt = stmt.Where("user_uid = ? AND id NOT IN (?)", m.UserUID, omit)
	}

	// Exclude client access tokens.
	stmt = stmt.Where("auth_provider NOT IN (?)", authn.ClientProviders)

	// Fetch sessions from database.
	sess := Sessions{}

	if err := stmt.Find(&sess).Error; err != nil {
		event.AuditErr([]string{"user %s", "failed to invalidate sessions", "%s"}, m.RefID, err)
		return 0
	}

	// Delete sessions from cache and database.
	for _, s := range sess {
		if err := s.Delete(); err != nil {
			event.AuditWarn([]string{"user %s", "failed to invalidate session %s", "%s"}, m.RefID, clean.Log(s.RefID), err)
		} else {
			deleted++
		}
	}

	// Return number of deleted sessions.
	return deleted
}

// SetPassword sets a new password stored as hash.
func (m *User) SetPassword(password string) error {
	if !m.IsRegistered() {
		return fmt.Errorf("only registered users can change their password")
	}

	if len([]rune(password)) < PasswordLength {
		return fmt.Errorf("password must have at least %d characters", PasswordLength)
	} else if len(password) > txt.ClipPassword {
		return fmt.Errorf("password must have less than %d characters", txt.ClipPassword)
	}

	pw := NewPassword(m.UserUID, password, false)

	if err := pw.Save(); err != nil {
		return err
	}

	return m.RegenerateTokens()
}

// DeletePassword removes the password of the user account, if one has been set.
func (m *User) DeletePassword() (err error) {
	// Make sure the user has a valid UID.
	if m.InvalidUID() {
		return fmt.Errorf("invalid user uid")
	}

	// Find local account password.
	pw := FindPassword(m.UserUID)

	// Return without error if no password has been set.
	if pw == nil {
		return nil
	}

	// Remove local account password.
	if err = pw.Delete(); err != nil {
		event.AuditErr([]string{"user %s", "failed to remove password", "%s"}, m.RefID, err)
	} else {
		event.AuditWarn([]string{"user %s", "password has been removed"}, m.RefID)
	}

	return err
}

// InvalidPassword returns true if the specified password is invalid or the account is not registered.
func (m *User) InvalidPassword(s string) bool {
	// Fail if not a registered user account.
	if !m.IsRegistered() {
		log.Warn("only registered users can log in")
		return true
	}

	// Fail if no password has been provided.
	if s == "" {
		return true
	}

	// Find local account password.
	pw := FindPassword(m.UserUID)

	// Fail if no password has been set.
	if pw == nil {
		return true
	}

	// Fail if the specified password is invalid.
	if pw.Invalid(s) {
		return true
	}

	// Password seems to be valid!
	return false
}

// VerifyPassword checks if the user has the specified password and the account is registered.
func (m *User) VerifyPassword(s string) bool {
	return !m.InvalidPassword(s)
}

// InvalidPasscode returns true if the specified passcode cannot be verified or is invalid.
func (m *User) InvalidPasscode(code string) bool {
	if m == nil {
		return true
	}

	valid, _, _ := m.VerifyPasscode(code)

	return !valid
}

// VerifyPasscode checks if the specified passcode could be verified and is valid.
func (m *User) VerifyPasscode(code string) (valid bool, passcode *Passcode, err error) {
	var recovery bool

	if m == nil {
		err = errors.New("user is nil")
	} else if passcode = m.Passcode(authn.KeyTOTP); passcode == nil {
		err = authn.ErrPasscodeNotSetUp
	} else if code == "" {
		err = authn.ErrPasscodeRequired
	} else if l := len(code); l < 1 || l > 255 {
		err = authn.ErrInvalidPasscode
	} else if valid, recovery, err = passcode.Valid(code); recovery {
		// Deactivate 2FA if recovery code has been used.
		passcode, err = m.DeactivatePasscode()
	}

	return valid, passcode, err
}

// ActivatePasscode activates two-factor authentication with a passcode.
func (m *User) ActivatePasscode() (passcode *Passcode, err error) {
	if m == nil {
		err = errors.New("user is nil")
	} else if !m.Provider().SupportsPasscodeAuthentication() {
		err = authn.ErrPasscodeNotSupported
	} else if passcode = m.Passcode(authn.KeyTOTP); passcode == nil {
		// Cannot enable 2FA if user has no passcode.
		err = authn.ErrPasscodeNotSetUp
	} else if err = passcode.Activate(); err == nil {
		// Enable 2FA after passcode has been activated.
		err = m.SetMethod(authn.Method2FA).Save()
		FlushSessionCache()
	} else if err != nil {
		log.Debugf("passcode: %s", clean.Error(err))
	}

	return passcode, err
}

// DeactivatePasscode disables two-factor authentication with a passcode.
func (m *User) DeactivatePasscode() (passcode *Passcode, err error) {
	if m == nil {
		err = errors.New("user is nil")
	} else if passcode = m.Passcode(authn.KeyTOTP); passcode == nil {
		// Disable 2FA if user has no passcode.
		err = m.SetMethod(authn.MethodDefault).Save()
		FlushSessionCache()
	} else if err = passcode.Delete(); err == nil {
		// Disable 2FA after passcode has been deleted.
		err = m.SetMethod(authn.MethodDefault).Save()
		FlushSessionCache()
	} else {
		log.Debugf("passcode: %s", clean.Error(err))
	}

	return passcode, err
}

// Validate checks if username, email and role are valid and returns an error otherwise.
func (m *User) Validate() (err error) {
	// Validate username.
	if userName, nameErr := authn.Username(m.UserName); nameErr != nil {
		return fmt.Errorf("username is %s", nameErr.Error())
	} else {
		m.UserName = userName
	}

	// Check if username also meets the length requirements.
	if len(m.Username()) < UsernameLength {
		return fmt.Errorf("username must have at least %d characters", UsernameLength)
	}

	// Check user role.
	if acl.UserRoles[m.UserRole] == "" {
		return fmt.Errorf("unsupported user role")
	}

	// Check if the username is unique.
	var duplicate = User{}

	if err = Db().
		Where("user_name = ? AND id <> ?", m.UserName, m.ID).
		First(&duplicate).Error; err == nil {
		return fmt.Errorf("user %s already exists", clean.LogQuote(m.UserName))
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
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
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

// SetFormValues sets the values specified in the form.
func (m *User) SetFormValues(frm form.User) *User {
	m.UserName = frm.Username()
	m.SetProvider(frm.Provider())
	m.SetMethod(frm.Method())
	m.SetAuthID(frm.AuthID, frm.AuthIssuer)
	m.UserEmail = frm.Email()
	m.DisplayName = frm.DisplayName
	m.SuperAdmin = frm.SuperAdmin
	m.CanLogin = frm.CanLogin
	m.WebDAV = frm.WebDAV
	m.SetRole(frm.Role())
	m.UserAttr = frm.Attr()
	m.SetBasePath(frm.BasePath)
	m.SetUploadPath(frm.UploadPath)

	// Set display name default if empty.
	if m.DisplayName == "" || m.DisplayName == AdminDisplayName && m.ID == 1 {
		m.DisplayName = m.FullName()
	}

	return m
}

// GenerateTokens generates preview and download tokens as needed.
func (m *User) GenerateTokens(force bool) *User {
	if m.ID < 0 {
		return m
	}

	if m.PreviewToken == "" || force {
		m.PreviewToken = GenerateToken()
	}

	if m.DownloadToken == "" || force {
		m.DownloadToken = GenerateToken()
	}

	return m
}

// RegenerateTokens replaces the existing preview and download tokens.
func (m *User) RegenerateTokens() error {
	if m.ID < 0 {
		return nil
	}

	m.GenerateTokens(true)

	return m.Updates(Map{"PreviewToken": m.PreviewToken, "DownloadToken": m.DownloadToken})
}

// RefreshShares updates the list of shares.
func (m *User) RefreshShares() *User {
	m.UserShares = FindUserShares(m.GetUID())
	return m
}

// NoShares checks if the user has no shares yet.
func (m *User) NoShares() bool {
	if m.NotRegistered() {
		return true
	}

	return m.UserShares.Empty()
}

// HasShares checks if the user has any shares.
func (m *User) HasShares() bool {
	return !m.NoShares()
}

// HasShare if a uid was shared with the user.
func (m *User) HasShare(uid string) bool {
	if m.NotRegistered() || m.NoShares() {
		return false
	}

	// Check if the share list contains the specified UID.
	return m.UserShares.Contains(uid)
}

// SharedUIDs returns shared entity UIDs.
func (m *User) SharedUIDs() UIDs {
	if m.IsRegistered() && m.NoShares() {
		m.RefreshShares()
	}

	return m.UserShares.UIDs()
}

// RedeemToken updates shared entity UIDs using the specified token.
func (m *User) RedeemToken(token string) (n int) {
	if !m.IsRegistered() {
		return 0
	}

	// Find links.
	links := FindValidLinks(token, "")

	// Found?
	if n = len(links); n == 0 {
		return n
	}

	// Find shares.
	for _, link := range links {
		if found := FindUserShare(UserShare{UserUID: m.GetUID(), ShareUID: link.ShareUID}); found == nil {
			share := NewUserShare(m.GetUID(), link.ShareUID, link.Perm, link.ExpiresAt())
			share.LinkUID = link.LinkUID
			share.Comment = link.Comment

			if err := share.Save(); err != nil {
				event.AuditErr([]string{"user %s", "token %s", "failed to redeem shares", "%s"}, m.RefID, clean.Log(token), err)
			} else {
				link.Redeem()
			}
		} else if err := found.UpdateLink(link); err != nil {
			event.AuditErr([]string{"user %s", "token %s", "failed to update shares", "%s"}, m.RefID, clean.Log(token), err)
		}
	}

	return n
}

// Form returns a populated user form to perform changes.
func (m *User) Form() (form.User, error) {
	frm := form.User{UserDetails: &form.UserDetails{}}

	if err := deepcopier.Copy(m).To(&frm); err != nil {
		return frm, err
	}

	if err := deepcopier.Copy(m.UserDetails).To(frm.UserDetails); err != nil {
		return frm, err
	}

	return frm, nil
}

// PrivilegeLevelChange checks if saving the form changes the user privileges.
func (m *User) PrivilegeLevelChange(f form.User) bool {
	return m.UserRole != f.Role() ||
		m.SuperAdmin != f.SuperAdmin ||
		m.CanLogin != f.CanLogin ||
		m.WebDAV != f.WebDAV ||
		m.UserAttr != f.Attr() ||
		m.AuthProvider != f.AuthProvider ||
		m.AuthMethod != f.AuthMethod ||
		m.AuthIssuer != f.AuthIssuer ||
		m.AuthID != f.AuthID ||
		m.BasePath != f.BasePath ||
		m.UploadPath != f.UploadPath
}

// SaveForm updates the entity using form data and stores it in the database.
func (m *User) SaveForm(f form.User, u *User) error {
	if m.UserName == "" || m.ID <= 0 {
		return fmt.Errorf("system users cannot be modified")
	} else if (m.ID == 1 || f.SuperAdmin) && acl.RoleAdmin.NotEqual(f.Role()) {
		return fmt.Errorf("super admin must not have a non-admin role")
	} else if f.BasePath != "" && clean.UserPath(f.BasePath) == "" {
		return fmt.Errorf("invalid base folder")
	} else if f.UploadPath != "" && clean.UserPath(f.UploadPath) == "" {
		return fmt.Errorf("invalid upload folder")
	}

	// Ignore details if not set.
	if f.UserDetails == nil {
		// Ignore.
	} else if err := deepcopier.Copy(f.UserDetails).To(m.UserDetails); err != nil {
		return err
	} else {
		m.UserDetails.UserAbout = txt.Clip(m.UserDetails.UserAbout, txt.ClipComment)
		m.UserDetails.UserBio = txt.Clip(m.UserDetails.UserBio, txt.ClipText)
	}

	// Sanitize display name.
	if n := clean.Name(f.DisplayName); n != "" && n != m.DisplayName {
		m.SetDisplayName(n, SrcManual)
	}

	// Set display name default if empty.
	if m.DisplayName == "" || m.DisplayName == AdminDisplayName && m.ID == 1 {
		m.DisplayName = m.FullName()
	}

	// Sanitize email address.
	if email := f.Email(); email != "" && email != m.UserEmail {
		m.UserEmail = email
		m.VerifiedAt = nil
		m.VerifyToken = GenerateToken()
	}

	// Change user privileges only if allowed.
	if u == nil {
		// Do nothing.
	} else if u.IsAdmin() {
		// Prevent admins from locking themselves out.
		if u.Equal(m) {
			m.SetRole(acl.RoleAdmin.String())
			m.CanLogin = true
		} else {
			m.SetRole(f.Role())
			m.CanLogin = f.CanLogin
		}

		m.WebDAV = f.WebDAV
		m.UserAttr = f.Attr()

		// Only allow super admins to change the authentication method and make other users super admins.
		if u.IsSuperAdmin() {
			if !u.Equal(m) {
				m.SuperAdmin = f.SuperAdmin
			}

			if !u.Equal(m) || f.Provider() != authn.ProviderNone {
				m.SetProvider(f.Provider())
			}

			m.SetMethod(f.Method())
			m.SetAuthID(f.AuthID, f.AuthIssuer)
		}

		m.SetBasePath(f.BasePath)
		m.SetUploadPath(f.UploadPath)
	}

	// Ensure super admins never have a non-admin role.
	if m.SuperAdmin {
		m.SetRole(acl.RoleAdmin.String())
	}

	// Make sure that the initial admin user cannot lock itself out.
	if m.ID == Admin.ID && (m.AclRole() != acl.RoleAdmin || !m.SuperAdmin || !m.CanLogin) {
		m.SetRole(acl.RoleAdmin.String())
		m.SuperAdmin = true
		m.CanLogin = true
	}

	return m.Save()
}

// SetDisplayName sets a new display name and, if possible, splits it into its components.
func (m *User) SetDisplayName(name, src string) *User {
	name = clean.Name(name)

	d := m.Details()
	priority := SrcPriority[src] >= SrcPriority[d.NameSrc]

	if name == "" || !priority && m.DisplayName != "" {
		return m
	}

	m.DisplayName = name

	if !priority {
		return m
	}

	d.NameSrc = src

	// Try to parse name into components.
	n := txt.ParseName(name)

	d.NameTitle = n.Title
	d.GivenName = n.Given
	d.MiddleName = n.Middle
	d.FamilyName = n.Family
	d.NameSuffix = n.Suffix
	d.NickName = n.Nick

	return m
}

// SetGivenName updates the user's given name.
func (m *User) SetGivenName(name string) *User {
	m.Details().SetGivenName(name)
	return m
}

// SetFamilyName updates the user's family name.
func (m *User) SetFamilyName(name string) *User {
	m.Details().SetFamilyName(name)
	return m
}

// HasAvatar reports if a user avatar image has been set.
func (m *User) HasAvatar() bool {
	if m == nil {
		return false
	}

	return m.Thumb != ""
}

// SetAvatar updates the user avatar image.
func (m *User) SetAvatar(thumb, thumbSrc string) error {
	if m.UserName == "" || m.ID <= 0 {
		return fmt.Errorf("system user avatars cannot be changed")
	}

	if SrcPriority[thumbSrc] < SrcPriority[m.ThumbSrc] && m.Thumb != "" {
		return fmt.Errorf("no permission to change avatar")
	}

	m.Thumb = thumb
	m.ThumbSrc = thumbSrc

	return m.Updates(Map{"Thumb": m.Thumb, "ThumbSrc": m.ThumbSrc})
}
