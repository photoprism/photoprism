package get

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "internal-photoprism-get")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)
	c := config.NewMinimalTestConfigWithDb("test", tempDir)

	SetConfig(c)
	defer c.CloseDb()

	code := m.Run()

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}
