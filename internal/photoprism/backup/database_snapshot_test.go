package backup

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/dsn"
)

// mariadbTestAdmin returns connection values for the administrative user of the MariaDB test server,
// or skips the test if the tests do not run against one.
func mariadbTestAdmin(t *testing.T) mariadbConn {
	t.Helper()

	if os.Getenv("PHOTOPRISM_TEST_DRIVER") != dsn.DriverMySQL {
		t.Skip("requires the MariaDB test database")
	}

	bin, err := exec.LookPath("mariadb")

	if err != nil {
		t.Skip("mariadb client not found")
	}

	d := dsn.Parse(os.Getenv("PHOTOPRISM_TEST_DSN"))

	return mariadbConn{Bin: bin, Host: d.Host(), Port: strconv.Itoa(d.Port()), User: d.User, Password: d.Password, Name: "mysql"}
}

// mariadbTestQuery runs sql with conn and returns its output without column names.
func mariadbTestQuery(t *testing.T, conn mariadbConn, sql string) string {
	t.Helper()

	out, err := conn.Cmd("-N", "-e", sql).CombinedOutput()
	require.NoError(t, err, string(out))

	return strings.TrimSpace(string(out))
}

func TestMariadbDumpArgs(t *testing.T) {
	assert.Equal(t, []string{"--single-transaction", "--skip-add-locks", "--skip-no-autocommit"}, mariadbDumpArgs())
}

func TestWarnNonInnodbTables(t *testing.T) {
	t.Run("QueryFailed", func(t *testing.T) {
		// A database without information_schema cannot be checked.
		db, err := gorm.Open(dsn.DriverSQLite3, ":memory:")
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })
		hook := captureLog(t)
		warnNonInnodbTables(db, "photoprism")
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "backup: failed to check the storage engine of the database tables")
	})
	t.Run("MariaDB", func(t *testing.T) {
		admin := mariadbTestAdmin(t)
		name := fmt.Sprintf("zz_snapshot_%d", time.Now().UnixNano())
		t.Cleanup(func() { _ = admin.Cmd("-e", "DROP DATABASE IF EXISTS "+name).Run() })
		mariadbTestQuery(t, admin, "CREATE DATABASE "+name+"; CREATE TABLE "+name+".t (id INT) ENGINE=InnoDB; "+
			"CREATE VIEW "+name+".v AS SELECT id FROM "+name+".t")
		db := get.Config().Db()

		hook := captureLog(t)
		warnNonInnodbTables(db, name)
		assert.Empty(t, hook.AllEntries())

		mariadbTestQuery(t, admin, "CREATE TABLE "+name+".u (id INT) ENGINE=MyISAM; CREATE TABLE "+name+".a (id INT) ENGINE=Aria")
		warnNonInnodbTables(db, name)
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Equal(t, "backup: found 2 tables without InnoDB, which the consistent snapshot does not cover (a, u)",
			hook.LastEntry().Message)

		// The warning names the first tables only.
		for i := 0; i < clean.LogNamesLimit+2; i++ {
			mariadbTestQuery(t, admin, fmt.Sprintf("CREATE TABLE %s.m%02d (id INT) ENGINE=MyISAM", name, i))
		}
		warnNonInnodbTables(db, name)
		assert.Equal(t, "backup: found 14 tables without InnoDB, which the consistent snapshot does not cover "+
			"(a, m00, m01, m02, m03, m04, m05, m06, m07, m08 and 4 more)", hook.LastEntry().Message)
	})
}

// TestMariadbDump_LimitedUser dumps and restores a database as a user with only the privileges PhotoPrism
// needs to run, which do not include LOCK TABLES.
func TestMariadbDump_LimitedUser(t *testing.T) {
	admin := mariadbTestAdmin(t)
	dumpBin, err := exec.LookPath("mariadb-dump")

	if err != nil {
		t.Skip("mariadb-dump client not found")
	}

	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	src, dst, user, password := "zz_dump_src_"+suffix, "zz_dump_dst_"+suffix, "zz_dump_"+suffix, "Lim1ted-"+suffix
	grants := "SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, INDEX, ALTER"

	t.Cleanup(func() {
		_ = admin.Cmd("-e", "DROP DATABASE IF EXISTS "+src+"; DROP DATABASE IF EXISTS "+dst+"; DROP USER IF EXISTS '"+user+"'@'%'").Run()
	})
	mariadbTestQuery(t, admin, "CREATE DATABASE "+src+"; CREATE DATABASE "+dst+"; "+
		"CREATE USER '"+user+"'@'%' IDENTIFIED BY '"+password+"'; "+
		"GRANT "+grants+" ON "+src+".* TO '"+user+"'@'%'; GRANT "+grants+" ON "+dst+".* TO '"+user+"'@'%'; "+
		"CREATE TABLE "+src+".t (id INT PRIMARY KEY, v VARCHAR(20)) ENGINE=InnoDB; "+
		"INSERT INTO "+src+".t VALUES (1,'val-a'),(2,'val-b'),(3,'val-c')")

	limited := admin
	limited.User, limited.Password = user, password

	for _, engine := range []string{"", "MyISAM"} {
		t.Run("Engine"+engine, func(t *testing.T) {
			// A table that does not use InnoDB is dumped as well, outside the snapshot.
			if engine != "" {
				mariadbTestQuery(t, admin, "CREATE TABLE "+src+".m (id INT) ENGINE="+engine+"; INSERT INTO "+src+".m VALUES (7)")
			}

			dumper := limited
			dumper.Bin, dumper.Name = dumpBin, src

			var dump bytes.Buffer
			require.NoError(t, runDump(dumper.Cmd(mariadbDumpArgs()...), &dump, password))
			for _, stmt := range []string{"LOCK TABLES", "AUTOCOMMIT=0"} {
				assert.False(t, strings.Contains(strings.ToUpper(dump.String()), stmt), "dump contains %s", stmt)
			}

			// Duplicating the INSERT statement fails that statement only, and the rows of the first remain.
			start := strings.Index(dump.String(), "INSERT INTO `t` VALUES")
			require.GreaterOrEqual(t, start, 0)
			insert := dump.String()[start:]
			insert = insert[:strings.Index(insert, ";\n")+2]
			duplicated := strings.Replace(dump.String(), insert, insert+insert, 1)

			restorer := limited
			restorer.Name = dst

			for _, tc := range []struct {
				dump   string
				failed int
			}{{dump.String(), 0}, {duplicated, 1}} {
				mariadbTestQuery(t, admin, "DROP TABLE IF EXISTS "+dst+".t, "+dst+".m")
				failed, restoreErr := runRestore(restorer.Cmd(mariadbRestoreArgs(restorer.Bin)...), strings.NewReader(tc.dump), password)
				require.NoError(t, restoreErr)
				assert.Equal(t, tc.failed, failed.Count, "%v", failed.Errors)
				assert.Equal(t, "3", mariadbTestQuery(t, admin, "SELECT COUNT(*) FROM "+dst+".t"))
				if engine != "" {
					assert.Equal(t, "7", mariadbTestQuery(t, admin, "SELECT id FROM "+dst+".m"))
				}
			}
		})
	}
}
