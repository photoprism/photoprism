package performancetest

import (
	"os"
	"path/filepath"

	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"

	"github.com/photoprism/photoprism/internal/event"

	"github.com/photoprism/photoprism/pkg/fs"
)

func Benchmark1k_SQLite(b *testing.B) {
	runSQLitePerfTest(1000, "1k", b)
}

func Benchmark10k_SQLite(b *testing.B) {
	runSQLitePerfTest(10000, "10k", b)
}

func Benchmark100k_SQLite(b *testing.B) {
	// Skip this as the setup and storage are excessive.
	if _, ok := os.LookupEnv("BENCH_GLACIAL"); !ok {
		b.Skip("skipping benchmark as BENCH_GLACIAL not set")
	} else {
		runSQLitePerfTest(100000, "100k", b)
	}
}

func runSQLitePerfTest(rows int, numberk string, b *testing.B) {
	// Setup here
	loglevel := event.Log.GetLevel()
	event.Log.SetLevel(logrus.ErrorLevel)
	testDbOriginal := "../../storage/test-" + numberk + ".original.sqlite"

	if !fs.FileExists(testDbOriginal) {
		log.Infof("Generating SQLite database with %d records", rows)
		require.NoError(b, generateDatabase(rows, "sqlite", testDbOriginal, true, true))
	}

	// Prepare temporary sqlite db.
	testDbTemp := "../../storage/test-" + numberk + ".db"
	dumpName, err := filepath.Abs(testDbTemp)
	_ = os.Remove(dumpName)
	if err != nil {
		b.Fatal(err)
	} else if err = fs.Copy(testDbOriginal, dumpName, true); err != nil {
		b.Fatal(err)
	}
	defer os.Remove(dumpName)

	// Force the dbConn to nil so that a new database can be connected to.
	entity.SetDbProvider(nil)

	// Create gorm.DB connection provider.
	db := &entity.DbConn{
		Driver: "sqlite",
		Dsn:    dumpName,
	}

	// Insert test fixtures into the database.
	entity.SetDbProvider(db)

	entity.InitDb(migrate.Opt(true, false, nil))

	defer db.Close()

	// tests here

	runTests(b)

	// teardown here
	event.Log.SetLevel(loglevel)
}
