package util

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProfileTime(t *testing.T) {
	start := time.Now()
	time.Sleep(1 * time.Millisecond)
	ProfileTime(start, fmt.Sprintf("%s", "Successful test"))
	assert.Contains(t, logBuffer.String(), "Successful test took")
}
