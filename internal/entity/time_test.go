package entity

import (
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {
	result := Timestamp()

	if result.Location() != time.UTC {
		t.Fatal("timestamp zone must be utc")
	}

	if result.After(time.Now().Add(time.Second)) {
		t.Fatal("timestamp should be in the past from now")
	}
}
