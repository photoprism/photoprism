package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCliFlags_Cli(t *testing.T) {
	cliFlags := Flags.Cli()
	standard := Flags.Find([]string{})

	assert.Equal(t, len(cliFlags), len(standard))
}

func TestCliFlags_Find(t *testing.T) {
	cliFlags := Flags.Cli()
	standard := Flags.Find([]string{})
	essentials := Flags.Find([]string{Essentials})
	other := Flags.Find([]string{"other"})

	assert.Equal(t, len(standard), len(other))
	assert.Equal(t, len(cliFlags), len(essentials))
	assert.Equal(t, len(other), len(essentials))
}
