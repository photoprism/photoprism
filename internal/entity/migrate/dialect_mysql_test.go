package migrate

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sirupsen/logrus"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDialectMysql(t *testing.T) {
	if dumpName, err := filepath.Abs("./testdata/migrate_mysql.sql"); err != nil {
		t.Fatal(err)
	} else if err = exec.Command("mysql", "-u", "migrate", "-pmigrate", "migrate",
		"-e", "source "+dumpName).Run(); err != nil {
		t.Fatal(err)
	}

	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	db, err := gorm.Open(mysql.Open(
		"migrate:migrate@tcp(mariadb:4001)/migrate?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true"),
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

	stmt := db.Table("photos").Where("photo_description = '' OR photo_description IS NULL")

	count := int64(0)

	// Fetch count from database.
	if err = stmt.Count(&count).Error; err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, int64(0), count)
	}
}
