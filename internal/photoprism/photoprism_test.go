package photoprism

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)

	db := entity.InitTestDb(os.Getenv("PHOTOPRISM_TEST_DSN"))

	code := m.Run()

	if db != nil {
		db.Close()
	}

	os.Exit(code)
}
