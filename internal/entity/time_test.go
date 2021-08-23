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

func TestTimestampPointer(t *testing.T) {
	result := TimestampPointer()

	if result == nil {
		t.Fatal("result must not be nil")
	}

	if result.Location() != time.UTC {
		t.Fatal("timestamp zone must be utc")
	}

	if result.After(time.Now().Add(time.Second)) {
		t.Fatal("timestamp should be in the past from now")
	}
}

func TestSeconds(t *testing.T) {
	result := Seconds(23)

	if result != 23*time.Second {
		t.Error("must be 23 seconds")
	}
}

func TestYesterday(t *testing.T) {
	now := time.Now()
	result := Yesterday()

	t.Logf("yesterday: %s", result)

	if result.After(now) {
		t.Error("yesterday is not before now")
	}

	if !result.Before(now) {
		t.Error("yesterday is before now")
	}
}
