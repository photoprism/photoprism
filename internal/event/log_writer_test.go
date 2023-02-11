package event

import (
	"log"
	"testing"
)

func TestLogWriter(t *testing.T) {
	l := log.Default()
	l.Println("Test 123")
}
