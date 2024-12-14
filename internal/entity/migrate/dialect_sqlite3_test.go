package migrate

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

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestDialectSQLite3(t *testing.T) {
	// Prepare temporary sqlite db.
	testDbOriginal := "./testdata/migrate_sqlite3"
	testDbTemp := "./testdata/migrate_sqlite3.db"
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
					SlowThreshold:             time.Second,   // Slow SQL threshold
					LogLevel:                  logger.Silent, // Log level
					IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
					ParameterizedQueries:      true,          // Don't include params in the SQL log
					Colorful:                  false,         // Disable color
				},
			),
		},
	)

	// Enable Foreign Keys on sqlite
	if db.Dialector.Name() == SQLite3 {
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

	stmt := db.Table("photos").Where("photo_description = '' OR photo_description IS NULL")

	count := int64(0)

	// Fetch count from database.
	if err = stmt.Count(&count).Error; err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, int64(0), count)
	}
}
