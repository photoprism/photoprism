package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/pkg/fs"
)

func TestDialectSQLite3(t *testing.T) {
	// Prepare temporary sqlite3 db.
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

	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	db, err := openTestDb(
		"sqlite3",
		dumpName,
	)

	if err != nil || db == nil {
		if err != nil {
			t.Fatal(err)
		}

		return
	}

	defer closeTestDb(db)

	opt := Opt(true, true, nil)

	// Run pre-migrations.
	if err = Run(db, opt.Pre()); err != nil {
		t.Error(err)
	}

	// Run migrations.
	if err = Run(db, opt); err != nil {
		t.Error(err)
	}

	stmt := db.Table("photos").Where("photo_caption = '' OR photo_caption IS NULL")

	count := int64(0)

	// Fetch count from database.
	if err = stmt.Count(&count).Error; err != nil {
		t.Error(err)
	} else {
		assert.EqualValues(t, 0, count)
	}
}
