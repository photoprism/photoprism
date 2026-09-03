package entity

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestDialectMysql(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	if entity.DbDialect() != dsn.DialectMySQL {
		t.Skip("skipping test as not MariaDB")
	}

	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()
	log.Info("Expect many table does not exist or no such table Error or SQLSTATE from migration.go")
	t.Run("ValidMigration", func(t *testing.T) {
		dbDSN := testextras.TestDbDSN(dsn.DriverMariaDB, "migrate")

		// Prepare migrate mariadb db.
		dumpName, err := filepath.Abs("../migrate/testdata/migrate_mysql.sql")
		if err != nil {
			t.Fatal(err)
		} else if err = testextras.ResetMariaDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
			t.Fatal(err)
		}

		//nolint:gosec // G204: dumpName comes from a fixed local fixture path in testdata.
		if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", dsn.Parse(dbDSN).Name,
			"-e", "source "+dumpName).Run(); err != nil {
			t.Fatal(err)
		}

		log = logrus.StandardLogger()
		log.SetLevel(logrus.TraceLevel)

		db, err := gorm.Open(mysql.Open(
			dbDSN),
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

		assert.Contains(t, buffer.String(), fmt.Sprintf("Table '%s.auth_sessions' doesn't exist", dsn.Parse(dbDSN).Name))
		assert.Contains(t, buffer.String(), fmt.Sprintf("Table '%s.auth_users_details' doesn't exist", dsn.Parse(dbDSN).Name))
		assert.Contains(t, buffer.String(), fmt.Sprintf("Table '%s.auth_users_settings' doesn't exist", dsn.Parse(dbDSN).Name))
		assert.Contains(t, buffer.String(), fmt.Sprintf("Table '%s.auth_users_shares' doesn't exist", dsn.Parse(dbDSN).Name))
		// There is a blank record.
		numberOfErrorsExpected := 7
		assert.Equal(t, numberOfErrorsExpected, len(strings.Split(buffer.String(), "\n")))
		if len(strings.Split(buffer.String(), "\n")) != numberOfErrorsExpected {
			for i := 0; i < len(strings.Split(buffer.String(), "\n")); i++ {
				assert.Empty(t, strings.Split(buffer.String(), "\n")[i])
			}
		}
		// Detect a foreign key issue
		assert.NotContains(t, buffer.String(), "errno: 150")

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
		dbDSN := testextras.TestDbDSN(dsn.DriverMariaDB, "migrate")

		// Prepare migrate mariadb db.
		dumpName, err := filepath.Abs("../migrate/testdata/migrate_mysql.sql")
		if err != nil {
			t.Fatal(err)
		} else if err = testextras.ResetMariaDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
			t.Fatal(err)
		}

		//nolint:gosec // G204: dumpName comes from a fixed local fixture path in testdata.
		if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", dsn.Parse(dbDSN).Name,
			"-e", "source "+dumpName).Run(); err != nil {
			t.Fatal(err)
		}

		log = logrus.StandardLogger()
		log.SetLevel(logrus.TraceLevel)

		db, err := gorm.Open(mysql.Open(
			dbDSN),
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

		assert.Contains(t, buffer.String(), fmt.Sprintf("Table '%s.auth_sessions' doesn't exist", dsn.Parse(dbDSN).Name))
		assert.Contains(t, buffer.String(), fmt.Sprintf("Table '%s.auth_users_details' doesn't exist", dsn.Parse(dbDSN).Name))
		assert.Contains(t, buffer.String(), fmt.Sprintf("Table '%s.auth_users_settings' doesn't exist", dsn.Parse(dbDSN).Name))
		assert.Contains(t, buffer.String(), fmt.Sprintf("Table '%s.auth_users_shares' doesn't exist", dsn.Parse(dbDSN).Name))
		// There is a blank record.
		numberOfErrorsExpected := 7
		assert.Equal(t, numberOfErrorsExpected, len(strings.Split(buffer.String(), "\n")))
		if len(strings.Split(buffer.String(), "\n")) != numberOfErrorsExpected {
			for i := 0; i < len(strings.Split(buffer.String(), "\n")); i++ {
				assert.Empty(t, strings.Split(buffer.String(), "\n")[i])
			}
		}

		stmt := db.Table("photos").Where("photo_caption = '' OR photo_caption IS NULL")

		count := int64(0)

		// Fetch count from database.
		if err = stmt.Count(&count).Error; err != nil {
			t.Error(err)
		} else {
			assert.Equal(t, int64(0), count)
		}
	})
	t.Run("AutoMigrationFailure", func(t *testing.T) {
		require.NoError(t, entity.Db().AutoMigrate(&entity.File{}))
		var results search.PhotoResults
		assert.NoError(t, entity.Db().Raw("SELECT ROW_NUMBER() OVER (PARTITION BY photos.id, files.id ORDER BY files.media_id) as rec_num, photos.id FROM files JOIN photos ON files.photo_id = photos.id AND files.media_id IS NOT NULL").Scan(&results).Error)
		// Test to make sure that after using search.Photo (via search.PhotoResults) that no errors happen in the AutoMigrate
		assert.NoError(t, entity.Db().AutoMigrate(&entity.File{}))
	})

	log.Info("End Expect many table does not exist or no such table Error or SQLSTATE from migration.go")
}
