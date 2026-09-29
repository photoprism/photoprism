package commands

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/leandro-lugaresi/hub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/fs"
)

// TestIndexCommand checks index output while preserving the package fixtures.
func TestIndexCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	if commandTestProcess(t) {
		return
	}
	previous := requireTestDb(t)
	camera := entity.CameraFixtures.Get("canon-eos-7d")
	lens := entity.LensFixtures.Get("4-37")
	session := entity.SessionFixtures.Get("client_analytics")
	t.Cleanup(func() {
		previous.RegisterDb()
		require.NoError(t, previous.Db().Where("id = ?", camera.ID).First(&entity.Camera{}).Error)
		require.NoError(t, previous.Db().Where("id = ?", lens.ID).First(&entity.Lens{}).Error)
		require.NoError(t, previous.Db().Where("id = ?", session.ID).First(&entity.Session{}).Error)
	})
	conf := resetConfigAndOpenDB(t)
	dir, err := os.MkdirTemp(conf.OriginalsPath(), "index-command-")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(dir))
		require.NoError(t, os.RemoveAll(filepath.Join(conf.SidecarPath(), filepath.Base(dir))))
	})
	fixture := os.Getenv("PHOTOPRISM_TEST_IMAGE_FILE")
	if fixture == "" {
		fixture = "../../pkg/fs/testdata/directory/example.jpg"
	}
	require.NoError(t, fs.Copy(fixture, filepath.Join(dir, "example.jpg"), false))

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
