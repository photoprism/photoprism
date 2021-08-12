package entity

import (
	"crypto/sha1"
	"encoding/base32"
	"time"
)

type PeopleFaces []PersonFace

// PersonFace represents the face of a Person.
type PersonFace struct {
	ID        string     `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	PersonUID string     `gorm:"type:VARBINARY(42);index;" json:"PersonUID" yaml:"PersonUID,omitempty"`
	Embedding string     `gorm:"type:LONGTEXT;" json:"Embedding" yaml:"Embedding,omitempty"`
	CreatedAt time.Time  `json:"CreatedAt" yaml:"CreatedAt,omitempty"`
	UpdatedAt time.Time  `json:"UpdatedAt" yaml:"UpdatedAt,omitempty"`
	DeletedAt *time.Time `sql:"index" json:"DeletedAt,omitempty" yaml:"-"`
}

// TableName returns the entity database table name.
func (PersonFace) TableName() string {
	return "people_faces_dev"
}

// NewPersonFace returns a new face.
func NewPersonFace(personUID, embedding string) *PersonFace {
	timeStamp := Timestamp()
	s := sha1.Sum([]byte(embedding))

	result := &PersonFace{
		ID:        base32.StdEncoding.EncodeToString(s[:]),
		PersonUID: personUID,
		Embedding: embedding,
		CreatedAt: timeStamp,
		UpdatedAt: timeStamp,
	}

	return result
}

// UnmarshalEmbedding parses the face embedding JSON string.
func (m *PersonFace) UnmarshalEmbedding() (result Embedding) {
	return UnmarshalEmbedding(m.Embedding)
}

// Save updates the existing or inserts a new face.
func (m *PersonFace) Save() error {
	peopleMutex.Lock()
	defer peopleMutex.Unlock()

	return Save(m, "ID")
}

// Create inserts the face to the database.
func (m *PersonFace) Create() error {
	peopleMutex.Lock()
	defer peopleMutex.Unlock()

	return Db().Create(m).Error
}

// Delete removes the face from the database.
func (m *PersonFace) Delete() error {
	return Db().Delete(m).Error
}

// Deleted returns true if the face is deleted.
func (m *PersonFace) Deleted() bool {
	return m.DeletedAt != nil
}

// Restore restores the face in the database.
func (m *PersonFace) Restore() error {
	if m.Deleted() {
		return UnscopedDb().Model(m).Update("DeletedAt", nil).Error
	}

	return nil
}

// Update a face property in the database.
func (m *PersonFace) Update(attr string, value interface{}) error {
	return UnscopedDb().Model(m).Update(attr, value).Error
}

// Updates face properties in the database.
func (m *PersonFace) Updates(values interface{}) error {
	return UnscopedDb().Model(m).Updates(values).Error
}
