package node

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/testextras"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	// Remove temporary SQLite files before running the tests.
	fs.PurgeTestDbFiles(".", false)
	// Remove temporary SQLite files after running the tests.
	defer fs.PurgeTestDbFiles(".", false)

	// Run unit tests.
	return testextras.TestDbCleanup(m.Run())
}
