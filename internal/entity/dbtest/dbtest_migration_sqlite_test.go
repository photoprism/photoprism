package entity

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestDialectSQLite3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if entity.DbDialect() != entity.SQLite3 {
		t.Skip("skipping test as not SQLite")
	}

	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()
	log.Info("Expect many table does not exist or no such table Error or SQLSTATE from migration.go")
	t.Run("ValidMigration", func(t *testing.T) {
		// Prepare temporary sqlite db.
		testDbOriginal := "../migrate/testdata/migrate_sqlite3"
		testDbTemp := "../migrate/testdata/migrate_sqlite3.db"
		if !fs.FileExists(testDbOriginal) {
			t.Fatal(testDbOriginal + " not found")
		}
		dumpName, err := filepath.Abs(testDbTemp)
		_ = os.Remove(dumpName)
		if err != nil {
			t.Fatal(err)
		} else if err = fs.Copy(testDbOriginal, dumpName, true); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(dumpName)

		log = logrus.StandardLogger()
		log.SetLevel(logrus.TraceLevel)

		dbDSN := dsn.DSN{Driver: dsn.DriverSQLite3, Server: filepath.Dir(dumpName), Name: filepath.Base(dumpName)}

		db, err := gorm.Open(sqlite.Open(dbDSN.ToString()),
			&gorm.Config{
				Logger: logger.New(
					log,
					logger.Config{
						SlowThreshold:             time.Second,  // Slow SQL threshold
						LogLevel:                  logger.Error, // Log level
						IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
						ParameterizedQueries:      false,        // Don't include params in the SQL log
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

		for i := 0; i < len(strings.Split(buffer.String(), "\n")); i++ {
			log.Info(strings.Split(buffer.String(), "\n")[i])
		}

		// Expect 5 errors (auth_? tables missing)
		// And a blank record.
		assert.Equal(t, 6, len(strings.Split(buffer.String(), "\n")))
		assert.Equal(t, 0, len(strings.Split(buffer.String(), "\n")[5]))

		stmt := db.Table("photos").Where("photo_caption = '' OR photo_caption IS NULL")

		count := int64(0)

		// Fetch count from database.
		if err = stmt.Count(&count).Error; err != nil {
			t.Error(err)
		} else {
			assert.Equal(t, int64(0), count)
		}

		// Test that the AutoIncrements are working
		populatePhotoPrismStructsWithAutoIncrement(t, db)

		// Test that the minimum values can be added to the database
		populatePhotoPrismStructsWithMin(t, db)

		// Test that the maximum values can be added to the database
		populatePhotoPrismStructsWithMax(t, db)

	})

	t.Run("InvalidDataUpgrade", func(t *testing.T) {
		// Prepare temporary sqlite db.
		testDbOriginal := "../migrate/testdata/migrate_sqlite3"
		testDbTemp := "../migrate/testdata/migrate_sqlite3.db"
		if !fs.FileExists(testDbOriginal) {
			t.Fatal(testDbOriginal + " not found")
		}
		dumpName, err := filepath.Abs(testDbTemp)
		_ = os.Remove(dumpName)
		if err != nil {
			t.Fatal(err)
		} else if err = fs.Copy(testDbOriginal, dumpName, true); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(dumpName)

		log = logrus.StandardLogger()
		log.SetLevel(logrus.TraceLevel)

		dbDSN := dsn.DSN{Driver: dsn.DriverSQLite3, Server: filepath.Dir(dumpName), Name: filepath.Base(dumpName)}

		db, err := gorm.Open(sqlite.Open(dbDSN.ToString()),
			&gorm.Config{
				Logger: logger.New(
					log,
					logger.Config{
						SlowThreshold:             time.Second,  // Slow SQL threshold
						LogLevel:                  logger.Error, // Log level
						IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
						ParameterizedQueries:      false,        // Don't include params in the SQL log
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

		// Load some invalid data into the database
		AlbumUID := byte('a')
		PhotoUID := byte('p')

		_, err = sqldb.Exec("INSERT INTO photos_albums (photo_uid, album_uid, `order`, hidden, missing, created_at, updated_at) VALUES (?, ?, '0', '0', '0', ?, ?)", rnd.GenerateUID(PhotoUID), rnd.GenerateUID(AlbumUID), time.Now().UTC(), time.Now().UTC())
		assert.Nil(t, err)

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

		// Expect 5 errors (auth_? tables missing)
		// And a blank record.
		assert.Equal(t, 6, len(strings.Split(buffer.String(), "\n")))
		assert.Equal(t, 0, len(strings.Split(buffer.String(), "\n")[5]))

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
