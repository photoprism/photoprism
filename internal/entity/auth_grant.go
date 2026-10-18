package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// GrantUID is the unique ID prefix of app grants.
const GrantUID = byte('g')

// Grants represents a list of app grants.
type Grants []Grant

// Grant represents a user's approval of an app connected with the OAuth authorization code flow. It holds
// the hash of the current refresh token; the app's access tokens are sessions whose AuthID is the grant UID.
type Grant struct {
	GrantUID    string    `gorm:"type:bytes;size:42;primaryKey;autoIncrement:false;" json:"UID" yaml:"UID"`
	UserUID     string    `gorm:"type:bytes;size:42;uniqueIndex:uix_auth_grants_user_client,priority:1;default:'';" json:"UserUID" yaml:"UserUID"`
	UserName    string    `gorm:"size:200;default:'';" json:"UserName" yaml:"UserName,omitempty"`
	ClientID    string    `gorm:"type:bytes;size:512;uniqueIndex:uix_auth_grants_user_client,priority:2;index;default:'';" json:"ClientID" yaml:"ClientID"`
	ClientName  string    `gorm:"size:200;default:'';" json:"ClientName" yaml:"ClientName,omitempty"`
	RedirectURI string    `gorm:"type:bytes;size:2048;default:'';" json:"-" yaml:"RedirectURI,omitempty"`
	Scope       string    `gorm:"size:1024;default:'';" json:"Scope" yaml:"Scope"`
	Resource    string    `gorm:"type:bytes;size:512;default:'';" json:"Resource" yaml:"Resource"`
	RefreshHash string    `gorm:"type:bytes;size:64;default:'';" json:"-" yaml:"-"`
	RefreshGen  int       `gorm:"default:0;" json:"-" yaml:"RefreshGen,omitempty"`
	UsedAt      time.Time `gorm:"index;" json:"UsedAt" yaml:"UsedAt"`
	CreatedAt   time.Time `json:"CreatedAt" yaml:"CreatedAt"`
	UpdatedAt   time.Time `json:"UpdatedAt" yaml:"UpdatedAt"`
}

// TableName returns the entity table name.
func (Grant) TableName() string {
	return "auth_grants"
}

// Create inserts a new grant with a new UID into the database.
func (m *Grant) Create() error {
	if m == nil {
		return errors.New("grant must not be nil")
	} else if m.UserUID == "" || m.ClientID == "" || m.RefreshHash == "" {
		return errors.New("grant requires a user, client, and refresh token")
	}

	if !rnd.IsUID(m.GrantUID, GrantUID) {
		m.GrantUID = rnd.GenerateUID(GrantUID)
	}

	if m.UsedAt.IsZero() {
		m.UsedAt = time.Now().UTC()
	}

	return Db().Create(m).Error
}

// Rotate atomically replaces the refresh token hash if it still equals oldHash, increments the generation
// and sets the time of last use, and reports whether this call changed it.
func (m *Grant) Rotate(oldHash, newHash string) (bool, error) {
	if m == nil || m.GrantUID == "" || oldHash == "" || newHash == "" {
		return false, errors.New("grant uid and refresh token hashes must not be empty")
	}

	now := time.Now().UTC()

	result := UnscopedDb().Model(&Grant{}).
		Where("grant_uid = ? AND refresh_hash = ?", m.GrantUID, oldHash).
		UpdateColumns(Values{"refresh_hash": newHash, "refresh_gen": gorm.Expr("refresh_gen + 1"), "used_at": now, "updated_at": now})

	if result.Error != nil {
		return false, result.Error
	} else if result.RowsAffected != 1 {
		return false, nil
	}

	m.RefreshHash = newHash
	m.RefreshGen++
	m.UsedAt = now
	m.UpdatedAt = now

	return true, nil
}

// Idle reports whether the grant has not been used since the specified time.
func (m *Grant) Idle(since time.Time) bool {
	if m == nil {
		return true
	}

	return m.UsedAt.Before(since)
}

// Delete removes the grant and then the access tokens issued for it.
func (m *Grant) Delete() error {
	if m == nil || m.GrantUID == "" {
		return errors.New("grant uid must not be empty")
	}

	// Delete the grant first, so a token issued meanwhile is caught by the check that the grant still exists.
	if err := UnscopedDb().Where("grant_uid = ?", m.GrantUID).Delete(&Grant{}).Error; err != nil {
		return err
	}

	return DeleteGrantSessions(m.GrantUID)
}

// FindGrant returns the grant with the specified UID, or nil if it does not exist.
func FindGrant(grantUID string) *Grant {
	if !rnd.IsUID(grantUID, GrantUID) {
		return nil
	}

	m := &Grant{}

	if err := UnscopedDb().Where("grant_uid = ?", grantUID).First(m).Error; err != nil {
		return nil
	}

	return m
}

// FindUserGrants returns the grants of the specified user, most recently used first.
func FindUserGrants(userUID string) (Grants, error) {
	result := Grants{}

	if userUID == "" {
		return result, nil
	}

	err := UnscopedDb().Where("user_uid = ?", userUID).Order("used_at DESC").Find(&result).Error

	return result, err
}

// FindUserClientGrant returns the grant of the user for the specified client, or nil if there is none.
func FindUserClientGrant(userUID, clientID string) *Grant {
	if userUID == "" || clientID == "" {
		return nil
	}

	m := &Grant{}

	if err := UnscopedDb().Where("user_uid = ? AND client_id = ?", userUID, clientID).First(m).Error; err != nil {
		return nil
	}

	return m
}

// DeleteGrants deletes the specified grants and their access tokens and returns the number of deleted grants.
func DeleteGrants(grants Grants) (deleted int) {
	for i := range grants {
		if err := grants[i].Delete(); err != nil {
			event.AuditErr([]string{"grant %s", "failed to delete", status.Error(err)}, grants[i].GrantUID)
		} else {
			deleted++
		}
	}

	return deleted
}

// DeleteUserGrants deletes the grants of a user and their access tokens, and returns the number of deleted grants.
func DeleteUserGrants(userUID string) int {
	grants, err := FindUserGrants(userUID)

	if err != nil {
		event.AuditErr([]string{"user %s", "failed to find grants", status.Error(err)}, userUID)
		return 0
	}

	return DeleteGrants(grants)
}

// DeleteClientGrants deletes the grants of a client and their access tokens, and returns the number of deleted grants.
func DeleteClientGrants(clientID string) int {
	if clientID == "" {
		return 0
	}

	grants := Grants{}

	if err := UnscopedDb().Where("client_id = ?", clientID).Find(&grants).Error; err != nil {
		event.AuditErr([]string{"client %s", "failed to find grants", status.Error(err)}, clientID)
		return 0
	}

	return DeleteGrants(grants)
}

// DeleteIdleGrants deletes the grants that have not been used since the specified time.
func DeleteIdleGrants(since time.Time) int {
	grants := Grants{}

	if err := UnscopedDb().Where("used_at < ?", since.UTC()).Find(&grants).Error; err != nil {
		event.AuditErr([]string{"failed to find idle grants", status.Error(err)})
		return 0
	}

	return DeleteGrants(grants)
}

// DeleteGrantSessions deletes the access tokens issued for the grant with the specified UID.
func DeleteGrantSessions(grantUID string) error {
	return deleteGrantSessions(grantUID, 0)
}

// DeleteGrantSessionsExceptLatest deletes the access tokens of the grant except the most recently created one.
func DeleteGrantSessionsExceptLatest(grantUID string) error {
	return deleteGrantSessions(grantUID, 1)
}

// deleteGrantSessions deletes the access tokens of a grant except the specified number of most recent ones.
func deleteGrantSessions(grantUID string, keep int) error {
	if grantUID == "" {
		return errors.New("grant uid must not be empty")
	}

	found := Sessions{}

	if err := UnscopedDb().Where("auth_id = ? AND auth_provider = ?", grantUID, authn.ProviderOAuthApp.String()).
		Order("created_at DESC").Find(&found).Error; err != nil {
		return err
	}

	for i := keep; i < len(found); i++ {
		if err := found[i].Delete(); err != nil {
			return err
		}
	}

	return nil
}
