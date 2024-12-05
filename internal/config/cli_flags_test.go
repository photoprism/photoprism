package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
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

func TestCliFlags_Replace(t *testing.T) {
	originalPublicFlag := CliFlag{Flag: &cli.BoolFlag{
		Name:    "public",
		Aliases: []string{"p"},
		Hidden:  true,
		Usage:   "disable authentication, advanced settings, and WebDAV remote access",
		EnvVars: EnvVars("PUBLIC"),
	}}

	newPublicFlag := CliFlag{Flag: &cli.BoolFlag{
		Name:    "public",
		Hidden:  false,
		Usage:   "disable authentication, advanced settings, and WebDAV remote access",
		EnvVars: EnvVars("PUBLIC"),
	}}

	cliFlags := CliFlags{
		{
			Flag: &cli.StringFlag{
				Name:    "auth-mode",
				Aliases: []string{"a"},
				Usage:   "authentication `MODE` (public, password)",
				Value:   "password",
				EnvVars: EnvVars("AUTH_MODE"),
			}},
		originalPublicFlag,
		{
			Flag: &cli.StringFlag{
				Name:    "admin-user",
				Aliases: []string{"login"},
				Usage:   "`USERNAME` of the superadmin account that is created on first startup",
				Value:   "admin",
				EnvVars: EnvVars("ADMIN_USER"),
			}}}

	assert.Equal(t, 3, len(cliFlags))
	assert.Equal(t, originalPublicFlag.String(), cliFlags[1].String())
	assert.Equal(t, originalPublicFlag.Hidden(), cliFlags[1].Hidden())

	t.Run("WrongName", func(t *testing.T) {

		r := cliFlags.Replace("xxx", newPublicFlag)

		assert.Equal(t, 3, len(r))
		assert.Equal(t, "auth-mode, a", r[0].String())
		assert.Equal(t, originalPublicFlag.String(), r[1].String())
		assert.Equal(t, "admin-user, login", r[2].String())
	})
	t.Run("Success", func(t *testing.T) {

		r := cliFlags.Replace("public, p", newPublicFlag)

		assert.Equal(t, 3, len(r))
		assert.Equal(t, newPublicFlag.String(), r[1].String())
		assert.Equal(t, newPublicFlag.Hidden(), r[1].Hidden())
	})
}

func TestCliFlags_Remove(t *testing.T) {
	cliFlags := CliFlags{
		{
			Flag: &cli.StringFlag{
				Name:    "auth-mode",
				Aliases: []string{"a"},
				Usage:   "authentication `MODE` (public, password)",
				Value:   "password",
				EnvVars: EnvVars("AUTH_MODE"),
			}},
		{
			Flag: &cli.StringFlag{
				Name:    "admin-user",
				Aliases: []string{"login"},
				Usage:   "`USERNAME` of the superadmin account that is created on first startup",
				Value:   "admin",
				EnvVars: EnvVars("ADMIN_USER"),
			}}}

	assert.Equal(t, 2, len(cliFlags))

	result := cliFlags.Remove([]string{"auth-mode, a"})

	assert.Equal(t, 1, len(result))
}

func TestCliFlags_Insert(t *testing.T) {
	PublicFlag := CliFlag{Flag: &cli.BoolFlag{
		Name:    "public",
		Aliases: []string{"p"},
		Hidden:  true,
		Usage:   "disable authentication, advanced settings, and WebDAV remote access",
		EnvVars: EnvVars("PUBLIC"),
	}}

	cliFlags := CliFlags{
		{
			Flag: &cli.StringFlag{
				Name:    "auth-mode",
				Aliases: []string{"a"},
				Usage:   "authentication `MODE` (public, password)",
				Value:   "password",
				EnvVars: EnvVars("AUTH_MODE"),
			}},
		{
			Flag: &cli.StringFlag{
				Name:    "admin-user",
				Aliases: []string{"login"},
				Usage:   "`USERNAME` of the superadmin account that is created on first startup",
				Value:   "admin",
				EnvVars: EnvVars("ADMIN_USER"),
			}}}

	assert.Equal(t, 2, len(cliFlags))

	t.Run("Success", func(t *testing.T) {
		result := cliFlags.Insert("auth-mode, a", []CliFlag{PublicFlag})

		assert.Equal(t, 3, len(result))

		assert.Equal(t, "auth-mode, a", result[0].String())
		assert.Equal(t, PublicFlag.String(), result[1].String())
		assert.Equal(t, "admin-user, login", result[2].String())
	})
	t.Run("WrongName", func(t *testing.T) {
		result := cliFlags.Insert("xxx", []CliFlag{PublicFlag})

		assert.Equal(t, 3, len(result))

		assert.Equal(t, "auth-mode, a", result[0].String())
		assert.Equal(t, "admin-user, login", result[1].String())
		assert.Equal(t, PublicFlag.String(), result[2].String())
	})
}

func TestCliFlags_InsertBefore(t *testing.T) {
	PublicFlag := CliFlag{Flag: &cli.BoolFlag{
		Name:    "public",
		Aliases: []string{"p"},
		Hidden:  true,
		Usage:   "disable authentication, advanced settings, and WebDAV remote access",
		EnvVars: EnvVars("PUBLIC"),
	}}

	cliFlags := CliFlags{
		{
			Flag: &cli.StringFlag{
				Name:    "auth-mode",
				Aliases: []string{"a"},
				Usage:   "authentication `MODE` (public, password)",
				Value:   "password",
				EnvVars: EnvVars("AUTH_MODE"),
			}},
		{
			Flag: &cli.StringFlag{
				Name:    "admin-user",
				Aliases: []string{"login"},
				Usage:   "`USERNAME` of the superadmin account that is created on first startup",
				Value:   "admin",
				EnvVars: EnvVars("ADMIN_USER"),
			}}}

	assert.Equal(t, 2, len(cliFlags))

	t.Run("Success", func(t *testing.T) {
		result := cliFlags.InsertBefore("auth-mode, a", []CliFlag{PublicFlag})

		assert.Equal(t, 3, len(result))

		assert.Equal(t, "auth-mode, a", result[1].String())
		assert.Equal(t, PublicFlag.String(), result[0].String())
		assert.Equal(t, "admin-user, login", result[2].String())
	})
	t.Run("WrongName", func(t *testing.T) {
		result := cliFlags.InsertBefore("xxx", []CliFlag{PublicFlag})

		assert.Equal(t, 3, len(result))

		t.Logf("flags: %#v", result)

		assert.Equal(t, "auth-mode, a", result[0].String())
		assert.Equal(t, "admin-user, login", result[1].String())
		assert.Equal(t, PublicFlag.String(), result[2].String())
	})
}

func TestCliFlags_Prepend(t *testing.T) {
	PublicFlag := CliFlag{Flag: &cli.BoolFlag{
		Name:    "public",
		Aliases: []string{"p"},
		Hidden:  true,
		Usage:   "disable authentication, advanced settings, and WebDAV remote access",
		EnvVars: EnvVars("PUBLIC"),
	}}

	cliFlags := CliFlags{
		{
			Flag: &cli.StringFlag{
				Name:    "auth-mode",
				Aliases: []string{"a"},
				Usage:   "authentication `MODE` (public, password)",
				Value:   "password",
				EnvVars: EnvVars("AUTH_MODE"),
			}},
		{
			Flag: &cli.StringFlag{
				Name:    "admin-user",
				Aliases: []string{"login"},
				Usage:   "`USERNAME` of the superadmin account that is created on first startup",
				Value:   "admin",
				EnvVars: EnvVars("ADMIN_USER"),
			}}}

	assert.Equal(t, 2, len(cliFlags))

	r := cliFlags.Prepend([]CliFlag{PublicFlag})

	assert.Equal(t, "auth-mode, a", r[1].String())
	assert.Equal(t, PublicFlag.String(), r[0].String())
	assert.Equal(t, "admin-user, login", r[2].String())
}
