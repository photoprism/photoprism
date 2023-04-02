package entity

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const PermDefault uint = 0

const (
	PermNone uint = 1 << iota
	PermView
	PermReact
	PermComment
	PermUpload
	PermEdit
	PermShare
	PermAll
)

// SharePrefix for RefID.
const (
	SharePrefix = "share"
)

// UserShares represents shared content.
type UserShares []UserShare

// UIDs returns shared UIDs.
func (m UserShares) UIDs() UIDs {
	result := make(UIDs, len(m))

	for i, share := range m {
		result[i] = share.ShareUID
	}

	return result
}

// Empty checks if there are no shares.
func (m UserShares) Empty() bool {
	return m == nil || len(m) == 0
}

// Contains checks the uid is shared.
func (m UserShares) Contains(uid string) bool {
	if len(m) == 0 {
		return false
	}

	for _, share := range m {
		if share.ShareUID == uid {
			return true
		}
	}

	return false
}

// UserShare represents content shared with a user.
type UserShare struct {
	UserUID   string     `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"-" yaml:"UserUID"`
	ShareUID  string     `gorm:"type:VARBINARY(42);primary_key;index;" json:"ShareUID" yaml:"ShareUID"`
	LinkUID   string     `gorm:"type:VARBINARY(42);" json:"LinkUID,omitempty" yaml:"LinkUID,omitempty"`
	ExpiresAt *time.Time `sql:"index" json:"ExpiresAt,omitempty" yaml:"ExpiresAt,omitempty"`
	Comment   string     `gorm:"size:512;" json:"Comment,omitempty" yaml:"Comment,omitempty"`
	Perm      uint       `json:"Perm,omitempty" yaml:"Perm,omitempty"`
	RefID     string     `gorm:"type:VARBINARY(16);" json:"-" yaml:"-"`
	CreatedAt time.Time  `json:"CreatedAt" yaml:"-"`
	UpdatedAt time.Time  `json:"UpdatedAt" yaml:"-"`
}

// TableName returns the entity table name.
func (UserShare) TableName() string {
	return "auth_users_shares"
}

// NewUserShare creates a new entity model.
func NewUserShare(userUID, shareUid string, perm uint, expires *time.Time) *UserShare {
	result := &UserShare{
		UserUID:   userUID,
		ShareUID:  shareUid,
		Perm:      perm,
		RefID:     rnd.RefID(SharePrefix),
		CreatedAt: TimeStamp(),
		UpdatedAt: TimeStamp(),
		ExpiresAt: expires,
	}

	return result
}

// FindUserShare fetches the matching record or returns null if it was not found.
func FindUserShare(find UserShare) *UserShare {
	if !find.HasID() {
		return nil
	}

	m := &UserShare{}

	// Find matching record.
	if UnscopedDb().First(m, "user_uid = ? AND share_uid = ?", find.UserUID, find.ShareUID).Error != nil {
		return nil
	}

	return m
}

// FindUserShares finds all shares to which the user has access.
func FindUserShares(userUid string) UserShares {
	found := UserShares{}

	if rnd.InvalidUID(userUid, UserUID) {
		return found
	}

	// Find matching record.
	if err := UnscopedDb().Find(&found, "user_uid = ? AND (expires_at IS NULL OR expires_at > ?)", userUid, TimeStamp()).Error; err != nil {
		event.AuditWarn([]string{"user %s", "find shares", "%s"}, clean.Log(userUid), err)
		return nil
	}

	return found
}

// HasID tests if the entity has a valid uid.
func (m *UserShare) HasID() bool {
	return rnd.IsUID(m.UserUID, UserUID) && rnd.IsUID(m.ShareUID, 0)
}

// Create inserts a new record into the database.
func (m *UserShare) Create() error {
	return Db().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *UserShare) Save() error {
	return Db().Save(m).Error
}

// Updates changes multiple record values.
func (m *UserShare) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}

// UpdateLink updates the share data using the Link provided.
func (m *UserShare) UpdateLink(link Link) error {
	if m.ShareUID != link.ShareUID {
		return fmt.Errorf("shared uid does not match")
	}

	m.LinkUID = link.LinkUID
	m.Comment = link.Comment
	m.Perm = link.Perm
	m.UpdatedAt = TimeStamp()
	m.ExpiresAt = link.ExpiresAt()

	values := Values{
		"link_uid":   m.LinkUID,
		"expires_at": m.ExpiresAt,
		"comment":    m.Comment,
		"perm":       m.Perm,
		"updated_at": m.UpdatedAt}

	return m.Updates(values)
}
