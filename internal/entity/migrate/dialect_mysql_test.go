package migrate

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestDialectMysql(t *testing.T) {
	if dumpName, err := filepath.Abs("./testdata/migrate_mysql.sql"); err != nil {
		t.Fatal(err)
	} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", "migrate",
		"-e", "source "+dumpName).Run(); err != nil {
		t.Fatal(err)
	}

	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	db, err := gorm.Open(
		"mysql",
		"migrate:migrate@tcp(mariadb:4001)/migrate?charset=utf8mb4,utf8&collation=utf8mb4_unicode_ci&parseTime=true",
	)

	if err != nil || db == nil {
		if err != nil {
			t.Fatal(err)
		}

		return
	}

	defer db.Close()

	db.LogMode(false)
	db.SetLogger(log)

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

	count := 0

	// Fetch count from database.
	if err = stmt.Count(&count).Error; err != nil {
		t.Error(err)
	} else {
		assert.Equal(t, 0, count)
	}
}
