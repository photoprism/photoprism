package entity

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/event"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	db := InitTestDb(
		os.Getenv("PHOTOPRISM_TEST_DRIVER"),
		os.Getenv("PHOTOPRISM_TEST_DSN"))

	defer db.Close()

	code := m.Run()

	os.Exit(code)
}

func TestTypeString(t *testing.T) {
	assert.Equal(t, "unknown", TypeString(""))
	assert.Equal(t, "foo", TypeString("foo"))
}
