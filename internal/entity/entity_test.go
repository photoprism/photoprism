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

	db := &Gorm{
		Driver: "mysql",
		Dsn:    "photoprism:photoprism@tcp(photoprism-db:4001)/photoprism?parseTime=true",
	}

	SetDbProvider(db)
	Migrate()

	code := m.Run()

	db.Close()

	os.Exit(code)
}
