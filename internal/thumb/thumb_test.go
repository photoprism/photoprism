package thumb

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/event"
)

// TestMain executes runTestMain returning it's results.  It is done this way so that defer can be used to cleanup.
func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)
	event.AuditLog = log

	defer Shutdown()
	// Remove generated test files and folders.
	defer os.RemoveAll("testdata/1")
	defer os.RemoveAll("testdata/cache")
	defer os.RemoveAll("testdata/vips")

	return m.Run()
}

func TestNew(t *testing.T) {
	fileHash := "d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2"
	contentUri := "/content"
	previewToken := "preview-token"

	t.Run("Fit1280", func(t *testing.T) {
		result := New(1920, 1080, fileHash, Sizes[Fit1280], contentUri, previewToken)

		assert.Equal(t, 1280, result.W)
		assert.Equal(t, 720, result.H)
		assert.Equal(t, "/content/t/d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2/preview-token/fit_1280", result.Src)
	})
	t.Run("Fit3840", func(t *testing.T) {
		result := New(1920, 1080, fileHash, Sizes[Fit3840], contentUri, previewToken)

		assert.Equal(t, 1920, result.W)
		assert.Equal(t, 1080, result.H)
		assert.Equal(t, "/content/t/d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2/preview-token/fit_3840", result.Src)
	})
	t.Run("Fit4096", func(t *testing.T) {
		result := New(1920, 1080, fileHash, Sizes[Fit4096], contentUri, previewToken)

		assert.Equal(t, 1920, result.W)
		assert.Equal(t, 1080, result.H)
		assert.Equal(t, "/content/t/d2b4a5d18276f96f1b5a1bf17fd82d6fab3807f2/preview-token/fit_4096", result.Src)
	})
}
