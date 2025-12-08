package vision

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/api/download"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/fs"
)

var assetsPath = fs.Abs("../../../assets")
var examplesPath = filepath.Join(assetsPath, "examples")

func TestMain(m *testing.M) {
	// Init test logger.
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log
	download.AllowedPaths = append(download.AllowedPaths, assetsPath)

	// Set test config values.
	DownloadUrl = "https://app.localssl.dev/api/v1/dl"
	ServiceUri = ""

	// Run unit tests.
	code := m.Run()

	os.Exit(code)
}
