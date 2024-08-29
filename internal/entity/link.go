package entity

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// LinkPrefix for RefID.
const (
	LinkUID    = byte('s')
	LinkPrefix = "link"
)

type Links []Link

// Link represents a link to share content.
type Link struct {
	LinkUID     string    `gorm:"type:VARBINARY(42);primary_key;" json:"UID,omitempty" yaml:"UID,omitempty"`
	ShareUID    string    `gorm:"type:VARBINARY(42);unique_index:idx_links_uid_token;" json:"ShareUID" yaml:"ShareUID"`
	ShareSlug   string    `gorm:"type:VARBINARY(160);index;" json:"Slug" yaml:"Slug,omitempty"`
	LinkToken   string    `gorm:"type:VARBINARY(160);unique_index:idx_links_uid_token;" json:"Token" yaml:"Token,omitempty"`
	LinkExpires int       `json:"Expires" yaml:"Expires,omitempty"`
	LinkViews   uint      `json:"Views" yaml:"-"`
	MaxViews    uint      `json:"MaxViews" yaml:"-"`
	HasPassword bool      `json:"VerifyPassword" yaml:"VerifyPassword,omitempty"`
	Comment     string    `gorm:"size:512;" json:"Comment,omitempty" yaml:"Comment,omitempty"`
	Perm        uint      `json:"Perm,omitempty" yaml:"Perm,omitempty"`
	RefID       string    `gorm:"type:VARBINARY(16);" json:"-" yaml:"-"`
	CreatedBy   string    `gorm:"type:VARBINARY(42);index" json:"CreatedBy,omitempty" yaml:"CreatedBy,omitempty"`
	CreatedAt   time.Time `deepcopier:"skip" json:"CreatedAt" yaml:"CreatedAt"`
	ModifiedAt  time.Time `deepcopier:"skip" json:"ModifiedAt" yaml:"ModifiedAt"`
}

// TableName returns the entity table name.
func (Link) TableName() string {
	return "links"
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Link) BeforeCreate(scope *gorm.Scope) error {
	if rnd.InvalidRefID(m.RefID) {
		m.RefID = rnd.RefID(LinkPrefix)
		Log("link", "set ref id", scope.SetColumn("RefID", m.RefID))
	}

	if rnd.IsUnique(m.LinkUID, LinkUID) {
		return nil
	}

	return scope.SetColumn("LinkUID", rnd.GenerateUID(LinkUID))
}

// NewLink creates a sharing link.
func NewLink(shareUid string, canComment, canEdit bool) Link {
	return NewUserLink(shareUid, OwnerUnknown)
}

// NewUserLink creates a sharing link owned by a user.
func NewUserLink(shareUid, userUid string) Link {
	now := Now()

	result := Link{
		LinkUID:    rnd.GenerateUID(LinkUID),
		ShareUID:   shareUid,
		LinkToken:  rnd.Base36(10),
		CreatedBy:  userUid,
		CreatedAt:  now,
		ModifiedAt: now,
	}

	return result
}

// Redeem increases the number of link visitors by one.
func (m *Link) Redeem() *Link {
	m.LinkViews += 1

	if err := Db().Model(m).UpdateColumn("link_views", gorm.Expr("link_views + 1")).Error; err != nil {
		event.AuditWarn([]string{"link %s", "failed to update view counter"}, clean.Log(m.RefID), err)
	}

	return m
}

// ExpiresAt returns the time when the share link expires or nil if it never expires.
func (m *Link) ExpiresAt() *time.Time {
	if m.LinkExpires <= 0 {
		return nil
	}

	expires := Now()
	expires = m.ModifiedAt.Add(Seconds(m.LinkExpires))

	return &expires
}

// Expired checks if the share link has expired.
func (m *Link) Expired() bool {
	if m.MaxViews > 0 && m.LinkViews >= m.MaxViews {
		return true
	}

	if expires := m.ExpiresAt(); expires == nil {
		return false
	} else {
		return Now().After(*expires)
	}
}

// SetSlug sets the URL slug of the link.
func (m *Link) SetSlug(s string) {
	m.ShareSlug = txt.Slug(s)
}

// SetPassword sets the password required to use the share link.
func (m *Link) SetPassword(password string) error {
	pw := NewPassword(m.LinkUID, password, false)

	if err := pw.Save(); err != nil {
		return err
	}

	m.HasPassword = true

	return nil
}

// InvalidPassword checks if the password provided to use the share link is invalid.
func (m *Link) InvalidPassword(password string) bool {
	if !m.HasPassword {
		return false
	}

	pw := FindPassword(m.LinkUID)

	if pw == nil {
		return password != ""
	}

	return pw.Invalid(password)
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *Link) Save() error {
	if !rnd.IsUID(m.ShareUID, 0) {
		return fmt.Errorf("invalid share uid")
	}

	if m.LinkToken == "" {
		return fmt.Errorf("empty link token")
	}

	m.ModifiedAt = Now()

	return Db().Save(m).Error
}

// Delete permanently deletes the link.
func (m *Link) Delete() error {
	if m.LinkToken == "" {
		return fmt.Errorf("empty link token")
	} else if m.LinkUID == "" {
		return fmt.Errorf("empty link uid")
	}

	// Remove related user shares.
	if err := UnscopedDb().Delete(UserShare{}, "link_uid = ?", m.LinkUID).Error; err != nil {
		event.AuditErr([]string{"link %s", "failed to remove related user shares", "%s"}, clean.Log(m.RefID), err)
	}

	return Db().Delete(m).Error
}

// DeleteShareLinks removes all links that match the shared UID.
func DeleteShareLinks(shareUid string) error {
	if shareUid == "" {
		return fmt.Errorf("empty share uid")
	}

	// Remove related user shares.
	if err := UnscopedDb().Delete(UserShare{}, "share_uid = ?", shareUid).Error; err != nil {
		event.AuditErr([]string{"share %s", "failed to remove related user shares", "%s"}, clean.Log(shareUid), err)
	}

	return Db().Delete(&Link{}, "share_uid = ?", shareUid).Error
}

// FindLink finds the link with the specified UID or nil if it is not found.
func FindLink(linkUid string) *Link {
	if rnd.InvalidUID(linkUid, LinkUID) {
		return nil
	}

	result := Link{}

	if Db().Where("link_uid = ?", linkUid).First(&result).Error != nil {
		event.AuditWarn([]string{"link %s", "not found"}, clean.Log(linkUid))
		return nil
	}

	return &result
}

// FindLinks returns a slice of links for a token and a share UID (at least one must be specified).
func FindLinks(token, shared string) (found Links) {
	found = Links{}
	token = clean.ShareToken(token)

	if token == "" && shared == "" {
		return found
	}

	q := Db()

	if token != "" {
		q = q.Where("link_token = ?", token)
	}

	if shared != "" {
		if rnd.IsUID(shared, 0) {
			q = q.Where("share_uid = ?", shared)
		} else {
			q = q.Where("share_slug = ?", shared)
		}
	}

	if err := q.Order("modified_at DESC").Find(&found).Error; err != nil {
		event.AuditErr([]string{"token %s", "%s"}, clean.Log(token), err)
	}

	return found
}

// FindValidLinks returns a slice of non-expired links for a token and share UID (at least one must be provided).
func FindValidLinks(token, shared string) (found Links) {
	found = Links{}

	for _, link := range FindLinks(token, shared) {
		if link.Expired() {
			continue
		}

		found = append(found, link)
	}

	return found
}

// String returns a human-readable identifier for use in logs.
func (m *Link) String() string {
	if m == nil {
		return "Link<nil>"
	}

	return clean.Log(m.LinkUID)
}
