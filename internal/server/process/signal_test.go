package process

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// receive waits for a signal on the server process channel.
func receive(t *testing.T) os.Signal {
	t.Helper()

	select {
	case sig := <-Signal:
		return sig
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for signal")
		return nil
	}
}

func TestRestart(t *testing.T) {
	t.Run("SendsUsr1", func(t *testing.T) {
		go Restart()
		assert.Equal(t, syscall.SIGUSR1, receive(t))
	})
}

func TestShutdown(t *testing.T) {
	t.Run("SendsTerm", func(t *testing.T) {
		go Shutdown()
		assert.Equal(t, syscall.SIGTERM, receive(t))
	})
}
