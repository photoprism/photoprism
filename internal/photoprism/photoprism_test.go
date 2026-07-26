package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

// useTestDb points the configs a test creates at an isolated SQLite database,
// so its rows stay separate from those of the package-wide test config.
//
// The driver must be set alongside the DSN. Left at the package default, a
// SQLite path is parsed as a MySQL DSN when the suite runs on MariaDB, and the
// resulting config error terminates the whole test binary.
func useTestDb(t *testing.T, name string) {
	t.Helper()

	t.Setenv("PHOTOPRISM_TEST_DRIVER", dsn.DriverSQLite3)
	t.Setenv("PHOTOPRISM_TEST_DSN", filepath.Join(t.TempDir(), name+".db"))
}

func runTestMain(m *testing.M) int {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	c := config.NewTestConfig("photoprism")
	config.OnceTestConfig(c)
	SetConfig(c)
	defer c.CleanupTestFolder()
	defer func() {
		if err := c.CloseDb(); err != nil {
			log.Warnf("close db: %v", err)
		}
		// Remove temporary SQLite files after running the tests.
		fs.PurgeTestDbFiles(".", false)
	}()

	return m.Run()
}
