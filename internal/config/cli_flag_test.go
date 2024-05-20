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

func TestCliFlag_Default(t *testing.T) {
	hasdefault := CliFlag{
		Flag: cli.StringFlag{
			Name:   "flag-with-default",
			Usage:  "`STRING`",
			EnvVar: "PHOTOPRISM_DEFAULT",
		},
		DocDefault: "default-value",
		Tags:       []string{"foo", "bar"},
	}

	nodefault := CliFlag{
		Flag: cli.StringFlag{
			Name:   "flag-without-default",
			Usage:  "`STRING`",
			EnvVar: "PHOTOPRISM_NODEFAULT",
		},
		Tags: []string{},
	}

	assert.Equal(t, "default-value", hasdefault.Default())
	assert.Equal(t, "", nodefault.Default())
}

func TestCliFlag_EnvVar(t *testing.T) {
	hasdefault := CliFlag{
		Flag: cli.StringFlag{
			Name:   "flag-with-default",
			Usage:  "`STRING`",
			EnvVar: "PHOTOPRISM_DEFAULT",
		},
		DocDefault: "default-value",
		Tags:       []string{"foo", "bar"},
	}

	assert.Equal(t, "PHOTOPRISM_DEFAULT", hasdefault.EnvVar())
}

func TestCliFlag_CommandFlag(t *testing.T) {
	hasdefault := CliFlag{
		Flag: cli.StringFlag{
			Name:   "flag-with-default",
			Usage:  "`STRING`",
			EnvVar: "PHOTOPRISM_DEFAULT",
		},
		DocDefault: "default-value",
		Tags:       []string{"foo", "bar"},
	}

	assert.Equal(t, "--flag-with-default", hasdefault.CommandFlag())
}

func TestCliFlag_Usage(t *testing.T) {
	community := CliFlag{
		Flag: cli.StringFlag{
			Name:  "flag-community",
			Usage: "`STRING`",
		},
		DocDefault: "default-value",
		Tags:       []string{"foo", "bar"},
	}

	essentials := CliFlag{
		Flag: cli.StringFlag{
			Name:  "flag-essentials",
			Usage: "`STRING`",
		},
		Tags: []string{"essentials"},
	}

	plus := CliFlag{
		Flag: cli.StringFlag{
			Name:  "flag-plus",
			Usage: "`STRING`",
		},
		Tags: []string{"plus"},
	}

	pro := CliFlag{
		Flag: cli.StringFlag{
			Name:  "flag-pro",
			Usage: "`STRING`",
		},
		Tags: []string{"pro"},
	}

	assert.Contains(t, community.Usage(), "STRING")
	assert.Contains(t, essentials.Usage(), "*essentials*")
	assert.Contains(t, plus.Usage(), "*plus*")
	assert.Contains(t, pro.Usage(), "*pro*")
}
