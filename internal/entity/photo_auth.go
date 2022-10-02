package entity

import "github.com/photoprism/photoprism/internal/event"

// PhotoAuth represents the ownership of a Photo and the corresponding permissions.
type PhotoAuth struct {
	UID     string `gorm:"type:VARBINARY(42);primary_key;auto_increment:false" json:"UID" yaml:"UID"`
	UserUID string `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;index" json:"UserUID" yaml:"UserUID"`
	TeamUID string `gorm:"type:VARBINARY(42);index" json:"TeamUID" yaml:"TeamUID"`
	Perm    uint   `json:"Perm,omitempty" yaml:"Perm,omitempty"`
	Changed int64  `json:"-" yaml:"-"`
}

// TableName returns the database table name.
func (PhotoAuth) TableName() string {
	return "photos_auth_dev"
}

// NewPhotoAuth creates a new entity model.
func NewPhotoAuth(uid, userUid, teamUid string, perm uint) *PhotoAuth {
	result := &PhotoAuth{
		UID:     uid,
		UserUID: userUid,
		TeamUID: teamUid,
		Perm:    perm,
		Changed: TimeStamp().Unix(),
	}

	return result
}

// Create inserts a new record into the database.
func (m *PhotoAuth) Create() error {
	m.Changed = TimeStamp().Unix()
	return Db().Create(m).Error
}

// Save updates the record in the database or inserts a new record if it does not already exist.
func (m *PhotoAuth) Save() error {
	m.Changed = TimeStamp().Unix()
	return Db().Save(m).Error
}

// FirstOrCreatePhotoUser returns the existing record or inserts a new record if it does not already exist.
func FirstOrCreatePhotoUser(m *PhotoAuth) *PhotoAuth {
	found := PhotoAuth{}

	if err := Db().Where("uid = ?", m.UID).First(&found).Error; err == nil {
		return &found
	} else if err = m.Create(); err != nil {
		event.AuditErr([]string{"photo %s", "failed to set owner and permissions", "%s"}, m.UID, err)
		return nil
	}

	return m
}
