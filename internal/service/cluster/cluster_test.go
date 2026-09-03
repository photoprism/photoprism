package cluster

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/testextras"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {

	// Run unit tests.
	return testextras.TestDbCleanup(m.Run())
}
