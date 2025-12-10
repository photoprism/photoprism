package capture

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	start := time.Now()
	time.Sleep(1 * time.Millisecond)
	result := Time(start, "Successful test")
	assert.Contains(t, result, "Successful test [")
}
