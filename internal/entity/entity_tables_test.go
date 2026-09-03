package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity/migrate"
)

// countTestRows returns the number of rows in the table with the specified name.
func countTestRows(t *testing.T, db *gorm.DB, name string) int64 {
	t.Helper()

	var count int64

	require.NoError(t, db.Table(name).Count(&count).Error)

	return count
}

func TestTables_Truncate(t *testing.T) {
	t.Run("KeepsSchemaTables", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}

		migrations := migrate.Migration{}.TableName()
		before := countTestRows(t, UnscopedDb(), migrations)
		require.NotZero(t, before, "migrations must not be empty")

		Tables{
			10:   {migrate.Migration{}.TableName(), &migrate.Migration{}},
			20:   {migrate.Version{}.TableName(), &migrate.Version{}},
			4020: {Details{}.TableName(), &Details{}},
		}.Truncate(UnscopedDb())

		if !assert.Equal(t, before, countTestRows(t, UnscopedDb(), migrations)) {
			t.Fatal("Truncate did something REALLY bad")
		}
		ResetTestFixtures()
	})
}
