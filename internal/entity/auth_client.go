package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ClientUID is the unique ID prefix.
const (
	ClientUID = byte('c')
)

// Clients represents a list of client applications.
type Clients []Client

// Client represents a client application.
type Client struct {
	ClientUID   string     `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"-" yaml:"ClientUID"`
	UserUID     string     `gorm:"type:VARBINARY(42);index;default:'';" json:"UserUID" yaml:"UserUID,omitempty"`
	UserName    string     `gorm:"size:64;index;" json:"UserName" yaml:"UserName,omitempty"`
	user        *User      `gorm:"-"`
	ClientName  string     `gorm:"size:200;" json:"ClientName" yaml:"ClientName,omitempty"`
	ClientType  string     `gorm:"type:VARBINARY(16)" json:"ClientType" yaml:"ClientType,omitempty"`
	ClientURL   string     `gorm:"type:VARBINARY(255);default:'';column:client_url;" json:"ClientURL" yaml:"ClientURL,omitempty"`
	CallbackURL string     `gorm:"type:VARBINARY(255);default:'';column:callback_url;" json:"CallbackURL" yaml:"CallbackURL,omitempty"`
	AuthMethod  string     `gorm:"type:VARBINARY(128);default:'';" json:"AuthMethod" yaml:"AuthMethod,omitempty"`
	AuthScope   string     `gorm:"size:1024;default:'';" json:"AuthScope" yaml:"AuthScope,omitempty"`
	AuthExpires int64      `json:"AuthExpires" yaml:"AuthExpires,omitempty"`
	AuthTokens  int64      `json:"AuthTokens" yaml:"AuthTokens,omitempty"` // TODO: Enforce limit for number of tokens.
	AuthEnabled bool       `json:"AuthEnabled" yaml:"AuthEnabled,omitempty"`
	LastActive  int64      `json:"LastActive" yaml:"LastActive,omitempty"`
	CreatedAt   time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt   time.Time  `json:"UpdatedAt" yaml:"-"`
	DeletedAt   *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity table name.
func (Client) TableName() string {
	return "auth_clients"
}

// NewClient returns a new client application instance.
func NewClient() *Client {
	return &Client{
		UserUID:     "",
		ClientName:  "",
		ClientType:  authn.ClientConfidential,
		ClientURL:   "",
		CallbackURL: "",
		AuthMethod:  authn.MethodOAuth2.String(),
		AuthScope:   "",
		AuthExpires: UnixHour,
		AuthTokens:  5,
		AuthEnabled: true,
		LastActive:  0,
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

// FindClient returns the matching client or nil if it was not found.
func FindClient(uid string) *Client {
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

// HasUID tests if the entity has a valid uid.
func (m *Client) HasUID() bool {
	return rnd.IsUID(m.ClientUID, ClientUID)
}

// User returns the related user account, if any.
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

// SetUser updates the related user account.
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

// Create new entity in the database.
func (m *Client) Create() error {
	return Db().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Client) Save() error {
	return Db().Save(m).Error
}

// Delete marks the entity as deleted.
func (m *Client) Delete() (err error) {
	if m.ClientUID == "" {
		return fmt.Errorf("client uid is missing")
	}

	if err = UnscopedDb().Delete(Session{}, "auth_id = ?", m.ClientUID).Error; err != nil {
		event.AuditErr([]string{"client %s", "delete", "failed to remove sessions", "%s"}, m.ClientUID, err)
	}

	err = Db().Delete(m).Error

	FlushSessionCache()

	return err
}

// Deleted checks if the client has been deleted.
func (m *Client) Deleted() bool {
	if m.DeletedAt == nil {
		return false
	}

	return !m.DeletedAt.IsZero()
}

// Updates multiple properties in the database.
func (m *Client) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// NewSecret sets a new secret stored as hash.
func (m *Client) NewSecret() (s string, err error) {
	if !m.HasUID() {
		return "", fmt.Errorf("invalid client uid")
	}

	s = rnd.Base62(32)

	pw := NewPassword(m.ClientUID, s, false)

	if err = pw.Save(); err != nil {
		return "", err
	}

	return s, nil
}

// HasSecret checks if the given client secret is correct.
func (m *Client) HasSecret(s string) bool {
	return !m.WrongSecret(s)
}

// WrongSecret checks if the given client secret is incorrect.
func (m *Client) WrongSecret(s string) bool {
	if !m.HasUID() {
		return true
	}

	// Empty secret?
	if s == "" {
		return true
	}

	// Fetch secret.
	pw := FindPassword(m.ClientUID)

	// Found?
	if pw == nil {
		return true
	}

	// Invalid?
	if pw.IsWrong(s) {
		return true
	}

	return false
}

// Method returns the client authentication method.
func (m *Client) Method() authn.MethodType {
	return authn.Method(m.AuthMethod)
}

// Scope returns the client authorization scope.
func (m *Client) Scope() string {
	return clean.Scope(m.AuthScope)
}

// SetScope sets the client authorization scope.
func (m *Client) SetScope(s string) *Client {
	m.AuthScope = clean.Scope(s)
	return m
}

// UpdateLastActive sets the last activity of the client to now.
func (m *Client) UpdateLastActive() *Client {
	if !m.HasUID() {
		return m
	}

	m.LastActive = UnixTime()

	if err := Db().Model(m).UpdateColumn("LastActive", m.LastActive).Error; err != nil {
		log.Debugf("client: failed to update %s timestamp (%s)", m.ClientUID, err)
	}

	return m
}

// NewSession creates a new client session.
func (m *Client) NewSession(c *gin.Context) *Session {
	// Create, initialize, and return new session.
	sess := NewSession(m.AuthExpires, 0).SetContext(c)
	sess.AuthID = m.UID()
	sess.AuthProvider = authn.ProviderClient.String()
	sess.AuthMethod = m.Method().String()
	sess.AuthScope = m.Scope()
	sess.SetUser(m.User())

	return sess
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

	return DeleteClientSessions(m.ClientUID, m.AuthTokens)
}

// Expires returns the auth expiration duration.
func (m *Client) Expires() time.Duration {
	return time.Duration(m.AuthExpires) * time.Second
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
		// Ignore.
	} else if u := FindUser(User{UserUID: frm.UserUID, UserName: frm.UserName}); u != nil {
		m.SetUser(u)
	}

	if frm.ClientName != "" {
		m.ClientName = frm.Name()
	}

	if frm.AuthMethod != "" {
		m.AuthMethod = frm.Method().String()
	}

	if frm.AuthScope != "" {
		m.SetScope(frm.AuthScope)
	}

	if frm.AuthExpires > UnixMonth {
		m.AuthExpires = UnixMonth
	} else if frm.AuthExpires > 0 {
		m.AuthExpires = frm.AuthExpires
	} else if m.AuthExpires <= 0 {
		m.AuthExpires = UnixHour
	}

	if frm.AuthTokens > 2147483647 {
		m.AuthTokens = 2147483647
	} else if frm.AuthTokens > 0 {
		m.AuthTokens = frm.AuthTokens
	} else if m.AuthTokens < 0 {
		m.AuthTokens = -1
	}

	if frm.AuthEnabled {
		m.AuthEnabled = true
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
