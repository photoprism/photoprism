package entity

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// SessionPrefix for RefID.
const (
	SessionPrefix = "sess"
)

// Sessions represents a list of sessions.
type Sessions []Session

// Session represents a User session.
type Session struct {
	ID            string          `gorm:"type:VARBINARY(2048);primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	Status        int             `gorm:"-"`
	AuthMethod    string          `gorm:"type:VARBINARY(128);default:'';" json:"AuthMethod,omitempty" yaml:"AuthMethod,omitempty"`
	AuthProvider  string          `gorm:"type:VARBINARY(128);default:'';" json:"AuthProvider,omitempty" yaml:"AuthProvider,omitempty"`
	AuthScope     string          `gorm:"size:1024;default:'';" json:"AuthScope,omitempty" yaml:"AuthScope,omitempty"`
	AuthID        string          `gorm:"type:VARBINARY(128);index;default:'';" json:"AuthID,omitempty" yaml:"AuthID,omitempty"`
	UserUID       string          `gorm:"type:VARBINARY(64);index;default:'';" json:"UserUID,omitempty" yaml:"UserUID,omitempty"`
	UserName      string          `gorm:"size:64;index;" json:"UserName,omitempty" yaml:"UserName,omitempty"`
	user          *User           `gorm:"-"`
	PreviewToken  string          `gorm:"type:VARBINARY(64);column:preview_token;default:'';" json:"-" yaml:"-"`
	DownloadToken string          `gorm:"type:VARBINARY(64);column:download_token;default:'';" json:"-" yaml:"-"`
	AccessToken   string          `gorm:"type:VARBINARY(4096);column:access_token;default:'';" json:"-" yaml:"-"`
	RefreshToken  string          `gorm:"type:VARBINARY(512);column:refresh_token;default:'';" json:"-" yaml:"-"`
	IdToken       string          `gorm:"type:VARBINARY(1024);column:id_token;default:'';" json:"IdToken,omitempty" yaml:"IdToken,omitempty"`
	UserAgent     string          `gorm:"size:512;" json:"UserAgent,omitempty" yaml:"UserAgent,omitempty"`
	ClientIP      string          `gorm:"size:64;column:client_ip;" json:"ClientIP,omitempty" yaml:"ClientIP,omitempty"`
	LoginIP       string          `gorm:"size:64;column:login_ip" json:"-" yaml:"-"`
	LoginAt       time.Time       `json:"-" yaml:"-"`
	DataJSON      json.RawMessage `gorm:"type:VARBINARY(4096);" json:"Data,omitempty" yaml:"Data,omitempty"`
	data          *SessionData    `gorm:"-"`
	RefID         string          `gorm:"type:VARBINARY(16);default:'';" json:"-" yaml:"-"`
	CreatedAt     time.Time       `json:"CreatedAt" yaml:"CreatedAt"`
	UpdatedAt     time.Time       `json:"UpdatedAt" yaml:"UpdatedAt"`
	ExpiresAt     time.Time       `sql:"index" json:"ExpiresAt,omitempty" yaml:"ExpiresAt,omitempty"`
}

// TableName returns the entity table name.
func (Session) TableName() string {
	return "auth_sessions_dev"
}

// NewSession creates a new session and returns it.
func NewSession(expiresAfter time.Duration) (m *Session) {
	m = &Session{
		ID:        rnd.SessionID(),
		RefID:     rnd.RefID(SessionPrefix),
		CreatedAt: TimeStamp(),
		UpdatedAt: TimeStamp(),
		ExpiresAt: time.Now().Add(expiresAfter),
	}

	return m
}

// SessionStatusUnauthorized returns a session with status unauthorized (401).
func SessionStatusUnauthorized() *Session {
	return &Session{Status: http.StatusUnauthorized}
}

// SessionStatusForbidden returns a session with status forbidden (403).
func SessionStatusForbidden() *Session {
	return &Session{Status: http.StatusForbidden}
}

// CacheDuration updates the session entity cache.
func (m *Session) CacheDuration(d time.Duration) {
	if !rnd.IsSessionID(m.ID) {
		return
	}

	sessionCache.Set(m.ID, *m, d)
}

// Cache caches the session with the default expiration duration.
func (m *Session) Cache() {
	m.CacheDuration(sessionCacheExpiration)
}

// Create new entity in the database.
func (m *Session) Create() (err error) {
	if err = Db().Create(m).Error; err == nil && rnd.IsSessionID(m.ID) {
		m.Cache()
	}

	return err
}

// Save entity properties.
func (m *Session) Save() (err error) {
	if err = Db().Save(m).Error; err == nil && rnd.IsSessionID(m.ID) {
		m.Cache()
	}

	return err
}

// Delete removes a session.
func (m *Session) Delete() error {
	DeleteFromSessionCache(m.ID)
	return UnscopedDb().Delete(m).Error
}

// Updates multiple properties in the database.
func (m *Session) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Session) BeforeCreate(scope *gorm.Scope) error {
	if rnd.InvalidRefID(m.RefID) {
		m.RefID = rnd.RefID(SessionPrefix)
		_ = scope.SetColumn("RefID", m.RefID)
	}

	if rnd.IsSessionID(m.ID) {
		return nil
	}

	m.ID = rnd.SessionID()

	return scope.SetColumn("ID", m.ID)
}

// User returns the session's user.
func (m *Session) User() *User {
	if m.user != nil {
		return m.user
	}

	if u := FindUserByUID(m.UserUID); u != nil {
		m.user = u
		return m.user
	}

	return &User{}
}

// RefreshUser updates the cached user data.
func (m *Session) RefreshUser() *Session {
	// Remove user if uid is nil.
	if m.UserUID == "" {
		m.user = nil
		return m
	}

	// Fetch matching record.
	if u := FindUserByUID(m.UserUID); u != nil {
		m.user = u
	}

	return m
}

// SetUser updates the session's user.
func (m *Session) SetUser(u *User) *Session {
	if u == nil {
		return m
	}

	m.user = u

	if u.UserUID != "" {
		m.UserUID = u.UserUID
		m.UserName = u.UserName
	}

	if u.DownloadToken != "" {
		m.DownloadToken = u.DownloadToken
	}

	if u.PreviewToken != "" {
		m.PreviewToken = u.PreviewToken
	}

	return m
}

// Data returns the session's data.
func (m *Session) Data() (data *SessionData) {
	if m.data != nil {
		data = m.data
	}

	data = NewSessionData()

	if len(m.DataJSON) == 0 {
		return data
	} else if err := json.Unmarshal(m.DataJSON, data); err != nil {
		log.Errorf("failed parsing session json: %s", err)
	} else {
		data.RefreshShares()
		m.data = data
	}

	return data
}

// SetData updates the session's data.
func (m *Session) SetData(data *SessionData) *Session {
	if data == nil {
		log.Debugf("auth: empty data passed to session %s", m.RefID)
		return m
	}

	// Refresh session data.
	data.RefreshShares()

	if j, err := json.Marshal(data); err != nil {
		log.Debugf("auth:  %s", err)
	} else {
		m.DataJSON = j
	}

	m.data = data

	return m
}

// SetContext updates the session's request context.
func (m *Session) SetContext(c *gin.Context) *Session {
	if c == nil {
		return m
	}

	m.SetClientIP(c.ClientIP())
	m.SetUserAgent(c.GetHeader("User-Agent"))

	return m
}

// IsVisitor checks if the session belongs to a sharing link visitor.
func (m *Session) IsVisitor() bool {
	return m.User().IsVisitor()
}

// IsRegistered checks if the session belongs to a registered user account.
func (m *Session) IsRegistered() bool {
	return m.User().IsRegistered()
}

// Unregistered checks if the session belongs to a unregistered user.
func (m *Session) Unregistered() bool {
	return !m.User().IsRegistered()
}

// NoShares checks if the session has no shares yet.
func (m *Session) NoShares() bool {
	return m.Data().NoShares()
}

// HasShares checks if the session has any shares.
func (m *Session) HasShares() bool {
	return m.Data().HasShares()
}

// HasShare if the session includes the specified share
func (m *Session) HasShare(uid string) bool {
	return m.Data().HasShare(uid)
}

// Expired checks if the session has expired.
func (m *Session) Expired() bool {
	if m.ExpiresAt.IsZero() {
		return false
	}

	return m.ExpiresAt.Before(time.Now())
}

// Invalid checks if the session does not belong to a registered user or a visitor with shares.
func (m *Session) Invalid() bool {
	return !m.Valid()
}

// Valid checks whether the session belongs to a registered user or a visitor with shares.
func (m *Session) Valid() bool {
	return m.User().IsRegistered() || m.IsVisitor() && m.HasShares()
}

// Abort aborts the request with the appropriate error code if access to the requested resource is denied.
func (m *Session) Abort(c *gin.Context) bool {
	if m.Valid() {
		return false
	}

	// Abort the request with the appropriate HTTP error code and message.
	switch m.Status {
	case http.StatusUnauthorized:
		c.AbortWithStatusJSON(m.Status, i18n.NewResponse(m.Status, i18n.ErrUnauthorized))
	default:
		c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
	}

	return true
}

// SharedUIDs returns shared entity UIDs.
func (m *Session) SharedUIDs() UIDs {
	data := m.Data()

	if data == nil {
		return UIDs{}
	}

	return data.SharedUIDs()
}

// SetUserAgent sets the client user agent.
func (m *Session) SetUserAgent(ua string) {
	if ua == "" {
		return
	} else if ua = txt.Clip(ua, 512); ua == "" {
		return
	} else if m.UserAgent != "" && m.UserAgent != ua {
		event.AuditWarn([]string{m.IP(), "session %s", "user agent has changed from %s to %s"}, m.RefID, clean.LogQuote(m.UserAgent), clean.LogQuote(ua))
	}

	m.UserAgent = ua
}

// SetClientIP sets the client IP address.
func (m *Session) SetClientIP(ip string) {
	if ip == "" {
		return
	} else if parsed := net.ParseIP(ip); parsed == nil {
		return
	} else if ip = parsed.String(); ip == "" {
		return
	} else if m.ClientIP != "" && m.ClientIP != ip {
		event.AuditWarn([]string{ip, "session %s", "client address has changed from %s to %s"}, m.RefID, clean.LogQuote(m.ClientIP), clean.LogQuote(ip))
	}

	m.ClientIP = ip

	if m.LoginIP == "" {
		m.LoginIP = ip
		m.LoginAt = TimeStamp()
	}
}

// IP returns the client IP address, or "unknown" if it is unknown.
func (m *Session) IP() string {
	if m.ClientIP != "" {
		return m.ClientIP
	} else {
		return "unknown"
	}
}
