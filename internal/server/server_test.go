package server

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/server/limiter"
)

func TestMain(m *testing.M) {
	// Init test logger.
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	// Init test config.
	c := config.TestConfig()
	get.SetConfig(c)

	// Increase login rate limit for testing.
	limiter.Login = limiter.NewLimit(1, 10000)

	// Run unit tests.
	code := m.Run()

	// Close database connection.
	_ = c.CloseDb()

	os.Exit(code)
}
