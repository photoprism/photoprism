package unix

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNow(t *testing.T) {
	result := Now()

	assert.Greater(t, result, int64(1706521797))
	assert.GreaterOrEqual(t, result, time.Now().UTC().Unix())
	assert.LessOrEqual(t, result, time.Now().UTC().Unix()+2)
}
