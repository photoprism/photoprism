package entity

import (
	"bytes"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var logBuffer bytes.Buffer

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.Out = &logBuffer
	log.SetLevel(logrus.DebugLevel)
	code := m.Run()
	os.Exit(code)
}

func TestID(t *testing.T) {
	for n := 0; n < 5; n++ {
		uuid := ID('x')
		t.Logf("id: %s", uuid)
		assert.Equal(t, len(uuid), 17)
	}
}

func BenchmarkID(b *testing.B) {
	for n := 0; n < b.N; n++ {
		ID('x')
	}
}
