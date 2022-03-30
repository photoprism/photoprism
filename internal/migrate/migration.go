package migrate

import (
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

// Migration represents a database schema migration.
type Migration struct {
	ID         string     `gorm:"size:16;primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	Dialect    string     `gorm:"size:16;" json:"Dialect" yaml:"Dialect,omitempty"`
	Error      string     `gorm:"size:255;" json:"Error" yaml:"Error,omitempty"`
	Source     string     `gorm:"size:16;" json:"Source" yaml:"Source,omitempty"`
	Statements []string   `gorm:"-" json:"Statements" yaml:"Statements,omitempty"`
	StartedAt  time.Time  `json:"StartedAt" yaml:"StartedAt,omitempty"`
	FinishedAt *time.Time `json:"FinishedAt" yaml:"FinishedAt,omitempty"`
}

// TableName returns the entity database table name.
func (Migration) TableName() string {
	return "migrations"
}

// Fail marks the migration as failed by adding an error message.
func (m *Migration) Fail(err error, db *gorm.DB) {
	if err == nil {
		return
	}

	m.Error = err.Error()

	db.Model(m).Updates(Values{"Error": m.Error})
}

// Finish updates the FinishedAt timestamp when the migration was successful.
func (m *Migration) Finish(db *gorm.DB) error {
	return db.Model(m).Updates(Values{"FinishedAt": time.Now().UTC()}).Error
}

// Execute runs the migration.
func (m *Migration) Execute(db *gorm.DB) error {
	for _, s := range m.Statements { //  ADD
		if err := db.Exec(s).Error; err != nil {
			// Normalize query and error for comparison.
			q := strings.ToUpper(s)
			e := strings.ToUpper(err.Error())

			// Log the errors triggered by ALTER and DROP statements
			// and otherwise ignore them, since some databases do not
			// support "IF EXISTS".
			if strings.HasPrefix(q, "ALTER TABLE ") &&
				strings.Contains(s, " ADD ") &&
				strings.Contains(e, "DUPLICATE") {
				log.Tracef("migrate: %s (ignored, column already exists)", err)
			} else if strings.HasPrefix(q, "DROP INDEX ") &&
				strings.Contains(e, "DROP") {
				log.Tracef("migrate: %s (ignored, probably didn't exist anymore)", err)
			} else if strings.HasPrefix(q, "DROP TABLE ") &&
				strings.Contains(e, "DROP") {
				log.Tracef("migrate: %s (ignored, probably didn't exist anymored)", err)
			} else {
				return err
			}
		}
	}

	return nil
}
