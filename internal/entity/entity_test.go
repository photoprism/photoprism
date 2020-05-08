package entity

import (
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)

	dsn := os.Getenv("PHOTOPRISM_TEST_DSN")

	if dsn == "" {
		panic("database dsn is empty")
	}

	db := InitTestDb(strings.Replace(dsn, "/photoprism", "/entity", 1))

	code := m.Run()

	if db != nil {
		db.Close()
	}

	os.Exit(code)
}
