package config

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service/cluster"
)

// systemLogHook records system log messages for the duration of a test.
func systemLogHook(t *testing.T) *logtest.Hook {
	t.Helper()
	logger, hook := logtest.NewNullLogger()
	previous := event.SystemLog
	event.SystemLog = logger
	t.Cleanup(func() { event.SystemLog = previous })
	return hook
}

// setUmask sets the process umask for the duration of a test.
func setUmask(t *testing.T, mask int) {
	t.Helper()
	previous := syscall.Umask(mask)
	t.Cleanup(func() { syscall.Umask(previous) })
}

// fileInode returns the inode number of a file.
func fileInode(t *testing.T, fileName string) uint64 {
	t.Helper()
	info, err := os.Stat(fileName)
	require.NoError(t, err)
	st, ok := info.Sys().(*syscall.Stat_t)
	require.True(t, ok)
	return st.Ino
}

// fileMode returns the permission bits of a file.
func fileMode(t *testing.T, fileName string) os.FileMode {
	t.Helper()
	info, err := os.Stat(fileName)
	require.NoError(t, err)
	return info.Mode().Perm()
}

// newOptionsModeConfig returns a config that writes options.yml to a temporary directory.
func newOptionsModeConfig(t *testing.T) *Config {
	t.Helper()
	c := NewConfig(CliTestContext())
	c.options.ConfigPath = t.TempDir()
	c.options.OptionsYaml = filepath.Join(c.options.ConfigPath, "options.yml")
	return c
}

// TestCredentialOptionKeys checks which options.yml keys count as credentials.
func TestCredentialOptionKeys(t *testing.T) {
	keys := credentialOptionKeys()
	for _, key := range []string{"AdminPassword", "OIDCSecret", "DownloadToken", "PreviewToken", "JoinToken",
		"NodeClientSecret", "DatabaseDSN", "DatabaseDsn", "DatabasePassword", "DatabaseProvisionDSN", "DatabaseProvisionProxyDSN", "VisionKey"} {
		assert.True(t, keys[key], key)
	}
	for _, key := range []string{"DatabaseName", "DatabaseUser", "ClusterUUID", "AuthSecret", "-", ""} {
		assert.False(t, keys[key], key)
	}
}

// TestHasCredentialOption checks whether options values contain a credential that is set.
func TestHasCredentialOption(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, hasCredentialOption(Values{"DatabasePassword": "secret"}))
		assert.True(t, hasCredentialOption(Values{"Existing": "value", "JoinToken": 1234}))
		assert.True(t, hasCredentialOption(Values{"DatabaseProvisionDSN": "root:secret@tcp(mariadb:4001)/"}))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, hasCredentialOption(Values{}))
		assert.False(t, hasCredentialOption(Values{"DatabaseName": "photoprism", "DatabasePassword": ""}))
		assert.False(t, hasCredentialOption(Values{"DatabaseDSN": nil}))
	})
}

// TestOptionsFileMode checks the mode options files are created with.
func TestOptionsFileMode(t *testing.T) {
	assert.Equal(t, os.FileMode(0o640), optionsFileMode(true))
	assert.Equal(t, os.FileMode(0o664), optionsFileMode(false))
}

// TestRestrictOptionsFile checks when access to an options file is restricted.
func TestRestrictOptionsFile(t *testing.T) {
	restrict := func(t *testing.T, fileName string) {
		f, err := os.Open(fileName) //nolint:gosec // test file in a temporary directory
		require.NoError(t, err)
		restrictOptionsFile(f, fileName)
		require.NoError(t, f.Close())
	}
	newFile := func(t *testing.T, mode os.FileMode) string {
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, os.WriteFile(fileName, []byte("DatabasePassword: secret\n"), mode))
		require.NoError(t, os.Chmod(fileName, mode))
		return fileName
	}
	t.Run("Narrowed", func(t *testing.T) {
		hook := systemLogHook(t)
		for mode, want := range map[os.FileMode]os.FileMode{0o664: 0o640, 0o666: 0o640, 0o604: 0o600, 0o660: 0o640} {
			fileName := newFile(t, mode)
			restrict(t, fileName)
			assert.Equal(t, want, fileMode(t, fileName), "%o", mode)
		}
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("StricterUnchanged", func(t *testing.T) {
		for _, mode := range []os.FileMode{0o600, 0o640, 0o400} {
			fileName := newFile(t, mode)
			restrict(t, fileName)
			assert.Equal(t, mode, fileMode(t, fileName))
		}
	})
	t.Run("OtherOwner", func(t *testing.T) {
		hook := systemLogHook(t)
		previous := optionsFileUid
		optionsFileUid = func(os.FileInfo) (int, bool) { return os.Geteuid() + 1, true }
		t.Cleanup(func() { optionsFileUid = previous })
		fileName := newFile(t, 0o664)
		restrict(t, fileName)
		assert.Equal(t, os.FileMode(0o664), fileMode(t, fileName))
		require.Len(t, hook.AllEntries(), 1)
		assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
		assert.Contains(t, hook.LastEntry().Message, "owned by another user")
	})
	t.Run("UnknownOwner", func(t *testing.T) {
		hook := systemLogHook(t)
		previous := optionsFileUid
		optionsFileUid = func(os.FileInfo) (int, bool) { return 0, false }
		t.Cleanup(func() { optionsFileUid = previous })
		fileName := newFile(t, 0o664)
		restrict(t, fileName)
		assert.Equal(t, os.FileMode(0o664), fileMode(t, fileName))
		assert.Empty(t, hook.AllEntries())
	})
	t.Run("ChmodFails", func(t *testing.T) {
		hook := systemLogHook(t)
		previous := chmodOptionsFile
		chmodOptionsFile = func(*os.File, os.FileMode) error { return errors.New("read-only file system") }
		t.Cleanup(func() { chmodOptionsFile = previous })
		fileName := newFile(t, 0o664)
		restrict(t, fileName)
		assert.Equal(t, os.FileMode(0o664), fileMode(t, fileName))
		require.Len(t, hook.AllEntries(), 1)
		assert.Contains(t, hook.LastEntry().Message, "read-only file system")
	})
}

// TestConfig_WriteOptionsYAMLMode checks the mode of options.yml after writes with and without credentials.
func TestConfig_WriteOptionsYAMLMode(t *testing.T) {
	t.Run("NewFileWithCredential", func(t *testing.T) {
		for _, mask := range []int{0o002, 0o022} {
			setUmask(t, mask)
			c := newOptionsModeConfig(t)
			_, err := c.SaveOptionsPatch(Values{"DatabasePassword": "secret"})
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o640), fileMode(t, c.OptionsYaml()), "umask %o", mask)
		}
	})
	t.Run("NewFileCreatedRestricted", func(t *testing.T) {
		// A new file holding a credential is created restricted rather than narrowed afterwards.
		setUmask(t, 0o002)
		systemLogHook(t)
		previous := chmodOptionsFile
		chmodOptionsFile = func(*os.File, os.FileMode) error { return errors.New("not permitted") }
		t.Cleanup(func() { chmodOptionsFile = previous })
		c := newOptionsModeConfig(t)
		_, err := c.SaveOptionsPatch(Values{"DatabasePassword": "secret"})
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o640), fileMode(t, c.OptionsYaml()))
	})
	t.Run("NewFileWithoutCredential", func(t *testing.T) {
		setUmask(t, 0o002)
		c := newOptionsModeConfig(t)
		_, err := c.SaveOptionsPatch(Values{"DatabaseName": "photoprism"})
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o664), fileMode(t, c.OptionsYaml()))
	})
	t.Run("ExistingFileNarrowed", func(t *testing.T) {
		hook := systemLogHook(t)
		c := newOptionsModeConfig(t)
		require.NoError(t, os.WriteFile(c.OptionsYaml(), []byte("Existing: value\n"), 0o664)) //nolint:gosec // mode under test
		require.NoError(t, os.Chmod(c.OptionsYaml(), 0o664))                                  //nolint:gosec // mode under test
		inode := fileInode(t, c.OptionsYaml())

		update := cluster.OptionsUpdate{}
		update.SetDatabaseName("cluster_d0123456789a")
		update.SetDatabasePassword("secret")
		wrote, err := c.SaveClusterOptionsUpdate(update)
		require.NoError(t, err)
		assert.True(t, wrote)

		assert.Equal(t, os.FileMode(0o640), fileMode(t, c.OptionsYaml()))
		assert.Equal(t, inode, fileInode(t, c.OptionsYaml()))
		assert.Empty(t, hook.AllEntries())

		content, err := os.ReadFile(c.OptionsYaml())
		require.NoError(t, err)
		var values map[string]any
		require.NoError(t, yaml.Unmarshal(content, &values))
		assert.Equal(t, "value", values["Existing"])
		assert.Equal(t, "secret", values["DatabasePassword"])
	})
	t.Run("ExistingStricterFile", func(t *testing.T) {
		c := newOptionsModeConfig(t)
		require.NoError(t, os.WriteFile(c.OptionsYaml(), []byte("Existing: value\n"), 0o600))
		require.NoError(t, os.Chmod(c.OptionsYaml(), 0o600))
		_, err := c.SaveOptionsPatch(Values{"JoinToken": "secret"})
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), fileMode(t, c.OptionsYaml()))
	})
	t.Run("ExistingFileWithoutCredential", func(t *testing.T) {
		c := newOptionsModeConfig(t)
		require.NoError(t, os.WriteFile(c.OptionsYaml(), []byte("Existing: value\n"), 0o664)) //nolint:gosec // mode under test
		require.NoError(t, os.Chmod(c.OptionsYaml(), 0o664))                                  //nolint:gosec // mode under test
		_, err := c.SaveOptionsPatch(Values{"DatabaseName": "photoprism"})
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o664), fileMode(t, c.OptionsYaml()))
	})
	t.Run("ReadOnlyFile", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root can write read-only files")
		}
		hook := systemLogHook(t)
		c := newOptionsModeConfig(t)
		require.NoError(t, os.WriteFile(c.OptionsYaml(), []byte("Existing: value\n"), 0o444)) //nolint:gosec // mode under test
		require.NoError(t, os.Chmod(c.OptionsYaml(), 0o444))                                  //nolint:gosec // mode under test
		_, err := c.SaveOptionsPatch(Values{"DatabasePassword": "secret"})
		assert.Error(t, err)
		assert.Equal(t, os.FileMode(0o444), fileMode(t, c.OptionsYaml()))
		assert.Empty(t, hook.AllEntries())
	})
}

// TestWriteOptionsFile checks that options files are written in place with the expected mode.
func TestWriteOptionsFile(t *testing.T) {
	t.Run("RestrictedBeforeWrite", func(t *testing.T) {
		// The mode is already restricted when the content is written.
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, os.WriteFile(fileName, []byte("Existing: value with more text\n"), 0o664)) //nolint:gosec // mode under test
		require.NoError(t, os.Chmod(fileName, 0o664))                                                 //nolint:gosec // mode under test
		var modeAtChmod os.FileMode
		previous := chmodOptionsFile
		chmodOptionsFile = func(f *os.File, mode os.FileMode) error {
			info, err := os.Stat(fileName)
			require.NoError(t, err)
			assert.Equal(t, int64(len("Existing: value with more text\n")), info.Size())
			modeAtChmod = mode
			return previous(f, mode)
		}
		t.Cleanup(func() { chmodOptionsFile = previous })
		require.NoError(t, writeOptionsFile(fileName, []byte("DatabasePassword: x\n"), true))
		assert.Equal(t, os.FileMode(0o640), modeAtChmod)
		assert.Equal(t, os.FileMode(0o640), fileMode(t, fileName))
		content, err := os.ReadFile(fileName) //nolint:gosec // test file in a temporary directory
		require.NoError(t, err)
		assert.Equal(t, "DatabasePassword: x\n", string(content))
	})
	t.Run("MissingDirectory", func(t *testing.T) {
		assert.Error(t, writeOptionsFile(filepath.Join(t.TempDir(), "missing", "options.yml"), []byte("x: 1\n"), false))
	})
}

// TestUrlHasCredential checks which URLs carry a credential.
func TestUrlHasCredential(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		for _, s := range []string{"http://user:secret@proxy.example.com:3128", "https://vision.example.com/api/v1/vision?api_key=x",
			"https://example.com/?Token=x", "https://example.com/?access_token=x", "https://example.com/?apikey=x",
			"https://example.com/?client-secret=x", "https://bucket.example.com/f?X-Amz-Credential=x&X-Amz-Signature=y",
			"https://example.com/?sig=x"} {
			assert.True(t, urlHasCredential(s), s)
		}
	})
	t.Run("False", func(t *testing.T) {
		for _, s := range []string{"", "http://proxy.example.com:3128", "http://user@proxy.example.com", "https://example.com/?page=2",
			"https://example.com/?author=x", "https://example.com/?keyword=x", "https://example.com/?monkey=1",
			"not a url", "user:secret@tcp(db:3306)/photoprism"} {
			assert.False(t, urlHasCredential(s), s)
		}
	})
}

// TestRestrictOptionsFileWithCredential checks that an existing options file with a credential is restricted.
func TestRestrictOptionsFileWithCredential(t *testing.T) {
	write := func(t *testing.T, content string) string {
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, os.WriteFile(fileName, []byte(content), 0o664)) //nolint:gosec // mode under test
		require.NoError(t, os.Chmod(fileName, 0o664))                      //nolint:gosec // mode under test
		return fileName
	}
	t.Run("Credential", func(t *testing.T) {
		fileName := write(t, "DatabasePassword: secret\n")
		restrictOptionsFileWithCredential(fileName)
		assert.Equal(t, os.FileMode(0o640), fileMode(t, fileName))
	})
	t.Run("UrlCredential", func(t *testing.T) {
		fileName := write(t, "HttpsProxy: http://user:secret@proxy.example.com:3128\n")
		restrictOptionsFileWithCredential(fileName)
		assert.Equal(t, os.FileMode(0o640), fileMode(t, fileName))
	})
	t.Run("NoCredential", func(t *testing.T) {
		fileName := write(t, "DatabaseName: photoprism\n")
		restrictOptionsFileWithCredential(fileName)
		assert.Equal(t, os.FileMode(0o664), fileMode(t, fileName))
	})
	t.Run("InvalidOrMissing", func(t *testing.T) {
		fileName := write(t, "DatabasePassword: [\n")
		restrictOptionsFileWithCredential(fileName)
		assert.Equal(t, os.FileMode(0o664), fileMode(t, fileName))
		restrictOptionsFileWithCredential(filepath.Join(t.TempDir(), "missing.yml"))
	})
	t.Run("NewConfig", func(t *testing.T) {
		// Loading the options restricts a file written by an earlier version.
		dir := t.TempDir()
		fileName := filepath.Join(dir, "options.yml")
		require.NoError(t, os.WriteFile(fileName, []byte("JoinToken: secret\n"), 0o664)) //nolint:gosec // mode under test
		require.NoError(t, os.Chmod(fileName, 0o664))                                    //nolint:gosec // mode under test
		ctx := CliTestContext()
		require.NoError(t, ctx.Set("config-path", dir))
		c := NewConfig(ctx)
		assert.Equal(t, fileName, c.OptionsYaml())
		assert.Equal(t, os.FileMode(0o640), fileMode(t, fileName))
	})
}

// TestRestrictOptionsFile_SpecialBits checks that restricting access keeps the setgid bit.
func TestRestrictOptionsFile_SpecialBits(t *testing.T) {
	fileName := filepath.Join(t.TempDir(), "options.yml")
	require.NoError(t, os.WriteFile(fileName, []byte("DatabasePassword: secret\n"), 0o664)) //nolint:gosec // mode under test
	require.NoError(t, os.Chmod(fileName, 0o664|os.ModeSetgid))                             //nolint:gosec // mode under test
	f, err := os.Open(fileName)                                                             //nolint:gosec // test file in a temporary directory
	require.NoError(t, err)
	restrictOptionsFile(f, fileName)
	require.NoError(t, f.Close())
	info, err := os.Stat(fileName)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640)|os.ModeSetgid, info.Mode()&optionsModeBits)
}

// TestRestrictOptionsFile_IgnoredChmod checks the warning when a filesystem accepts a mode change without applying it.
func TestRestrictOptionsFile_IgnoredChmod(t *testing.T) {
	hook := systemLogHook(t)
	previous := chmodOptionsFile
	chmodOptionsFile = func(*os.File, os.FileMode) error { return nil }
	t.Cleanup(func() { chmodOptionsFile = previous })
	fileName := filepath.Join(t.TempDir(), "options.yml")
	require.NoError(t, os.WriteFile(fileName, []byte("DatabasePassword: secret\n"), 0o664)) //nolint:gosec // mode under test
	require.NoError(t, os.Chmod(fileName, 0o664))                                           //nolint:gosec // mode under test
	f, err := os.Open(fileName)                                                             //nolint:gosec // test file in a temporary directory
	require.NoError(t, err)
	restrictOptionsFile(f, fileName)
	require.NoError(t, f.Close())
	require.Len(t, hook.AllEntries(), 1)
	assert.Contains(t, hook.LastEntry().Message, "not supported by the filesystem")
}

// TestOptionsFile_Root checks that a process running as root never restricts access to an options file.
func TestOptionsFile_Root(t *testing.T) {
	hook := systemLogHook(t)
	prevUid := optionsProcessUid
	optionsProcessUid = func() int { return 0 }
	t.Cleanup(func() { optionsProcessUid = prevUid })
	setUmask(t, 0o002)

	t.Run("NewFile", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, writeOptionsFile(fileName, []byte("DatabasePassword: x\n"), true))
		assert.Equal(t, os.FileMode(0o664), fileMode(t, fileName))
	})
	t.Run("ExistingFile", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, os.WriteFile(fileName, []byte("Existing: value\n"), 0o664)) //nolint:gosec // mode under test
		require.NoError(t, os.Chmod(fileName, 0o664))                                  //nolint:gosec // mode under test
		require.NoError(t, writeOptionsFile(fileName, []byte("DatabasePassword: x\n"), true))
		assert.Equal(t, os.FileMode(0o664), fileMode(t, fileName))
		restrictOptionsFileWithCredential(fileName)
		assert.Equal(t, os.FileMode(0o664), fileMode(t, fileName))
	})
	assert.Empty(t, hook.AllEntries())
}

// TestRestrictOptionsFileWithCredential_Symlink checks that loading the options leaves a symbolic link, e.g. a ConfigMap, as it is.
func TestRestrictOptionsFileWithCredential_Symlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.yml")
	require.NoError(t, os.WriteFile(target, []byte("DatabasePassword: secret\n"), 0o664)) //nolint:gosec // mode under test
	require.NoError(t, os.Chmod(target, 0o664))                                           //nolint:gosec // mode under test
	link := filepath.Join(dir, "options.yml")
	require.NoError(t, os.Symlink(target, link))
	restrictOptionsFileWithCredential(link)
	assert.Equal(t, os.FileMode(0o664), fileMode(t, target))
}
