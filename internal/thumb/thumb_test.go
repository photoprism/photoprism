package thumb

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
	log.SetLevel(logrus.TraceLevel)

	code := m.Run()

	// remove temporary test files
	_ = os.RemoveAll("testdata/1")

	os.Exit(code)
}
