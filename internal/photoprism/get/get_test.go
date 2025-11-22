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
	c := config.NewMinimalTestConfigWithDb("test", tempDir)

	SetConfig(c)

	code := m.Run()

	if err = c.CloseDb(); err != nil {
		log.Errorf("close db: %v", err)
	}

	if err = os.RemoveAll(tempDir); err != nil {
		log.Errorf("remove temp dir: %v", err)
	}

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}
