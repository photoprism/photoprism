package commands

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/dsn"
)

func TestMigrationCommand(t *testing.T) {
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
		dbDrv, dbDSN := dsn.PhotoPrismTestToDriverDSN(0)
		// Run command with test context.
		appArgs := []string{"photoprism",
			"--database-driver", dbDrv,
			"--database-dsn", dbDSN}
		cmdArgs := []string{"migrations", "run", "--trace", "--failed"}

		ctx := NewTestContextWithParse(appArgs, cmdArgs)

		s := event.Subscribe("log.*")
		defer event.Unsubscribe(s)

		var l string

		assert.IsType(t, hub.Subscription{}, s)

		go func() {
			for msg := range s.Receiver {
				l += msg.Fields["message"].(string) + "\n"
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
		assert.Contains(t, l, "migrate: running database migrations")
		assert.Contains(t, buffer.String(), "migrate: enabled trace mode")
		assert.Contains(t, buffer.String(), "migrate: running previously failed migrations")
	})

	t.Run("TargetPopulated", func(t *testing.T) {
		dbDSN := dsn.DSN{Driver: dsn.DriverMariaDB, Net: "tcp", Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "mariadb:4001", User: "migrate", Password: "migrate"}

		// Setup target database
		os.Remove("/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db")
		if err := copyFile("/go/src/github.com/photoprism/photoprism/internal/commands/testdata/transfer_sqlite3", "/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db"); err != nil {
			t.Fatal(err.Error())
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "mysql",
			"--database-dsn", dbDSN.ToString(),
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
			os.Remove("/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db")
		}
	})

	t.Run("TargetPopulatedBatch500", func(t *testing.T) {
		dbDSN := dsn.DSN{Driver: dsn.DriverMariaDB, Net: "tcp", Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "mariadb:4001", User: "migrate", Password: "migrate"}

		// Setup target database
		os.Remove("/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db")
		if err := copyFile("/go/src/github.com/photoprism/photoprism/internal/commands/testdata/transfer_sqlite3", "/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db"); err != nil {
			t.Fatal(err.Error())
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "mysql",
			"--database-dsn", dbDSN.ToString(),
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
			os.Remove("/go/src/github.com/photoprism/photoprism/storage/targetpopulated.test.db")
		}
	})

	t.Run("MySQLtoPostgreSQL", func(t *testing.T) {
		dbDSN := dsn.DSN{Driver: dsn.DriverMariaDB, Net: "tcp", Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "mariadb:4001", User: "migrate", Password: "migrate"}
		tfDSN := dsn.DSN{Driver: dsn.DriverPostgreSQL, Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "postgres:5432", User: "migrate", Password: "migrate"}

		// Load migrate database as source
		if dumpName, err := filepath.Abs("./testdata/transfer_mysql"); err != nil {
			t.Fatal(err)
		} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()),
			"-e", "source "+dumpName).Run(); err != nil {
			t.Fatal(err)
		}

		// Clear PostgreSQL target (migrate)
		if err := testextras.ResetPostgresDB("migrate", testextras.GetDBMutexID()); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "mysql",
			"--database-dsn", dbDSN.ToString(),
			"--transfer-driver", "postgres",
			"--transfer-dsn", tfDSN.ToString()}
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
		//t.Logf(output)
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
		assert.Contains(t, l, "migrate: number of albums transfered 31")
		assert.Contains(t, l, "migrate: number of albumusers transfered 0")
		assert.Contains(t, l, "migrate: number of cameras transfered 6")
		assert.Contains(t, l, "migrate: number of categories transfered 1")
		assert.Contains(t, l, "migrate: number of cells transfered 9")
		assert.Contains(t, l, "migrate: number of clients transfered 7")
		assert.Contains(t, l, "migrate: number of countries transfered 1")
		assert.Contains(t, l, "migrate: number of duplicates transfered 0")
		assert.Contains(t, l, "migrate: number of errors transfered 0")
		assert.Contains(t, l, "migrate: number of faces transfered 7")
		assert.Contains(t, l, "migrate: number of files transfered 71")
		assert.Contains(t, l, "migrate: number of fileshares transfered 2")
		assert.Contains(t, l, "migrate: number of filesyncs transfered 3")
		assert.Contains(t, l, "migrate: number of folders transfered 3")
		assert.Contains(t, l, "migrate: number of keywords transfered 26")
		assert.Contains(t, l, "migrate: number of labels transfered 32")
		assert.Contains(t, l, "migrate: number of lenses transfered 2")
		assert.Contains(t, l, "migrate: number of links transfered 5")
		assert.Contains(t, l, "migrate: number of markers transfered 18")
		assert.Contains(t, l, "migrate: number of passcodes transfered 3")
		assert.Contains(t, l, "migrate: number of passwords transfered 11")
		assert.Contains(t, l, "migrate: number of photos transfered 58")
		assert.Contains(t, l, "migrate: number of photousers transfered 0")
		assert.Contains(t, l, "migrate: number of places transfered 10")
		assert.Contains(t, l, "migrate: number of reactions transfered 3")
		assert.Contains(t, l, "migrate: number of sessions transfered 21")
		assert.Contains(t, l, "migrate: number of services transfered 2")
		assert.Contains(t, l, "migrate: number of subjects transfered 6")
		assert.Contains(t, l, "migrate: number of users transfered 11")
		assert.Contains(t, l, "migrate: number of userdetails transfered 9")
		assert.Contains(t, l, "migrate: number of usersettings transfered 13")
		assert.Contains(t, l, "migrate: number of usershares transfered 1")

		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(postgres.Open(tfDSN.ToString()), &gorm.Config{})
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
		dbDSN := dsn.DSN{Driver: dsn.DriverMariaDB, Net: "tcp", Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "mariadb:4001", User: "migrate", Password: "migrate"}

		// Remove target database file
		os.Remove("/go/src/github.com/photoprism/photoprism/storage/mysqltosqlite.test.db")

		// Load migrate database as source
		if dumpName, err := filepath.Abs("./testdata/transfer_mysql"); err != nil {
			t.Fatal(err)
		} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()),
			"-e", "source "+dumpName).Run(); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "mysql",
			"--database-dsn", dbDSN.ToString(),
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
		//t.Logf(output)
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
		assert.Contains(t, l, "migrate: number of albums transfered 31")
		assert.Contains(t, l, "migrate: number of albumusers transfered 0")
		assert.Contains(t, l, "migrate: number of cameras transfered 6")
		assert.Contains(t, l, "migrate: number of categories transfered 1")
		assert.Contains(t, l, "migrate: number of cells transfered 9")
		assert.Contains(t, l, "migrate: number of clients transfered 7")
		assert.Contains(t, l, "migrate: number of countries transfered 1")
		assert.Contains(t, l, "migrate: number of duplicates transfered 0")
		assert.Contains(t, l, "migrate: number of errors transfered 0")
		assert.Contains(t, l, "migrate: number of faces transfered 7")
		assert.Contains(t, l, "migrate: number of files transfered 71")
		assert.Contains(t, l, "migrate: number of fileshares transfered 2")
		assert.Contains(t, l, "migrate: number of filesyncs transfered 3")
		assert.Contains(t, l, "migrate: number of folders transfered 3")
		assert.Contains(t, l, "migrate: number of keywords transfered 26")
		assert.Contains(t, l, "migrate: number of labels transfered 32")
		assert.Contains(t, l, "migrate: number of lenses transfered 2")
		assert.Contains(t, l, "migrate: number of links transfered 5")
		assert.Contains(t, l, "migrate: number of markers transfered 18")
		assert.Contains(t, l, "migrate: number of passcodes transfered 3")
		assert.Contains(t, l, "migrate: number of passwords transfered 11")
		assert.Contains(t, l, "migrate: number of photos transfered 58")
		assert.Contains(t, l, "migrate: number of photousers transfered 0")
		assert.Contains(t, l, "migrate: number of places transfered 10")
		assert.Contains(t, l, "migrate: number of reactions transfered 3")
		assert.Contains(t, l, "migrate: number of sessions transfered 21")
		assert.Contains(t, l, "migrate: number of services transfered 2")
		assert.Contains(t, l, "migrate: number of subjects transfered 6")
		assert.Contains(t, l, "migrate: number of users transfered 11")
		assert.Contains(t, l, "migrate: number of userdetails transfered 9")
		assert.Contains(t, l, "migrate: number of usersettings transfered 13")
		assert.Contains(t, l, "migrate: number of usershares transfered 1")
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
			os.Remove("/go/src/github.com/photoprism/photoprism/storage/mysqltosqlite.test.db")
		}
	})

	t.Run("MySQLtoSQLitePopulated", func(t *testing.T) {
		dbDSN := dsn.DSN{Driver: dsn.DriverMariaDB, Net: "tcp", Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "mariadb:4001", User: "migrate", Password: "migrate"}

		// Remove target database file
		os.Remove("/go/src/github.com/photoprism/photoprism/storage/mysqltosqlitepopulated.test.db")
		if err := copyFile("/go/src/github.com/photoprism/photoprism/internal/commands/testdata/transfer_sqlite3", "/go/src/github.com/photoprism/photoprism/storage/mysqltosqlitepopulated.test.db"); err != nil {
			t.Fatal(err.Error())
		}

		// Load migrate database as source
		if dumpName, err := filepath.Abs("./testdata/transfer_mysql"); err != nil {
			t.Fatal(err)
		} else if err = exec.Command("mariadb", "-u", "migrate", "-pmigrate", fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()),
			"-e", "source "+dumpName).Run(); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "mysql",
			"--database-dsn", dbDSN.ToString(),
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
		//t.Logf(output)
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

		assert.Contains(t, l, "migrate: number of albums transfered 31")
		assert.Contains(t, l, "migrate: number of albumusers transfered 0")
		assert.Contains(t, l, "migrate: number of cameras transfered 6")
		assert.Contains(t, l, "migrate: number of categories transfered 1")
		assert.Contains(t, l, "migrate: number of cells transfered 9")
		assert.Contains(t, l, "migrate: number of clients transfered 7")
		assert.Contains(t, l, "migrate: number of countries transfered 1")
		assert.Contains(t, l, "migrate: number of duplicates transfered 0")
		assert.Contains(t, l, "migrate: number of errors transfered 0")
		assert.Contains(t, l, "migrate: number of faces transfered 7")
		assert.Contains(t, l, "migrate: number of files transfered 71")
		assert.Contains(t, l, "migrate: number of fileshares transfered 2")
		assert.Contains(t, l, "migrate: number of filesyncs transfered 3")
		assert.Contains(t, l, "migrate: number of folders transfered 3")
		assert.Contains(t, l, "migrate: number of keywords transfered 26")
		assert.Contains(t, l, "migrate: number of labels transfered 32")
		assert.Contains(t, l, "migrate: number of lenses transfered 2")
		assert.Contains(t, l, "migrate: number of links transfered 5")
		assert.Contains(t, l, "migrate: number of markers transfered 18")
		assert.Contains(t, l, "migrate: number of passcodes transfered 3")
		assert.Contains(t, l, "migrate: number of passwords transfered 11")
		assert.Contains(t, l, "migrate: number of photos transfered 58")
		assert.Contains(t, l, "migrate: number of photousers transfered 0")
		assert.Contains(t, l, "migrate: number of places transfered 10")
		assert.Contains(t, l, "migrate: number of reactions transfered 3")
		assert.Contains(t, l, "migrate: number of sessions transfered 21")
		assert.Contains(t, l, "migrate: number of services transfered 2")
		assert.Contains(t, l, "migrate: number of subjects transfered 6")
		assert.Contains(t, l, "migrate: number of users transfered 11")
		assert.Contains(t, l, "migrate: number of userdetails transfered 9")
		assert.Contains(t, l, "migrate: number of usersettings transfered 13")
		assert.Contains(t, l, "migrate: number of usershares transfered 1")

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
			os.Remove("/go/src/github.com/photoprism/photoprism/storage/mysqltosqlitepopulated.test.db")
		}
	})

	t.Run("PostgreSQLtoMySQL", func(t *testing.T) {
		dbDSN := dsn.DSN{Driver: dsn.DriverPostgreSQL, Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "postgres:5432", User: "migrate", Password: "migrate"}
		tfDSN := dsn.DSN{Driver: dsn.DriverMariaDB, Net: "tcp", Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "mariadb:4001", User: "migrate", Password: "migrate"}

		// Load migrate database as source
		if dumpName, err := filepath.Abs("./testdata/transfer_postgresql"); err != nil {
			t.Fatal(err)
		} else {
			// Clear Postgres source (migrate)
			if err := testextras.ResetPostgresDB("migrate", testextras.GetDBMutexID()); err != nil {
				t.Fatal(err)
			}
			if err = exec.Command("psql", dbDSN.ForPSQL(), "--file="+dumpName).Run(); err != nil {
				t.Fatal(err)
			}
		}

		// Clear MySQL target (migrate)
		if err := testextras.ResetMariaDB("migrate", testextras.GetDBMutexID()); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "postgres",
			"--database-dsn", dbDSN.ToString(),
			"--transfer-driver", "mysql",
			"--transfer-dsn", tfDSN.ToString()}
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
		//t.Logf(output)
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

		assert.Contains(t, l, "migrate: number of albums transfered 31")
		assert.Contains(t, l, "migrate: number of albumusers transfered 0")
		assert.Contains(t, l, "migrate: number of cameras transfered 6")
		assert.Contains(t, l, "migrate: number of categories transfered 1")
		assert.Contains(t, l, "migrate: number of cells transfered 9")
		assert.Contains(t, l, "migrate: number of clients transfered 7")
		assert.Contains(t, l, "migrate: number of countries transfered 1")
		assert.Contains(t, l, "migrate: number of duplicates transfered 0")
		assert.Contains(t, l, "migrate: number of errors transfered 0")
		assert.Contains(t, l, "migrate: number of faces transfered 7")
		assert.Contains(t, l, "migrate: number of files transfered 71")
		assert.Contains(t, l, "migrate: number of fileshares transfered 2")
		assert.Contains(t, l, "migrate: number of filesyncs transfered 3")
		assert.Contains(t, l, "migrate: number of folders transfered 3")
		assert.Contains(t, l, "migrate: number of keywords transfered 26")
		assert.Contains(t, l, "migrate: number of labels transfered 32")
		assert.Contains(t, l, "migrate: number of lenses transfered 2")
		assert.Contains(t, l, "migrate: number of links transfered 5")
		assert.Contains(t, l, "migrate: number of markers transfered 18")
		assert.Contains(t, l, "migrate: number of passcodes transfered 3")
		assert.Contains(t, l, "migrate: number of passwords transfered 11")
		assert.Contains(t, l, "migrate: number of photos transfered 58")
		assert.Contains(t, l, "migrate: number of photousers transfered 0")
		assert.Contains(t, l, "migrate: number of places transfered 10")
		assert.Contains(t, l, "migrate: number of reactions transfered 3")
		assert.Contains(t, l, "migrate: number of sessions transfered 21")
		assert.Contains(t, l, "migrate: number of services transfered 2")
		assert.Contains(t, l, "migrate: number of subjects transfered 6")
		assert.Contains(t, l, "migrate: number of users transfered 11")
		assert.Contains(t, l, "migrate: number of userdetails transfered 9")
		assert.Contains(t, l, "migrate: number of usersettings transfered 13")
		assert.Contains(t, l, "migrate: number of usershares transfered 1")

		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(mysql.Open(tfDSN.ToString()), &gorm.Config{})
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
		dbDSN := dsn.DSN{Driver: dsn.DriverPostgreSQL, Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "postgres:5432", User: "migrate", Password: "migrate"}

		// Remove target database file
		os.Remove("/go/src/github.com/photoprism/photoprism/storage/postgresqltosqlite.test.db")

		// Load migrate database as source
		if dumpName, err := filepath.Abs("./testdata/transfer_postgresql"); err != nil {
			t.Fatal(err)
		} else {
			// Clear Postgres source (migrate)
			if err := testextras.ResetPostgresDB("migrate", testextras.GetDBMutexID()); err != nil {
				t.Fatal(err)
			}

			if err = exec.Command("psql", dbDSN.ForPSQL(), "--file="+dumpName).Run(); err != nil {
				t.Fatal(err)
			}
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "postgres",
			"--database-dsn", dbDSN.ToString(),
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
		//t.Logf(output)
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

		assert.Contains(t, l, "migrate: number of albums transfered 31")
		assert.Contains(t, l, "migrate: number of albumusers transfered 0")
		assert.Contains(t, l, "migrate: number of cameras transfered 6")
		assert.Contains(t, l, "migrate: number of categories transfered 1")
		assert.Contains(t, l, "migrate: number of cells transfered 9")
		assert.Contains(t, l, "migrate: number of clients transfered 7")
		assert.Contains(t, l, "migrate: number of countries transfered 1")
		assert.Contains(t, l, "migrate: number of duplicates transfered 0")
		assert.Contains(t, l, "migrate: number of errors transfered 0")
		assert.Contains(t, l, "migrate: number of faces transfered 7")
		assert.Contains(t, l, "migrate: number of files transfered 71")
		assert.Contains(t, l, "migrate: number of fileshares transfered 2")
		assert.Contains(t, l, "migrate: number of filesyncs transfered 3")
		assert.Contains(t, l, "migrate: number of folders transfered 3")
		assert.Contains(t, l, "migrate: number of keywords transfered 26")
		assert.Contains(t, l, "migrate: number of labels transfered 32")
		assert.Contains(t, l, "migrate: number of lenses transfered 2")
		assert.Contains(t, l, "migrate: number of links transfered 5")
		assert.Contains(t, l, "migrate: number of markers transfered 18")
		assert.Contains(t, l, "migrate: number of passcodes transfered 3")
		assert.Contains(t, l, "migrate: number of passwords transfered 11")
		assert.Contains(t, l, "migrate: number of photos transfered 58")
		assert.Contains(t, l, "migrate: number of photousers transfered 0")
		assert.Contains(t, l, "migrate: number of places transfered 10")
		assert.Contains(t, l, "migrate: number of reactions transfered 3")
		assert.Contains(t, l, "migrate: number of sessions transfered 21")
		assert.Contains(t, l, "migrate: number of services transfered 2")
		assert.Contains(t, l, "migrate: number of subjects transfered 6")
		assert.Contains(t, l, "migrate: number of users transfered 11")
		assert.Contains(t, l, "migrate: number of userdetails transfered 9")
		assert.Contains(t, l, "migrate: number of usersettings transfered 13")
		assert.Contains(t, l, "migrate: number of usershares transfered 1")

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
			os.Remove("/go/src/github.com/photoprism/photoprism/storage/postgresqltosqlite.test.db")
		}
	})

	t.Run("SQLiteToMySQL", func(t *testing.T) {
		tfDSN := dsn.DSN{Driver: dsn.DriverMariaDB, Net: "tcp", Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "mariadb:4001", User: "migrate", Password: "migrate"}

		// Remove target database file
		os.Remove("/go/src/github.com/photoprism/photoprism/storage/sqlitetomysql.test.db")

		// Load migrate database as source
		if err := copyFile("/go/src/github.com/photoprism/photoprism/internal/commands/testdata/transfer_sqlite3", "/go/src/github.com/photoprism/photoprism/storage/sqlitetomysql.test.db"); err != nil {
			t.Fatal(err.Error())
		}

		// Clear MySQL target (migrate)
		if err := testextras.ResetMariaDB("migrate", testextras.GetDBMutexID()); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "sqlite",
			"--database-dsn", "/go/src/github.com/photoprism/photoprism/storage/sqlitetomysql.test.db?_busy_timeout=5000&_foreign_keys=on",
			"--transfer-driver", "mysql",
			"--transfer-dsn", tfDSN.ToString()}
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
		//t.Logf(output)
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

		assert.Contains(t, l, "migrate: number of albums transfered 31")
		assert.Contains(t, l, "migrate: number of albumusers transfered 0")
		assert.Contains(t, l, "migrate: number of cameras transfered 6")
		assert.Contains(t, l, "migrate: number of categories transfered 1")
		assert.Contains(t, l, "migrate: number of cells transfered 9")
		assert.Contains(t, l, "migrate: number of clients transfered 7")
		assert.Contains(t, l, "migrate: number of countries transfered 1")
		assert.Contains(t, l, "migrate: number of duplicates transfered 0")
		assert.Contains(t, l, "migrate: number of errors transfered 0")
		assert.Contains(t, l, "migrate: number of faces transfered 7")
		assert.Contains(t, l, "migrate: number of files transfered 71")
		assert.Contains(t, l, "migrate: number of fileshares transfered 2")
		assert.Contains(t, l, "migrate: number of filesyncs transfered 3")
		assert.Contains(t, l, "migrate: number of folders transfered 3")
		assert.Contains(t, l, "migrate: number of keywords transfered 26")
		assert.Contains(t, l, "migrate: number of labels transfered 32")
		assert.Contains(t, l, "migrate: number of lenses transfered 2")
		assert.Contains(t, l, "migrate: number of links transfered 5")
		assert.Contains(t, l, "migrate: number of markers transfered 18")
		assert.Contains(t, l, "migrate: number of passcodes transfered 3")
		assert.Contains(t, l, "migrate: number of passwords transfered 11")
		assert.Contains(t, l, "migrate: number of photos transfered 58")
		assert.Contains(t, l, "migrate: number of photousers transfered 0")
		assert.Contains(t, l, "migrate: number of places transfered 10")
		assert.Contains(t, l, "migrate: number of reactions transfered 3")
		assert.Contains(t, l, "migrate: number of sessions transfered 21")
		assert.Contains(t, l, "migrate: number of services transfered 2")
		assert.Contains(t, l, "migrate: number of subjects transfered 6")
		assert.Contains(t, l, "migrate: number of users transfered 11")
		assert.Contains(t, l, "migrate: number of userdetails transfered 9")
		assert.Contains(t, l, "migrate: number of usersettings transfered 13")
		assert.Contains(t, l, "migrate: number of usershares transfered 1")

		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(mysql.Open(tfDSN.ToString()), &gorm.Config{})
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
			os.Remove("/go/src/github.com/photoprism/photoprism/storage/sqlitetomysql.test.db")
		}
	})

	t.Run("SQLiteToPostgreSQL", func(t *testing.T) {
		tfDSN := dsn.DSN{Driver: dsn.DriverPostgreSQL, Name: fmt.Sprintf("migrate_%02d", testextras.GetDBMutexID()), Server: "postgres:5432", User: "migrate", Password: "migrate"}

		// Remove target database file
		os.Remove("/go/src/github.com/photoprism/photoprism/storage/sqlitetopostgresql.test.db")

		// Load migrate database as source
		if err := copyFile("/go/src/github.com/photoprism/photoprism/internal/commands/testdata/transfer_sqlite3", "/go/src/github.com/photoprism/photoprism/storage/sqlitetopostgresql.test.db"); err != nil {
			t.Fatal(err.Error())
		}

		// Clear PostgreSQL target (migrate)
		if err := testextras.ResetPostgresDB("migrate", testextras.GetDBMutexID()); err != nil {
			t.Fatal(err)
		}

		// Run command with test context.
		log = event.Log

		appArgs := []string{"photoprism",
			"--database-driver", "sqlite",
			"--database-dsn", "/go/src/github.com/photoprism/photoprism/storage/sqlitetopostgresql.test.db?_busy_timeout=5000&_foreign_keys=on",
			"--transfer-driver", "postgres",
			"--transfer-dsn", tfDSN.ToString()}
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
		//t.Logf(output)
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

		assert.Contains(t, l, "migrate: number of albums transfered 31")
		assert.Contains(t, l, "migrate: number of albumusers transfered 0")
		assert.Contains(t, l, "migrate: number of cameras transfered 6")
		assert.Contains(t, l, "migrate: number of categories transfered 1")
		assert.Contains(t, l, "migrate: number of cells transfered 9")
		assert.Contains(t, l, "migrate: number of clients transfered 7")
		assert.Contains(t, l, "migrate: number of countries transfered 1")
		assert.Contains(t, l, "migrate: number of duplicates transfered 0")
		assert.Contains(t, l, "migrate: number of errors transfered 0")
		assert.Contains(t, l, "migrate: number of faces transfered 7")
		assert.Contains(t, l, "migrate: number of files transfered 71")
		assert.Contains(t, l, "migrate: number of fileshares transfered 2")
		assert.Contains(t, l, "migrate: number of filesyncs transfered 3")
		assert.Contains(t, l, "migrate: number of folders transfered 3")
		assert.Contains(t, l, "migrate: number of keywords transfered 26")
		assert.Contains(t, l, "migrate: number of labels transfered 32")
		assert.Contains(t, l, "migrate: number of lenses transfered 2")
		assert.Contains(t, l, "migrate: number of links transfered 5")
		assert.Contains(t, l, "migrate: number of markers transfered 18")
		assert.Contains(t, l, "migrate: number of passcodes transfered 3")
		assert.Contains(t, l, "migrate: number of passwords transfered 11")
		assert.Contains(t, l, "migrate: number of photos transfered 58")
		assert.Contains(t, l, "migrate: number of photousers transfered 0")
		assert.Contains(t, l, "migrate: number of places transfered 10")
		assert.Contains(t, l, "migrate: number of reactions transfered 3")
		assert.Contains(t, l, "migrate: number of sessions transfered 21")
		assert.Contains(t, l, "migrate: number of services transfered 2")
		assert.Contains(t, l, "migrate: number of subjects transfered 6")
		assert.Contains(t, l, "migrate: number of users transfered 11")
		assert.Contains(t, l, "migrate: number of userdetails transfered 9")
		assert.Contains(t, l, "migrate: number of usersettings transfered 13")
		assert.Contains(t, l, "migrate: number of usershares transfered 1")

		// Make sure that a sequence update has worked.
		testdb, err := gorm.Open(postgres.Open(tfDSN.ToString()), &gorm.Config{})
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
			os.Remove("/go/src/github.com/photoprism/photoprism/storage/sqlitetopostgresql.test.db")
		}
	})

}

func copyFile(source, target string) error {
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("copyFile: source file %s is required", source)
	}

	if _, err := os.Stat(target); err != nil {
		if err = os.Remove(target); err != nil {
			if !strings.Contains(err.Error(), "no such file or directory") {
				return fmt.Errorf("copyFile: target file %s can not be removed with error %s", target, err.Error())
			}
		}
	}

	sourceFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("copyFile: source file %s can not be opened with error %s", source, err.Error())
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("copyFile: target file %s can not be opened with error %s", target, err.Error())
	}

	defer func() {
		closeErr := targetFile.Close()
		if err == nil {
			err = closeErr
		}
	}()

	if _, err = io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("copyFile: copy failed with error %s", err.Error())
	}

	if err = targetFile.Sync(); err != nil {
		return fmt.Errorf("copyFile: target sync failed with error %s", err.Error())
	}

	return nil
}
