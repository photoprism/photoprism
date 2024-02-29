package entity

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/i18n"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/header"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/report"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/photoprism/photoprism/pkg/unix"
)

// SessionPrefix for RefID.
const (
	SessionPrefix = "sess"
	UnknownIP     = "0.0.0.0"
)

// Sessions represents a list of sessions.
type Sessions []Session

// Session represents a User session.
type Session struct {
	ID            string          `gorm:"type:VARBINARY(2048);primary_key;auto_increment:false;" json:"-" yaml:"ID"`
	authToken     string          `gorm:"-" yaml:"-"`
	UserUID       string          `gorm:"type:VARBINARY(42);index;default:'';" json:"UserUID" yaml:"UserUID,omitempty"`
	UserName      string          `gorm:"size:200;index;" json:"UserName" yaml:"UserName,omitempty"`
	user          *User           `gorm:"-" yaml:"-"`
	ClientUID     string          `gorm:"type:VARBINARY(42);index;default:'';" json:"ClientUID" yaml:"ClientUID,omitempty"`
	ClientName    string          `gorm:"size:200;default:'';" json:"ClientName" yaml:"ClientName,omitempty"`
	ClientIP      string          `gorm:"size:64;column:client_ip;index" json:"ClientIP" yaml:"ClientIP,omitempty"`
	client        *Client         `gorm:"-" yaml:"-"`
	AuthProvider  string          `gorm:"type:VARBINARY(128);default:'';" json:"AuthProvider" yaml:"AuthProvider,omitempty"`
	AuthMethod    string          `gorm:"type:VARBINARY(128);default:'';" json:"AuthMethod" yaml:"AuthMethod,omitempty"`
	AuthDomain    string          `gorm:"type:VARBINARY(255);default:'';" json:"AuthDomain" yaml:"AuthDomain,omitempty"`
	AuthID        string          `gorm:"type:VARBINARY(255);index;default:'';" json:"AuthID" yaml:"AuthID,omitempty"`
	AuthScope     string          `gorm:"size:1024;default:'';" json:"AuthScope" yaml:"AuthScope,omitempty"`
	LastActive    int64           `json:"LastActive" yaml:"LastActive,omitempty"`
	SessExpires   int64           `gorm:"index" json:"Expires" yaml:"Expires,omitempty"`
	SessTimeout   int64           `json:"Timeout" yaml:"Timeout,omitempty"`
	PreviewToken  string          `gorm:"type:VARBINARY(64);column:preview_token;default:'';" json:"-" yaml:"-"`
	DownloadToken string          `gorm:"type:VARBINARY(64);column:download_token;default:'';" json:"-" yaml:"-"`
	AccessToken   string          `gorm:"type:VARBINARY(4096);column:access_token;default:'';" json:"-" yaml:"-"`
	RefreshToken  string          `gorm:"type:VARBINARY(512);column:refresh_token;default:'';" json:"-" yaml:"-"`
	IdToken       string          `gorm:"type:VARBINARY(1024);column:id_token;default:'';" json:"IdToken,omitempty" yaml:"IdToken,omitempty"`
	UserAgent     string          `gorm:"size:512;" json:"UserAgent" yaml:"UserAgent,omitempty"`
	DataJSON      json.RawMessage `gorm:"type:VARBINARY(4096);" json:"-" yaml:"Data,omitempty"`
	data          *SessionData    `gorm:"-" yaml:"-"`
	RefID         string          `gorm:"type:VARBINARY(16);default:'';" json:"ID" yaml:"-"`
	LoginIP       string          `gorm:"size:64;column:login_ip" json:"LoginIP" yaml:"-"`
	LoginAt       time.Time       `json:"LoginAt" yaml:"-"`
	CreatedAt     time.Time       `json:"CreatedAt" yaml:"CreatedAt"`
	UpdatedAt     time.Time       `json:"UpdatedAt" yaml:"UpdatedAt"`
	Status        int             `gorm:"-" json:"Status" yaml:"-"`
}

// TableName returns the entity table name.
func (Session) TableName() string {
	return "auth_sessions"
}

// NewSession creates a new session using the maxAge and timeout in seconds.
func NewSession(lifetime, timeout int64) (m *Session) {
	m = &Session{}

	m.Regenerate()

	if lifetime > 0 {
		m.SessExpires = TimeStamp().Unix() + lifetime
	}

	if timeout > 0 {
		m.SessTimeout = timeout
	}

	return m
}

// Expires sets an explicit expiration time.
func (m *Session) Expires(t time.Time) *Session {
	if t.IsZero() {
		return m
	}

	m.SessExpires = t.Unix()
	return m
}

// DeleteExpiredSessions deletes all expired sessions.
func DeleteExpiredSessions() (deleted int) {
	found := Sessions{}

	if err := Db().Where("sess_expires > 0 AND sess_expires < ?", unix.Time()).Find(&found).Error; err != nil {
		event.AuditErr([]string{"failed to fetch expired sessions", "%s"}, err)
		return deleted
	}

	for _, sess := range found {
		if err := sess.Delete(); err != nil {
			event.AuditErr([]string{sess.IP(), "session %s", "failed to delete", "%s"}, sess.RefID, err)
		} else {
			deleted++
		}
	}

	return deleted
}

// DeleteClientSessions deletes client sessions above the specified limit.
func DeleteClientSessions(client *Client, authMethod authn.MethodType, limit int64) (deleted int) {
	if limit < 0 {
		return 0
	} else if client == nil {
		return 0
	}

	q := Db()

	if client.HasUID() {
		q = q.Where("client_uid = ?", client.UID())
	} else if client.HasName() {
		q = q.Where("client_name = ?", client.Name())
	} else {
		return 0
	}

	if client.HasUser() {
		q = q.Where("user_uid = ?", client.UserUID)
	}

	if !authMethod.IsDefault() {
		q = q.Where("auth_method = ?", authMethod.String())
	}

	q = q.Order("created_at DESC").Limit(1000000000).Offset(limit)

	found := Sessions{}

	if err := q.Find(&found).Error; err != nil {
		event.AuditErr([]string{"failed to fetch client sessions", "%s"}, err)
		return deleted
	}

	for _, sess := range found {
		if err := sess.Delete(); err != nil {
			event.AuditErr([]string{sess.IP(), "session %s", "failed to delete", "%s"}, sess.RefID, err)
		} else {
			deleted++
		}
	}

	return deleted
}

// SessionStatusUnauthorized returns a session with status unauthorized (401).
func SessionStatusUnauthorized() *Session {
	return &Session{Status: http.StatusUnauthorized}
}

// SessionStatusForbidden returns a session with status forbidden (403).
func SessionStatusForbidden() *Session {
	return &Session{Status: http.StatusForbidden}
}

// SessionStatusTooManyRequests returns a session with status too many requests (429).
func SessionStatusTooManyRequests() *Session {
	return &Session{Status: http.StatusTooManyRequests}
}

// FindSessionByRefID finds an existing session by ref ID.
func FindSessionByRefID(refId string) *Session {
	if !rnd.IsRefID(refId) {
		return nil
	}

	m := &Session{}

	// Build query.
	if err := UnscopedDb().Where("ref_id = ?", refId).First(m).Error; err != nil {
		return nil
	}

	return m
}

// AuthToken returns the secret client authentication token.
func (m *Session) AuthToken() string {
	return m.authToken
}

// SetAuthToken sets a custom authentication token.
func (m *Session) SetAuthToken(authToken string) *Session {
	m.authToken = authToken
	m.ID = rnd.SessionID(authToken)

	return m
}

// AuthTokenType returns the authentication token type.
func (m *Session) AuthTokenType() string {
	return header.AuthBearer
}

// Regenerate (re-)initializes the session with a random auth token, ID, and RefID.
func (m *Session) Regenerate() *Session {
	if !rnd.IsSessionID(m.ID) {
		// Do not delete the old session if no ID is set yet.
	} else if err := m.Delete(); err != nil {
		event.AuditErr([]string{m.IP(), "session %s", "failed to delete", "%s"}, m.RefID, err)
	} else {
		event.AuditErr([]string{m.IP(), "session %s", "deleted"}, m.RefID)
	}

	generated := TimeStamp()

	m.SetAuthToken(rnd.AuthToken())
	m.RefID = rnd.RefID(SessionPrefix)
	m.CreatedAt = generated
	m.UpdatedAt = generated

	return m
}

// CacheDuration updates the session entity cache.
func (m *Session) CacheDuration(d time.Duration) {
	if !rnd.IsSessionID(m.ID) {
		return
	}

	CacheSession(m, d)
}

// Cache caches the session with the default expiration duration.
func (m *Session) Cache() {
	m.CacheDuration(sessionCacheExpiration)
}

// ClearCache deletes the session from the cache.
func (m *Session) ClearCache() {
	DeleteFromSessionCache(m.ID)
}

// Create new entity in the database.
func (m *Session) Create() (err error) {
	if err = Db().Create(m).Error; err == nil && rnd.IsSessionID(m.ID) {
		m.Cache()
	}

	return err
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Session) Save() error {
	if err := Db().Save(m).Error; err != nil {
		return err
	} else if rnd.IsSessionID(m.ID) {
		m.Cache()
	}

	// Limit the number of sessions that are created with an app password.
	if !m.Method().IsSession() {
		return nil
	} else if !m.Provider().IsApplication() {
		return nil
	} else if client := m.Client(); client.NoName() || client.Tokens() < 1 {
		return nil
	} else if deleted := DeleteClientSessions(client, authn.MethodSession, client.Tokens()); deleted > 0 {
		event.AuditInfo([]string{m.IP(), "session %s", "deleted %s"}, m.RefID, english.Plural(deleted, "previously created client session", "previously created client sessions"))
	}

	return nil
}

// Delete permanently deletes a session.
func (m *Session) Delete() error {
	return DeleteSession(m)
}

// Updates multiple properties in the database.
func (m *Session) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Session) BeforeCreate(scope *gorm.Scope) error {
	if rnd.InvalidRefID(m.RefID) {
		m.RefID = rnd.RefID(SessionPrefix)
		Log("session", "set ref id", scope.SetColumn("RefID", m.RefID))
	}

	if rnd.IsSessionID(m.ID) {
		return nil
	}

	m.Regenerate()

	return scope.SetColumn("ID", m.ID)
}

// SetClient updates the client of this session.
func (m *Session) SetClient(c *Client) *Session {
	if c == nil {
		return m
	}

	m.client = c
	m.ClientUID = c.UID()
	m.ClientName = c.ClientName
	m.AuthProvider = c.Provider().String()
	m.AuthMethod = c.Method().String()
	m.AuthScope = c.Scope()
	m.SetUser(c.User())

	return m
}

// SetClientName changes the session's client name.
func (m *Session) SetClientName(s string) *Session {
	if s == "" {
		return m
	}

	m.ClientName = clean.Name(s)

	return m
}

// Client returns the session's client.
func (m *Session) Client() *Client {
	if m == nil {
		return &Client{}
	} else if m.client != nil {
		return m.client
	} else if c := FindClientByUID(m.ClientUID); c != nil {
		m.SetClient(c)
		return m.client
	}

	return &Client{
		UserUID:    m.UserUID,
		UserName:   m.UserName,
		ClientUID:  m.ClientUID,
		ClientName: m.ClientName,
		ClientRole: m.ClientRole().String(),
		AuthScope:  m.Scope(),
		AuthMethod: m.AuthMethod,
	}
}

// ClientRole returns the session's client ACL role.
func (m *Session) ClientRole() acl.Role {
	if m.HasClient() {
		return m.Client().AclRole()
	} else if m.IsClient() {
		return acl.RoleClient
	}

	return acl.RoleNone
}

// ClientInfo returns the session's client identifier string.
func (m *Session) ClientInfo() string {
	if m.HasClient() {
		return m.Client().String()
	} else if m.ClientName != "" {
		return m.ClientName
	}

	return report.NotAssigned
}

// HasClient checks if a client entity is assigned to the session.
func (m *Session) HasClient() bool {
	if m == nil {
		return false
	}

	return m.ClientUID != ""
}

// NoClient if this session has no client entity assigned.
func (m *Session) NoClient() bool {
	return !m.HasClient()
}

// IsClient checks if this session authenticates an API client.
func (m *Session) IsClient() bool {
	return authn.Provider(m.AuthProvider).IsClient()
}

// User returns the session's user entity.
func (m *Session) User() *User {
	if m == nil {
		return &User{}
	} else if m.user != nil {
		return m.user
	} else if m.UserUID == "" {
		return &User{}
	}

	if u := FindUserByUID(m.UserUID); u != nil {
		m.SetUser(u)
		return m.user
	}

	return &User{}
}

// UserRole returns the session's user ACL role.
func (m *Session) UserRole() acl.Role {
	return m.User().AclRole()
}

// UserInfo returns the session's user information.
func (m *Session) UserInfo() string {
	name := m.Username()

	if name != "" {
		return name
	}

	return m.UserRole().String()
}

// SetUser updates the user entity of this session.
func (m *Session) SetUser(u *User) *Session {
	if u == nil {
		return m
	}

	// Update user.
	m.user = u
	m.UserUID = u.UserUID
	m.UserName = u.UserName

	// Update tokens.
	m.SetPreviewToken(u.PreviewToken)
	m.SetDownloadToken(u.DownloadToken)

	return m
}

// HasUser checks if a user entity is assigned to the session.
func (m *Session) HasUser() bool {
	if m == nil {
		return false
	}

	return m.UserUID != ""
}

// NoUser checks if this session has no user entity assigned.
func (m *Session) NoUser() bool {
	return !m.HasUser()
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
		m.SetUser(u)
	}

	return m
}

// Username returns the login name.
func (m *Session) Username() string {
	return m.UserName
}

// AuthInfo returns information about the authentication type.
func (m *Session) AuthInfo() string {
	provider := m.Provider()
	method := m.Method()

	if method.IsDefault() {
		return provider.Pretty()
	}

	return fmt.Sprintf("%s (%s)", provider.Pretty(), method.Pretty())
}

// SetAuthID sets a custom authentication identifier.
func (m *Session) SetAuthID(id string) *Session {
	if id == "" {
		return m
	}

	m.AuthID = clean.Name(id)

	return m
}

// Scope returns the authorization scope as a sanitized string.
func (m *Session) Scope() string {
	return clean.Scope(m.AuthScope)
}

// HasScope checks if the session has the given authorization scope.
func (m *Session) HasScope(scope string) bool {
	return list.ParseAttr(m.Scope()).Contains(scope)
}

// SetScope sets a custom authentication scope.
func (m *Session) SetScope(scope string) *Session {
	if scope == "" {
		return m
	}

	m.AuthScope = clean.Scope(scope)

	return m
}

// Method returns the authentication method.
func (m *Session) Method() authn.MethodType {
	return authn.Method(m.AuthMethod)
}

// SetMethod sets a custom authentication method.
func (m *Session) SetMethod(method authn.MethodType) *Session {
	if method == "" {
		return m
	}

	m.AuthMethod = method.String()

	return m
}

// Provider returns the authentication provider.
func (m *Session) Provider() authn.ProviderType {
	return authn.Provider(m.AuthProvider)
}

// SetProvider updates the session's authentication provider.
func (m *Session) SetProvider(provider authn.ProviderType) *Session {
	if provider == "" {
		return m
	}

	m.AuthProvider = provider.String()

	return m
}

// ChangePassword changes the password of the current user.
func (m *Session) ChangePassword(newPw string) (err error) {
	u := m.User()

	if u == nil {
		return fmt.Errorf("unknown user")
	}

	// Change password.
	err = u.SetPassword(newPw)

	m.SetPreviewToken(u.PreviewToken)
	m.SetDownloadToken(u.DownloadToken)

	return nil
}

// SetPreviewToken updates the preview token if not empty.
func (m *Session) SetPreviewToken(token string) *Session {
	if m.ID == "" {
		return m
	}

	if token != "" {
		m.PreviewToken = token
		PreviewToken.Set(token, m.ID)
	} else if m.PreviewToken == "" {
		m.PreviewToken = GenerateToken()
		PreviewToken.Set(token, m.ID)
	}

	return m
}

// SetDownloadToken updates the download token if not empty.
func (m *Session) SetDownloadToken(token string) *Session {
	if m.ID == "" {
		return m
	}

	if token != "" {
		m.DownloadToken = token
		DownloadToken.Set(token, m.ID)
	} else if m.DownloadToken == "" {
		m.DownloadToken = GenerateToken()
		DownloadToken.Set(token, m.ID)
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

// SetContext sets the session request context.
func (m *Session) SetContext(c *gin.Context) *Session {
	if c == nil || m == nil {
		return m
	}

	// Set client ip address from request context.
	if clientIp := header.ClientIP(c); clientIp != "" {
		m.SetClientIP(clientIp)
	} else if m.ClientIP == "" {
		// Unit tests often do not set a client IP.
		m.SetClientIP(UnknownIP)
	}

	// Set client user agent from request context.
	if ua := header.UserAgent(c); ua != "" {
		m.SetUserAgent(ua)
	}

	return m
}

// UpdateContext sets the session request context and updates the session entry in the database if it has changed.
func (m *Session) UpdateContext(c *gin.Context) *Session {
	if c == nil || m == nil {
		return m
	}

	changed := false

	// Set client ip address from request context.
	if clientIp := header.ClientIP(c); clientIp != "" && (clientIp != m.ClientIP || m.LoginIP == "") {
		m.SetClientIP(clientIp)
		changed = true
	} else if m.ClientIP == "" {
		// Unit tests often do not set a client IP.
		m.SetClientIP(UnknownIP)
		changed = true
	}

	// Set client user agent from request context.
	if ua := header.UserAgent(c); ua != "" && ua != m.UserAgent {
		m.SetUserAgent(ua)
		changed = true
	}

	if !changed {
		return m
	} else if err := m.Save(); err != nil {
		log.Debugf("auth:  %s while updating session context", err)
	}

	return m
}

// IsVisitor checks if the session belongs to a sharing link visitor.
func (m *Session) IsVisitor() bool {
	return m.User().IsVisitor()
}

// IsSuperAdmin checks if the session belongs to a registered super admin user.
func (m *Session) IsSuperAdmin() bool {
	if !m.IsRegistered() {
		return false
	}

	return m.User().IsSuperAdmin()
}

// IsRegistered checks if the session belongs to a registered user account.
func (m *Session) IsRegistered() bool {
	if m == nil || m.user == nil || rnd.InvalidUID(m.UserUID, UserUID) {
		return false
	}

	return m.User().IsRegistered()
}

// NotRegistered checks if the user is not registered with an own account.
func (m *Session) NotRegistered() bool {
	return !m.IsRegistered()
}

// NoShares checks if the session has no shares yet.
func (m *Session) NoShares() bool {
	return !m.HasShares()
}

// HasShares checks if the session has any shares.
func (m *Session) HasShares() bool {
	if user := m.User(); user.IsRegistered() {
		return user.HasShares()
	} else if data := m.Data(); data == nil {
		return false
	} else {
		return data.HasShares()
	}
}

// HasRegisteredUser checks if the session belongs to a registered user.
func (m *Session) HasRegisteredUser() bool {
	if !m.HasUser() {
		return false
	}

	return m.User().IsRegistered()
}

// HasShare if the session includes the specified share
func (m *Session) HasShare(uid string) bool {
	if user := m.User(); user.IsRegistered() {
		return user.HasShare(uid)
	} else if data := m.Data(); data == nil {
		return false
	} else {
		return data.HasShare(uid)
	}
}

// SharedUIDs returns shared entity UIDs.
func (m *Session) SharedUIDs() UIDs {
	if user := m.User(); user.IsRegistered() {
		return user.SharedUIDs()
	} else if data := m.Data(); data == nil {
		return UIDs{}
	} else {
		return data.SharedUIDs()
	}
}

// RedeemToken updates shared entity UIDs using the specified token.
func (m *Session) RedeemToken(token string) (n int) {
	if user := m.User(); user.IsRegistered() {
		return user.RedeemToken(token)
	} else if data := m.Data(); data == nil {
		return 0
	} else {
		return data.RedeemToken(token)
	}
}

// ExpiresAt returns the time when the session expires.
func (m *Session) ExpiresAt() time.Time {
	if m.SessExpires <= 0 {
		return time.Time{}
	}

	return time.Unix(m.SessExpires, 0)
}

// ExpiresIn returns the expiration time in seconds.
func (m *Session) ExpiresIn() int64 {
	if m.SessExpires <= 0 {
		return 0
	}

	return m.SessExpires - unix.Time()
}

// TimeoutAt returns the time at which the session will expire due to inactivity.
func (m *Session) TimeoutAt() time.Time {
	if m.SessTimeout <= 0 || m.LastActive <= 0 {
		return m.ExpiresAt()
	} else if t := m.LastActive + m.SessTimeout; t <= m.SessExpires || m.SessExpires <= 0 {
		return time.Unix(m.LastActive+m.SessTimeout, 0)
	} else {
		return m.ExpiresAt()
	}
}

// TimedOut checks if the session has expired due to inactivity..
func (m *Session) TimedOut() bool {
	if at := m.TimeoutAt(); at.IsZero() {
		return false
	} else {
		return at.Before(UTC())
	}
}

// Expired checks if the session has expired.
func (m *Session) Expired() bool {
	if m.SessExpires <= 0 {
		return m.TimedOut()
	} else if at := m.ExpiresAt(); at.IsZero() {
		return false
	} else {
		return at.Before(UTC())
	}
}

// UpdateLastActive sets the last activity of the session to now.
func (m *Session) UpdateLastActive() *Session {
	if m.Invalid() {
		return m
	}

	m.LastActive = unix.Time()

	if err := Db().Model(m).UpdateColumn("LastActive", m.LastActive).Error; err != nil {
		event.AuditWarn([]string{m.IP(), "session %s", "failed to update last active time", "%s"}, m.RefID, err)
	}

	return m
}

// Invalid checks if the session does not belong to a registered user or a visitor with shares.
func (m *Session) Invalid() bool {
	return !m.Valid()
}

// Valid checks whether the session belongs to a registered user or a visitor with shares.
func (m *Session) Valid() bool {
	if m.IsClient() {
		return true
	}

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
	case http.StatusTooManyRequests:
		c.AbortWithStatusJSON(m.Status, gin.H{"error": "rate limit exceeded", "code": http.StatusTooManyRequests})
	default:
		c.AbortWithStatusJSON(http.StatusForbidden, i18n.NewResponse(http.StatusForbidden, i18n.ErrForbidden))
	}

	return true
}

// SetUserAgent sets the client user agent.
func (m *Session) SetUserAgent(ua string) {
	if m == nil || ua == "" {
		return
	} else if ua = txt.Clip(ua, 512); ua == "" {
		return
	} else if m.UserAgent != "" && m.UserAgent != ua {
		event.AuditWarn([]string{m.IP(), "session %s", "user agent has changed from %s to %s"}, m.RefID, clean.LogQuote(m.UserAgent), clean.LogQuote(ua))
	}

	m.UserAgent = ua

	return
}

// SetClientIP sets the client IP address.
func (m *Session) SetClientIP(ip string) {
	if m == nil || ip == "" {
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

	return
}

// IP returns the client IP address, or "unknown" if it is unknown.
func (m *Session) IP() string {
	if m.ClientIP != "" {
		return m.ClientIP
	} else {
		return "0.0.0.0"
	}
}

// HttpStatus returns the session status as HTTP code for the client.
func (m *Session) HttpStatus() int {
	if m.Status > 0 {
		return m.Status
	} else if m.Valid() {
		return http.StatusOK
	}

	return http.StatusUnauthorized
}
