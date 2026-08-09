package performancetest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/migrate"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func Benchmark1k_MySQL(b *testing.B) {
	runMySQLPerfTest(1000, "1k", b)
}

func Benchmark10k_MySQL(b *testing.B) {
	runMySQLPerfTest(10000, "10k", b)
}

func Benchmark100k_MySQL(b *testing.B) {
	// Skip this as the setup and storage are excessive.
	if _, ok := os.LookupEnv("BENCH_GLACIAL"); !ok {
		b.Skip("skipping benchmark as BENCH_GLACIAL not set")
	} else {
		runMySQLPerfTest(100000, "100k", b)
	}
}

func runMySQLPerfTest(rows int, numberk string, b *testing.B) {
	// Setup here
	loglevel := event.Log.GetLevel()
	event.Log.SetLevel(logrus.ErrorLevel)
	testDbOriginal := "../../storage/test-" + numberk + ".original.mysql"
	mDSN := dsn.Parse(testextras.TestDbDSN(dsn.DriverMariaDB, "migrate"))
	mysqlD := dsn.TestDSNFromEnv(dsn.DriverMariaDB, "migrate")
	mysqlDSN := mysqlD.ToString()

	// Prepare temporary mariadb db.
	if !fs.FileExists(testDbOriginal) {
		log.Info("Generating Mariadb database with %d records", rows)
		require.NoError(b, generateDatabase(rows, "mysql", mDSN.ToString(), true, true))
		resultFile := "--result-file=" + testDbOriginal
		if err := exec.Command("mariadb-dump", "--user=migrate", "--password=migrate", "--lock-tables", "--no-create-db", mDSN.Name, resultFile).Run(); err != nil { //nolint:gosec // test generated input, test only credentials
			b.Fatal(err)
		}
	}

	// Prepare migrate mariadb db.
	if dumpName, err := filepath.Abs(testDbOriginal); err != nil {
		b.Fatal(err)
	} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", mDSN.Name, //nolint:gosec // test generated input, test only credentials
		"-e", "source "+dumpName).Run(); err != nil {
		b.Fatal(err)
	}
	defer testextras.TestDbRemoveByName(dsn.DriverMySQL, "migrate")

	// Force the dbConn to nil so that a new database can be connected to.
	entity.SetDbProvider(nil)

	// Create gorm.DB connection provider.
	db := &entity.DbConn{
		Driver: "mysql",
		Dsn:    mysqlDSN,
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
