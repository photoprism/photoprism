package migrate

import (
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

	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestGorm2Migration(t *testing.T) {
	driver, _ := dsn.PhotoPrismTestToDriverDSN(0)
	if driver != "sqlite" {
		t.Skip("skipping test as not SQLite")
	}
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

	require.NoError(t, ConvertSQLiteDataTypes(db))

	type ResultSQL struct {
		SQL string
	}
	var createstatement ResultSQL
	db.Raw("SELECT sql FROM sqlite_master WHERE type = 'table' AND tbl_name = 'photos' AND name = 'photos';").Scan(&createstatement)
	cs := strings.ToLower(createstatement.SQL)
	assert.NotContains(t, cs, "varchar")
	assert.NotContains(t, cs, "varbinary")
	assert.NotContains(t, cs, "mediumblob")
	assert.NotContains(t, cs, "bigint")
	assert.NotContains(t, cs, "bool")
	assert.NotContains(t, cs, "float")
}
