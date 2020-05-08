package entity

import (
	"bytes"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

var logBuffer bytes.Buffer

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.Out = &logBuffer
	log.SetLevel(logrus.DebugLevel)

	dsn := os.Getenv("PHOTOPRISM_TEST_DSN")

	if dsn == "" {
		panic("database dsn is empty")
	}

	db := InitTestDb(dsn)

	code := m.Run()

	if db != nil {
		db.Close()
	}

	os.Exit(code)
}
