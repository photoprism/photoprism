package capture

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	start := time.Now()
	time.Sleep(1 * time.Millisecond)
	Time(start, fmt.Sprintf("%s", "Successful test"))
	assert.Contains(t, logBuffer.String(), "Successful test [")
}
