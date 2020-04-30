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

	db := InitTestDb("photoprism:photoprism@tcp(photoprism-db:4001)/photoprism?parseTime=true")

	code := m.Run()

	db.Close()

	os.Exit(code)
}
