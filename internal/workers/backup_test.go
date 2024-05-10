package workers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/mutex"
)

func TestBackup_Start(t *testing.T) {
	conf := config.TestConfig()

	t.Logf("database-dsn: %s", conf.DatabaseDsn())

	worker := NewBackup(conf)

	assert.IsType(t, &Backup{}, worker)

	if err := mutex.BackupWorker.Start(); err != nil {
		t.Fatal(err)
	}

	// Mutex should prevent worker from starting.
	if err := worker.Start(true, true, true); err == nil {
		t.Fatal("error expected")
	}

	mutex.BackupWorker.Stop()

	// Start worker.
	if err := worker.Start(true, true, false); err != nil {
		t.Fatal(err)
	}

	// Rerun worker.
	if err := worker.Start(true, true, false); err != nil {
		t.Fatal(err)
	}
}
