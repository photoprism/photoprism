package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestCliContext(t *testing.T) {
	result := CliContext()

	assert.IsType(t, new(cli.Context), result)
}
