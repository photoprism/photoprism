package entity

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/log/status"
	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

// LinkPrefix for RefID.
const (
	LinkUID    = byte('s')
	LinkPrefix = "link"
)

// Conditions that select the links and user shares deleted by deleteLinks.
const (
	linkUidCond  = "link_uid IN (?)"
	shareUidCond = "share_uid IN (?)"
)

// Links represents a list of share links.
type Links []Link

// Link represents a link to share content.
type Link struct {
	LinkUID     string    `gorm:"type:bytes;size:42;primaryKey;" json:"UID,omitempty" yaml:"UID,omitempty"`
	ShareUID    string    `gorm:"type:bytes;size:42;uniqueIndex:idx_links_uid_token;" json:"ShareUID" yaml:"ShareUID"`
	ShareSlug   string    `gorm:"type:bytes;size:160;index;" json:"Slug" yaml:"Slug,omitempty"`
	LinkToken   string    `gorm:"type:bytes;size:160;uniqueIndex:idx_links_uid_token;" json:"Token" yaml:"Token,omitempty"`
	LinkExpires int       `gorm:"size:32" json:"Expires" yaml:"Expires,omitempty"`
	LinkViews   uint      `gorm:"size:32" json:"Views" yaml:"-"`    // Counts the redemptions of the link.
	MaxViews    uint      `gorm:"size:32" json:"MaxViews" yaml:"-"` // Reserved, should be 0.
	Comment     string    `gorm:"size:512;" json:"Comment,omitempty" yaml:"Comment,omitempty"`
	Perm        uint      `gorm:"size:32" json:"Perm,omitempty" yaml:"Perm,omitempty"`
	RefID       string    `gorm:"type:bytes;size:16;" json:"-" yaml:"-"`
	CreatedBy   string    `gorm:"type:bytes;size:42;index" json:"CreatedBy,omitempty" yaml:"CreatedBy,omitempty"`
	CreatedAt   time.Time `deepcopier:"skip" json:"CreatedAt" yaml:"CreatedAt"`
	ModifiedAt  time.Time `deepcopier:"skip" json:"ModifiedAt" yaml:"ModifiedAt"`
}

// TableName returns the entity table name.
func (Link) TableName() string {
	return "links"
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Link) BeforeCreate(tx *gorm.DB) error {
	if rnd.InvalidRefID(m.RefID) {
		m.RefID = rnd.RefID(LinkPrefix)
		tx.Statement.SetColumn("RefID", m.RefID)
	}

	if rnd.IsUnique(m.LinkUID, LinkUID) {
		return nil
	}

	tx.Statement.SetColumn("LinkUID", rnd.GenerateUID(LinkUID))

	return nil
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
	m.LinkViews++

	if err := Db().Model(m).UpdateColumn("link_views", gorm.Expr("link_views + 1")).Error; err != nil {
		event.AuditWarn([]string{"link %s", "update views", status.Error(err)}, clean.Log(m.RefID))
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

// Expired checks if the share link has passed its expiration time.
func (m *Link) Expired() bool {
	if expires := m.ExpiresAt(); expires == nil {
		return false
	} else {
		return Now().After(*expires)
	}
}

// Redeemable checks if the link admits a new redemption. The view limit is read here rather than by
// Expired, which decides what an earlier redemption keeps.
func (m *Link) Redeemable() bool {
	if m.MaxViews > 0 && m.LinkViews >= m.MaxViews {
		return false
	}

	return !m.Expired()
}

// SetSlug sets the URL slug of the link.
func (m *Link) SetSlug(s string) {
	m.ShareSlug = txt.Slug(s)
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

// Delete permanently deletes the link with its user shares and the password stored for its UID, if any.
func (m *Link) Delete() error {
	if m.LinkToken == "" {
		return fmt.Errorf("empty link token")
	} else if m.LinkUID == "" {
		return fmt.Errorf("empty link uid")
	}

	if err := Db().Transaction(func(tx *gorm.DB) error { return deleteLinks(tx, linkUidCond, []string{m.LinkUID}) }); err != nil {
		event.AuditErr([]string{"link %s", "failed to remove", status.Error(err)}, clean.Log(m.RefID))
		return err
	}

	return nil
}

// DeleteShareLinks removes all links that match the shared UID, with their user shares and passwords.
func DeleteShareLinks(shareUid string) error {
	if shareUid == "" {
		return fmt.Errorf("empty share uid")
	}

	if err := Db().Transaction(func(tx *gorm.DB) error { return DeleteShareLinksTx(tx, []string{shareUid}) }); err != nil {
		event.AuditErr([]string{"share %s", "failed to remove links", status.Error(err)}, clean.Log(shareUid))
		return err
	}

	return nil
}

// DeleteShareLinksTx removes the links of the share UIDs, with their user shares and passwords,
// within the transaction tx.
func DeleteShareLinksTx(tx *gorm.DB, shareUids []string) error {
	return deleteLinks(tx, shareUidCond, shareUids)
}

// deleteLinks removes the links and user shares matching cond, which must be linkUidCond or
// shareUidCond, and the passwords stored for the link UIDs. The passwords go
// first, so the links still identify them; run it in a transaction, so a failure keeps all three.
func deleteLinks(tx *gorm.DB, cond string, values []string) error {
	if len(values) == 0 {
		return nil
	} else if slices.Contains(values, "") {
		return fmt.Errorf("empty link or share uid")
	}

	if err := tx.Delete(&UserShare{}, cond, values).Error; err != nil {
		return err
	}

	linkUids := tx.Model(&Link{}).Select("link_uid").Where(cond, values)

	if err := tx.Delete(&Password{}, "uid IN (?)", linkUids).Error; err != nil {
		return err
	}

	return tx.Delete(&Link{}, cond, values).Error
}

// FindLink finds the link with the specified UID or nil if it is not found.
func FindLink(linkUid string) *Link {
	if rnd.InvalidUID(linkUid, LinkUID) {
		return nil
	}

	result := Link{}

	if Db().Where("link_uid = ?", linkUid).First(&result).Error != nil {
		event.AuditWarn([]string{"link %s", status.NotFound}, clean.Log(linkUid))
		return nil
	}

	return &result
}

// FindRedeemableLinksByToken returns the links a request-supplied share token may be redeemed
// through. The token is required: it is sanitized first and must still be non-empty, so a caller
// cannot reach the lookup without one. Internal callers listing a record's links use FindLinks.
func FindRedeemableLinksByToken(token, shared string) Links {
	if token = clean.ShareToken(token); token == "" {
		return Links{}
	}

	return FindRedeemableLinks(token, shared)
}

// FindRedeemedLinksByToken returns the links a token resolves that are within their expiration time,
// which is what a redemption keeps. The token is sanitized as it is for a request-supplied one.
func FindRedeemedLinksByToken(token, shared string) Links {
	if token = clean.ShareToken(token); token == "" {
		return Links{}
	}

	return FindRedeemedLinks(token, shared)
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
		event.AuditErr([]string{"share token lookup failed", status.Error(err)})
	} else if len(found) == 0 {
		// SQLite schemas of earlier releases stored numeric tokens as numbers, without leading zeros.
		if numeric := sqliteNumericToken(token); numeric != token {
			return FindLinks(numeric, shared)
		}
	}

	return found
}

// sqliteNumericToken returns a numeric token without leading zeros if the database is SQLite, and the
// token unchanged otherwise.
func sqliteNumericToken(token string) string {
	if token == "" || DbDialect() != dsn.DialectSQLite || strings.Trim(token, "0123456789") != "" {
		return token
	} else if numeric := strings.TrimLeft(token, "0"); numeric != "" {
		return numeric
	}

	return "0"
}

// FindRedeemableLinks returns the links for a token and share UID that admit a new redemption
// (at least one of the two must be provided).
func FindRedeemableLinks(token, shared string) (found Links) {
	found = Links{}

	for _, link := range FindLinks(token, shared) {
		if !link.Redeemable() {
			continue
		}

		found = append(found, link)
	}

	return found
}

// FindRedeemedLinks returns the links for a token and share UID that are still within their
// expiration time, which is the access a redemption keeps (at least one of the two must be provided).
func FindRedeemedLinks(token, shared string) (found Links) {
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
