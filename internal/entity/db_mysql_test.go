package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestTextDefaultColumn_DefaultValue(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		value, ok := textDefaultColumn{defaultValue: "image"}.DefaultValue()
		assert.True(t, ok)
		assert.Equal(t, "image", value)
	})
	t.Run("Empty", func(t *testing.T) {
		value, ok := textDefaultColumn{}.DefaultValue()
		assert.True(t, ok)
		assert.Equal(t, "", value)
	})
}

func TestMysqlHexDefault(t *testing.T) {
	assert.Equal(t, []string{"x'696d616765'", "696d616765"}, mysqlHexDefault.FindStringSubmatch("x'696d616765'"))
	assert.Equal(t, []string{"X'7A7A'", "7A7A"}, mysqlHexDefault.FindStringSubmatch("X'7A7A'"))
	assert.Nil(t, mysqlHexDefault.FindStringSubmatch("'zz'"))
	assert.Nil(t, mysqlHexDefault.FindStringSubmatch("x'zz'"))
}

func TestMysqlMigrator_ColumnTypes(t *testing.T) {
	if DbDialect() != dsn.DialectMySQL {
		t.Skip("requires the MariaDB test database")
	}

	t.Run("Migrator", func(t *testing.T) {
		_, ok := Db().Migrator().(mysqlMigrator)
		assert.True(t, ok)
	})
	t.Run("BinaryDefaults", func(t *testing.T) {
		columns, err := Db().Migrator().ColumnTypes(&Photo{})
		require.NoError(t, err)

		defaults := make(map[string]string)

		for _, c := range columns {
			if value, ok := c.DefaultValue(); ok {
				defaults[c.Name()] = value
			}
		}

		assert.Equal(t, "image", defaults["photo_type"])
		assert.Equal(t, "zz", defaults["place_id"])
		assert.Equal(t, "Local", defaults["time_zone"])
	})
}
