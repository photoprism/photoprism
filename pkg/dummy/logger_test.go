package dummy

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	log := NewLogger()

	log.Fatal("foo", 1, []string{}, nil)
	log.Fatalf("foo", 1, []string{}, nil)
	log.Fatalln()
	log.Panic("foo", 1, []string{}, nil)
	log.Panicf("foo", 1, []string{}, nil)
	log.Panicln()

	assert.Equal(t, logrus.PanicLevel, log.GetLevel())
	log.SetLevel(logrus.TraceLevel)
	assert.Equal(t, logrus.TraceLevel, log.GetLevel())

	log.Fatal("foo", 1, []string{}, nil)
	log.Fatalf("foo", 1, []string{}, nil)
	log.Fatalln()
	log.Panic("foo", 1, []string{}, nil)
	log.Panicf("foo", 1, []string{}, nil)
	log.Panicln()
}
