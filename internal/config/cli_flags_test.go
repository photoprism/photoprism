package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCliFlags_Cli(t *testing.T) {
	cliFlags := Flags.Cli()
	standard := Flags.Find([]string{})

	assert.Greater(t, len(cliFlags), len(standard))
}

func TestCliFlags_Find(t *testing.T) {
	cliFlags := Flags.Cli()
	standard := Flags.Find([]string{})
	sponsor := Flags.Find([]string{EnvSponsor})
	other := Flags.Find([]string{"other"})

	assert.Equal(t, len(standard), len(other))
	assert.Equal(t, len(cliFlags), len(sponsor))
	assert.Less(t, len(other), len(sponsor))
}
