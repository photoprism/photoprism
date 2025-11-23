package provisioner

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/pkg/fs"
)

// TestMain executes testMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

// TestMain ensures SQLite test DB artifacts are purged after the suite runs.
func testMain(m *testing.M) int {
	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)
	// Remove temporary SQLite files after running the tests.
	defer fs.PurgeTestDbFiles(".", false)

	// Run unit tests.
	return m.Run()
}
