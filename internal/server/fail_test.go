package server

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/server/process"
)

// receiveSignal waits for Fail to send a signal to the server process channel.
func receiveSignal(t *testing.T) os.Signal {
	t.Helper()

	select {
	case sig := <-process.Signal:
		return sig
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for shutdown signal")
		return nil
	}
}

func TestFail(t *testing.T) {
	t.Run("LogsErrorAndShutsDown", func(t *testing.T) {
		hook := captureRecoveryLog(t, logrus.InfoLevel)

		go Fail("server: %s socket %s already exists", "unix", "/tmp/server.sock")

		assert.Equal(t, syscall.SIGTERM, receiveSignal(t))
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
		assert.Equal(t, "server: unix socket /tmp/server.sock already exists", hook.LastEntry().Message)
	})
	t.Run("EmptyMessage", func(t *testing.T) {
		hook := captureRecoveryLog(t, logrus.InfoLevel)

		go Fail("")

		assert.Equal(t, syscall.SIGTERM, receiveSignal(t))
		assert.Empty(t, hook.AllEntries())
	})
}
