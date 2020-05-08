package query

import (
	"os"
	"testing"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)

	dsn := os.Getenv("PHOTOPRISM_TEST_DSN")

	if dsn == "" {
		panic("database dsn is empty")
	}

	db := entity.InitTestDb(dsn)

	code := m.Run()

	if db != nil {
		db.Close()
	}

	os.Exit(code)
}
