package entity

import (
	"testing"
	"time"
)

// TestEntity is an entity dedicated to test database management functionality.
type TestEntity struct {
	ID        string    `gorm:"type:VARBINARY(42);primary_key;auto_increment:false;" json:"TestID" yaml:"TestID"`
	TestLabel string    `gorm:"type:VARCHAR(400);unique_index;" json:"Label" yaml:"Label"`
	TestCount int       `gorm:"default:1" json:"Count" yaml:"-"`
	CreatedAt time.Time `json:"CreatedAt" yaml:"-"`
	UpdatedAt time.Time `json:"UpdatedAt" yaml:"-"`
}

// TableName returns the entity database table name.
func (TestEntity) TableName() string {
	return "test_ignore"
}

func TestRecreateTable(t *testing.T) {
	t.Run("TestEntity", func(t *testing.T) {
		if err := Db().CreateTable(TestEntity{}).Error; err != nil {
			t.Fatal(err)
		}

		if err := RecreateTable(TestEntity{}); err != nil {
			t.Fatal(err)
		}
	})
}
