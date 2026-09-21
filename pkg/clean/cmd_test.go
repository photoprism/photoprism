package clean

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCmd(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", Cmd(nil, "Sup3r$ecret"))
	})
	t.Run("NoArgs", func(t *testing.T) {
		assert.Equal(t, "/usr/bin/sqlite3", Cmd(exec.Command("/usr/bin/sqlite3"), "Sup3r$ecret"))
	})
	t.Run("NoSecrets", func(t *testing.T) {
		cmd := exec.Command("/usr/bin/mariadb-dump", "-u", "photoprism", "photoprism")
		assert.Equal(t, "/usr/bin/mariadb-dump -u photoprism photoprism", Cmd(cmd))
	})
	t.Run("AttachedValue", func(t *testing.T) {
		cmd := exec.Command("/usr/bin/mariadb-dump", "-u", "photoprism", "-pSup3r$ecret", "photoprism")
		s := Cmd(cmd, "Sup3r$ecret")
		assert.Equal(t, "/usr/bin/mariadb-dump -u photoprism -p*** photoprism", s)
		assert.NotContains(t, s, "Sup3r$ecret")
	})
	t.Run("LongOption", func(t *testing.T) {
		cmd := exec.Command("/usr/bin/mariadb", "--password=Sup3r$ecret", "photoprism")
		assert.Equal(t, "/usr/bin/mariadb --password=*** photoprism", Cmd(cmd, "Sup3r$ecret"))
	})
	t.Run("SeparateValue", func(t *testing.T) {
		cmd := exec.Command("/usr/bin/mariadb", "--password", "Sup3r$ecret")
		assert.Equal(t, "/usr/bin/mariadb --password ***", Cmd(cmd, "Sup3r$ecret"))
	})
	t.Run("ValueWithSpaces", func(t *testing.T) {
		cmd := exec.Command("/usr/bin/mariadb-dump", "-ptwo words", "photoprism")
		assert.Equal(t, "/usr/bin/mariadb-dump -p*** photoprism", Cmd(cmd, "two words"))
	})
	t.Run("MultipleSecrets", func(t *testing.T) {
		cmd := exec.Command("/usr/bin/mariadb", "-pSup3r$ecret", "--ssl-key=k3y")
		assert.Equal(t, "/usr/bin/mariadb -p*** --ssl-key=***", Cmd(cmd, "Sup3r$ecret", "k3y"))
	})
	t.Run("EmptySecretIgnored", func(t *testing.T) {
		cmd := exec.Command("/usr/bin/mariadb-dump", "-u", "photoprism", "photoprism")
		assert.Equal(t, "/usr/bin/mariadb-dump -u photoprism photoprism", Cmd(cmd, ""))
	})
	t.Run("ValueMatchingAnotherArgument", func(t *testing.T) {
		// A value identical to another argument masks both, as nothing tells them apart.
		cmd := exec.Command("/usr/bin/mariadb-dump", "-u", "photoprism", "-pphotoprism", "photoprism")
		assert.Equal(t, "/usr/bin/mariadb-dump -u *** -p*** ***", Cmd(cmd, "photoprism"))
	})
	t.Run("FlagNamesKeptReadable", func(t *testing.T) {
		cmd := exec.Command("/usr/bin/mariadb-dump", "--protocol", "tcp", "--skip-ssl", "-u", "root")
		assert.Equal(t, "/usr/bin/mariadb-dump --protocol tcp --skip-ssl -u root", Cmd(cmd, "l", "ssl", "u"))
	})
	t.Run("UriArgument", func(t *testing.T) {
		// An argument naming no secret still has the credentials of a URI it holds removed.
		cmd := exec.Command("/usr/bin/ffmpeg", "-metadata", "comment=https://alice:hunter2@videos.example.com/v/1?token=T")
		s := Cmd(cmd)
		assert.Equal(t, "/usr/bin/ffmpeg -metadata comment=https://alice:***@videos.example.com/v/1?token=***", s)
	})
	t.Run("PathContainingSecret", func(t *testing.T) {
		// Only a whole argument or a flag value is recognized, so a path keeps its name.
		cmd := exec.Command("/usr/bin/Sup3r$ecret-dump", "-pSup3r$ecret")
		assert.Equal(t, "/usr/bin/Sup3r$ecret-dump -p***", Cmd(cmd, "Sup3r$ecret"))
	})
}

func TestMaskedArg(t *testing.T) {
	secrets := []string{"s3cret"}

	for _, tc := range []struct{ name, arg, want string }{
		{"WholeArgument", "s3cret", "***"},
		{"ShortFlagValue", "-ps3cret", "-p***"},
		{"LongFlagValue", "--password=s3cret", "--password=***"},
		{"RepeatedInValue", "--password=s3crets3cret", "--password=***"},
		{"WithinFlagValue", "--default-character-set=s3cretUTF8", "--default-character-set=***"},
		{"TrailingFlagName", "--secret-s3cret", "--secret-s3cret"},
		{"WithinValue", "s3cretdb.example.com", "s3cretdb.example.com"},
		{"ShortFlagOnly", "-p", "-p"},
		{"NotAFlag", "photoprism", "photoprism"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, maskedArg(tc.arg, secrets))
		})
	}
	t.Run("EmptySecret", func(t *testing.T) {
		assert.Equal(t, "-p", maskedArg("-p", []string{""}))
		assert.Equal(t, "--protocol", maskedArg("--protocol", []string{""}))
	})
	t.Run("NoSecrets", func(t *testing.T) {
		assert.Equal(t, "-ps3cret", maskedArg("-ps3cret", nil))
	})
}

func TestSecrets(t *testing.T) {
	t.Run("EveryOccurrence", func(t *testing.T) {
		s := Secrets("Access denied for user 's3cret'@'host' using s3cret", "s3cret")
		assert.Equal(t, "Access denied for user '***'@'host' using ***", s)
	})
	t.Run("MultipleSecrets", func(t *testing.T) {
		assert.Equal(t, "*** and ***", Secrets("s3cret and k3yfile", "s3cret", "k3yfile"))
	})
	t.Run("EmptySecretIgnored", func(t *testing.T) {
		assert.Equal(t, "unchanged", Secrets("unchanged", "", ""))
	})
	t.Run("ShortSecretIgnored", func(t *testing.T) {
		// A one-letter value would replace letters throughout and leave the text unreadable.
		denied := "Access denied for user 'photoprism'@'db' (using password: YES)"
		assert.Equal(t, denied, Secrets(denied, "e"))
		assert.Equal(t, denied, Secrets(denied, "den"))
	})
	t.Run("ShortestMaskedSecret", func(t *testing.T) {
		secret := "pass"
		assert.Len(t, secret, SecretMinLength)
		assert.Equal(t, "user ***", Secrets("user "+secret, secret))
	})
	t.Run("NoSecrets", func(t *testing.T) {
		assert.Equal(t, "unchanged", Secrets("unchanged"))
	})
	t.Run("EmptyText", func(t *testing.T) {
		assert.Equal(t, "", Secrets("", "s3cret"))
	})
}

func TestFlagValueIndex(t *testing.T) {
	for _, tc := range []struct {
		name, arg string
		want      int
	}{
		{"ShortFlag", "-ps3cret", 2},
		{"LongFlagWithEquals", "--password=s3cret", 11},
		{"LongFlagWithoutValue", "--password", 0},
		{"ShortFlagWithoutValue", "-p", 0},
		{"Positional", "photoprism", 0},
		{"Empty", "", 0},
		{"DashOnly", "-", 0},
		{"EqualsFirst", "-=x", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, flagValueIndex(tc.arg))
		})
	}
}
