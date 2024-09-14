package migrate

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

// Migration represents a database schema migration.
type Migration struct {
	ID         string     `gorm:"size:16;primary_key;auto_increment:false;" json:"ID" yaml:"ID"`
	Dialect    string     `gorm:"size:16;" json:"Dialect" yaml:"Dialect,omitempty"`
	Stage      string     `gorm:"size:16;" json:"Stage" yaml:"Stage,omitempty"`
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

// StageName returns the stage name.
func (m *Migration) StageName() string {
	if m.Stage == "" {
		return StageMain
	} else {
		return m.Stage
	}
}

// RunStage checks if the run stage matches.
func (m *Migration) RunStage(name string) bool {
	return m.StageName() == name
}

// Skip checks if the migration should be skipped based on the options.
func (m *Migration) Skip(opt Options) bool {
	return !m.RunStage(opt.StageName())
}

// Finished tests if the migration has been finished yet.
func (m *Migration) Finished() bool {
	if m.FinishedAt == nil {
		return false
	}

	return !m.FinishedAt.IsZero()
}

// RunDuration returns the run duration of started migrations.
func (m *Migration) RunDuration() time.Duration {
	if m.Error != "" || m.StartedAt.IsZero() {
		return time.Duration(0)
	}

	if m.Finished() {
		return m.FinishedAt.UTC().Sub(m.StartedAt.UTC())
	}

	return time.Now().UTC().Sub(m.StartedAt.UTC())
}

// Repeat tests if the migration should be repeated.
func (m *Migration) Repeat(runFailed bool) bool {
	if runFailed && m.Error != "" {
		return true
	}

	// Don't repeat if finished.
	if m.Finished() {
		return false
	}

	// Repeat not started yet.
	if m.StartedAt.IsZero() {
		return true
	}

	// Repeat if "running" for more than 60 minutes.
	return m.RunDuration().Minutes() >= 60
}

// Fail marks the migration as failed by adding an error message and removing the FinishedAt timestamp.
func (m *Migration) Fail(err error, db *gorm.DB) {
	if err == nil || db == nil {
		return
	}

	m.FinishedAt = nil

	if err.Error() == "" {
		m.Error = "unknown error"
	} else {
		m.Error = err.Error()
	}

	db.Model(m).Updates(Map{"FinishedAt": m.FinishedAt, "Error": m.Error})
}

// Finish updates the FinishedAt timestamp and removes the error message when the migration was successful.
func (m *Migration) Finish(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	finished := time.Now().UTC().Truncate(time.Second)
	m.FinishedAt = &finished
	m.Error = ""
	return db.Model(m).Updates(Map{"FinishedAt": m.FinishedAt, "Error": m.Error}).Error
}

// Execute runs the migration.
func (m *Migration) Execute(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	for _, s := range m.Statements { //  ADD
		if err := db.Exec(s).Error; err != nil {
			// Log the errors triggered by ALTER and DROP statements
			// and otherwise ignore them, since some databases do not
			// support "IF EXISTS".
			if IgnoreErr.Matches(s, err.Error()) {
				log.Tracef("migrate: ignored %s", err)
			} else {
				return err
			}
		}
	}

	return nil
}
