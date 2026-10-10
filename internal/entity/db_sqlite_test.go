package entity

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm/schema"
)

// sqliteTypeTest is a model for testing SQLite column types.
type sqliteTypeTest struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"size:160"`
	Note    string
	Slug    string `gorm:"type:bytes;size:42"`
	Data    []byte `gorm:"type:bytes;size:16777215"`
	Text    string `gorm:"size:4294967295"`
	Count   int    `gorm:"size:32"`
	Enabled bool
}

// TableName returns the entity table name.
func (sqliteTypeTest) TableName() string {
	return "type_tests"
}

func TestSqliteDialector_DataTypeOf(t *testing.T) {
	s, err := schema.Parse(&sqliteTypeTest{}, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)

	d := sqliteDialector{Dialector: &sqlite.Dialector{}}

	assert.Equal(t, "varchar(160)", d.DataTypeOf(s.LookUpField("Name")))
	assert.Equal(t, "varchar(255)", d.DataTypeOf(s.LookUpField("Note")))
	assert.Equal(t, "blob(0)", d.DataTypeOf(s.LookUpField("Slug")))
	assert.Equal(t, "blob", d.DataTypeOf(s.LookUpField("Data")))
	assert.Equal(t, "text", d.DataTypeOf(s.LookUpField("Text")))
	assert.Equal(t, "integer", d.DataTypeOf(s.LookUpField("Count")))
	assert.Equal(t, "numeric", d.DataTypeOf(s.LookUpField("Enabled")))
}

func TestSqliteDialector_Migrator(t *testing.T) {
	db := openTestDb(t)
	require.NoError(t, db.AutoMigrate(&sqliteTypeTest{}))

	var ddl string
	require.NoError(t, db.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?", "type_tests").Row().Scan(&ddl))
	assert.Contains(t, ddl, "`name` varchar(160)")
	assert.Contains(t, ddl, "`note` varchar(255)")
	assert.Contains(t, ddl, "`slug` blob(0)")

	// A second migration finds nothing to change.
	counter := &statementCounter{}
	db.Logger = counter
	require.NoError(t, db.AutoMigrate(&sqliteTypeTest{}))

	for _, stmt := range counter.sql {
		assert.NotContains(t, stmt, "ALTER TABLE", stmt)
		assert.NotContains(t, stmt, "CREATE TABLE", stmt)
	}
}
