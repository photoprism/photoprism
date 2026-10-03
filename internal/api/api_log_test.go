package api

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/event"
)

// captureLog replaces the package logger for the duration of a test.
func captureLog(t *testing.T) *test.Hook {
	t.Helper()

	orig := log
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	log = logger

	t.Cleanup(func() { log = orig })

	return hook
}

func TestLogErr(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		hook := captureLog(t)
		logErr("upload", nil)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("PathIsRemoved", func(t *testing.T) {
		hook := captureLog(t)
		logErr("upload", &iofs.PathError{Op: "remove", Path: "/photoprism/storage/users/uqxc08w3d0ej2283/upload/x", Err: iofs.ErrNotExist})
		entries := hook.AllEntries()

		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		assert.Equal(t, logrus.ErrorLevel, entries[0].Level)
		assert.Contains(t, entries[0].Message, "upload: ")
		assert.Contains(t, entries[0].Message, "***")
		assert.NotContains(t, entries[0].Message, "/photoprism/storage")
		assert.NotContains(t, entries[0].Message, "uqxc08w3d0ej2283")
	})
	t.Run("WrappedPathIsRemoved", func(t *testing.T) {
		hook := captureLog(t)
		cause := &os.LinkError{Op: "rename", Old: "/srv/data/a.jpg", New: "/srv/data/b.jpg", Err: iofs.ErrPermission}
		logErr("upload", fmt.Errorf("could not move the file: %w", cause))
		entries := hook.AllEntries()

		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		assert.NotContains(t, entries[0].Message, "/srv/data")
		assert.Contains(t, entries[0].Message, "could not move the file")
	})
	t.Run("ControlCharactersAreRemoved", func(t *testing.T) {
		hook := captureLog(t)
		logErr("upload", errors.New("rejected\n2026-01-01 ERRO forged entry"))
		entries := hook.AllEntries()

		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		assert.NotContains(t, entries[0].Message, "\n")
		assert.NotContains(t, entries[0].Message, "\r")
	})
	t.Run("LengthIsBounded", func(t *testing.T) {
		hook := captureLog(t)
		logErr("upload", errors.New(strings.Repeat("a", 64000)))
		entries := hook.AllEntries()

		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		assert.Less(t, len(entries[0].Message), 8192)
	})
}

func TestLogWarn(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		hook := captureLog(t)
		logWarn("upload", nil)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("PathIsRemoved", func(t *testing.T) {
		hook := captureLog(t)
		logWarn("upload", &iofs.PathError{Op: "remove", Path: "/photoprism/storage/users/uqxc08w3d0ej2283/upload/x.zip", Err: iofs.ErrNotExist})
		entries := hook.AllEntries()

		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}

		assert.Equal(t, logrus.WarnLevel, entries[0].Level)
		assert.Contains(t, entries[0].Message, "***")
		assert.NotContains(t, entries[0].Message, "/photoprism/storage")
		assert.NotContains(t, entries[0].Message, "uqxc08w3d0ej2283")
	})
	t.Run("TypedNilDoesNotPanic", func(t *testing.T) {
		hook := captureLog(t)
		var cause *os.SyscallError
		assert.NotPanics(t, func() { logWarn("upload", fmt.Errorf("cleanup: %w", cause)) })
		assert.Len(t, hook.AllEntries(), 1)
	})
}

// captureSystemLog replaces the console logger for the duration of a test.
func captureSystemLog(t *testing.T) *test.Hook {
	t.Helper()

	orig := event.SystemLog
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	event.SystemLog = logger

	t.Cleanup(func() { event.SystemLog = orig })

	return hook
}

func TestSystemErr(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		hook := captureLog(t)
		systemErr("metrics", nil)
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("ReachesTheConsoleAndNotTheBrowser", func(t *testing.T) {
		// Both copies are asserted. System renders the console one outside Format, so a segment
		// list written for Format alone renders there as a literal verb; and the package logger
		// is what feeds a browser, so systemErr must not reach it.
		hook := captureLog(t)
		console := captureSystemLog(t)

		s := event.Subscribe("system.log.error")
		defer event.Unsubscribe(s)

		systemErr("metrics", &net.OpError{
			Op:   "write",
			Net:  "unix",
			Addr: &net.UnixAddr{Name: "/photoprism/storage/vision.sock", Net: "unix"},
			Err:  errors.New("broken pipe"),
		})

		assert.Empty(t, hook.AllEntries())

		entries := console.AllEntries()
		require.Len(t, entries, 1)
		assert.Equal(t, "metrics: write unix /photoprism/storage/vision.sock: broken pipe", entries[0].Message)

		select {
		case msg := <-s.Receiver:
			// ErrorFull, because the reader of this channel is the operator who acts on the path.
			assert.Contains(t, msg.Fields["message"], "/photoprism/storage/vision.sock")
			assert.Contains(t, msg.Fields["message"], "metrics")
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for the system event")
		}
	})
}
