package session

import (
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	Monitor(time.Minute)
	Shutdown()
}
