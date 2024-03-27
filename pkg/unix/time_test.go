package unix

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTime(t *testing.T) {
	result := Time()

	assert.Greater(t, result, int64(1706521797))
	assert.GreaterOrEqual(t, result, time.Now().UTC().Unix())
	assert.LessOrEqual(t, result, time.Now().UTC().Unix()+2)
}
