package mutex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkersRunning(t *testing.T) {
	assert.False(t, WorkersRunning())
}
