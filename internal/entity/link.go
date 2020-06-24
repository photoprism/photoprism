package entity

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/pkg/rnd"
)

type Links []Link

// Link represents a sharing link.
type Link struct {
	LinkUID      string    `gorm:"type:varbinary(42);primary_key;" json:"UID,omitempty" yaml:"UID,omitempty"`
	ShareUID     string    `gorm:"type:varbinary(42);unique_index:idx_links_uid_token;" json:"ShareUID"`
	ShareToken   string    `gorm:"type:varbinary(255);unique_index:idx_links_uid_token;" json:"ShareToken"`
	ShareExpires int       `json:"ShareExpires"`
	HasPassword  bool      `json:"HasPassword"`
	CanComment   bool      `json:"CanComment"`
	CanEdit      bool      `json:"CanEdit"`
	CreatedAt    time.Time `deepcopier:"skip" json:"CreatedAt"`
	UpdatedAt    time.Time `deepcopier:"skip" json:"UpdatedAt"`
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
	result := Link{
		LinkUID:    rnd.PPID('s'),
		ShareUID:   shareUID,
		ShareToken: rnd.Token(10),
		CanComment: canComment,
		CanEdit:    canEdit,
	}

	return result
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

// Create inserts a new row to the database.
func (m *Link) Create() error {
	if !rnd.IsPPID(m.ShareUID, 0) {
		return fmt.Errorf("link: invalid share uid (%s)", m.ShareUID)
	}

	if m.ShareToken == "" {
		return fmt.Errorf("link: empty share token")
	}

	return Db().Create(m).Error
}

// Save inserts a new row to the database or updates a row if the primary key already exists.
func (m *Link) Save() error {
	if !rnd.IsPPID(m.ShareUID, 0) {
		return fmt.Errorf("link: invalid share uid (%s)", m.ShareUID)
	}

	if m.ShareToken == "" {
		return fmt.Errorf("link: empty share token")
	}

	return Db().Save(m).Error
}

func (m *Link) Delete() error {
	if m.ShareToken == "" {
		return fmt.Errorf("link: empty share token")
	}

	return Db().Delete(m).Error
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
func FindLinks(shareToken, shareUID string) (result Links) {
	if shareToken == "" && shareUID == "" {
		log.Errorf("link: share token and uid must not be empty at the same time (find links)")
		return []Link{}
	}

	q := Db()

	if shareToken != "" {
		q = q.Where("share_token = ?", shareToken)
	}

	if shareUID != "" {
		q = q.Where("share_uid = ?", shareUID)
	}

	if err := q.Find(&result).Error; err != nil {
		log.Errorf("link: %s (not found)", err)
	}

	return result
}

// String returns an human readable identifier for logging.
func (m *Link) String() string {
	return fmt.Sprintf("%s/%s", m.ShareUID, m.ShareToken)
}
