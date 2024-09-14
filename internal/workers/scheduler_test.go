package workers

import (
	"testing"

	"github.com/photoprism/photoprism/internal/config"
)

func TestNewJob(t *testing.T) {
	if err := NewJob("", config.DefaultBackupSchedule, func() {}); err == nil {
		t.Fatal("expected error")
	}

	if err := NewJob("test", "", func() {}); err != nil {
		t.Fatal(err)
	}
}
