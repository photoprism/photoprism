package avatar

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/fs"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	tempDir, err := os.MkdirTemp("", "avatar-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	c := config.NewMinimalTestConfigWithDb("avatar", tempDir)
	get.SetConfig(c)
	photoprism.SetConfig(c)
	defer c.CloseDb()

	code := m.Run()

	// Remove temporary SQLite files after running the tests.
	fs.PurgeTestDbFiles(".", false)

	os.Exit(code)
}
