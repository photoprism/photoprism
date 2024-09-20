package entity

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log

// Log logs the error if any and keeps quiet otherwise.
func Log(model, action string, err error) {
	if err != nil {
		log.Errorf("%s: %s (%s)", model, err, action)
	}
}

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	driver := os.Getenv("PHOTOPRISM_TEST_DRIVER")
	dsn := ""

	if driver == "mysql" {
		dsn = os.Getenv("PHOTOPRISM_DATABASE_USER") + ":" + os.Getenv("PHOTOPRISM_DATABASE_PASSWORD") + "@(mariadb:4001)/pptest?charset=utf8mb4&parseTime=True&loc=Local"
	} else if driver == "sqlite" {
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN")
	} else {
		driver = os.Getenv("PHOTOPRISM_TEST_DRIVER")
		dsn = os.Getenv("PHOTOPRISM_TEST_DSN")
	}

	db := entity.InitTestDb(
		driver,
		dsn)

	defer db.Close()

	code := m.Run()

	os.Exit(code)
}
