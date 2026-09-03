package migrate

import (
	"testing"
	"time"

	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestVersion(t *testing.T) {
	dbDriver, _ := dsn.PhotoPrismTestToDriverDSN()
	dbDsn := testextras.TestDbDSN(dbDriver, "migrate")

	log := logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	db, err := gorm.Open(dsn.GormDrivers[dbDriver](dbDsn),
		&gorm.Config{
			Logger: logger.New(
				log,
				logger.Config{
					SlowThreshold:             time.Second,  // Slow SQL threshold
					LogLevel:                  logger.Error, // Log level
					IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
					ParameterizedQueries:      true,         // Don't include params in the SQL log
					Colorful:                  false,        // Disable color
				},
			),
		},
	)

	if err != nil || db == nil {
		if err != nil {
			t.Fatal(err)
		}

		return
	}

	sqldb, _ := db.DB()
	defer sqldb.Close()
	defer testextras.TestDbRemoveByName(dbDriver, "migrate")

	require.NoError(t, db.AutoMigrate(&Version{}))

	t.Run("TableName", func(t *testing.T) {
		assert.Equal(t, "versions", Version{}.TableName())
	})
	t.Run("Unknown", func(t *testing.T) {
		var version Version
		assert.True(t, version.Unknown())
		version = Version{Version: ""}
		assert.True(t, version.Unknown())
		version.Version = UnknownVersion.Version
		assert.True(t, version.Unknown())
		version.Version = "1.0.0"
		assert.False(t, version.Unknown())
	})
	t.Run("NeedsMigration", func(t *testing.T) {
		require.NoError(t, db.Where("1=1").Delete(&Version{}).Error)
		var version Version
		assert.True(t, UnknownVersion.NeedsMigration())
		assert.True(t, version.NeedsMigration())
		version = Version{
			Version:    "1.0.0",
			Edition:    "dev",
			MigratedAt: nil,
			CreatedAt:  time.Now(),
		}
		assert.True(t, version.NeedsMigration())
		now := time.Now()
		version.MigratedAt = &now
		version.CreatedAt = time.Time{}
		assert.True(t, version.NeedsMigration())
		version.CreatedAt = time.Now()
		version.MigratedAt = &time.Time{}
		assert.True(t, version.NeedsMigration())
		version.MigratedAt = &now
		assert.False(t, version.NeedsMigration())
	})
	t.Run("CreateTable", func(t *testing.T) {
		version := &Version{
			Version:    "1.0.0",
			Edition:    "dev",
			MigratedAt: nil,
		}
		require.NoError(t, version.CreateTable(db))
		assert.True(t, db.Migrator().HasTable(&Version{}))
	})
	t.Run("Create", func(t *testing.T) {
		version := &Version{
			Version:    "1.0.0",
			Edition:    "dev",
			MigratedAt: nil,
		}
		require.NoError(t, version.Create(db))
		t.Cleanup(func() {
			require.NoError(t, db.Unscoped().Delete(&version).Error)
		})
	})
	t.Run("FirstOrCreateVersion", func(t *testing.T) {
		assert.Equal(t, &UnknownVersion, FirstOrCreateVersion(db, nil))
		assert.Equal(t, &UnknownVersion, FirstOrCreateVersion(db, &Version{Version: ""}))

		version := FirstOrCreateVersion(db, &Version{
			Version:    "1.0.0",
			Edition:    "dev",
			MigratedAt: nil,
		})
		require.NotNil(t, version)
		assert.Equal(t, "1.0.0", version.Version)
		assert.Equal(t, "dev", version.Edition)
		assert.NotEqual(t, version.CreatedAt, time.Time{})
		assert.NotEqual(t, 1000, version.CreatedAt.Year())
		t.Cleanup(func() {
			require.NoError(t, db.Unscoped().Delete(&version).Error)
		})
	})
	t.Run("Find", func(t *testing.T) {
		version := &Version{
			Version:    "1.0.0",
			Edition:    "dev",
			MigratedAt: nil,
		}
		require.NoError(t, version.Create(db))
		t.Cleanup(func() {
			require.NoError(t, db.Unscoped().Delete(&version).Error)
		})

		v2 := &Version{
			Version:    "1.0.0",
			Edition:    "dev",
			MigratedAt: nil,
		}
		found := v2.Find(db)
		assert.NotNil(t, found)
	})
	t.Run("SaveVersion", func(t *testing.T) {
		version := FirstOrCreateVersion(db, &Version{
			Version:    "1.0.0",
			Edition:    "dev",
			MigratedAt: nil,
		})
		require.NotNil(t, version)
		t.Cleanup(func() {
			require.NoError(t, db.Unscoped().Delete(&version).Error)
		})
		now := time.Now()
		version.MigratedAt = &now
		beforeUpdated := version.UpdatedAt
		time.Sleep(2 * time.Second)
		require.NoError(t, version.Save(db))
		assert.NotEqual(t, beforeUpdated, version.UpdatedAt)
	})
}
