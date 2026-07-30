package migrate

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDialectPostgreSQL(t *testing.T) {
	driver, _ := dsn.PhotoPrismTestToDriverDSN()
	if driver != "postgres" {
		t.Skip("skipping test as not PostgreSQL")
	}
	t.Run("Existing", func(t *testing.T) {
		dbDSN := testextras.TestDbDSN(dsn.DriverPostgreSQL, "migrate")

		if dumpName, err := filepath.Abs("./testdata/migrate_postgres.sql"); err != nil {
			t.Fatal(err)
		} else {
			// Clear Postgres source (migrate)
			if err := testextras.ResetPostgresDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
				t.Fatal(err)
			}
			pDSN := dsn.Parse(dbDSN)
			if err = exec.Command("psql", pDSN.ForPSQL(), "--file="+dumpName).Run(); err != nil {
				t.Fatal(err)
			}
		}

		log.Info("Expect potential table does not exist or no such table Error or SQLSTATE from migration.go")
		log = logrus.StandardLogger()
		log.SetLevel(logrus.TraceLevel)

		db, err := gorm.Open(postgres.Open(
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
			assert.Equal(t, int64(77), count)
		}

		count = int64(0)

		// Fetch count of collation name from database.
		if err = db.Table("pg_collation").Where("collname = ?", "caseinsensitive").Count(&count).Error; err != nil {
			t.Error(err)
		} else {
			assert.Equal(t, int64(1), count)
		}
		log.Info("End Expect potential table does not exist or no such table Error or SQLSTATE from migration.go")
	})

	t.Run("New", func(t *testing.T) {
		dbDSN := testextras.TestDbDSN(dsn.DriverPostgreSQL, "migrate")

		if dumpName, err := filepath.Abs("./testdata/migrate_postgres.sql"); err != nil {
			t.Fatal(err)
		} else {
			// Clear Postgres source (migrate)
			if err := testextras.ResetPostgresDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
				t.Fatal(err)
			}
			pDSN := dsn.Parse(dbDSN)
			if err = exec.Command("psql", pDSN.ForPSQL(), "--file="+dumpName).Run(); err != nil { //nolint:gosec // parameters created in test earlier
				t.Fatal(err)
			}
		}

		log.Info("Expect potential table does not exist or no such table Error or SQLSTATE from migration.go")
		log = logrus.StandardLogger()
		log.SetLevel(logrus.TraceLevel)

		db, err := gorm.Open(postgres.Open(
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
		opt.NewDatabase = true

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

		// Use the old column name as Pre and Main shouldn't run
		stmt := db.Table("photos").Where("photo_description = '' OR photo_description IS NULL")

		count := int64(0)

		// Fetch count from database.
		if err = stmt.Count(&count).Error; err != nil {
			t.Error(err)
		} else {
			assert.Equal(t, int64(77), count)
		}

		count = int64(0)

		// Fetch count of collation name from database.
		if err = db.Table("pg_collation").Where("collname = ?", "caseinsensitive").Count(&count).Error; err != nil {
			t.Error(err)
		} else {
			assert.Equal(t, int64(1), count)
		}
		log.Info("End Expect potential table does not exist or no such table Error or SQLSTATE from migration.go")
	})

}
