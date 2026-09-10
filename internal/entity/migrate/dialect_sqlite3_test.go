package migrate

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestDialectSQLite3(t *testing.T) {
	driver, _ := dsn.PhotoPrismTestToDriverDSN()
	if driver != "sqlite" {
		t.Skip("skipping test as not SQLite")
	}
	// Prepare temporary sqlite db.
	testDbOriginal := "./testdata/migrate_sqlite3"
	testDbTemp := filepath.Join(t.TempDir(), "migrate_sqlite3.db")
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

	log.Info("Expect many table does not exist or no such table Error or SQLSTATE from migration.go")
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
					ParameterizedQueries:      true,         // Don't include params in the SQL log
					Colorful:                  false,        // Disable color
				},
			),
		},
	)

	// Enable Foreign Keys on sqlite
	if db.Dialector.Name() == dsn.DialectSQLite {
		db.Exec("PRAGMA foreign_keys = ON")
		log.Info("sqlite foreign keys enabled")
	}

	if err != nil || db == nil {
		if err != nil {
			t.Fatal(err)
		}

		return
	}

	sqldb, _ := db.DB()
	defer sqldb.Close()

	opt := Opt(true, true, nil)

	// Run pre-migrations.
	if err = Run(db, opt.Pre()); err != nil {
		t.Error(err)
	}

	// Run migrations.
	if err = Run(db, opt); err != nil {
		t.Error(err)
	}

	// Run post-migrations.
	if err = Run(db, opt.Post()); err != nil {
		t.Error(err)
	}

	stmt := db.Table("photos").Where("photo_caption = '' OR photo_caption IS NULL")

	count := int64(0)

	// Fetch count from database.
	if err = stmt.Count(&count).Error; err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, int64(0), count)
	}
	log.Info("End Expect many table does not exist or no such table Error or SQLSTATE from migration.go")
}
