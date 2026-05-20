package performancetest

import (
	"os/exec"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

func BenchmarkMigration_SQLite(b *testing.B) {
	driver, _ := dsn.PhotoPrismTestToDriverDSN(0)
	if driver != "sqlite" {
		b.Skip("skipping benchmark as not SQLite")
	}

	// Setup here
	loglevel := event.Log.GetLevel()

	// tests here

	b.Run("OneKUpgradeTest_Custom", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-1k.original.sqlite") {
			log.Info("Generating SQLite database with 1000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			require.NoError(b, generateDatabase(1000, "sqlite", "../../storage/test-1k.original.sqlite", true, true))
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			sqliteMigration("../../storage/test-1k.original.sqlite", "../../storage/test-1k.db", 1000, false, "OneKUpgradeTest_Custom", time.Minute, b)
		}
	})

	b.Run("OneKUpgradeTest_Auto", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-1k.original.sqlite") {
			log.Info("Generating SQLite database with 1000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			require.NoError(b, generateDatabase(1000, "sqlite", "../../storage/test-1k.original.sqlite", true, true))
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			sqliteMigration("../../storage/test-1k.original.sqlite", "../../storage/test-1k.db", 1000, true, "OneKUpgradeTest_Auto", time.Minute, b)
		}
	})

	b.Run("TenKUpgradeTest_Custom", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-10k.original.sqlite") {
			log.Info("Generating SQLite database with 10000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			require.NoError(b, generateDatabase(10000, "sqlite", "../../storage/test-10k.original.sqlite", true, true))
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			sqliteMigration("../../storage/test-10k.original.sqlite", "../../storage/test-10k.db", 10000, false, "TenKUpgradeTest_Custom", 2*time.Minute, b)
		}
	})

	b.Run("TenKUpgradeTest_Auto", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-10k.original.sqlite") {
			log.Info("Generating SQLite database with 10000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			require.NoError(b, generateDatabase(10000, "sqlite", "../../storage/test-10k.original.sqlite", true, true))
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			sqliteMigration("../../storage/test-10k.original.sqlite", "../../storage/test-10k.db", 10000, true, "TenKUpgradeTest_Auto", 2*time.Minute, b)
		}
	})

	b.Run("OneHundredKUpgradeTest_Custom", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-100k.original.sqlite") {
			log.Info("Generating SQLite database with 100000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			require.NoError(b, generateDatabase(100000, "sqlite", "../../storage/test-100k.original.sqlite", true, true))
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			sqliteMigration("../../storage/test-100k.original.sqlite", "../../storage/test-100k.db", 100000, false, "OneHundredKUpgradeTest_Custom", 20*time.Minute, b)
		}
	})

	b.Run("OneHundredKUpgradeTest_Auto", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-100k.original.sqlite") {
			log.Info("Generating SQLite database with 100000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			require.NoError(b, generateDatabase(100000, "sqlite", "../../storage/test-100k.original.sqlite", true, true))
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			sqliteMigration("../../storage/test-100k.original.sqlite", "../../storage/test-100k.db", 100000, true, "OneHundredKUpgradeTest_Auto", 20*time.Minute, b)
		}
	})

	// teardown here
	event.Log.SetLevel(loglevel)
}

func BenchmarkMigration_MySQL(b *testing.B) {
	driver, _ := dsn.PhotoPrismTestToDriverDSN(0)
	if driver != "mysql" {
		b.Skip("skipping benchmark as not MariaDB")
	}

	// Setup here
	loglevel := event.Log.GetLevel()

	b.Run("OneKUpgradeTest", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-1k.original.mysql") {
			log.Info("Generating Mariadb database with 1000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			d := dsn.TestDSNPortFromEnv(dsn.DriverMariaDB, "migrate", 1)
			require.NoError(b, generateDatabase(1000, "mysql", d.ToString(), true, true))
			resultFile := "--result-file=" + "../../storage/test-1k.original.mysql"
			if err := exec.Command("mariadb-dump", "--user=migrate", "--password=migrate", "--lock-tables", "--add-drop-database", "--databases", "migrate", resultFile).Run(); err != nil {
				b.Fatal(err)
			}
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			mysqlMigration("../../storage/test-1k.original.mysql", 1000, "OneKUpgradeTest", time.Minute, b)
		}
	})

	b.Run("TenKUpgradeTest", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-10k.original.mysql") {
			log.Info("Generating Mariadb database with 10000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			d := dsn.TestDSNPortFromEnv(dsn.DriverMariaDB, "migrate", 1)
			require.NoError(b, generateDatabase(10000, "mysql", d.ToString(), true, true))
			resultFile := "--result-file=" + "../../storage/test-10k.original.mysql"
			if err := exec.Command("mariadb-dump", "--user=migrate", "--password=migrate", "--lock-tables", "--add-drop-database", "--databases", "migrate", resultFile).Run(); err != nil {
				b.Fatal(err)
			}
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			mysqlMigration("../../storage/test-10k.original.mysql", 10000, "TenKUpgradeTest", 2*time.Minute, b)
		}
	})

	b.Run("OneHundredKUpgradeTest", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-100k.original.mysql") {
			log.Info("Generating Mariadb database with 100000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			d := dsn.TestDSNPortFromEnv(dsn.DriverMariaDB, "migrate", 1)
			require.NoError(b, generateDatabase(100000, "mysql", d.ToString(), true, true))
			resultFile := "--result-file=" + "../../storage/test-100k.original.mysql"
			if err := exec.Command("mariadb-dump", "--user=migrate", "--password=migrate", "--lock-tables", "--add-drop-database", "--databases", "migrate", resultFile).Run(); err != nil {
				b.Fatal(err)
			}
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			mysqlMigration("../../storage/test-100k.original.mysql", 100000, "OneHundredKUpgradeTest", 20*time.Minute, b)
		}
	})
	// teardown here
	event.Log.SetLevel(loglevel)
}

func BenchmarkMigration_PostgreSQL(b *testing.B) {
	driver, _ := dsn.PhotoPrismTestToDriverDSN(0)
	if driver != "postgres" {
		b.Skip("skipping benchmark as not PostgreSQL")
	}

	// Setup here
	loglevel := event.Log.GetLevel()
	mDSN := dsn.TestDSNPortFromEnv(dsn.DriverPostgreSQL, "migrate", 1)
	// tests here

	b.Run("OneKUpgradeTest", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-1k.original.postgresql") {
			log.Info("Generating PostgreSQL database with 1000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			require.NoError(b, generateDatabase(1000, Postgres, mDSN.ToString(), true, true))
			resultFile := "../../storage/test-1k.original.postgresql"
			if err := exec.Command("pg_dump", "-d", mDSN.ForPSQL(), "-F c", "-f", resultFile).Run(); err != nil { //nolint:gosec // test generated input
				b.Fatal(err)
			}
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			postgresqlMigration("../../storage/test-1k.original.postgresql", 1000, "OneKUpgradeTest", time.Minute, b)
		}
	})

	b.Run("TenKUpgradeTest", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-10k.original.postgresql") {
			log.Info("Generating PostgreSQL database with 10000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			require.NoError(b, generateDatabase(10000, Postgres, mDSN.ToString(), true, true))
			resultFile := "../../storage/test-10k.original.postgresql"
			if err := exec.Command("pg_dump", "-d", mDSN.ForPSQL(), "-F c", "-f", resultFile).Run(); err != nil { //nolint:gosec // test generated input
				b.Fatal(err)
			}
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			postgresqlMigration("../../storage/test-10k.original.postgresql", 10000, "TenKUpgradeTest", 2*time.Minute, b)
		}
	})

	b.Run("OneHundredKUpgradeTest", func(b *testing.B) {
		if !fs.FileExists("../../storage/test-100k.original.postgresql") {
			log.Info("Generating PostgreSQL database with 100000 records")
			event.Log.SetLevel(logrus.ErrorLevel)
			require.NoError(b, generateDatabase(100000, Postgres, mDSN.ToString(), true, true))
			resultFile := "../../storage/test-100k.original.postgresql"
			if err := exec.Command("pg_dump", "-d", mDSN.ForPSQL(), "-F c", "-f", resultFile).Run(); err != nil { //nolint:gosec // test generated input
				b.Fatal(err)
			}
			event.Log.SetLevel(loglevel)
		}
		for b.Loop() {
			postgresqlMigration("../../storage/test-100k.original.postgresql", 100000, "OneHundredKUpgradeTest", 20*time.Minute, b)
		}
	})
	// teardown here
	event.Log.SetLevel(loglevel)
}
