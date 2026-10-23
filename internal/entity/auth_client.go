package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/list"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/time/unix"
	"github.com/photoprism/photoprism/pkg/txt/report"
)

// ClientUID is the unique ID prefix.
const (
	ClientUID = byte('c')
)

// clientChangeFields lists the settings SaveChanges writes, so that identifiers and activity data changed
// meanwhile are kept. Updates with Select("*") would insert a missing row.
var clientChangeFields = []string{"ClientName", "ClientRole", "ClientType", "AuthProvider", "AuthMethod", "AuthScope",
	"AuthExpires", "AuthTokens", "AuthEnabled", "DataJSON", "UpdatedAt"}

// ErrClientNotFound is returned when a client does not exist or has been deleted.
var ErrClientNotFound = errors.New("client not found")

// Clients is a convenience alias for slices of Client.
type Clients []Client

// Client represents an OAuth/OpenID client registered with PhotoPrism.
type Client struct {
	ClientUID    string          `gorm:"type:bytes;size:42;primaryKey;autoIncrement:false;" json:"-" yaml:"ClientUID"`
	NodeUUID     string          `gorm:"type:bytes;size:64;index;default:'';" json:"NodeUUID,omitempty" yaml:"NodeUUID,omitempty"`
	UserUID      string          `gorm:"type:bytes;size:42;index;default:'';" json:"UserUID" yaml:"UserUID,omitempty"`
	UserName     string          `gorm:"size:200;index;" json:"UserName" yaml:"UserName,omitempty"`
	user         *User           `gorm:"-" yaml:"-"`
	AppName      string          `gorm:"size:64;" json:"AppName" yaml:"AppName,omitempty"`
	AppVersion   string          `gorm:"size:64;" json:"AppVersion" yaml:"AppVersion,omitempty"`
	ClientName   string          `gorm:"size:200;" json:"ClientName" yaml:"ClientName,omitempty"`
	DisplayName  string          `gorm:"size:200;" json:"DisplayName" yaml:"DisplayName,omitempty"`
	NameSrc      string          `gorm:"type:bytes;size:8;default:'';" json:"NameSrc,omitempty" yaml:"NameSrc,omitempty"`
	ClientRole   string          `gorm:"size:64;default:'';" json:"ClientRole" yaml:"ClientRole,omitempty"`
	ClientType   string          `gorm:"type:bytes;size:16" json:"ClientType" yaml:"ClientType,omitempty"`
	ClientURL    string          `gorm:"type:bytes;size:255;default:'';column:client_url;" json:"ClientURL" yaml:"ClientURL,omitempty"`
	AuthProvider string          `gorm:"type:bytes;size:128;default:'';" json:"AuthProvider" yaml:"AuthProvider,omitempty"`
	AuthMethod   string          `gorm:"type:bytes;size:128;default:'';" json:"AuthMethod" yaml:"AuthMethod,omitempty"`
	AuthScope    string          `gorm:"size:1024;default:'';" json:"AuthScope" yaml:"AuthScope,omitempty"`
	AuthExpires  int64           `json:"AuthExpires" yaml:"AuthExpires,omitempty"`
	AuthTokens   int64           `json:"AuthTokens" yaml:"AuthTokens,omitempty"`
	AuthEnabled  bool            `json:"AuthEnabled" yaml:"AuthEnabled,omitempty"`
	RefreshToken string          `gorm:"type:bytes;size:2048;column:refresh_token;default:'';" json:"-" yaml:"-"`
	IdToken      string          `gorm:"type:bytes;size:2048;column:id_token;default:'';" json:"IdToken,omitempty" yaml:"IdToken,omitempty"`
	DataJSON     json.RawMessage `gorm:"type:bytes;size:4096;" json:"-" yaml:"Data,omitempty"`
	data         *ClientData     `gorm:"-" yaml:"-"`
	secretHash   string          `gorm:"-" yaml:"-"`
	LastActive   int64           `json:"LastActive" yaml:"LastActive,omitempty"`
	CreatedAt    time.Time       `json:"CreatedAt" yaml:"-"`
	UpdatedAt    time.Time       `json:"UpdatedAt" yaml:"-"`
	DeletedAt    DeletedAt       `gorm:"index" json:"DeletedAt,omitzero" yaml:"-"`
}

// TableName returns the entity table name.
func (Client) TableName() string {
	return "auth_clients"
}

// NewClient returns a new client application instance with default ACL role and authentication settings.
func NewClient() *Client {
	return &Client{
		UserUID:      "",
		ClientName:   "",
		ClientRole:   acl.RoleClient.String(),
		ClientType:   authn.ClientConfidential,
		ClientURL:    "",
		AuthProvider: authn.ProviderClient.String(),
		AuthMethod:   authn.MethodOAuth2.String(),
		AuthScope:    "",
		AuthExpires:  unix.Hour,
		AuthTokens:   5,
		AuthEnabled:  true,
		LastActive:   0,
	}
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Client) BeforeCreate(tx *gorm.DB) error {
	if rnd.IsUID(m.ClientUID, ClientUID) {
		return nil
	}

	m.ClientUID = rnd.GenerateUID(ClientUID)

	tx.Statement.SetColumn("ClientUID", m.ClientUID)

	return nil
}

// FindClientByUID returns the matching client or nil if it was not found. The lookup reads
// deleted records too, and MUST keep doing so: an ownership check needs to see that a retired
// record still holds an identifier, and a caller that resolves a client gates on Deleted().
func FindClientByUID(uid string) *Client {
	if uid == "" {
		return nil
	} else if rnd.InvalidUID(uid, ClientUID) {
		return nil
	}

	m := &Client{}

	// Find matching record.
	if err := UnscopedDb().First(m, "client_uid = ?", uid).Error; err != nil {
		return nil
	}

	return m
}

// FindClient returns the client with the given client UID or node UUID, including a deleted
// one, or nil if neither matches. Operator tooling resolves its argument through this, so a
// cluster node can be named by the UUID its operator sees rather than by its client UID.
// When several records share a node UUID the most recently updated one is returned.
func FindClient(id string) *Client {
	if id == "" {
		return nil
	}

	if m := FindClientByUID(id); m != nil {
		return m
	}

	// A node UUID must carry the canonical layout, so an arbitrary argument is never
	// used as one: auth_clients.node_uuid is VARBINARY and storage may reinterpret a
	// value without separators.
	if !rnd.IsCanonicalUUID(id) {
		return nil
	}

	if list := FindClientsByNodeUUID(id); len(list) > 0 {
		return &list[0]
	}

	return nil
}

// FindClientByNodeUUID returns the current client with the given NodeUUID or nil if not found.
// Callers resolve a node through this to act on it, so it skips deleted records rather than
// returning whichever row sorts first; use FindClientsByNodeUUID to test whether a UUID is taken.
func FindClientByNodeUUID(nodeUUID string) *Client {
	if nodeUUID == "" {
		return nil
	}

	m := &Client{}

	if err := Db().Where("node_uuid = ?", nodeUUID).Order("updated_at DESC").First(m).Error; err != nil {
		return nil
	}

	return m
}

// FindClientsByNodeUUID returns all client rows matching the given NodeUUID ordered by UpdatedAt
// descending, including deleted ones, so that a retired identifier still counts as taken.
func FindClientsByNodeUUID(nodeUUID string) []Client {
	if nodeUUID == "" {
		return nil
	}

	var list []Client

	if err := UnscopedDb().Where("node_uuid = ?", nodeUUID).Order("updated_at DESC").Find(&list).Error; err != nil {
		return nil
	}

	return list
}

// clientScope returns the sanitized scope, reading an empty scope as "*", which SetFormValues stores for it.
func clientScope(s string) string {
	if s = clean.Scope(s); s == "" {
		return list.Any
	}

	return s
}

// CountClientSessions returns the number of sessions per client UID for the specified client UIDs, leaving out
// those whose expiration time has passed.
func CountClientSessions(clientUIDs []string) (map[string]int, error) {
	result := make(map[string]int, len(clientUIDs))

	if len(clientUIDs) == 0 {
		return result, nil
	}

	rows := []struct {
		ClientUID string
		Count     int
	}{}

	if err := Db().Model(&Session{}).Select("client_uid, COUNT(*) AS count").
		Where("client_uid IN (?)", clientUIDs).Where("(sess_expires <= 0 OR sess_expires >= ?)", unix.Now()).
		Group("client_uid").Scan(&rows).Error; err != nil {
		return result, err
	}

	for _, row := range rows {
		result[row.ClientUID] = row.Count
	}

	return result, nil
}

// GetUID returns the client uid string.
func (m *Client) GetUID() string {
	return m.ClientUID
}

// HasUID tests if the client has a valid uid.
func (m *Client) HasUID() bool {
	return rnd.IsUID(m.ClientUID, ClientUID)
}

// InvalidUID tests if the client does not have a valid uid.
func (m *Client) InvalidUID() bool {
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
	switch {
	case m == nil:
		return report.NotAssigned
	case m.HasUID():
		return m.GetUID()
	case m.HasName():
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

// SetDisplayName sets the human-friendly DisplayName when src has at least the
// priority of the current NameSrc, mirroring User.SetDisplayName. An empty name
// from a manual (admin) source clears the override so the next instance
// registration can repopulate it; an empty name from any other source is ignored.
func (m *Client) SetDisplayName(name, src string) *Client {
	name = clean.TypeUnicode(name)

	if SrcPriority[src] < SrcPriority[m.NameSrc] {
		return m
	}

	if name == "" {
		if src == SrcManual {
			m.DisplayName = ""
			m.NameSrc = SrcAuto
		}
		return m
	}

	m.DisplayName = name
	m.NameSrc = src

	return m
}

// SetRole sets the client role specified as string.
func (m *Client) SetRole(role string) *Client {
	r := clean.Role(role)

	// Map known roles (includes aliases like "none" or empty); fall back to client if unknown.
	if mapped, ok := acl.ClientRoles[r]; ok {
		m.ClientRole = mapped.String()
	} else {
		m.ClientRole = acl.RoleClient.String()
	}

	return m
}

// HasRole checks the client role specified as string.
func (m *Client) HasRole(role acl.Role) bool {
	return m.AclRole() == role
}

// AclRole returns the client role for ACL permission checks. A deleted client holds
// no permissions, so its role is resolved as none wherever a session still names it.
func (m *Client) AclRole() acl.Role {
	if m.Deleted() {
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

// HasInactiveUser checks if the client belongs to a user account that no longer exists or whose
// bound clients must be refused, see User.DenyClientAccess.
func (m *Client) HasInactiveUser() bool {
	if m == nil || m.UserUID == "" {
		return false
	} else if !m.HasUser() {
		return true
	}

	u := m.User()

	return !u.HasUID() || u.DenyClientAccess()
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
	switch {
	case m == nil:
		return ""
	case m.UserUID == "":
		return report.NotAssigned
	case m.UserName != "":
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
	err := Db().Create(m).Error
	flushClientCache(m.ClientUID)

	return err
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Client) Save() error {
	err := Db().Save(m).Error
	flushClientCache(m.ClientUID)

	if err != nil {
		return err
	}

	return m.endDisabledSessions()
}

// SaveChanges writes the specified settings, or all of clientChangeFields if none are specified, of a client that
// exists and is not deleted, and returns ErrClientNotFound otherwise, as Save would insert a missing row. It ends
// the sessions of a disabled client if it writes AuthEnabled.
func (m *Client) SaveChanges(fields ...string) error {
	if m == nil || !m.HasUID() {
		return ErrClientNotFound
	}

	selected := clientChangeFields

	if len(fields) > 0 {
		selected = make([]string, 0, len(fields)+1)

		for _, field := range fields {
			if !slices.Contains(clientChangeFields, field) {
				return fmt.Errorf("unsupported client field %s", clean.Log(field))
			} else if field != "UpdatedAt" && !slices.Contains(selected, field) {
				selected = append(selected, field)
			}
		}

		selected = append(selected, "UpdatedAt")
	}

	result := Db().Model(m).Select(selected).Updates(m)
	flushClientCache(m.ClientUID)

	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		// MariaDB reports no affected rows for an update that changes nothing.
		if stored := m.Stored(); stored == nil || stored.Deleted() {
			return ErrClientNotFound
		}
	}

	if !slices.Contains(selected, "AuthEnabled") {
		return nil
	}

	return m.endDisabledSessions()
}

// endDisabledSessions deletes the sessions and app access tokens of a client whose authentication is disabled.
func (m *Client) endDisabledSessions() error {
	if m.AuthEnabled {
		return nil
	} else if _, err := m.DeleteSessions(); err != nil {
		return err
	} else if m.IsApp() {
		SuspendClientGrants(m.ClientUID)
	}

	return nil
}

// RevokeChangedAccess revokes the sessions of a client that is not an app client if its role or scope
// differs from the specified previous values, and returns the number of revoked sessions. App clients are
// skipped, since their next access token takes the new scope.
func (m *Client) RevokeChangedAccess(prevRole, prevScope string) int {
	if m == nil || m.IsApp() {
		return 0
	} else if m.ClientRole == prevRole && m.Scope() == clientScope(prevScope) {
		return 0
	}

	return m.RevokeSessions(nil)
}

// Delete removes the secret, sessions, and grants of the client, and then marks it as deleted. The record
// is retained, so the identifiers it holds stay reserved and remain unavailable to any other record. Use
// Purge to release them.
func (m *Client) Delete() (err error) {
	if m.ClientUID == "" {
		return fmt.Errorf("client uid is empty")
	}

	if err = m.DeleteSecret(); err != nil {
		return err
	}

	if _, err = m.DeleteSessions(); err != nil {
		return err
	}

	// Removing an app client ends the grants of the users who connected it.
	DeleteClientGrants(m.ClientUID)

	err = Db().Delete(m).Error
	flushClientCache(m.ClientUID)

	return err
}

// Restore removes the deletion mark, so the client can be used again once a new secret is set. A
// deletion releases the client name, so this refuses when a current record has taken an identifier
// since, rather than leaving two live holders of it.
func (m *Client) Restore() error {
	if m == nil {
		return fmt.Errorf("client is nil")
	} else if m.ClientUID == "" {
		return fmt.Errorf("client uid is empty")
	}

	if conflict := m.RestoreConflict(); conflict != "" {
		return fmt.Errorf("%s is already in use by client %s", conflict, m.conflictingClient(conflict))
	}

	// A secret stored before the deletion is not carried over.
	if stored := m.Stored(); stored != nil && stored.Deleted() {
		if err := m.DeleteSecret(); err != nil {
			return err
		}
	}

	err := UnscopedDb().Model(m).Update("DeletedAt", nil).Error
	flushClientCache(m.ClientUID)

	if err != nil {
		return err
	}

	m.DeletedAt = DeletedAt{}

	return nil
}

// RestoreConflict reports the identifier that keeps this client from being restored, or an empty
// string when none does. Lookups resolve the most recently updated record, so a second live holder
// could take an identifier over rather than merely duplicate it. The stored record decides, since
// an object read before the deletion does not carry the mark.
func (m *Client) RestoreConflict() string {
	stored := m.Stored()

	if stored == nil || !stored.Deleted() {
		return ""
	}

	if stored.ClientName != "" && stored.currentClientWith("client_name = ?", stored.ClientName) != nil {
		return "client name"
	}

	if stored.NodeUUID != "" && stored.currentClientWith("node_uuid = ?", stored.NodeUUID) != nil {
		return "node uuid"
	}

	return ""
}

// Stored returns the record as it is saved, or nil when there is none. A receiver read before a
// change does not reflect it, so a decision about the record's state is made on this rather than
// on the copy in hand.
func (m *Client) Stored() *Client {
	if m == nil || m.ClientUID == "" {
		return nil
	}

	return FindClientByUID(m.ClientUID)
}

// conflictingClient names the current record that holds the given identifier, for reporting.
func (m *Client) conflictingClient(conflict string) string {
	stored := m.Stored()

	if stored == nil {
		return report.NotAssigned
	}

	switch conflict {
	case "client name":
		return stored.currentClientWith("client_name = ?", stored.ClientName).String()
	case "node uuid":
		return stored.currentClientWith("node_uuid = ?", stored.NodeUUID).String()
	}

	return report.NotAssigned
}

// currentClientWith returns another client that is not deleted and matches the given condition.
func (m *Client) currentClientWith(query string, args ...any) *Client {
	other := &Client{}

	if err := Db().Where(query, args...).Where("client_uid <> ?", m.ClientUID).First(other).Error; err != nil {
		return nil
	}

	return other
}

// Purge permanently removes the client, so that the identifiers it holds become available
// again. For a cluster node that includes its node UUID, which the provisioned database and the
// access grants are named after, so a caller releases those to the next holder along with it.
func (m *Client) Purge() (err error) {
	if m == nil {
		return fmt.Errorf("client is nil")
	} else if m.ClientUID == "" {
		return fmt.Errorf("client uid is empty")
	}

	// The client UID becomes available again, so its stored secret does not outlive it.
	if err = m.DeleteSecret(); err != nil {
		return err
	}

	if _, err = m.DeleteSessions(); err != nil {
		return err
	}

	DeleteClientGrants(m.ClientUID)

	err = UnscopedDb().Delete(m).Error
	flushClientCache(m.ClientUID)

	return err
}

// DeleteSecret removes the stored client secret, if any.
func (m *Client) DeleteSecret() error {
	if !m.HasUID() {
		return fmt.Errorf("invalid client uid")
	}

	if pw := FindPassword(m.ClientUID); pw != nil {
		err := pw.Delete()
		flushClientCache(m.ClientUID)

		return err
	}

	return nil
}

// DeleteSessions deletes all sessions that belong to this client.
func (m *Client) DeleteSessions() (deleted int, err error) {
	if m.ClientUID == "" {
		return 0, fmt.Errorf("client uid is empty")
	}

	if deleted = DeleteClientSessions(m, "", 0, ""); deleted > 0 {
		event.AuditInfo([]string{"client %s", "deleted %s"}, m.String(), english.Plural(deleted, "session", "sessions"))
	}

	return deleted, nil
}

// RevokeSessions deletes the sessions of this client, except for the ones specified to omit.
func (m *Client) RevokeSessions(omit []string) (deleted int) {
	if m == nil || m.ClientUID == "" {
		return 0
	}

	stmt := Db().Where("client_uid = ?", m.ClientUID)

	if len(omit) > 0 {
		stmt = stmt.Where("id NOT IN (?)", omit)
	}

	found := Sessions{}

	if err := stmt.Find(&found).Error; err != nil {
		event.AuditErr([]string{"client %s", "revoke sessions", status.Error(err)}, m.String())
		return 0
	}

	for _, sess := range found {
		if err := sess.Delete(); err != nil {
			event.AuditWarn([]string{"client %s", "revoke session %s", status.Error(err)}, m.String(), clean.Log(sess.RefID))
		} else {
			deleted++
		}
	}

	if deleted > 0 {
		event.AuditInfo([]string{"client %s", "revoke %s", status.Deleted}, m.String(), english.Plural(deleted, "session", "sessions"))
	}

	return deleted
}

// Deleted checks if the client has been deleted. A missing client reads as deleted,
// so a failed lookup and a retired record reach the same decision.
func (m *Client) Deleted() bool {
	if m == nil {
		return true
	} else if !m.DeletedAt.Valid {
		return false
	}

	return !m.DeletedAt.Time.IsZero()
}

// Disabled checks if the client has been deleted or its authentication is turned off.
func (m *Client) Disabled() bool {
	if m == nil {
		return true
	}

	return m.Deleted() || !m.AuthEnabled
}

// Updates multiple properties in the database.
func (m *Client) Updates(values any) error {
	err := UnscopedDb().Model(m).Updates(values).Error
	flushClientCache(m.ClientUID)

	return err
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
	err = pw.Save()
	flushClientCache(m.ClientUID)

	return err
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

	m.secretHash = pw.Hash

	return false
}

// AuthChanged reports whether the stored client differs from the settings this copy was
// authenticated with: it is missing, deleted, or disabled, its role or scope differs, or its secret
// hash differs from the one InvalidSecret matched. Secrets are not compared again.
func (m *Client) AuthChanged() bool {
	stored := m.Stored()

	switch {
	case stored == nil || stored.Disabled():
		return true
	case stored.AclRole() != m.AclRole() || clientScope(stored.AuthScope) != clientScope(m.AuthScope):
		return true
	case m.secretHash == "":
		return true
	}

	pw := FindPassword(m.ClientUID)

	return pw == nil || pw.Hash != m.secretHash
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

// UpdateLastActive sets the time of last activity to now and optionally also updates the auth_clients table.
// It does not flush the client cache, since no cached read depends on LastActive.
func (m *Client) UpdateLastActive(save bool) *Client {
	if m == nil {
		return &Client{}
	} else if !m.HasUID() {
		return m
	}

	// Set time of last activity to now (Unix timestamp).
	m.LastActive = unix.Now()

	// Update activity timestamp of this client in the auth_clients table.
	if !save {
		return m
	} else if err := Db().Model(m).UpdateColumn("last_active", m.LastActive).Error; err != nil {
		event.AuditWarn([]string{"client %s", "failed to update activity timestamp", status.Error(err)}, m.ClientUID)
	}

	return m
}

// NewSession creates a new client session.
func (m *Client) NewSession(c *gin.Context, t authn.GrantType) *Session {
	// Create, initialize, and return new session.
	return NewSession(m.AuthExpires, 0).SetContext(c).SetClient(m).SetGrantType(t)
}

// EnforceAuthTokenLimit deletes client sessions above the configured limit and returns the
// number of deleted sessions. The session with the specified ID ranks first within its
// creation second, so a newly issued token is among the retained ones.
func (m *Client) EnforceAuthTokenLimit(keepID string) (deleted int) {
	if m == nil || !m.HasUID() || m.AuthTokens < 0 {
		return 0
	}

	return DeleteClientSessions(m, authn.MethodOAuth2, m.Tokens(), keepID)
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
		s := reportValue(v)

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

	// Set values from form; an empty role keeps the current one.
	m.SetName(frm.Name())

	if role := frm.Role(); role != "" {
		m.SetRole(role)
	}

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
		m.AuthProvider = authn.ProviderClient.String()
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

	// Set the redirect URIs of app clients, which authenticate without a client secret.
	if frm.RedirectURIs != nil {
		data := m.GetData()
		data.RedirectURIs = frm.RedirectURIs
		m.SetData(data)
	}

	if m.IsApp() {
		m.ClientType = authn.ClientPublic
		m.ClientRole = acl.RoleClient.String()
	}

	return m
}

// IsApp reports whether this is an app client that users connect with the authorization code flow.
func (m *Client) IsApp() bool {
	if m == nil {
		return false
	}

	return m.Method() == authn.MethodPKCE
}

// RedirectURIs returns the registered redirect URIs of the client.
func (m *Client) RedirectURIs() []string {
	if m == nil {
		return nil
	}

	return m.GetData().RedirectURIs
}

// Validate checks the client application properties before saving them.
func (m *Client) Validate() (err error) {
	// Deleted client?
	if m.Deleted() {
		return errors.New("client has been deleted")
	}

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

	// App clients need at least one redirect URI.
	if m.IsApp() && len(m.RedirectURIs()) == 0 {
		return errors.New("app clients require at least one redirect URI")
	}

	// App client data, which holds the redirect URIs, exceeds the column size?
	if m.IsApp() && len(m.DataJSON) > ClientDataMaxBytes {
		return errors.New("client data is too large")
	}

	return nil
}
