package performancetest

import (
	"fmt"
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

func Benchmark1k_Postgres(b *testing.B) {
	runPostgresPerfTest(1000, "1k", b)
}

func Benchmark10k_Postgres(b *testing.B) {
	runPostgresPerfTest(10000, "10k", b)
}

func Benchmark100k_Postgres(b *testing.B) {
	// Skip this as the setup and storage are excessive.
	if _, ok := os.LookupEnv("BENCH_GLACIAL"); !ok {
		b.Skip("skipping benchmark as BENCH_GLACIAL not set")
	} else {
		runPostgresPerfTest(100000, "100k", b)
	}
}

func runPostgresPerfTest(rows int, numberk string, b *testing.B) {
	// Setup here
	loglevel := event.Log.GetLevel()
	event.Log.SetLevel(logrus.ErrorLevel)
	testDbOriginal := "../../storage/test-" + numberk + ".original.postgresql"
	oDSN := testextras.TestDbDSN(dsn.DriverPostgreSQL, "migrate")
	mDSN := dsn.Parse(oDSN)
	_, admindsn := dsn.PhotoPrismDriverToDriverDSN(dsn.DriverPostgres)
	pDSN := dsn.Parse(admindsn)

	// Prepare temporary PostgreSQL db.
	if !fs.FileExists(testDbOriginal) {
		log.Info("Generating PostgreSQL database with %d records", rows)
		require.NoError(b, generateDatabase(rows, dsn.DriverPostgres, mDSN.ToString(), true, true))
		if err := exec.Command("pg_dump", "-d", mDSN.ForPSQL(), "-F", "c", "-f", testDbOriginal).Run(); err != nil { //nolint:gosec // test generated input
			b.Fatal(err)
		}
	}

	// Prepare migrate PostgreSQL db.
	if dumpName, err := filepath.Abs(testDbOriginal); err != nil {
		b.Fatal(err)
	} else if err = exec.Command("dropdb", fmt.Sprintf("--maintenance-db=%s", pDSN.ForPSQL()), "--force", "--if-exists", mDSN.Name).Run(); err != nil { //nolint:gosec // test generated input, test only credentials
		b.Fatal(err)
	} else if err = exec.Command("createdb", fmt.Sprintf("--maintenance-db=%s", pDSN.ForPSQL()), "-O", "migrate", "-T", "template0", mDSN.Name).Run(); err != nil { //nolint:gosec // test generated input, test only credentials
		b.Fatal(err)
	} else if err = exec.Command("pg_restore", "-d", mDSN.ForPSQL(), dumpName).Run(); err != nil { //nolint:gosec // test generated input, test only credentials
		b.Fatal(err)
	}
	defer testextras.TestDbRemoveByName(dsn.DriverPostgreSQL, "migrate")

	// Force the dbConn to nil so that a new database can be connected to.
	entity.SetDbProvider(nil)

	// Create gorm.DB connection provider.
	db := &entity.DbConn{
		Driver: dsn.DriverPostgres,
		Dsn:    mDSN.ToString(),
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
