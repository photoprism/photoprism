package photoprism

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	os.Exit(runMain(m))
}

// runMain initializes package-level test state and returns the test exit code.
func runMain(m *testing.M) int {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	// Isolate package fixtures per process so parallel package test runs do not
	// race on shared storage/testdata directories.
	testRoot, err := os.MkdirTemp("", "photoprism-test-*")
	if err != nil {
		log.Errorf("create test root: %v", err)
		return 1
	}
	defer os.RemoveAll(testRoot)

	if err = os.Setenv("PHOTOPRISM_STORAGE_PATH", filepath.Join(testRoot, "storage")); err != nil {
		log.Errorf("set PHOTOPRISM_STORAGE_PATH: %v", err)
		return 1
	}

	if err = os.Setenv("PHOTOPRISM_TEST_DRIVER", "sqlite"); err != nil {
		log.Errorf("set PHOTOPRISM_TEST_DRIVER: %v", err)
		return 1
	}

	testDsn := filepath.Join(testRoot, "photoprism-test.db") + "?_busy_timeout=5000"
	if err = os.Setenv("PHOTOPRISM_TEST_DSN", testDsn); err != nil {
		log.Errorf("set PHOTOPRISM_TEST_DSN: %v", err)
		return 1
	}

	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)

	c := config.NewTestConfig("photoprism")
	SetConfig(c)

	defer func() {
		if err = c.CloseDb(); err != nil {
			log.Warnf("close db: %v", err)
		}

		// Remove temporary SQLite files after running the tests.
		fs.PurgeTestDbFiles(".", false)
	}()

	return m.Run()
}
