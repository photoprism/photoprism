package commands

import (
	"testing"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/photoprism/photoprism/internal/event"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/capture"
)

func TestIndexCommand(t *testing.T) {
	var err error

	ctx := config.CliTestContext()

	s := event.Subscribe("log.info")
	defer event.Unsubscribe(s)
	logs := ""

	assert.IsType(t, hub.Subscription{}, s)

	go func() {
		for msg := range s.Receiver {
			logs += msg.Fields["message"].(string) + "\n"
		}
	}()

	stdout := capture.Output(func() {
		err = IndexCommand.Run(ctx)
	})

	if err != nil {
		t.Fatal(err)
	}

	if stdout != "" {
		t.Errorf("unexpected stdout output: %s", stdout)
	}

	time.Sleep(time.Second)

	if output := logs; output != "" {
		// Expected index command output.
		assert.Contains(t, output, "indexing originals")
		assert.Contains(t, output, "classify: loading labels")
		assert.Contains(t, output, "index: no .ppignore file found")
		assert.Contains(t, output, "index: updating primary files")
		assert.Contains(t, output, "index: flagging hidden files")
	} else {
		t.Fatal("log output missing")
	}
}
