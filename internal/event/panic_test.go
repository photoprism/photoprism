package event

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// newTestLogger returns a logrus logger writing to buf at error level.
func newTestLogger(buf *bytes.Buffer) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(buf)
	logger.SetLevel(logrus.ErrorLevel)
	return logger
}

func TestLogPanic(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var sysBuf, logBuf bytes.Buffer
		origSys, origLog := SystemLog, Log
		SystemLog = newTestLogger(&sysBuf)
		Log = newTestLogger(&logBuf)
		t.Cleanup(func() { SystemLog = origSys; Log = origLog })

		LogPanic("boom")

		out := sysBuf.String()
		assert.Contains(t, out, "panic")
		assert.Contains(t, out, "boom")
		assert.Contains(t, out, "goroutine")
		// Must not be routed through the default, database-persisted logger.
		assert.Empty(t, logBuf.String())
	})
	t.Run("Nil", func(t *testing.T) {
		var buf bytes.Buffer
		orig := SystemLog
		SystemLog = newTestLogger(&buf)
		t.Cleanup(func() { SystemLog = orig })

		LogPanic(nil)

		assert.Empty(t, buf.String())
	})
	t.Run("NoLogger", func(t *testing.T) {
		orig := SystemLog
		SystemLog = nil
		t.Cleanup(func() { SystemLog = orig })

		assert.NotPanics(t, func() { LogPanic("boom") })
	})
}

func TestRecover(t *testing.T) {
	t.Run("Panic", func(t *testing.T) {
		var buf bytes.Buffer
		orig := SystemLog
		SystemLog = newTestLogger(&buf)
		t.Cleanup(func() { SystemLog = orig })

		func() {
			defer Recover()
			panic("kaboom")
		}()

		out := buf.String()
		assert.Contains(t, out, "panic")
		assert.Contains(t, out, "kaboom")
		assert.Contains(t, out, "goroutine")
	})
	t.Run("NoPanic", func(t *testing.T) {
		var buf bytes.Buffer
		orig := SystemLog
		SystemLog = newTestLogger(&buf)
		t.Cleanup(func() { SystemLog = orig })

		func() { defer Recover() }()

		assert.Empty(t, buf.String())
	})
}

func TestSafe(t *testing.T) {
	t.Run("RecoversAndContinues", func(t *testing.T) {
		var buf bytes.Buffer
		orig := SystemLog
		SystemLog = newTestLogger(&buf)
		t.Cleanup(func() { SystemLog = orig })

		ran := 0
		assert.NotPanics(t, func() {
			Safe(func() { panic("boom") })
			Safe(func() { ran++ })
		})

		assert.Equal(t, 1, ran)
		assert.Contains(t, buf.String(), "panic")
		assert.Contains(t, buf.String(), "boom")
	})
	t.Run("Success", func(t *testing.T) {
		called := false
		Safe(func() { called = true })
		assert.True(t, called)
	})
}
