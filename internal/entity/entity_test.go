package entity

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)

	db := InitTestDb(os.Getenv("PHOTOPRISM_TEST_DRIVER"), os.Getenv("PHOTOPRISM_TEST_DSN"))
	defer db.Close()

	code := m.Run()

	os.Exit(code)
}
