package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/report"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/unix"
)

// ClientUID is the unique ID prefix.
const (
	ClientUID = byte('c')
)

// Clients represents a list of client applications.
type Clients []Client

// Client represents a client application.
type Client struct {
	ClientUID    string    `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"-" yaml:"ClientUID"`
	UserUID      string    `gorm:"type:VARBINARY(42);index;default:'';" json:"UserUID" yaml:"UserUID,omitempty"`
	UserName     string    `gorm:"size:200;index;" json:"UserName" yaml:"UserName,omitempty"`
	user         *User     `gorm:"-" yaml:"-"`
	ClientName   string    `gorm:"size:200;" json:"ClientName" yaml:"ClientName,omitempty"`
	ClientRole   string    `gorm:"size:64;default:'';" json:"ClientRole" yaml:"ClientRole,omitempty"`
	ClientType   string    `gorm:"type:VARBINARY(16)" json:"ClientType" yaml:"ClientType,omitempty"`
	ClientURL    string    `gorm:"type:VARBINARY(255);default:'';column:client_url;" json:"ClientURL" yaml:"ClientURL,omitempty"`
	CallbackURL  string    `gorm:"type:VARBINARY(255);default:'';column:callback_url;" json:"CallbackURL" yaml:"CallbackURL,omitempty"`
	AuthProvider string    `gorm:"type:VARBINARY(128);default:'';" json:"AuthProvider" yaml:"AuthProvider,omitempty"`
	AuthMethod   string    `gorm:"type:VARBINARY(128);default:'';" json:"AuthMethod" yaml:"AuthMethod,omitempty"`
	AuthScope    string    `gorm:"size:1024;default:'';" json:"AuthScope" yaml:"AuthScope,omitempty"`
	AuthExpires  int64     `json:"AuthExpires" yaml:"AuthExpires,omitempty"`
	AuthTokens   int64     `json:"AuthTokens" yaml:"AuthTokens,omitempty"`
	AuthEnabled  bool      `json:"AuthEnabled" yaml:"AuthEnabled,omitempty"`
	LastActive   int64     `json:"LastActive" yaml:"LastActive,omitempty"`
	CreatedAt    time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt    time.Time `json:"UpdatedAt" yaml:"-"`
}

// TableName returns the entity table name.
func (Client) TableName() string {
	return "auth_clients"
}

// NewClient returns a new client application instance.
func NewClient() *Client {
	return &Client{
		UserUID:      "",
		ClientName:   "",
		ClientRole:   acl.RoleClient.String(),
		ClientType:   authn.ClientConfidential,
		ClientURL:    "",
		CallbackURL:  "",
		AuthProvider: authn.ProviderClientCredentials.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "",
		AuthExpires:  unix.Hour,
		AuthTokens:   5,
		AuthEnabled:  true,
		LastActive:   0,
	}
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Client) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.ClientUID, ClientUID) {
		return nil
	}

	m.ClientUID = rnd.GenerateUID(ClientUID)

	return scope.SetColumn("ClientUID", m.ClientUID)
}

// FindClientByUID returns the matching client or nil if it was not found.
func FindClientByUID(uid string) *Client {
	if rnd.InvalidUID(uid, ClientUID) {
		return nil
	}

	m := &Client{}

	// Find matching record.
	if err := UnscopedDb().First(m, "client_uid = ?", uid).Error; err != nil {
		return nil
	}

	return m
}

// UID returns the client uid string.
func (m *Client) UID() string {
	return m.ClientUID
}

// HasUID tests if the client has a valid uid.
func (m *Client) HasUID() bool {
	return rnd.IsUID(m.ClientUID, ClientUID)
}

// NoUID tests if the client does not have a valid uid.
func (m *Client) NoUID() bool {
	return !m.HasUID()
}

// Name returns the client name string.
func (m *Client) Name() string {
	return m.ClientName
}

// HasName tests if the client has a name.
func (m *Client) HasName() bool {
	return m.ClientName != ""
}

// NoName tests if the client does not have a name.
func (m *Client) NoName() bool {
	return !m.HasName()
}

// String returns the client id or name for use in logs and reports.
func (m *Client) String() string {
	if m == nil {
		return report.NotAssigned
	} else if m.HasUID() {
		return m.UID()
	} else if m.HasName() {
		return m.Name()
	}

	return report.NotAssigned
}

// SetName sets a custom client name.
func (m *Client) SetName(s string) *Client {
	if s = clean.Name(s); s != "" {
		m.ClientName = s
	}

	return m
}

// SetRole sets the client role specified as string.
func (m *Client) SetRole(role string) *Client {
	if role != "" {
		m.ClientRole = acl.ClientRoles[clean.Role(role)].String()
	}

	return m
}

// HasRole checks the client role specified as string.
func (m *Client) HasRole(role acl.Role) bool {
	return m.AclRole() == role
}

// AclRole returns the client role for ACL permission checks.
func (m *Client) AclRole() acl.Role {
	if m == nil {
		return acl.RoleNone
	}

	if role, ok := acl.ClientRoles[clean.Role(m.ClientRole)]; ok {
		return role
	}

	return acl.RoleNone
}

// User returns the user who owns the client, if any.
func (m *Client) User() *User {
	if m.user != nil {
		return m.user
	} else if m.UserUID == "" {
		return &User{}
	}

	if u := FindUserByUID(m.UserUID); u != nil {
		m.user = u
		return m.user
	}

	return &User{}
}

// HasUser checks the client belongs to a user.
func (m *Client) HasUser() bool {
	return rnd.IsUID(m.UserUID, UserUID)
}

// SetUser sets the user to which the client belongs.
func (m *Client) SetUser(u *User) *Client {
	if u == nil {
		return m
	}

	// Update user references.
	m.user = u
	m.UserUID = u.UserUID
	m.UserName = u.UserName

	return m
}

// UserInfo reports the user that is assigned to this client.
func (m *Client) UserInfo() string {
	if m == nil {
		return ""
	} else if m.UserUID == "" {
		return report.NotAssigned
	} else if m.UserName != "" {
		return m.UserName
	}

	return m.UserUID
}

// AuthInfo reports the authentication configured for this client.
func (m *Client) AuthInfo() string {
	if m == nil {
		return ""
	}

	provider := m.Provider()
	method := m.Method()

	if method.IsDefault() {
		return provider.Pretty()
	}

	if provider.IsDefault() {
		return method.Pretty()
	}

	return fmt.Sprintf("%s (%s)", provider.Pretty(), method.Pretty())
}

// Create new entity in the database.
func (m *Client) Create() error {
	return Db().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Client) Save() error {
	if err := Db().Save(m).Error; err != nil {
		return err
	}

	// Delete related sessions if authentication is disabled.
	if m.AuthEnabled {
		return nil
	} else if _, err := m.DeleteSessions(); err != nil {
		return err
	}

	return nil
}

// Delete marks the entity as deleted.
func (m *Client) Delete() (err error) {
	if m.ClientUID == "" {
		return fmt.Errorf("client uid is missing")
	}

	if _, err = m.DeleteSessions(); err != nil {
		return err
	}

	err = Db().Delete(m).Error

	return err
}

// DeleteSessions deletes all sessions that belong to this client.
func (m *Client) DeleteSessions() (deleted int, err error) {
	if m.ClientUID == "" {
		return 0, fmt.Errorf("client uid is missing")
	}

	if deleted = DeleteClientSessions(m, "", 0); deleted > 0 {
		event.AuditInfo([]string{"client %s", "deleted %s"}, m.String(), english.Plural(deleted, "session", "sessions"))
	}

	return deleted, nil
}

// Deleted checks if the client has been deleted.
func (m *Client) Deleted() bool {
	if m == nil {
		return true
	}

	return false
}

// Disabled checks if the client authentication has been disabled.
func (m *Client) Disabled() bool {
	if m == nil {
		return true
	}

	return !m.AuthEnabled
}

// Updates multiple properties in the database.
func (m *Client) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// NewSecret sets a random client secret and returns it if successful.
func (m *Client) NewSecret() (secret string, err error) {
	if !m.HasUID() {
		return "", fmt.Errorf("invalid client uid")
	}

	secret = rnd.ClientSecret()

	if err = m.SetSecret(secret); err != nil {
		return "", err
	}

	return secret, nil
}

// SetSecret updates the current client secret or returns an error otherwise.
func (m *Client) SetSecret(secret string) (err error) {
	if !m.HasUID() {
		return fmt.Errorf("invalid client uid")
	} else if !rnd.IsClientSecret(secret) {
		return fmt.Errorf("invalid client secret")
	}

	pw := NewPassword(m.ClientUID, secret, false)

	if err = pw.Save(); err != nil {
		return err
	}

	return nil
}

// VerifySecret checks if the given client secret is correct.
func (m *Client) VerifySecret(s string) bool {
	return !m.InvalidSecret(s)
}

// InvalidSecret checks if the given client secret is invalid.
func (m *Client) InvalidSecret(s string) bool {
	// Check client UID.
	if !m.HasUID() {
		// Invalid, ID is missing.
		return true
	}

	// Check if secret is empty.
	if s == "" {
		// Invalid, no secret provided.
		return true
	}

	// Find secret.
	pw := FindPassword(m.ClientUID)

	// Found?
	if pw == nil {
		// Invalid, not found.
		return true
	}

	// Matches?
	if pw.Invalid(s) {
		// Invalid, does not match.
		return true
	}

	return false
}

// Provider returns the client authentication provider.
func (m *Client) Provider() authn.ProviderType {
	return authn.Provider(m.AuthProvider)
}

// SetProvider sets a custom client authentication provider.
func (m *Client) SetProvider(provider authn.ProviderType) *Client {
	if !provider.IsDefault() {
		m.AuthProvider = provider.String()
	}
	return m
}

// Method returns the client authentication method.
func (m *Client) Method() authn.MethodType {
	return authn.Method(m.AuthMethod)
}

// SetMethod sets a custom client authentication method.
func (m *Client) SetMethod(method authn.MethodType) *Client {
	if !method.IsDefault() {
		m.AuthMethod = method.String()
	}
	return m
}

// Scope returns the client authorization scope.
func (m *Client) Scope() string {
	return clean.Scope(m.AuthScope)
}

// SetScope sets a custom client authorization scope.
func (m *Client) SetScope(s string) *Client {
	if s = clean.Scope(s); s != "" {
		m.AuthScope = clean.Scope(s)
	}
	return m
}

// UpdateLastActive sets the last activity of the client to now.
func (m *Client) UpdateLastActive() *Client {
	if !m.HasUID() {
		return m
	}

	m.LastActive = unix.Time()

	if err := Db().Model(m).UpdateColumn("LastActive", m.LastActive).Error; err != nil {
		log.Debugf("client: failed to update %s timestamp (%s)", m.ClientUID, err)
	}

	return m
}

// NewSession creates a new client session.
func (m *Client) NewSession(c *gin.Context, t authn.GrantType) *Session {
	// Create, initialize, and return new session.
	return NewSession(m.AuthExpires, 0).SetContext(c).SetClient(m).SetGrantType(t)
}

// EnforceAuthTokenLimit deletes client sessions above the configured limit and returns the number of deleted sessions.
func (m *Client) EnforceAuthTokenLimit() (deleted int) {
	if m == nil {
		return 0
	} else if !m.HasUID() {
		return 0
	} else if m.AuthTokens < 0 {
		return 0
	}

	return DeleteClientSessions(m, authn.MethodOAuth2, m.AuthTokens)
}

// Expires returns the auth expiration duration.
func (m *Client) Expires() time.Duration {
	return time.Duration(m.AuthExpires) * time.Second
}

// SetExpires sets a custom auth expiration time in seconds.
func (m *Client) SetExpires(i int64) *Client {
	if i != 0 {
		m.AuthExpires = i
	}

	return m
}

// Tokens returns maximum number of access tokens this client can create.
func (m *Client) Tokens() int64 {
	if m.AuthTokens == 0 {
		return 1
	}

	return m.AuthTokens
}

// SetTokens sets a custom access token limit for this client.
func (m *Client) SetTokens(i int64) *Client {
	if i != 0 {
		m.AuthTokens = i
	}

	return m
}

// Report returns the entity values as rows.
func (m *Client) Report(skipEmpty bool) (rows [][]string, cols []string) {
	cols = []string{"Name", "Value"}

	// Extract model values.
	values, _, err := ModelValues(m, "ClientUID")

	// Ok?
	if err != nil {
		return rows, cols
	}

	rows = make([][]string, 0, len(values))

	for k, v := range values {
		s := fmt.Sprintf("%#v", v)

		// Skip empty values?
		if !skipEmpty || s != "" {
			rows = append(rows, []string{k, s})
		}
	}

	return rows, cols
}

// SetFormValues sets the values specified in the form.
func (m *Client) SetFormValues(frm form.Client) *Client {
	if frm.UserUID == "" && frm.UserName == "" {
		// Client does not belong to a specific user or the user remains unchanged.
	} else if u := FindUser(User{UserUID: frm.UserUID, UserName: frm.UserName}); u != nil {
		m.SetUser(u)
	}

	// Set custom client UID?
	if id := frm.ID(); m.ClientUID == "" && id != "" {
		m.ClientUID = id
	}

	// Set values from form.
	m.SetName(frm.Name())
	m.SetProvider(frm.Provider())
	m.SetMethod(frm.Method())
	m.SetScope(frm.Scope())
	m.SetTokens(frm.Tokens())
	m.SetExpires(frm.Expires())

	// Enable authentication?
	if frm.AuthEnabled {
		m.AuthEnabled = true
	}

	// Replace empty values with defaults.
	if m.AuthProvider == "" {
		m.AuthProvider = authn.ProviderClientCredentials.String()
	}

	if m.AuthMethod == "" {
		m.AuthMethod = authn.MethodOAuth2.String()
	}

	if m.AuthScope == "" {
		m.AuthScope = "*"
	}

	if m.AuthExpires <= 0 {
		m.AuthExpires = unix.Hour
	}

	if m.AuthTokens <= 0 {
		m.AuthTokens = -1
	}

	return m
}

// Validate checks the client application properties before saving them.
func (m *Client) Validate() (err error) {
	// Empty client name?
	if m.ClientName == "" {
		return errors.New("client name must not be empty")
	}

	// Empty client type?
	if m.ClientType == "" {
		return errors.New("client type must not be empty")
	}

	// Empty authorization method?
	if m.AuthMethod == "" {
		return errors.New("authorization method must not be empty")
	}

	// Empty authorization scope?
	if m.AuthScope == "" {
		return errors.New("authorization scope must not be empty")
	}

	return nil
}
