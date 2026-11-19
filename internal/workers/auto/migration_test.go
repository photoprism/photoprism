package auto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism/get"
)

// TestFaceMigrationRearms pins that an automatic run held off by a face migration announces
// nothing and sets its trigger again, so the files are processed once the migration has completed.
func TestFaceMigrationRearms(t *testing.T) {
	conf := config.TestConfig()
	get.SetConfig(conf)

	lock, err := mutex.AcquireFileLock(conf.FacesLockFile(), "faces migration")
	require.NoError(t, err)
	t.Cleanup(lock.Release)
	t.Cleanup(ResetIndex)
	t.Cleanup(ResetImport)

	t.Run("Index", func(t *testing.T) {
		ResetIndex()
		success := event.Subscribe("notify.success")
		defer event.Unsubscribe(success)

		require.NoError(t, Index())

		indexMutex.Lock()
		armed := !autoIndex.IsZero()
		indexMutex.Unlock()
		assert.True(t, armed, "the trigger is set again")

		select {
		case msg := <-success.Receiver:
			t.Fatalf("unexpected notification %v", msg.Fields)
		case <-time.After(100 * time.Millisecond):
		}
	})
	t.Run("Import", func(t *testing.T) {
		ResetImport()
		info := event.Subscribe("notify.*")
		defer event.Unsubscribe(info)

		require.NoError(t, Import())

		importMutex.Lock()
		armed := !autoImport.IsZero()
		importMutex.Unlock()
		assert.True(t, armed, "the trigger is set again")

		select {
		case msg := <-info.Receiver:
			t.Fatalf("unexpected notification %s %v", msg.Name, msg.Fields)
		case <-time.After(100 * time.Millisecond):
		}
	})
}
