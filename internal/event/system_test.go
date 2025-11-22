package event

import (
	"io"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type systemTestLogger struct {
	*logrus.Logger
	mu      sync.Mutex
	entries []logEntry
}

type logEntry struct {
	level logrus.Level
	args  []interface{}
}

func newSystemTestLogger() *systemTestLogger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	return &systemTestLogger{Logger: logger}
}

func (l *systemTestLogger) Log(level logrus.Level, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entryArgs := append([]interface{}(nil), args...)
	l.entries = append(l.entries, logEntry{level: level, args: entryArgs})
}

func (l *systemTestLogger) lastEntry() (logEntry, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.entries) == 0 {
		return logEntry{}, false
	}

	return l.entries[len(l.entries)-1], true
}

func (l *systemTestLogger) entryCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.entries)
}

func TestSystemLoggingFunctions(t *testing.T) {
	tests := []struct {
		name  string
		level logrus.Level
		call  func(ev []string, args ...interface{})
	}{
		{name: "Info", level: logrus.InfoLevel, call: func(ev []string, args ...interface{}) { SystemInfo(ev, args...) }},
		{name: "Warn", level: logrus.WarnLevel, call: func(ev []string, args ...interface{}) { SystemWarn(ev, args...) }},
		{name: "Debug", level: logrus.DebugLevel, call: func(ev []string, args ...interface{}) { SystemDebug(ev, args...) }},
		{name: "Error", level: logrus.ErrorLevel, call: func(ev []string, args ...interface{}) { SystemError(ev, args...) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newSystemTestLogger()
			orig := SystemLog
			SystemLog = logger
			t.Cleanup(func() { SystemLog = orig })

			events := []string{"cleanup %s", "finished"}
			args := []interface{}{"cache"}
			expectedMessage := Format(events, args...)
			topic := "system.log." + tt.level.String()

			subscription := Subscribe(topic)
			t.Cleanup(func() { Unsubscribe(subscription) })

			tt.call(events, args...)

			var msg Message
			select {
			case msg = <-subscription.Receiver:
			case <-time.After(100 * time.Millisecond):
				t.Fatalf("timeout waiting for %s event", topic)
			}

			require.Equal(t, topic, msg.Name)
			require.Equal(t, tt.level.String(), msg.Fields["level"])
			require.Equal(t, expectedMessage, msg.Fields["message"])

			require.IsType(t, time.Time{}, msg.Fields["time"])
			timeValue := msg.Fields["time"].(time.Time)
			assert.False(t, timeValue.IsZero())

			entry, ok := logger.lastEntry()
			require.True(t, ok)
			assert.Equal(t, tt.level, entry.level)
			require.Len(t, entry.args, 1)
			assert.Equal(t, SystemPrefix+expectedMessage, entry.args[0])
		})
	}
}

func TestSystemSkipsEmptyEvents(t *testing.T) {
	logger := newSystemTestLogger()
	orig := SystemLog
	SystemLog = logger
	defer func() { SystemLog = orig }()

	topic := "system.log." + logrus.InfoLevel.String()
	subscription := Subscribe(topic)
	defer Unsubscribe(subscription)

	System(logrus.InfoLevel, nil)
	System(logrus.InfoLevel, []string{})
	SystemInfo(nil)

	assert.Equal(t, 0, logger.entryCount())

	select {
	case msg := <-subscription.Receiver:
		t.Fatalf("unexpected message received: %#v", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestSystemPublishesWithoutLogger(t *testing.T) {
	orig := SystemLog
	SystemLog = nil
	defer func() { SystemLog = orig }()

	events := []string{"maintenance"}
	expectedMessage := Format(events)
	topic := "system.log." + logrus.InfoLevel.String()

	subscription := Subscribe(topic)
	defer Unsubscribe(subscription)

	System(logrus.InfoLevel, events)

	var msg Message
	select {
	case msg = <-subscription.Receiver:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout waiting for %s event", topic)
	}

	require.Equal(t, topic, msg.Name)
	require.Equal(t, logrus.InfoLevel.String(), msg.Fields["level"])
	require.Equal(t, expectedMessage, msg.Fields["message"])

	require.IsType(t, time.Time{}, msg.Fields["time"])
	timeValue := msg.Fields["time"].(time.Time)
	assert.False(t, timeValue.IsZero())
}
