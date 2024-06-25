package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
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
	originalPublicFlag := CliFlag{Flag: cli.BoolFlag{
		Name:   "public, p",
		Hidden: true,
		Usage:  "disable authentication, advanced settings, and WebDAV remote access",
		EnvVar: EnvVar("PUBLIC"),
	}}

	newPublicFlag := CliFlag{Flag: cli.BoolFlag{
		Name:   "public",
		Hidden: false,
		Usage:  "disable authentication, advanced settings, and WebDAV remote access",
		EnvVar: EnvVar("PUBLIC"),
	}}

	cliFlags := CliFlags{
		{
			Flag: cli.StringFlag{
				Name:   "auth-mode, a",
				Usage:  "authentication `MODE` (public, password)",
				Value:  "password",
				EnvVar: EnvVar("AUTH_MODE"),
			}},
		originalPublicFlag,
		{
			Flag: cli.StringFlag{
				Name:   "admin-user, login",
				Usage:  "`USERNAME` of the superadmin account that is created on first startup",
				Value:  "admin",
				EnvVar: EnvVar("ADMIN_USER"),
			}}}

	assert.Equal(t, 3, len(cliFlags))
	assert.Equal(t, originalPublicFlag.Name(), cliFlags[1].Name())
	assert.Equal(t, originalPublicFlag.Hidden(), cliFlags[1].Hidden())

	t.Run("WrongName", func(t *testing.T) {

		r := cliFlags.Replace("xxx", newPublicFlag)

		assert.Equal(t, 3, len(r))
		assert.Equal(t, "auth-mode, a", r[0].Name())
		assert.Equal(t, originalPublicFlag.Name(), r[1].Name())
		assert.Equal(t, "admin-user, login", r[2].Name())
	})
	t.Run("Success", func(t *testing.T) {

		r := cliFlags.Replace("public, p", newPublicFlag)

		assert.Equal(t, 3, len(r))
		assert.Equal(t, newPublicFlag.Name(), r[1].Name())
		assert.Equal(t, newPublicFlag.Hidden(), r[1].Hidden())
	})
}

func TestCliFlags_Remove(t *testing.T) {
	cliFlags := CliFlags{
		{
			Flag: cli.StringFlag{
				Name:   "auth-mode, a",
				Usage:  "authentication `MODE` (public, password)",
				Value:  "password",
				EnvVar: EnvVar("AUTH_MODE"),
			}},
		{
			Flag: cli.StringFlag{
				Name:   "admin-user, login",
				Usage:  "`USERNAME` of the superadmin account that is created on first startup",
				Value:  "admin",
				EnvVar: EnvVar("ADMIN_USER"),
			}}}

	assert.Equal(t, 2, len(cliFlags))

	r := cliFlags.Remove([]string{"auth-mode, a"})

	assert.Equal(t, 1, len(r))
}

func TestCliFlags_Insert(t *testing.T) {
	PublicFlag := CliFlag{Flag: cli.BoolFlag{
		Name:   "public, p",
		Hidden: true,
		Usage:  "disable authentication, advanced settings, and WebDAV remote access",
		EnvVar: EnvVar("PUBLIC"),
	}}

	cliFlags := CliFlags{
		{
			Flag: cli.StringFlag{
				Name:   "auth-mode, a",
				Usage:  "authentication `MODE` (public, password)",
				Value:  "password",
				EnvVar: EnvVar("AUTH_MODE"),
			}},
		{
			Flag: cli.StringFlag{
				Name:   "admin-user, login",
				Usage:  "`USERNAME` of the superadmin account that is created on first startup",
				Value:  "admin",
				EnvVar: EnvVar("ADMIN_USER"),
			}}}

	assert.Equal(t, 2, len(cliFlags))

	t.Run("Success", func(t *testing.T) {
		r := cliFlags.Insert("auth-mode, a", []CliFlag{PublicFlag})

		assert.Equal(t, 3, len(r))

		assert.Equal(t, "auth-mode, a", r[0].Name())
		assert.Equal(t, PublicFlag.Name(), r[1].Name())
		assert.Equal(t, "admin-user, login", r[2].Name())
	})
	t.Run("WrongName", func(t *testing.T) {
		r := cliFlags.Insert("xxx", []CliFlag{PublicFlag})

		assert.Equal(t, 3, len(r))

		assert.Equal(t, "auth-mode, a", r[0].Name())
		assert.Equal(t, "admin-user, login", r[1].Name())
		assert.Equal(t, PublicFlag.Name(), r[2].Name())
	})
}

func TestCliFlags_InsertBefore(t *testing.T) {
	PublicFlag := CliFlag{Flag: cli.BoolFlag{
		Name:   "public, p",
		Hidden: true,
		Usage:  "disable authentication, advanced settings, and WebDAV remote access",
		EnvVar: EnvVar("PUBLIC"),
	}}

	cliFlags := CliFlags{
		{
			Flag: cli.StringFlag{
				Name:   "auth-mode, a",
				Usage:  "authentication `MODE` (public, password)",
				Value:  "password",
				EnvVar: EnvVar("AUTH_MODE"),
			}},
		{
			Flag: cli.StringFlag{
				Name:   "admin-user, login",
				Usage:  "`USERNAME` of the superadmin account that is created on first startup",
				Value:  "admin",
				EnvVar: EnvVar("ADMIN_USER"),
			}}}

	assert.Equal(t, 2, len(cliFlags))

	t.Run("Success", func(t *testing.T) {
		r := cliFlags.InsertBefore("auth-mode, a", []CliFlag{PublicFlag})

		assert.Equal(t, 3, len(r))

		assert.Equal(t, "auth-mode, a", r[1].Name())
		assert.Equal(t, PublicFlag.Name(), r[0].Name())
		assert.Equal(t, "admin-user, login", r[2].Name())
	})
	t.Run("WrongName", func(t *testing.T) {
		r := cliFlags.InsertBefore("xxx", []CliFlag{PublicFlag})

		assert.Equal(t, 3, len(r))

		assert.Equal(t, "auth-mode, a", r[0].Name())
		assert.Equal(t, "admin-user, login", r[1].Name())
		assert.Equal(t, PublicFlag.Name(), r[2].Name())
	})
}

func TestCliFlags_Prepend(t *testing.T) {
	PublicFlag := CliFlag{Flag: cli.BoolFlag{
		Name:   "public, p",
		Hidden: true,
		Usage:  "disable authentication, advanced settings, and WebDAV remote access",
		EnvVar: EnvVar("PUBLIC"),
	}}

	cliFlags := CliFlags{
		{
			Flag: cli.StringFlag{
				Name:   "auth-mode, a",
				Usage:  "authentication `MODE` (public, password)",
				Value:  "password",
				EnvVar: EnvVar("AUTH_MODE"),
			}},
		{
			Flag: cli.StringFlag{
				Name:   "admin-user, login",
				Usage:  "`USERNAME` of the superadmin account that is created on first startup",
				Value:  "admin",
				EnvVar: EnvVar("ADMIN_USER"),
			}}}

	assert.Equal(t, 2, len(cliFlags))

	r := cliFlags.Prepend([]CliFlag{PublicFlag})

	assert.Equal(t, "auth-mode, a", r[1].Name())
	assert.Equal(t, PublicFlag.Name(), r[0].Name())
	assert.Equal(t, "admin-user, login", r[2].Name())
}
