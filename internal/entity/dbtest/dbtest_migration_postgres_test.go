package entity

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestDialectPostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if entity.DbDialect() != dsn.DialectPostgreSQL {
		t.Skip("skipping test as not PostgreSQL")
	}

	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()
	log.Info("Expect many table does not exist or no such table Error or SQLSTATE from migration.go")
	t.Run("ValidMigration", func(t *testing.T) {
		dbDSN := dsn.TestDSNPortFromEnv(dsn.DriverPostgreSQL, "migrate", testextras.GetDBMutexID())

		if dumpName, err := filepath.Abs("../migrate/testdata/migrate_postgres.sql"); err != nil {
			t.Fatal(err)
		} else {
			// Clear Postgres source (migrate)
			if err := testextras.ResetPostgresDB("migrate", testextras.GetDBMutexID()); err != nil {
				t.Fatal(err)
			}

			if err = exec.Command("psql", dbDSN.ForPSQL(), "--file="+dumpName).Run(); err != nil { //nolint:gosec // test generated input
				t.Fatal(err)
			}
		}

		log = logrus.StandardLogger()
		log.SetLevel(logrus.TraceLevel)

		db, err := gorm.Open(postgres.Open(
			dbDSN.ToString()),
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

		opt := migrate.Opt(true, true, nil)

		// Make sure that migrate and version is done, as the Once doesn't work as it has already been set before we opened the new database..
		if err = db.AutoMigrate(&migrate.Migration{}); err != nil {
			t.Error(err)
		}
		if err = db.AutoMigrate(&migrate.Version{}); err != nil {
			t.Error(err)
		}

		// Setup and capture SQL Logging output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)

		entity.Entities.Migrate(db, opt)
		// The bad thing is that the above panics, but doesn't return an error.

		// Reset logger
		log.SetOutput(os.Stdout)

		// There is a blank record.
		assert.Equal(t, 1, len(strings.Split(buffer.String(), "\n")))
		if len(strings.Split(buffer.String(), "\n")) != 1 {
			for i := 0; i < len(strings.Split(buffer.String(), "\n")); i++ {
				assert.Empty(t, strings.Split(buffer.String(), "\n")[i])
			}
		}
		// Detect a foreign key issue
		assert.NotContains(t, buffer.String(), "23503")

		stmt := db.Table("photos").Where("photo_caption <> '' AND photo_caption IS NOT NULL")

		count := int64(0)

		// Fetch count from database.
		if err = stmt.Count(&count).Error; err != nil {
			t.Error(err)
		} else {
			assert.Equal(t, int64(3), count)
		}

		// Test that the AutoIncrements are working
		populatePhotoPrismStructsWithAutoIncrement(t, db)

		// Test that the minimum values can be added to the database
		populatePhotoPrismStructsWithMin(t, db)

		// Test that the maximum values can be added to the database
		populatePhotoPrismStructsWithMax(t, db)

	})

	t.Run("EmptyDB", func(t *testing.T) {
		dbDSN := dsn.TestDSNPortFromEnv(dsn.DriverPostgreSQL, "migrate", testextras.GetDBMutexID())

		// Clear Postgres source (migrate)
		if err := testextras.ResetPostgresDB("migrate", testextras.GetDBMutexID()); err != nil {
			t.Fatal(err)
		}

		log = logrus.StandardLogger()
		log.SetLevel(logrus.TraceLevel)

		db, err := gorm.Open(postgres.Open(
			dbDSN.ToString()),
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

		opt := migrate.Opt(true, true, nil)

		// Make sure that migrate and version is done, as the Once doesn't work as it has already been set before we opened the new database..
		require.NoError(t, db.AutoMigrate(&migrate.Migration{}))
		require.NoError(t, db.AutoMigrate(&migrate.Version{}))

		// Setup and capture SQL Logging output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)

		entity.Entities.Migrate(db, opt)
		// The bad thing is that the above panics, but doesn't return an error.

		// Reset logger
		log.SetOutput(os.Stdout)

		// There is a blank record.
		assert.Equal(t, 1, len(strings.Split(buffer.String(), "\n")))
		if len(strings.Split(buffer.String(), "\n")) != 1 {
			for i := 0; i < len(strings.Split(buffer.String(), "\n")); i++ {
				assert.Empty(t, strings.Split(buffer.String(), "\n")[i])
			}
		}
		// Detect a foreign key issue
		assert.NotContains(t, buffer.String(), "23503")

		stmt := db.Table("photos").Where("photo_caption = '' OR photo_caption IS NULL")

		count := int64(0)

		// Fetch count from database.
		if err = stmt.Count(&count).Error; err != nil {
			t.Error(err)
		} else {
			assert.Equal(t, int64(0), count)
		}
	})
	log.Info("End Expect many table does not exist or no such table Error or SQLSTATE from migration.go")
}
