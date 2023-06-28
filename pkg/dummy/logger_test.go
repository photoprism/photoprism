package dummy

import (
	"errors"
	"testing"
	"time"

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

func TestLogger_WithField(t *testing.T) {
	log := NewLogger()
	assert.Equal(t, "unit", log.WithField("test", "unit").Data["test"])
}

func TestLogger_WithFields(t *testing.T) {
	log := NewLogger()
	fields := logrus.Fields{"test": "unit", "color": "blue"}
	assert.Equal(t, "unit", log.WithFields(fields).Data["test"])
	assert.Equal(t, "blue", log.WithFields(fields).Data["color"])
}

func TestLogger_WithError(t *testing.T) {
	t.Run("Error for logger test", func(t *testing.T) {
		log := NewLogger()
		err := errors.New("Error for logger test")
		assert.Equal(t, "Error for logger test", log.WithError(err).Message)
	})
}

func TestLogger_WithTime(t *testing.T) {
	log := NewLogger()
	time := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, "2020-01-01 00:00:00 +0000 UTC", log.WithTime(time).Time.String())
}
