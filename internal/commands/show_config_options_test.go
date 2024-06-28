package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShowConfigOptionsCommand(t *testing.T) {
	var err error

	ctx := NewTestContext([]string{})

	err = ShowConfigOptionsCommand.Run(ctx)

	assert.NoError(t, err)
}
