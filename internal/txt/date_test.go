package txt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMonths(t *testing.T) {
	assert.Equal(t, "Unknown", Months[0])
	assert.Equal(t, "January", Months[1])
	assert.Equal(t, "December", Months[12])
}
