package session

import (
	"testing"
	"time"
)

func TestCleanup(t *testing.T) {
	Cleanup(time.Minute)
	Shutdown()
}
