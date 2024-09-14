package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	assert.Equal(t, "yes", Bool(true, "yes", "no"))
	assert.Equal(t, "no", Bool(false, "yes", "no"))

}
