package commands

import (
	"sync"
	"testing"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/capture"
)

func TestIndexCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	s := event.Subscribe("log.info")
	defer event.Unsubscribe(s)

	// The receiver runs until the subscription closes, so the log it collects is
	// read while that goroutine may still be appending to it.
	var mu sync.Mutex
	var l string

	assert.IsType(t, hub.Subscription{}, s)

	go func() {
		for msg := range s.Receiver {
			mu.Lock()
			l += msg.Fields["message"].(string) + "\n"
			mu.Unlock()
		}
	}()

	stdout := capture.Output(func() {
		err = IndexCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	if stdout != "" {
		t.Logf("stdout: %s", stdout)
	}

	time.Sleep(time.Second)

	mu.Lock()
	logged := l
	mu.Unlock()

	// Check command output.
	if logged != "" {
		assert.NotContains(t, logged, "error")
		assert.NotContains(t, logged, "warning")
	} else {
		t.Fatal("log output missing")
	}
}
