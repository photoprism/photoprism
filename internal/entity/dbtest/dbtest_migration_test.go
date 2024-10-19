package entity

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestDialectSQLite3(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	dbtestMutex.Lock()
	defer dbtestMutex.Unlock()

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
		} else if err = fs.Copy(testDbOriginal, dumpName); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(dumpName)

		log = logrus.StandardLogger()
		log.SetLevel(logrus.TraceLevel)

		dsn := fmt.Sprintf("%v?_foreign_keys=on&_busy_timeout=5000", dumpName)

		db, err := gorm.Open(sqlite.Open(dsn),
			&gorm.Config{
				Logger: logger.New(
					log,
					logger.Config{
						SlowThreshold:             time.Second, // Slow SQL threshold
						LogLevel:                  logger.Info, // Log level
						IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
						ParameterizedQueries:      false,       // Don't include params in the SQL log
						Colorful:                  false,       // Disable color
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

		// INSERT INTO "photos_albums" ("photo_uid", "album_uid", "order", "hidden", "missing", "created_at", "updated_at") VALUES ('pksx180k4pog19ig', 'aksx1801bih9b9w1', '0', '0', '0', '2024-10-07 10:23:25', '2024-10-07 10:23:25');

		opt := migrate.Opt(true, true, nil)

		// Make sure that migrate and version is done, as the Once doesn't work as it has already been set before we opened the new database..
		err = db.AutoMigrate(&migrate.Migration{})
		err = db.AutoMigrate(&migrate.Version{})

		// Run pre-migrations.
		if err = migrate.Run(db.Debug(), opt.Pre()); err != nil {
			t.Error(err)
		}

		entity.Entities.Migrate(db, opt)

		// Run migrations.
		if err = migrate.Run(db, opt); err != nil {
			t.Error(err)
		}

		stmt := db.Table("photos").Where("photo_description = '' OR photo_description IS NULL")

		count := int64(0)

		// Fetch count from database.
		if err = stmt.Count(&count).Error; err != nil {
			t.Error(err)
		} else {
			assert.Equal(t, int64(0), count)
		}
	})

}
