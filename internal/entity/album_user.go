package entity

import (
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// AlbumUser maps an album to a user or team and stores the associated permissions.
type AlbumUser struct {
	UID     string `gorm:"type:VARBINARY(42);primary_key;auto_increment:false" json:"UID" yaml:"UID"`
	UserUID string `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;index" json:"UserUID,omitempty" yaml:"UserUID,omitempty"`
	TeamUID string `gorm:"type:VARBINARY(42);index" json:"TeamUID,omitempty" yaml:"TeamUID,omitempty"`
	Perm    uint   `json:"Perm,omitempty" yaml:"Perm,omitempty"`
}

// TableName returns the database table name.
func (AlbumUser) TableName() string {
	return "albums_users"
}

// NewAlbumUser creates a new ownership/permission entry for an album.
func NewAlbumUser(uid, userUid, teamUid string, perm uint) *AlbumUser {
	result := &AlbumUser{
		UID:     uid,
		UserUID: userUid,
		TeamUID: teamUid,
		Perm:    perm,
	}

	return result
}

// Create inserts a new record into the database.
func (m *AlbumUser) Create() error {
	return Db().Create(m).Error
}

// Save updates the record or inserts it when no row exists yet.
func (m *AlbumUser) Save() error {
	return Db().Save(m).Error
}

// FirstOrCreateAlbumUser returns the existing record or inserts it when missing, auditing failures.
func FirstOrCreateAlbumUser(m *AlbumUser) *AlbumUser {
	found := AlbumUser{}

	if err := Db().Where("uid = ?", m.UID).First(&found).Error; err == nil {
		return &found
	} else if err = m.Create(); err != nil {
		event.AuditErr([]string{"album %s", "failed to set owner and permissions", status.Error(err)}, m.UID)
		return nil
	}

	return m
}
