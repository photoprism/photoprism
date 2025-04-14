package vision

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/api/download"
	"github.com/photoprism/photoprism/internal/event"
)

func TestMain(m *testing.M) {
	// Init test logger.
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log
	download.AllowedPaths = append(download.AllowedPaths, AssetsPath)

	// Set test config values.
	DownloadUrl = "https://app.localssl.dev/api/v1/dl"
	ServiceUri = ""

	// Run unit tests.
	code := m.Run()

	os.Exit(code)
}
