package commands

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMigrationCommand(t *testing.T) {
	mdbDSN := dsn.Parse(testextras.TestDbDSN(dsn.DriverMariaDB, "migrate"))
	pdbDSN := dsn.Parse(testextras.TestDbDSN(dsn.DriverPostgreSQL, "migrate"))
	defer func() {
		require.NoError(t, testextras.TestDbRemoveByName(mdbDSN.Driver, "migrate"))
		require.NoError(t, testextras.TestDbRemoveByName(pdbDSN.Driver, "migrate"))
	}()

	t.Run("NoMigrateSettings", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(MigrationsCommands, []string{"migrations", "transfer"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "config: transfer config must be provided")
		}
		assert.Equal(t, "", output)
	})

	t.Run("InvalidCommand", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(MigrationsCommands, []string{"migrations", "--magles"})

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "flag provided but not defined: -magles")
		}
		assert.Contains(t, output, "flag provided but not defined: -magles")
	})

	t.Run("Status", func(t *testing.T) {
		// Run command with test context.
		output, err := RunWithTestContext(MigrationsCommands, []string{"migrations", "status"})

		// Check command output for plausibility.
		// t.Log(output)
		assert.Empty(t, err)
		assert.Contains(t, output, "Dialect")
		assert.Contains(t, output, "Stage")
		assert.Contains(t, output, "Status")
	})

	t.Run("RunTraceAndFailed", func(t *testing.T) {
		dbDrv, dbDSN := dsn.PhotoPrismTestToDriverDSN()
		// Run command with test context.
		appArgs := []string{"photoprism",
			"--database-driver", dbDrv,
			"--database-dsn", dbDSN}
		cmdArgs := []string{"migrations", "run", "--trace", "--failed"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.*")
		defer event.Unsubscribe(s)

		var l strings.Builder

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l.WriteString(msg.Fields["message"].(string) + "\n")
			}
		}()

		// Setup and capture SQL Logging output
		buffer := bytes.Buffer{}
		log.SetOutput(&buffer)

		output, err := RunWithProvidedTestContext(ctx, MigrationsCommands, cmdArgs)
		// Reset logger
		log.SetOutput(os.Stdout)

		// Check command output for plausibility.
		// t.Logf("buffer = %s", buffer.String())
		assert.Empty(t, err)
		assert.Empty(t, output)
		assert.Contains(t, l.String(), "migrate: running database migrations")
		assert.Contains(t, buffer.String(), "migrate: enabled trace mode")
		assert.Contains(t, buffer.String(), "migrate: running previously failed migrations")
	})

	t.Run("TargetPopulated", func(t *testing.T) {
		dbDSN := testextras.TestDbDSN(dsn.DriverMariaDB, "migrate")
		if err := testextras.ResetMariaDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
			t.Fatal(err.Error())
		}

		// Setup target database
		_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db")
		if err := fs.Copy("/go/src/github.com/photoprism/photoprism/internal/commands/testdata/transfer_sqlite3", "/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db", true); err != nil {
			t.Fatal(err.Error())
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "mysql",
			"--database-dsn", dbDSN,
			"--transfer-driver", "sqlite",
			"--transfer-dsn", "/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db?_busy_timeout=5000&_foreign_keys=on"}
		cmdArgs := []string{"migrations", "transfer"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.info")
		defer event.Unsubscribe(s)

		var l string

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l += msg.Fields["message"].(string) + "\n"
			}
		}()

		output, err := RunWithProvidedTestContext(ctx, MigrationsCommands, cmdArgs)

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "migrate: transfer target database is not empty")
		assert.NotContains(t, output, "Usage")

		time.Sleep(time.Second)

		// Check command output.
		if l == "" {
			t.Fatal("log output missing")
		}

		assert.Contains(t, l, "migrate: transfer batch size set to 100")

		if !t.Failed() {
			_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db")
		}
	})

	t.Run("TargetPopulatedBatch500", func(t *testing.T) {
		dbDSN := testextras.TestDbDSN(dsn.DriverMariaDB, "migrate")
		if err := testextras.ResetMariaDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
			t.Fatal(err.Error())
		}

		// Setup target database
		_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db")
		if err := fs.Copy("/go/src/github.com/photoprism/photoprism/internal/commands/testdata/transfer_sqlite3", "/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db", true); err != nil {
			t.Fatal(err.Error())
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "mysql",
			"--database-dsn", dbDSN,
			"--transfer-driver", "sqlite",
			"--transfer-dsn", "/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db?_busy_timeout=5000&_foreign_keys=on"}
		cmdArgs := []string{"migrations", "transfer", "-batch", "500"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.info")
		defer event.Unsubscribe(s)

		var l string

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l += msg.Fields["message"].(string) + "\n"
			}
		}()

		output, err := RunWithProvidedTestContext(ctx, MigrationsCommands, cmdArgs)

		// Check command output for plausibility.
		// t.Logf(output)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "migrate: transfer target database is not empty")
		assert.NotContains(t, output, "Usage")

		time.Sleep(time.Second)

		// Check command output.
		if l == "" {
			t.Fatal("log output missing")
		}

		assert.Contains(t, l, "migrate: transfer batch size set to 500")

		if !t.Failed() {
			_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db")
		}
	})

	t.Run("MySQLtoPostgreSQL", func(t *testing.T) {
		dbDSN := testextras.TestDbDSN(dsn.DriverMariaDB, "migrate")
		tfDSN := testextras.TestDbDSN(dsn.DriverPostgreSQL, "migrate")
		if err := testextras.ResetMariaDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
			t.Fatal(err.Error())
		}

		// Load migrate database as source
		if dumpName, err := filepath.Abs("./testdata/transfer_mysql"); err != nil {
			t.Fatal(err)
		} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", dsn.Parse(dbDSN).Name, //nolint:gosec // test generated input
			"-e", "source "+dumpName).Run(); err != nil {
			t.Fatal(err)
		}

		// Clear PostgreSQL target (migrate)
		if err := testextras.ResetPostgresDB(dsn.Parse(tfDSN).Name, "migrate"); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "mysql",
			"--database-dsn", dbDSN,
			"--transfer-driver", "postgres",
			"--transfer-dsn", tfDSN}
		cmdArgs := []string{"migrations", "transfer", "-batch", "10"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.info")
		defer event.Unsubscribe(s)

		var l string

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l += msg.Fields["message"].(string) + "\n"
			}
		}()

		output, err := RunWithProvidedTestContext(ctx, MigrationsCommands, cmdArgs)

		// Check command output for plausibility.
		// t.Logf(output)
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		assert.NotContains(t, output, "Usage")

		time.Sleep(time.Second)

		// Check command output.
		if l == "" {
			t.Fatal("log output missing")
		}
		// t.Logf(l)

		assert.Contains(t, l, "migrate: transfer batch size set to 10")
		assert.Contains(t, l, "migrate: number of albums transferred 31")
		assert.Contains(t, l, "migrate: number of albumusers transferred 0")
		assert.Contains(t, l, "migrate: number of cameras transferred 6")
		assert.Contains(t, l, "migrate: number of categories transferred 1")
		assert.Contains(t, l, "migrate: number of cells transferred 9")
		assert.Contains(t, l, "migrate: number of clients transferred 7")
		assert.Contains(t, l, "migrate: number of countries transferred 1")
		assert.Contains(t, l, "migrate: number of duplicates transferred 0")
		assert.Contains(t, l, "migrate: number of errors transferred 0")
		assert.Contains(t, l, "migrate: number of faces transferred 7")
		assert.Contains(t, l, "migrate: number of files transferred 71")
		assert.Contains(t, l, "migrate: number of fileshares transferred 2")
		assert.Contains(t, l, "migrate: number of filesyncs transferred 3")
		assert.Contains(t, l, "migrate: number of folders transferred 3")
		assert.Contains(t, l, "migrate: number of keywords transferred 26")
		assert.Contains(t, l, "migrate: number of labels transferred 32")
		assert.Contains(t, l, "migrate: number of lenses transferred 2")
		assert.Contains(t, l, "migrate: number of links transferred 5")
		assert.Contains(t, l, "migrate: number of markers transferred 18")
		assert.Contains(t, l, "migrate: number of passcodes transferred 3")
		assert.Contains(t, l, "migrate: number of passwords transferred 11")
		assert.Contains(t, l, "migrate: number of photos transferred 58")
		assert.Contains(t, l, "migrate: number of photousers transferred 0")
		assert.Contains(t, l, "migrate: number of places transferred 10")
		assert.Contains(t, l, "migrate: number of reactions transferred 3")
		assert.Contains(t, l, "migrate: number of sessions transferred 21")
		assert.Contains(t, l, "migrate: number of services transferred 2")
		assert.Contains(t, l, "migrate: number of subjects transferred 6")
		assert.Contains(t, l, "migrate: number of users transferred 11")
		assert.Contains(t, l, "migrate: number of userdetails transferred 9")
		assert.Contains(t, l, "migrate: number of usersettings transferred 13")
		assert.Contains(t, l, "migrate: number of usershares transferred 1")

		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(postgres.Open(tfDSN), &gorm.Config{})
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		lens := entity.Lens{LensSlug: "PhotoPrismTest Data Slug For Lens", LensName: "PhotoPrism Biocular", LensMake: "PhotoPrism", LensModel: "Short", LensType: "Mono", LensDescription: "Special Test Lens"}
		if result := testdb.Create(&lens); result.Error != nil {
			assert.NoError(t, result.Error)
			t.FailNow()
		}
	})

	t.Run("MySQLtoSQLite", func(t *testing.T) {
		dbDSN := testextras.TestDbDSN(dsn.DriverMariaDB, "migrate")
		if err := testextras.ResetMariaDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
			t.Fatal(err.Error())
		}

		// Remove target database file
		_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/mysqltosqlite.test.db")

		// Load migrate database as source
		if dumpName, err := filepath.Abs("./testdata/transfer_mysql"); err != nil {
			t.Fatal(err)
		} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", dsn.Parse(dbDSN).Name, //nolint:gosec // test generated input
			"-e", "source "+dumpName).Run(); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "mysql",
			"--database-dsn", dbDSN,
			"--transfer-driver", "sqlite",
			"--transfer-dsn", "/go/src/github.com/photoprism/photoprism/storage/mysqltosqlite.test.db?_busy_timeout=5000&_foreign_keys=on"}
		cmdArgs := []string{"migrations", "transfer", "-batch", "1000"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.info")
		defer event.Unsubscribe(s)

		var l string

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l += msg.Fields["message"].(string) + "\n"
			}
		}()

		output, err := RunWithProvidedTestContext(ctx, MigrationsCommands, cmdArgs)

		// Check command output for plausibility.
		// t.Logf(output)
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		assert.NotContains(t, output, "Usage")

		time.Sleep(time.Second)

		// Check command output.
		if l == "" {
			t.Fatal("log output missing")
		}
		// t.Logf(l)

		assert.Contains(t, l, "migrate: transfer batch size set to 1000")
		assert.Contains(t, l, "migrate: number of albums transferred 31")
		assert.Contains(t, l, "migrate: number of albumusers transferred 0")
		assert.Contains(t, l, "migrate: number of cameras transferred 6")
		assert.Contains(t, l, "migrate: number of categories transferred 1")
		assert.Contains(t, l, "migrate: number of cells transferred 9")
		assert.Contains(t, l, "migrate: number of clients transferred 7")
		assert.Contains(t, l, "migrate: number of countries transferred 1")
		assert.Contains(t, l, "migrate: number of duplicates transferred 0")
		assert.Contains(t, l, "migrate: number of errors transferred 0")
		assert.Contains(t, l, "migrate: number of faces transferred 7")
		assert.Contains(t, l, "migrate: number of files transferred 71")
		assert.Contains(t, l, "migrate: number of fileshares transferred 2")
		assert.Contains(t, l, "migrate: number of filesyncs transferred 3")
		assert.Contains(t, l, "migrate: number of folders transferred 3")
		assert.Contains(t, l, "migrate: number of keywords transferred 26")
		assert.Contains(t, l, "migrate: number of labels transferred 32")
		assert.Contains(t, l, "migrate: number of lenses transferred 2")
		assert.Contains(t, l, "migrate: number of links transferred 5")
		assert.Contains(t, l, "migrate: number of markers transferred 18")
		assert.Contains(t, l, "migrate: number of passcodes transferred 3")
		assert.Contains(t, l, "migrate: number of passwords transferred 11")
		assert.Contains(t, l, "migrate: number of photos transferred 58")
		assert.Contains(t, l, "migrate: number of photousers transferred 0")
		assert.Contains(t, l, "migrate: number of places transferred 10")
		assert.Contains(t, l, "migrate: number of reactions transferred 3")
		assert.Contains(t, l, "migrate: number of sessions transferred 21")
		assert.Contains(t, l, "migrate: number of services transferred 2")
		assert.Contains(t, l, "migrate: number of subjects transferred 6")
		assert.Contains(t, l, "migrate: number of users transferred 11")
		assert.Contains(t, l, "migrate: number of userdetails transferred 9")
		assert.Contains(t, l, "migrate: number of usersettings transferred 13")
		assert.Contains(t, l, "migrate: number of usershares transferred 1")
		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(sqlite.Open("/go/src/github.com/photoprism/photoprism/storage/mysqltosqlite.test.db?_busy_timeout=5000&_foreign_keys=on"), &gorm.Config{})
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		lens := entity.Lens{LensSlug: "PhotoPrismTest Data Slug For Lens", LensName: "PhotoPrism Biocular", LensMake: "PhotoPrism", LensModel: "Short", LensType: "Mono", LensDescription: "Special Test Lens"}
		if result := testdb.Create(&lens); result.Error != nil {
			assert.NoError(t, result.Error)
			t.FailNow()
		}

		// Remove target database file
		if !t.Failed() {
			_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/mysqltosqlite.test.db")
		}
	})

	t.Run("MySQLtoSQLitePopulated", func(t *testing.T) {
		dbDSN := testextras.TestDbDSN(dsn.DriverMariaDB, "migrate")
		if err := testextras.ResetMariaDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
			t.Fatal(err.Error())
		}

		// Remove target database file
		_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/mysqltosqlitepopulated.test.db")
		if err := fs.Copy("/go/src/github.com/photoprism/photoprism/internal/commands/testdata/transfer_sqlite3", "/go/src/github.com/photoprism/photoprism/storage/mysqltosqlitepopulated.test.db", true); err != nil {
			t.Fatal(err.Error())
		}

		// Load migrate database as source
		if dumpName, err := filepath.Abs("./testdata/transfer_mysql"); err != nil {
			t.Fatal(err)
		} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", dsn.Parse(dbDSN).Name, //nolint:gosec // G204 test generated input
			"-e", "source "+dumpName).Run(); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "mysql",
			"--database-dsn", dbDSN,
			"--transfer-driver", "sqlite",
			"--transfer-dsn", "/go/src/github.com/photoprism/photoprism/storage/mysqltosqlitepopulated.test.db?_busy_timeout=5000&_foreign_keys=on"}
		cmdArgs := []string{"migrations", "transfer", "-force"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.info")
		defer event.Unsubscribe(s)

		var l string

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l += msg.Fields["message"].(string) + "\n"
			}
		}()

		output, err := RunWithProvidedTestContext(ctx, MigrationsCommands, cmdArgs)

		// Check command output for plausibility.
		// t.Logf(output)
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		assert.NotContains(t, output, "Usage")

		time.Sleep(time.Second)

		// Check command output.
		if l == "" {
			t.Fatal("log output missing")
		}
		// t.Logf(l)

		assert.Contains(t, l, "migrate: number of albums transferred 31")
		assert.Contains(t, l, "migrate: number of albumusers transferred 0")
		assert.Contains(t, l, "migrate: number of cameras transferred 6")
		assert.Contains(t, l, "migrate: number of categories transferred 1")
		assert.Contains(t, l, "migrate: number of cells transferred 9")
		assert.Contains(t, l, "migrate: number of clients transferred 7")
		assert.Contains(t, l, "migrate: number of countries transferred 1")
		assert.Contains(t, l, "migrate: number of duplicates transferred 0")
		assert.Contains(t, l, "migrate: number of errors transferred 0")
		assert.Contains(t, l, "migrate: number of faces transferred 7")
		assert.Contains(t, l, "migrate: number of files transferred 71")
		assert.Contains(t, l, "migrate: number of fileshares transferred 2")
		assert.Contains(t, l, "migrate: number of filesyncs transferred 3")
		assert.Contains(t, l, "migrate: number of folders transferred 3")
		assert.Contains(t, l, "migrate: number of keywords transferred 26")
		assert.Contains(t, l, "migrate: number of labels transferred 32")
		assert.Contains(t, l, "migrate: number of lenses transferred 2")
		assert.Contains(t, l, "migrate: number of links transferred 5")
		assert.Contains(t, l, "migrate: number of markers transferred 18")
		assert.Contains(t, l, "migrate: number of passcodes transferred 3")
		assert.Contains(t, l, "migrate: number of passwords transferred 11")
		assert.Contains(t, l, "migrate: number of photos transferred 58")
		assert.Contains(t, l, "migrate: number of photousers transferred 0")
		assert.Contains(t, l, "migrate: number of places transferred 10")
		assert.Contains(t, l, "migrate: number of reactions transferred 3")
		assert.Contains(t, l, "migrate: number of sessions transferred 21")
		assert.Contains(t, l, "migrate: number of services transferred 2")
		assert.Contains(t, l, "migrate: number of subjects transferred 6")
		assert.Contains(t, l, "migrate: number of users transferred 11")
		assert.Contains(t, l, "migrate: number of userdetails transferred 9")
		assert.Contains(t, l, "migrate: number of usersettings transferred 13")
		assert.Contains(t, l, "migrate: number of usershares transferred 1")

		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(sqlite.Open("/go/src/github.com/photoprism/photoprism/storage/mysqltosqlitepopulated.test.db?_busy_timeout=5000&_foreign_keys=on"), &gorm.Config{})
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		lens := entity.Lens{LensSlug: "PhotoPrismTest Data Slug For Lens", LensName: "PhotoPrism Biocular", LensMake: "PhotoPrism", LensModel: "Short", LensType: "Mono", LensDescription: "Special Test Lens"}
		if result := testdb.Create(&lens); result.Error != nil {
			assert.NoError(t, result.Error)
			t.FailNow()
		}

		// Remove target database file
		if !t.Failed() {
			_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/mysqltosqlitepopulated.test.db")
		}
	})

	t.Run("PostgreSQLtoMySQL", func(t *testing.T) {
		dbDSN := testextras.TestDbDSN(dsn.DriverPostgreSQL, "migrate")
		tfDSN := testextras.TestDbDSN(dsn.DriverMariaDB, "migrate")
		if err := testextras.ResetMariaDB(dsn.Parse(tfDSN).Name, "migrate"); err != nil {
			t.Fatal(err.Error())
		}

		// Load migrate database as source
		if dumpName, err := filepath.Abs("./testdata/transfer_postgresql"); err != nil {
			t.Fatal(err)
		} else {
			// Clear Postgres source (migrate)
			if err := testextras.ResetPostgresDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
				t.Fatal(err)
			}
			pDSN := dsn.Parse(dbDSN)
			if err = exec.Command("psql", pDSN.ForPSQL(), "--file="+dumpName).Run(); err != nil { //nolint:gosec // test generated input
				t.Fatal(err)
			}
		}

		// Clear MySQL target (migrate)
		if err := testextras.ResetMariaDB(dsn.Parse(tfDSN).Name, "migrate"); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "postgres",
			"--database-dsn", dbDSN,
			"--transfer-driver", "mysql",
			"--transfer-dsn", tfDSN}
		cmdArgs := []string{"migrations", "transfer"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.info")
		defer event.Unsubscribe(s)

		var l string

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l += msg.Fields["message"].(string) + "\n"
			}
		}()

		output, err := RunWithProvidedTestContext(ctx, MigrationsCommands, cmdArgs)

		// Check command output for plausibility.
		// t.Logf(output)
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		assert.NotContains(t, output, "Usage")

		time.Sleep(time.Second)

		// Check command output.
		if l == "" {
			t.Fatal("log output missing")
		}
		// t.Logf(l)

		assert.Contains(t, l, "migrate: number of albums transferred 31")
		assert.Contains(t, l, "migrate: number of albumusers transferred 0")
		assert.Contains(t, l, "migrate: number of cameras transferred 6")
		assert.Contains(t, l, "migrate: number of categories transferred 1")
		assert.Contains(t, l, "migrate: number of cells transferred 9")
		assert.Contains(t, l, "migrate: number of clients transferred 7")
		assert.Contains(t, l, "migrate: number of countries transferred 1")
		assert.Contains(t, l, "migrate: number of duplicates transferred 0")
		assert.Contains(t, l, "migrate: number of errors transferred 0")
		assert.Contains(t, l, "migrate: number of faces transferred 7")
		assert.Contains(t, l, "migrate: number of files transferred 71")
		assert.Contains(t, l, "migrate: number of fileshares transferred 2")
		assert.Contains(t, l, "migrate: number of filesyncs transferred 3")
		assert.Contains(t, l, "migrate: number of folders transferred 3")
		assert.Contains(t, l, "migrate: number of keywords transferred 26")
		assert.Contains(t, l, "migrate: number of labels transferred 32")
		assert.Contains(t, l, "migrate: number of lenses transferred 2")
		assert.Contains(t, l, "migrate: number of links transferred 5")
		assert.Contains(t, l, "migrate: number of markers transferred 18")
		assert.Contains(t, l, "migrate: number of passcodes transferred 3")
		assert.Contains(t, l, "migrate: number of passwords transferred 11")
		assert.Contains(t, l, "migrate: number of photos transferred 58")
		assert.Contains(t, l, "migrate: number of photousers transferred 0")
		assert.Contains(t, l, "migrate: number of places transferred 10")
		assert.Contains(t, l, "migrate: number of reactions transferred 3")
		assert.Contains(t, l, "migrate: number of sessions transferred 21")
		assert.Contains(t, l, "migrate: number of services transferred 2")
		assert.Contains(t, l, "migrate: number of subjects transferred 6")
		assert.Contains(t, l, "migrate: number of users transferred 11")
		assert.Contains(t, l, "migrate: number of userdetails transferred 9")
		assert.Contains(t, l, "migrate: number of usersettings transferred 13")
		assert.Contains(t, l, "migrate: number of usershares transferred 1")

		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(mysql.Open(tfDSN), &gorm.Config{})
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		lens := entity.Lens{LensSlug: "PhotoPrismTest Data Slug For Lens", LensName: "PhotoPrism Biocular", LensMake: "PhotoPrism", LensModel: "Short", LensType: "Mono", LensDescription: "Special Test Lens"}
		if result := testdb.Create(&lens); result.Error != nil {
			assert.NoError(t, result.Error)
			t.FailNow()
		}
	})

	t.Run("PostgreSQLtoSQLite", func(t *testing.T) {
		dbDSN := testextras.TestDbDSN(dsn.DriverPostgreSQL, "migrate")
		if err := testextras.ResetMariaDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
			t.Fatal(err.Error())
		}

		// Remove target database file
		_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/postgresqltosqlite.test.db")

		// Load migrate database as source
		if dumpName, err := filepath.Abs("./testdata/transfer_postgresql"); err != nil {
			t.Fatal(err)
		} else {
			// Clear Postgres source (migrate)
			if err := testextras.ResetPostgresDB(dsn.Parse(dbDSN).Name, "migrate"); err != nil {
				t.Fatal(err)
			}
			pDSN := dsn.Parse(dbDSN)
			if err = exec.Command("psql", pDSN.ForPSQL(), "--file="+dumpName).Run(); err != nil { //nolint:gosec // test generated input
				t.Fatal(err)
			}
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "postgres",
			"--database-dsn", dbDSN,
			"--transfer-driver", "sqlite",
			"--transfer-dsn", "/go/src/github.com/photoprism/photoprism/storage/postgresqltosqlite.test.db?_busy_timeout=5000&_foreign_keys=on"}
		cmdArgs := []string{"migrations", "transfer"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.info")
		defer event.Unsubscribe(s)

		var l string

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l += msg.Fields["message"].(string) + "\n"
			}
		}()

		output, err := RunWithProvidedTestContext(ctx, MigrationsCommands, cmdArgs)

		// Check command output for plausibility.
		// t.Logf(output)
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		assert.NotContains(t, output, "Usage")

		time.Sleep(time.Second)

		// Check command output.
		if l == "" {
			t.Fatal("log output missing")
		}
		// t.Logf(l)

		assert.Contains(t, l, "migrate: number of albums transferred 31")
		assert.Contains(t, l, "migrate: number of albumusers transferred 0")
		assert.Contains(t, l, "migrate: number of cameras transferred 6")
		assert.Contains(t, l, "migrate: number of categories transferred 1")
		assert.Contains(t, l, "migrate: number of cells transferred 9")
		assert.Contains(t, l, "migrate: number of clients transferred 7")
		assert.Contains(t, l, "migrate: number of countries transferred 1")
		assert.Contains(t, l, "migrate: number of duplicates transferred 0")
		assert.Contains(t, l, "migrate: number of errors transferred 0")
		assert.Contains(t, l, "migrate: number of faces transferred 7")
		assert.Contains(t, l, "migrate: number of files transferred 71")
		assert.Contains(t, l, "migrate: number of fileshares transferred 2")
		assert.Contains(t, l, "migrate: number of filesyncs transferred 3")
		assert.Contains(t, l, "migrate: number of folders transferred 3")
		assert.Contains(t, l, "migrate: number of keywords transferred 26")
		assert.Contains(t, l, "migrate: number of labels transferred 32")
		assert.Contains(t, l, "migrate: number of lenses transferred 2")
		assert.Contains(t, l, "migrate: number of links transferred 5")
		assert.Contains(t, l, "migrate: number of markers transferred 18")
		assert.Contains(t, l, "migrate: number of passcodes transferred 3")
		assert.Contains(t, l, "migrate: number of passwords transferred 11")
		assert.Contains(t, l, "migrate: number of photos transferred 58")
		assert.Contains(t, l, "migrate: number of photousers transferred 0")
		assert.Contains(t, l, "migrate: number of places transferred 10")
		assert.Contains(t, l, "migrate: number of reactions transferred 3")
		assert.Contains(t, l, "migrate: number of sessions transferred 21")
		assert.Contains(t, l, "migrate: number of services transferred 2")
		assert.Contains(t, l, "migrate: number of subjects transferred 6")
		assert.Contains(t, l, "migrate: number of users transferred 11")
		assert.Contains(t, l, "migrate: number of userdetails transferred 9")
		assert.Contains(t, l, "migrate: number of usersettings transferred 13")
		assert.Contains(t, l, "migrate: number of usershares transferred 1")

		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(sqlite.Open("/go/src/github.com/photoprism/photoprism/storage/postgresqltosqlite.test.db?_busy_timeout=5000&_foreign_keys=on"), &gorm.Config{})
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		lens := entity.Lens{LensSlug: "PhotoPrismTest Data Slug For Lens", LensName: "PhotoPrism Biocular", LensMake: "PhotoPrism", LensModel: "Short", LensType: "Mono", LensDescription: "Special Test Lens"}
		if result := testdb.Create(&lens); result.Error != nil {
			assert.NoError(t, result.Error)
			t.FailNow()
		}

		// Remove target database file
		if !t.Failed() {
			_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/postgresqltosqlite.test.db")
		}
	})

	t.Run("SQLiteToMySQL", func(t *testing.T) {
		tfDSN := testextras.TestDbDSN(dsn.DriverMariaDB, "migrate")

		// Remove target database file
		_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/sqlitetomysql.test.db")

		// Load migrate database as source
		if err := fs.Copy("/go/src/github.com/photoprism/photoprism/internal/commands/testdata/transfer_sqlite3", "/go/src/github.com/photoprism/photoprism/storage/sqlitetomysql.test.db", true); err != nil {
			t.Fatal(err.Error())
		}

		// Clear MySQL target (migrate)
		if err := testextras.ResetMariaDB(dsn.Parse(tfDSN).Name, "migrate"); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "sqlite",
			"--database-dsn", "/go/src/github.com/photoprism/photoprism/storage/sqlitetomysql.test.db?_busy_timeout=5000&_foreign_keys=on",
			"--transfer-driver", "mysql",
			"--transfer-dsn", tfDSN}
		cmdArgs := []string{"migrations", "transfer"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.info")
		defer event.Unsubscribe(s)

		var l string

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l += msg.Fields["message"].(string) + "\n"
			}
		}()

		output, err := RunWithProvidedTestContext(ctx, MigrationsCommands, cmdArgs)

		// Check command output for plausibility.
		// t.Logf(output)
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		assert.NotContains(t, output, "Usage")

		time.Sleep(time.Second)

		// Check command output.
		if l == "" {
			t.Fatal("log output missing")
		}
		// t.Logf(l)

		assert.Contains(t, l, "migrate: number of albums transferred 31")
		assert.Contains(t, l, "migrate: number of albumusers transferred 0")
		assert.Contains(t, l, "migrate: number of cameras transferred 6")
		assert.Contains(t, l, "migrate: number of categories transferred 1")
		assert.Contains(t, l, "migrate: number of cells transferred 9")
		assert.Contains(t, l, "migrate: number of clients transferred 7")
		assert.Contains(t, l, "migrate: number of countries transferred 1")
		assert.Contains(t, l, "migrate: number of duplicates transferred 0")
		assert.Contains(t, l, "migrate: number of errors transferred 0")
		assert.Contains(t, l, "migrate: number of faces transferred 7")
		assert.Contains(t, l, "migrate: number of files transferred 71")
		assert.Contains(t, l, "migrate: number of fileshares transferred 2")
		assert.Contains(t, l, "migrate: number of filesyncs transferred 3")
		assert.Contains(t, l, "migrate: number of folders transferred 3")
		assert.Contains(t, l, "migrate: number of keywords transferred 26")
		assert.Contains(t, l, "migrate: number of labels transferred 32")
		assert.Contains(t, l, "migrate: number of lenses transferred 2")
		assert.Contains(t, l, "migrate: number of links transferred 5")
		assert.Contains(t, l, "migrate: number of markers transferred 18")
		assert.Contains(t, l, "migrate: number of passcodes transferred 3")
		assert.Contains(t, l, "migrate: number of passwords transferred 11")
		assert.Contains(t, l, "migrate: number of photos transferred 58")
		assert.Contains(t, l, "migrate: number of photousers transferred 0")
		assert.Contains(t, l, "migrate: number of places transferred 10")
		assert.Contains(t, l, "migrate: number of reactions transferred 3")
		assert.Contains(t, l, "migrate: number of sessions transferred 21")
		assert.Contains(t, l, "migrate: number of services transferred 2")
		assert.Contains(t, l, "migrate: number of subjects transferred 6")
		assert.Contains(t, l, "migrate: number of users transferred 11")
		assert.Contains(t, l, "migrate: number of userdetails transferred 9")
		assert.Contains(t, l, "migrate: number of usersettings transferred 13")
		assert.Contains(t, l, "migrate: number of usershares transferred 1")

		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(mysql.Open(tfDSN), &gorm.Config{})
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		lens := entity.Lens{LensSlug: "PhotoPrismTest Data Slug For Lens", LensName: "PhotoPrism Biocular", LensMake: "PhotoPrism", LensModel: "Short", LensType: "Mono", LensDescription: "Special Test Lens"}
		if result := testdb.Create(&lens); result.Error != nil {
			assert.NoError(t, result.Error)
			t.FailNow()
		}

		// Remove target database file
		if !t.Failed() {
			_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/sqlitetomysql.test.db")
		}
	})

	t.Run("SQLiteToPostgreSQL", func(t *testing.T) {
		tfDSN := testextras.TestDbDSN(dsn.DriverPostgreSQL, "migrate")

		// Remove target database file
		_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/sqlitetopostgresql.test.db")

		// Load migrate database as source
		if err := fs.Copy("/go/src/github.com/photoprism/photoprism/internal/commands/testdata/transfer_sqlite3", "/go/src/github.com/photoprism/photoprism/storage/sqlitetopostgresql.test.db", true); err != nil {
			t.Fatal(err.Error())
		}

		// Clear PostgreSQL target (migrate)
		if err := testextras.ResetPostgresDB(dsn.Parse(tfDSN).Name, "migrate"); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "sqlite",
			"--database-dsn", "/go/src/github.com/photoprism/photoprism/storage/sqlitetopostgresql.test.db?_busy_timeout=5000&_foreign_keys=on",
			"--transfer-driver", "postgres",
			"--transfer-dsn", tfDSN}
		cmdArgs := []string{"migrations", "transfer"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.info")
		defer event.Unsubscribe(s)

		var l string

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l += msg.Fields["message"].(string) + "\n"
			}
		}()

		output, err := RunWithProvidedTestContext(ctx, MigrationsCommands, cmdArgs)

		// Check command output for plausibility.
		// t.Logf(output)
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		assert.NotContains(t, output, "Usage")

		time.Sleep(time.Second)

		// Check command output.
		if l == "" {
			t.Fatal("log output missing")
		}
		// t.Logf(l)

		assert.Contains(t, l, "migrate: number of albums transferred 31")
		assert.Contains(t, l, "migrate: number of albumusers transferred 0")
		assert.Contains(t, l, "migrate: number of cameras transferred 6")
		assert.Contains(t, l, "migrate: number of categories transferred 1")
		assert.Contains(t, l, "migrate: number of cells transferred 9")
		assert.Contains(t, l, "migrate: number of clients transferred 7")
		assert.Contains(t, l, "migrate: number of countries transferred 1")
		assert.Contains(t, l, "migrate: number of duplicates transferred 0")
		assert.Contains(t, l, "migrate: number of errors transferred 0")
		assert.Contains(t, l, "migrate: number of faces transferred 7")
		assert.Contains(t, l, "migrate: number of files transferred 71")
		assert.Contains(t, l, "migrate: number of fileshares transferred 2")
		assert.Contains(t, l, "migrate: number of filesyncs transferred 3")
		assert.Contains(t, l, "migrate: number of folders transferred 3")
		assert.Contains(t, l, "migrate: number of keywords transferred 26")
		assert.Contains(t, l, "migrate: number of labels transferred 32")
		assert.Contains(t, l, "migrate: number of lenses transferred 2")
		assert.Contains(t, l, "migrate: number of links transferred 5")
		assert.Contains(t, l, "migrate: number of markers transferred 18")
		assert.Contains(t, l, "migrate: number of passcodes transferred 3")
		assert.Contains(t, l, "migrate: number of passwords transferred 11")
		assert.Contains(t, l, "migrate: number of photos transferred 58")
		assert.Contains(t, l, "migrate: number of photousers transferred 0")
		assert.Contains(t, l, "migrate: number of places transferred 10")
		assert.Contains(t, l, "migrate: number of reactions transferred 3")
		assert.Contains(t, l, "migrate: number of sessions transferred 21")
		assert.Contains(t, l, "migrate: number of services transferred 2")
		assert.Contains(t, l, "migrate: number of subjects transferred 6")
		assert.Contains(t, l, "migrate: number of users transferred 11")
		assert.Contains(t, l, "migrate: number of userdetails transferred 9")
		assert.Contains(t, l, "migrate: number of usersettings transferred 13")
		assert.Contains(t, l, "migrate: number of usershares transferred 1")

		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(postgres.Open(tfDSN), &gorm.Config{})
		if err != nil {
			assert.NoError(t, err)
			t.FailNow()
		}
		lens := entity.Lens{LensSlug: "PhotoPrismTest Data Slug For Lens", LensName: "PhotoPrism Biocular", LensMake: "PhotoPrism", LensModel: "Short", LensType: "Mono", LensDescription: "Special Test Lens"}
		if result := testdb.Create(&lens); result.Error != nil {
			assert.NoError(t, result.Error)
			t.FailNow()
		}

		// Remove target database file
		if !t.Failed() {
			_ = os.Remove("/go/src/github.com/photoprism/photoprism/storage/sqlitetopostgresql.test.db")
		}
	})

}
