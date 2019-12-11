package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProfileTime(t *testing.T) {
	start := time.Now()
	time.Sleep(1 * time.Millisecond)
	ProfileTime(start, fmt.Sprintf("%s", "Successful test"))
	assert.Contains(t, logBuffer.String(), "Successful test [")
}
