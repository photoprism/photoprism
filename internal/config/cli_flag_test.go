package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestCliFlag_Skip(t *testing.T) {
	withTags := CliFlag{
		Flag: cli.StringFlag{
			Name:   "with-tags",
			Usage:  "`STRING`",
			EnvVar: "PHOTOPRISM_WITH_TAGS",
		},
		Tags: []string{"foo", "bar"},
	}

	noTags := CliFlag{
		Flag: cli.StringFlag{
			Name:   "no-tags",
			Usage:  "`STRING`",
			EnvVar: "PHOTOPRISM_NO_TAGS",
		},
		Tags: []string{},
	}

	t.Run("True", func(t *testing.T) {
		assert.True(t, withTags.Skip([]string{"baz"}))
		assert.False(t, noTags.Skip([]string{"baz"}))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, withTags.Skip([]string{"foo"}))
		assert.False(t, noTags.Skip([]string{"foo"}))
	})
}

func TestCliFlag_Hidden(t *testing.T) {
	hidden := CliFlag{
		Flag: cli.StringFlag{
			Name:   "is-hidden",
			Usage:  "`STRING`",
			EnvVar: "PHOTOPRISM_HIDDEN",
			Hidden: true,
		},
		Tags: []string{"foo", "bar"},
	}

	visible := CliFlag{
		Flag: cli.StringFlag{
			Name:   "is-visible",
			Usage:  "`STRING`",
			EnvVar: "PHOTOPRISM_VISIBLE",
			Hidden: false,
		},
		Tags: []string{},
	}

	t.Run("True", func(t *testing.T) {
		assert.True(t, hidden.Hidden())
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, visible.Hidden())
	})
}
