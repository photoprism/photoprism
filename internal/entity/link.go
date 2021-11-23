package entity

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/rnd"
	"github.com/photoprism/photoprism/pkg/txt"
)

type Links []Link

// Link represents a sharing link.
type Link struct {
	LinkUID     string    `gorm:"type:VARBINARY(42);primary_key;" json:"UID,omitempty" yaml:"UID,omitempty"`
	ShareUID    string    `gorm:"type:VARBINARY(42);unique_index:idx_links_uid_token;" json:"Share" yaml:"Share"`
	ShareSlug   string    `gorm:"type:VARBINARY(160);index;" json:"Slug" yaml:"Slug,omitempty"`
	LinkToken   string    `gorm:"type:VARBINARY(160);unique_index:idx_links_uid_token;" json:"Token" yaml:"Token,omitempty"`
	LinkExpires int       `json:"Expires" yaml:"Expires,omitempty"`
	LinkViews   uint      `json:"Views" yaml:"-"`
	MaxViews    uint      `json:"MaxViews" yaml:"-"`
	HasPassword bool      `json:"HasPassword" yaml:"HasPassword,omitempty"`
	CanComment  bool      `json:"CanComment" yaml:"CanComment,omitempty"`
	CanEdit     bool      `json:"CanEdit" yaml:"CanEdit,omitempty"`
	CreatedAt   time.Time `deepcopier:"skip" json:"CreatedAt" yaml:"CreatedAt"`
	ModifiedAt  time.Time `deepcopier:"skip" json:"ModifiedAt" yaml:"ModifiedAt"`
}

// BeforeCreate creates a random UID if needed before inserting a new row to the database.
func (m *Link) BeforeCreate(scope *gorm.Scope) error {
	if rnd.IsUID(m.LinkUID, 's') {
		return nil
	}

	return scope.SetColumn("LinkUID", rnd.PPID('s'))
}

// NewLink creates a sharing link.
func NewLink(shareUID string, canComment, canEdit bool) Link {
	now := TimeStamp()

	result := Link{
		LinkUID:    rnd.PPID('s'),
		ShareUID:   shareUID,
		LinkToken:  rnd.Token(10),
		CanComment: canComment,
		CanEdit:    canEdit,
		CreatedAt:  now,
		ModifiedAt: now,
	}

	return result
}

func (m *Link) Redeem() {
	m.LinkViews += 1

	result := Db().Model(m).UpdateColumn("LinkViews", m.LinkViews)

	if result.RowsAffected == 0 {
		log.Warnf("link: failed updating share view counter for %s", m.LinkUID)
	}
}

func (m *Link) Expired() bool {
	if m.MaxViews > 0 && m.LinkViews >= m.MaxViews {
		return true
	}

	if m.LinkExpires <= 0 {
		return false
	}

	now := TimeStamp()
	expires := m.ModifiedAt.Add(Seconds(m.LinkExpires))

	return now.After(expires)
}

func (m *Link) SetSlug(s string) {
	m.ShareSlug = txt.Slug(s)
}

func (m *Link) SetPassword(password string) error {
	pw := NewPassword(m.LinkUID, password)

	if err := pw.Save(); err != nil {
		return err
	}

	m.HasPassword = true

	return nil
}

func (m *Link) InvalidPassword(password string) bool {
	if !m.HasPassword {
		return false
	}

	pw := FindPassword(m.LinkUID)

	if pw == nil {
		return password != ""
	}

	return pw.InvalidPassword(password)
}

// Save inserts a new row to the database or updates a row if the primary key already exists.
func (m *Link) Save() error {
	if !rnd.IsPPID(m.ShareUID, 0) {
		return fmt.Errorf("link: invalid share uid (%s)", m.ShareUID)
	}

	if m.LinkToken == "" {
		return fmt.Errorf("link: empty share token")
	}

	m.ModifiedAt = TimeStamp()

	return Db().Save(m).Error
}

// Delete the link.
func (m *Link) Delete() error {
	if m.LinkToken == "" {
		return fmt.Errorf("link: empty share token")
	}

	return Db().Delete(m).Error
}

// DeleteShareLinks removed all links matching the share uid.
func DeleteShareLinks(shareUID string) error {
	return Db().Delete(&Link{}, "share_uid = ?", shareUID).Error
}

// FindLink returns an entity pointer if exists.
func FindLink(linkUID string) *Link {
	result := Link{}

	if err := Db().Where("link_uid = ?", linkUID).First(&result).Error; err == nil {
		return &result
	} else {
		log.Errorf("link: %s (not found)", err)
	}

	return nil
}

// FindLinks returns a slice of links for a token and share UID (at least one must be provided).
func FindLinks(token, share string) (result Links) {
	if token == "" && share == "" {
		log.Errorf("link: share token and uid must not be empty at the same time (find links)")
		return []Link{}
	}

	q := Db()

	if token != "" {
		q = q.Where("link_token = ?", token)
	}

	if share != "" {
		if rnd.IsPPID(share, 'a') {
			q = q.Where("share_uid = ?", share)
		} else {
			q = q.Where("share_slug = ?", share)
		}
	}

	if err := q.Order("modified_at DESC").Find(&result).Error; err != nil {
		log.Errorf("link: %s (not found)", err)
	}

	return result
}

// FindValidLinks returns a slice of non-expired links for a token and share UID (at least one must be provided).
func FindValidLinks(token, share string) (result Links) {
	for _, link := range FindLinks(token, share) {
		if !link.Expired() {
			result = append(result, link)
		}
	}

	return result
}

// String returns an human readable identifier for logging.
func (m *Link) String() string {
	return m.LinkUID
}
