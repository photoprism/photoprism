package workers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
)

// TestIndex_StartFaceMigration pins that a run held off by a face migration reports the wait and
// neither purges, updates moments nor announces a completed run.
func TestIndex_StartFaceMigration(t *testing.T) {
	conf := config.TestConfig()

	lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)

	updating := event.Subscribe("index.*")
	defer event.Unsubscribe(updating)
	success := event.Subscribe("notify.success")
	defer event.Unsubscribe(success)

	err = NewIndex(conf).Start()
	require.ErrorIs(t, err, photoprism.ErrFaceMigration)

	select {
	case msg := <-updating.Receiver:
		t.Fatalf("unexpected index event %s", msg.Name)
	case msg := <-success.Receiver:
		t.Fatalf("unexpected notification %v", msg.Fields)
	case <-time.After(100 * time.Millisecond):
	}

	assert.False(t, mutex.IndexWorker.Running())
}
