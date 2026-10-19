package mcp

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

// runTestMain opens the test database with fixtures, which the data tools read.
func runTestMain(m *testing.M) int {
	// Remove temporary SQLite files before and after running the tests.
	fs.PurgeTestDbFiles(".", false)
	defer fs.PurgeTestDbFiles(".", false)

	db := entity.InitTestDb(os.Getenv("PHOTOPRISM_TEST_DRIVER"), os.Getenv("PHOTOPRISM_TEST_DSN"))
	defer db.Close()

	return m.Run()
}
