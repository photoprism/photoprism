package migrate

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDialectMysql(t *testing.T) {
	driver, _ := dsn.PhotoPrismTestToDriverDSN()
	if driver != "mysql" {
		t.Skip("skipping test as not MariaDB")
	}

	dbDSN := testextras.TestDbDSN(dsn.DriverMySQL, "migrate")
	dumpName, err := filepath.Abs("./testdata/migrate_mysql.sql")
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

	log.Info("Expect many table does not exist or no such table Error or SQLSTATE from migration.go")
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	port := os.Getenv("MARIADB_PORT")
	if port == "" {
		port = "4001"
	}

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
