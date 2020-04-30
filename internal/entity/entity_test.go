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

	db := InitTestDb(os.Getenv("PHOTOPRISM_TEST_DSN"))

	code := m.Run()

	db.Close()

	os.Exit(code)
}
