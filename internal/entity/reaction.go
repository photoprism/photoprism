package entity

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/photoprism/photoprism/pkg/react"
)

// Reaction represents a human response to content such as photos and albums.
type Reaction struct {
	EntityUID string     `gorm:"type:VARBINARY(64);primary_key;auto_increment:false" json:"EntityUID,omitempty" yaml:"EntityUID,omitempty"`
	UserUID   string     `gorm:"type:VARBINARY(64);primary_key;auto_increment:false" json:"UserUID,omitempty" yaml:"UserUID,omitempty"`
	Reaction  string     `gorm:"type:VARBINARY(64);primary_key;auto_increment:false" json:"Reaction,omitempty" yaml:"Reaction,omitempty"`
	Reacted   int        `json:"Reacted,omitempty" yaml:"Reacted,omitempty"`
	ReactedAt *time.Time `sql:"index" json:"ReactedAt,omitempty" yaml:"ReactedAt,omitempty"`
}

// TableName returns the entity table name.
func (Reaction) TableName() string {
	return "reactions_dev"
}

// NewReaction creates a new Reaction struct.
func NewReaction(entityUID, userUID string) *Reaction {
	return &Reaction{
		EntityUID: entityUID,
		UserUID:   userUID,
	}
}

// FindReaction returns the matching Reaction record or nil if it was not found.
func FindReaction(entityUID, userUID string) (m *Reaction) {
	if entityUID == "" || userUID == "" {
		return nil
	}

	m = &Reaction{}

	if Db().First(m, "entity_uid = ? AND user_uid = ?", entityUID, userUID).RecordNotFound() {
		return nil
	}

	return m
}

// React adds a react.Emoji reaction.
func (m *Reaction) React(emo react.Emoji) *Reaction {
	m.Reaction = emo.String()
	m.Reacted += 1
	return m
}

// Emoji returns the reaction Emoji.
func (m *Reaction) Emoji() react.Emoji {
	return react.Emoji(m.Reaction)
}

// String returns the user reaction as string.
func (m *Reaction) String() string {
	return m.Reaction
}

// InvalidUID checks if the entity or user uid are missing or incorrect.
func (m *Reaction) InvalidUID() bool {
	return m.EntityUID == "" || m.UserUID == ""
}

// Unknown checks if the reaction data is missing or incorrect.
func (m *Reaction) Unknown() bool {
	if m.InvalidUID() {
		return true
	}

	return len(m.Reaction) == 0
}

// Save saves the Reaction.
func (m *Reaction) Save() (err error) {
	if m.Unknown() {
		return fmt.Errorf("unknown reaction")
	}

	if m.ReactedAt == nil {
		return m.Create()
	}

	reactedAt := TimePointer()

	cols := Values{"reaction": m.Reaction, "reacted": gorm.Expr("reacted + 1"), "reacted_at": reactedAt}

	if err = Db().Model(Reaction{}).
		Where("entity_uid = ? AND user_uid = ?", m.EntityUID, m.UserUID).
		UpdateColumns(cols).Error; err == nil {
		m.Reacted += 1
		m.ReactedAt = reactedAt
	}

	return err
}

// Create inserts a new Reaction.
func (m *Reaction) Create() (err error) {
	if m.Unknown() {
		return fmt.Errorf("reaction invalid")
	}

	r := &Reaction{EntityUID: m.EntityUID, UserUID: m.UserUID, Reaction: m.Reaction, Reacted: m.Reacted, ReactedAt: TimePointer()}

	if err = Db().Create(r).Error; err == nil {
		m.ReactedAt = r.ReactedAt
	}

	return err
}

// Delete deletes the Reaction.
func (m *Reaction) Delete() error {
	if m.InvalidUID() {
		return fmt.Errorf("reaction invalid")
	}

	// Delete record.
	err := Db().Delete(m, "entity_uid = ? AND user_uid = ?", m.EntityUID, m.UserUID).Error

	// Ok?
	if err == nil {
		m.ReactedAt = nil
	}

	return err
}
