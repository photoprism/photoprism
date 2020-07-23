package mutex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkersBusy(t *testing.T) {
	assert.False(t, WorkersBusy())
}
