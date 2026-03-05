package performancetest

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func sqliteMigration(original string, temp string, numberOfRecords int, skipSpeedup bool, testname string, expectedDuration time.Duration, b *testing.B) {

	b.StopTimer()
	// Prepare temporary sqlite db.
	testDbOriginal := original
	testDbTemp := temp
	dumpName, err := filepath.Abs(testDbTemp)
	_ = os.Remove(dumpName)
	if err != nil {
		b.Fatal(err)
	} else if err = fs.Copy(testDbOriginal, dumpName, true); err != nil {
		b.Fatal(err)
	}
	defer os.Remove(dumpName)
	b.StartTimer()

	log = logrus.StandardLogger()
	log.SetLevel(logrus.ErrorLevel)

	start := time.Now()
	dbDSN := dsn.DSN{Driver: dsn.DriverSQLite3, Server: filepath.Dir(dumpName), Name: filepath.Base(dumpName)}

	db, err := gorm.Open(sqlite.Open(dbDSN.ToString()),
		&gorm.Config{
			Logger: logger.New(
				log,
				logger.Config{
					SlowThreshold:             time.Second,  // Slow SQL threshold
					LogLevel:                  logger.Error, // Log level
					IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
					ParameterizedQueries:      false,        // Don't include params in the SQL log
					Colorful:                  false,        // Disable color
				},
			),
		},
	)

	if err != nil || db == nil {
		if err != nil {
			b.Fatal(err)
		}

		return
	}

	sqldb, _ := db.DB()
	defer sqldb.Close()

	opt := migrate.Opt(true, true, nil)

	// Make sure that migrate and version is done, as the Once doesn't work as it has already been set before we opened the new database..
	if err = db.AutoMigrate(&migrate.Migration{}); err != nil {
		b.Fatal(err)
	}
	if err = db.AutoMigrate(&migrate.Version{}); err != nil {
		b.Fatal(err)
	}

	if skipSpeedup {
		// Skip the Gorm Migration Speedup.
		version := migrate.FirstOrCreateVersion(db, migrate.NewVersion("Gorm For SQLite", "V2 Upgrade"))
		require.NoError(b, version.Migrated(db))
	}

	// Setup and capture SQL Logging output
	buffer := bytes.Buffer{}
	log.SetOutput(&buffer)

	entity.Entities.Migrate(db, opt)
	// The bad thing is that the above panics, but doesn't return an error.

	// Reset logger
	log.SetOutput(os.Stdout)

	// Expect 0 errors (no such table accounts, and missing account_id in files_sync and files_share)
	// And a blank record.
	assert.Equal(b, 1, len(strings.Split(buffer.String(), "\n")))
	if len(strings.Split(buffer.String(), "\n")) == 1 {
		assert.Equal(b, 0, len(strings.Split(buffer.String(), "\n")[0]))
	} else {
		log.Error("Migration result not as expected.  Results follow:")
		for i := 0; i < len(strings.Split(buffer.String(), "\n")); i++ {
			log.Error(strings.Split(buffer.String(), "\n")[i])
		}
	}

	elapsed := time.Since(start)

	stmt := db.Table("photos").Where("photo_uid IS NOT NULL")

	count := int64(0)

	// Fetch count from database.
	if err = stmt.Count(&count).Error; err != nil {
		b.Error(err)
	} else {
		assert.Equal(b, int64(numberOfRecords), count)
	}

	log.Info(testname, " sqlite took ", elapsed)
	assert.LessOrEqual(b, elapsed, expectedDuration)
}

func mysqlMigration(testDbOriginal string, numberOfRecords int, testname string, expectedDuration time.Duration, b *testing.B) {
	b.StopTimer()
	// Prepare migrate mariadb db.
	if dumpName, err := filepath.Abs(testDbOriginal); err != nil {
		b.Fatal(err)
	} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", "migrate", //nolint:gosec // test generated input
		"-e", "source "+dumpName).Run(); err != nil {
		b.Fatal(err)
	}

	b.StartTimer()
	start := time.Now()

	log = logrus.StandardLogger()
	log.SetLevel(logrus.ErrorLevel)

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
			b.Fatal(err)
		}

		return
	}

	sqldb, _ := db.DB()
	defer sqldb.Close()

	opt := migrate.Opt(true, true, nil)

	// Make sure that migrate and version is done, as the Once doesn't work as it has already been set before we opened the new database..
	if err = db.AutoMigrate(&migrate.Migration{}); err != nil {
		b.Fatal(err)
	}
	if err = db.AutoMigrate(&migrate.Version{}); err != nil {
		b.Fatal(err)
	}

	// Setup and capture SQL Logging output
	buffer := bytes.Buffer{}
	log.SetLevel(logrus.TraceLevel)
	log.SetOutput(&buffer)

	entity.Entities.Migrate(db, opt)
	// The bad thing is that the above panics, but doesn't return an error.

	// Reset logger
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.ErrorLevel)

	// Expect 3 errors (no such table accounts, and missing account_id in files_sync and files_share)
	// And a blank record.
	assert.Equal(b, 4, len(strings.Split(buffer.String(), "\n")))
	if len(strings.Split(buffer.String(), "\n")) == 4 {
		assert.Equal(b, 0, len(strings.Split(buffer.String(), "\n")[3]))
	} else {
		log.Error("Migration result not as expected.  Results follow:")
		for i := 0; i < len(strings.Split(buffer.String(), "\n")); i++ {
			log.Error(strings.Split(buffer.String(), "\n")[i])
		}
	}

	elapsed := time.Since(start)

	stmt := db.Table("photos").Where("photo_uid IS NOT NULL")

	count := int64(0)

	// Fetch count from database.
	if err = stmt.Count(&count).Error; err != nil {
		b.Error(err)
	} else {
		assert.Equal(b, int64(numberOfRecords), count)
	}

	log.Info(testname, " mysql took ", elapsed)
	assert.LessOrEqual(b, elapsed, expectedDuration)
}

func postgresqlMigration(testDbOriginal string, numberOfRecords int, testname string, expectedDuration time.Duration, b *testing.B) {
	b.StopTimer()
	postgresqlDSN := "postgresql://migrate:migrate@postgres:5432/migrate" //nolint:gosec // test specific credentials
	postgresqlParams := "?TimeZone=UTC&connect_timeout=15&lock_timeout=5000&sslmode=disable"

	// Prepare migrate PostgreSQL db.
	if dumpName, err := filepath.Abs(testDbOriginal); err != nil {
		b.Fatal(err)
	} else if err = exec.Command("dropdb", "--maintenance-db=postgresql://photoprism:photoprism@postgres:5432/postgres", "--force", "--if-exists", "migrate").Run(); err != nil {
		b.Fatal(err)
	} else if err = exec.Command("createdb", "--maintenance-db=postgresql://photoprism:photoprism@postgres:5432/postgres", "-O", "migrate", "-T", "template0", "migrate").Run(); err != nil {
		b.Fatal(err)
	} else if err = exec.Command("pg_restore", "-d", postgresqlDSN, dumpName).Run(); err != nil { //nolint:gosec // test generated parameters
		b.Fatal(err)
	}

	b.StartTimer()
	start := time.Now()

	log = logrus.StandardLogger()
	log.SetLevel(logrus.ErrorLevel)

	db, err := gorm.Open(postgres.Open(
		postgresqlDSN+postgresqlParams),
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
			b.Fatal(err)
		}

		return
	}

	sqldb, _ := db.DB()
	defer sqldb.Close()

	opt := migrate.Opt(true, true, nil)

	// Make sure that migrate and version is done, as the Once doesn't work as it has already been set before we opened the new database..
	if err = db.AutoMigrate(&migrate.Migration{}); err != nil {
		b.Fatal(err)
	}
	if err = db.AutoMigrate(&migrate.Version{}); err != nil {
		b.Fatal(err)
	}

	// Setup and capture SQL Logging output
	buffer := bytes.Buffer{}
	log.SetLevel(logrus.TraceLevel)
	log.SetOutput(&buffer)

	entity.Entities.Migrate(db, opt)
	// The bad thing is that the above panics, but doesn't return an error.

	// Reset logger
	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.ErrorLevel)

	// Expect 3 errors (no such table accounts, and missing account_id in files_sync and files_share)
	// And a blank record.
	assert.Equal(b, 4, len(strings.Split(buffer.String(), "\n")))
	if len(strings.Split(buffer.String(), "\n")) == 4 {
		assert.Equal(b, 0, len(strings.Split(buffer.String(), "\n")[3]))
	} else {
		log.Error("Migration result not as expected.  Results follow:")
		for i := 0; i < len(strings.Split(buffer.String(), "\n")); i++ {
			log.Error(strings.Split(buffer.String(), "\n")[i])
		}
	}

	elapsed := time.Since(start)

	stmt := db.Table("photos").Where("photo_uid IS NOT NULL")

	count := int64(0)

	// Fetch count from database.
	if err = stmt.Count(&count).Error; err != nil {
		b.Error(err)
	} else {
		assert.Equal(b, int64(numberOfRecords), count)
	}

	log.Info(testname, " postgresql took ", elapsed)
	assert.LessOrEqual(b, elapsed, expectedDuration)
}
