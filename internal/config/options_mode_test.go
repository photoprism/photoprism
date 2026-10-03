package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

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

// asOptionsUser makes the options file helpers see a non-root process user that owns the files the test
// creates if the tests run as root, so tests expecting a restricted file also pass as root. The stand-in
// user ID is negative, so no real file owner can match it.
func asOptionsUser(t *testing.T) {
	t.Helper()

	if os.Geteuid() != 0 {
		return
	}

	const uid = -1
	prevProcessUid, prevFileUid := optionsProcessUid, optionsFileUid
	optionsProcessUid = func() int { return uid }
	optionsFileUid = func(info os.FileInfo) (int, bool) {
		if owner, known := prevFileUid(info); !known || owner != 0 {
			return owner, known
		}

		return uid, true
	}
	t.Cleanup(func() { optionsProcessUid, optionsFileUid = prevProcessUid, prevFileUid })
}

// withinTimeout fails the test if fn does not return within a few seconds, e.g. because it opened a named pipe.
func withinTimeout(t *testing.T, fn func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		fn()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("did not return")
	}
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
		assert.True(t, hasCredentialOption(Values{"Existing": []any{"value", map[any]any{"url": "https://token@example.com/"}}}))
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
	asOptionsUser(t)
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
		optionsFileUid = func(os.FileInfo) (int, bool) { return optionsProcessUid() + 1, true }
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
	asOptionsUser(t)
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
	t.Run("NamedPipe", func(t *testing.T) {
		systemLogHook(t)
		c := newOptionsModeConfig(t)
		require.NoError(t, syscall.Mkfifo(c.OptionsYaml(), 0o664))
		withinTimeout(t, func() {
			_, err := c.SaveOptionsPatch(Values{"DatabaseName": "photoprism"})
			assert.ErrorIs(t, err, errOptionsFileType)
		})
		withinTimeout(t, func() { assert.Empty(t, c.SupersededFaceModel()) })
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
	asOptionsUser(t)
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
	t.Run("TooLarge", func(t *testing.T) {
		// Options the reader would refuse are not written, and an existing file keeps its content.
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, os.WriteFile(fileName, []byte("Existing: value\n"), 0o600))
		data := []byte(strings.Repeat("x", optionsFileMaxBytes+1))
		assert.ErrorIs(t, writeOptionsFile(fileName, data, false), ErrOptionsTooLarge)
		content, err := os.ReadFile(fileName) //nolint:gosec // test file in a temporary directory
		require.NoError(t, err)
		assert.Equal(t, "Existing: value\n", string(content))
		missing := filepath.Join(t.TempDir(), "options.yml")
		assert.ErrorIs(t, writeOptionsFile(missing, data, false), ErrOptionsTooLarge)
		assert.NoFileExists(t, missing)
	})
	t.Run("MaxSize", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "options.yml")
		data := []byte("x: " + strings.Repeat("y", optionsFileMaxBytes-4) + "\n")
		require.Len(t, data, optionsFileMaxBytes)
		require.NoError(t, writeOptionsFile(fileName, data, false))
		read, err := readOptionsFile(fileName)
		require.NoError(t, err)
		assert.Equal(t, data, read)
	})
	t.Run("NamedPipe", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, syscall.Mkfifo(fileName, 0o664))
		r, err := os.OpenFile(fileName, os.O_RDONLY|syscall.O_NONBLOCK, 0) //nolint:gosec // test file in a temporary directory
		require.NoError(t, err)
		t.Cleanup(func() { _ = r.Close() })
		assert.ErrorIs(t, writeOptionsFile(fileName, []byte("x: 1\n"), false), errOptionsFileType)
	})
}

// TestUrlHasCredential checks which URLs carry a credential.
func TestUrlHasCredential(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		for _, s := range []string{"http://user:secret@proxy.example.com:3128", "https://vision.example.com/api/v1/vision?api_key=x",
			"https://example.com/?Token=x", "https://example.com/?access_token=x", "https://example.com/?apikey=x",
			"https://example.com/?client-secret=x", "https://bucket.example.com/f?X-Amz-Credential=x&X-Amz-Signature=y",
			"https://example.com/?sig=x", "https://token@example.com/", "http://user@proxy.example.com",
			"user:secret@proxy.example.com:3128", "user:secret@tcp(db:3306)/photoprism", "https://example.com/?accessToken=x",
			"https://example.com/?authToken=x", "https://example.com/?accessKey=x", "https://example.com/?clientSecret=x",
			"https://example.com/?pwd=x", "https://example.com/cb#access_token=x", "https://example.com/?monkey=1",
			"http://user:pa#ss@proxy.example.com:3128", "http://user:p%zz@proxy.example.com:3128", "https://example.com/?x=1;api_key=y",
			"https://example.com/?api_key=abc%", "https://example.com/?authtoken=x", "https://example.com/?pass=x",
			"https://example.com/#/cb?access_token=x", "https://example.com/#!access_token=x", "sip-proxy:secret@proxy.example.com"} {
			assert.True(t, urlHasCredential(s), s)
		}
	})
	t.Run("False", func(t *testing.T) {
		for _, s := range []string{"", "http://proxy.example.com:3128", "https://example.com/?page=2", "https://example.com/?author=x",
			"https://example.com/?keyword=x", "https://example.com/?pageSize=10", "https://example.com/?api_key=",
			"https://example.com/docs#section-2", "https://example.com/docs#api-key", "not a url", "admin@example.com",
			"mailto:legal@example.com", "sip:user:x@example.com", "proxy.example.com:3128", "/photoprism/storage"} {
			assert.False(t, urlHasCredential(s), s)
		}
	})
}

// TestValueHasCredential checks options values that are, or contain, a URL with a credential.
func TestValueHasCredential(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, valueHasCredential("https://token@example.com/"))
		assert.True(t, valueHasCredential([]any{"value", "https://example.com/?api_key=x"}))
		assert.True(t, valueHasCredential(map[any]any{"a": map[any]any{"b": []any{"user:secret@proxy.example.com:3128"}}}))
		assert.True(t, valueHasCredential(map[string]any{"url": "https://token@example.com/"}))
		assert.True(t, valueHasCredential(map[any]any{"https://token@example.com/": 1}))
		assert.True(t, valueHasCredential(map[string]any{"https://token@example.com/": 1}))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, valueHasCredential(nil))
		assert.False(t, valueHasCredential(42))
		assert.False(t, valueHasCredential("https://example.com/"))
		assert.False(t, valueHasCredential([]any{"value", 1, map[any]any{"url": "https://example.com/"}}))
	})
}

// TestWarnOptionsFile checks that a warning is logged once per options file and reason.
func TestWarnOptionsFile(t *testing.T) {
	hook := systemLogHook(t)
	fileName := filepath.Join(t.TempDir(), "options.yml")
	warnOptionsFile(fileName, "owned by another user")
	warnOptionsFile(fileName, "owned by another user")
	warnOptionsFile(fileName, "not supported by the filesystem")
	warnOptionsFile(filepath.Join(t.TempDir(), "options.yml"), "owned by another user")
	require.Len(t, hook.AllEntries(), 3)
	assert.Contains(t, hook.AllEntries()[0].Message, "owned by another user")
	assert.Contains(t, hook.AllEntries()[1].Message, "not supported by the filesystem")
}

// TestRestrictOptionsFileWithCredential checks that an existing options file with a credential is restricted.
func TestRestrictOptionsFileWithCredential(t *testing.T) {
	asOptionsUser(t)
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
	t.Run("NamedPipe", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, syscall.Mkfifo(fileName, 0o664))
		done := make(chan struct{})
		go func() {
			restrictOptionsFileWithCredential(fileName)
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("named pipe was read")
		}
	})
	t.Run("NamedPipeWithWriter", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, syscall.Mkfifo(fileName, 0o664))
		require.NoError(t, os.Chmod(fileName, 0o664)) //nolint:gosec // mode under test
		w, err := os.OpenFile(fileName, os.O_RDWR, 0) //nolint:gosec // test file in a temporary directory
		require.NoError(t, err)
		t.Cleanup(func() { _ = w.Close() })
		_, err = w.WriteString("DatabasePassword: secret\n")
		require.NoError(t, err)
		done := make(chan struct{})
		go func() {
			restrictOptionsFileWithCredential(fileName)
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("named pipe was read")
		}
		assert.Equal(t, os.FileMode(0o664), fileMode(t, fileName)&os.ModePerm)
	})
	t.Run("NestedUrlCredential", func(t *testing.T) {
		fileName := write(t, "Existing:\n  - https://token@example.com/\n")
		restrictOptionsFileWithCredential(fileName)
		assert.Equal(t, os.FileMode(0o640), fileMode(t, fileName))
	})
	t.Run("Invalid", func(t *testing.T) {
		// A file that cannot be parsed may hold a credential.
		fileName := write(t, "DatabaseName: [\n")
		restrictOptionsFileWithCredential(fileName)
		assert.Equal(t, os.FileMode(0o640), fileMode(t, fileName))
		restrictOptionsFileWithCredential(filepath.Join(t.TempDir(), "missing.yml"))
	})
	t.Run("TooLarge", func(t *testing.T) {
		fileName := write(t, "# "+strings.Repeat("x", optionsFileMaxBytes)+"\nDatabaseName: photoprism\n")
		restrictOptionsFileWithCredential(fileName)
		assert.Equal(t, os.FileMode(0o640), fileMode(t, fileName))
	})
	t.Run("NewConfigNamedPipe", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, syscall.Mkfifo(filepath.Join(dir, "options.yml"), 0o664))
		ctx := CliTestContext()
		require.NoError(t, ctx.Set("config-path", dir))
		systemLogHook(t)
		done := make(chan struct{})
		go func() {
			NewConfig(ctx)
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			t.Fatal("named pipe was read")
		}
	})
	t.Run("NewConfigInvalidValue", func(t *testing.T) {
		// A value that does not match its option still leaves a file with a credential restricted.
		dir := t.TempDir()
		fileName := filepath.Join(dir, "options.yml")
		require.NoError(t, os.WriteFile(fileName, []byte("JoinToken: secret\nHttpPort: 2342x\n"), 0o664)) //nolint:gosec // mode under test
		require.NoError(t, os.Chmod(fileName, 0o664))                                                     //nolint:gosec // mode under test
		ctx := CliTestContext()
		require.NoError(t, ctx.Set("config-path", dir))
		systemLogHook(t)
		NewConfig(ctx)
		assert.Equal(t, os.FileMode(0o640), fileMode(t, fileName))
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
	asOptionsUser(t)
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
	asOptionsUser(t)
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
	t.Run("RestrictOptionsFile", func(t *testing.T) {
		prevFileUid := optionsFileUid
		optionsFileUid = func(os.FileInfo) (int, bool) { return 0, true }
		t.Cleanup(func() { optionsFileUid = prevFileUid })
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, os.WriteFile(fileName, []byte("DatabasePassword: x\n"), 0o664)) //nolint:gosec // mode under test
		require.NoError(t, os.Chmod(fileName, 0o664))                                      //nolint:gosec // mode under test
		f, err := os.Open(fileName)                                                        //nolint:gosec // test file in a temporary directory
		require.NoError(t, err)
		restrictOptionsFile(f, fileName)
		require.NoError(t, f.Close())
		assert.Equal(t, os.FileMode(0o664), fileMode(t, fileName))
	})
	assert.Empty(t, hook.AllEntries())
}

// TestRestrictOptionsFileWithCredential_Symlink checks that loading the options leaves a symbolic link, e.g. a ConfigMap, as it is.
func TestRestrictOptionsFileWithCredential_Symlink(t *testing.T) {
	asOptionsUser(t)
	dir := t.TempDir()
	target := filepath.Join(dir, "target.yml")
	require.NoError(t, os.WriteFile(target, []byte("DatabasePassword: secret\n"), 0o664)) //nolint:gosec // mode under test
	require.NoError(t, os.Chmod(target, 0o664))                                           //nolint:gosec // mode under test
	link := filepath.Join(dir, "options.yml")
	require.NoError(t, os.Symlink(target, link))
	restrictOptionsFileWithCredential(link)
	assert.Equal(t, os.FileMode(0o664), fileMode(t, target))
}

// TestParamsHaveCredential checks which URL parameters name a credential.
func TestParamsHaveCredential(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		for _, s := range []string{"api_key=x", "a=1&accessToken=x", "a=1;client_secret=x", "Api%5FKey=x", "api_ke%79=x", "sig=abc%"} {
			assert.True(t, paramsHaveCredential(s), s)
		}
	})
	t.Run("False", func(t *testing.T) {
		for _, s := range []string{"", "api_key=", "api_key", "page=2&author=x", "keyword=x;w=100"} {
			assert.False(t, paramsHaveCredential(s), s)
		}
	})
}

// TestReadOptionsFile checks which options files are read.
func TestReadOptionsFile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dir := t.TempDir()
		fileName := filepath.Join(dir, "target.yml")
		require.NoError(t, os.WriteFile(fileName, []byte("DatabaseName: photoprism\n"), 0o600))
		data, err := readOptionsFile(fileName)
		require.NoError(t, err)
		assert.Equal(t, "DatabaseName: photoprism\n", string(data))
		link := filepath.Join(dir, "options.yml")
		require.NoError(t, os.Symlink(fileName, link))
		data, err = readOptionsFile(link)
		require.NoError(t, err)
		assert.Equal(t, "DatabaseName: photoprism\n", string(data))
	})
	t.Run("NamedPipe", func(t *testing.T) {
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, syscall.Mkfifo(fileName, 0o600))
		withinTimeout(t, func() {
			_, err := readOptionsFile(fileName)
			assert.ErrorIs(t, err, errOptionsFileType)
		})
	})
	t.Run("Missing", func(t *testing.T) {
		_, err := readOptionsFile(filepath.Join(t.TempDir(), "options.yml"))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

// TestReadOptionsData checks the size and type limits for an open options file.
func TestReadOptionsData(t *testing.T) {
	read := func(t *testing.T, size int) ([]byte, error) {
		fileName := filepath.Join(t.TempDir(), "options.yml")
		require.NoError(t, os.WriteFile(fileName, []byte(strings.Repeat("#", size)), 0o600))
		f, err := os.Open(fileName) //nolint:gosec // test file in a temporary directory
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.Close() })
		return readOptionsData(f)
	}
	t.Run("Success", func(t *testing.T) {
		data, err := read(t, optionsFileMaxBytes)
		require.NoError(t, err)
		assert.Len(t, data, optionsFileMaxBytes)
	})
	t.Run("TooLarge", func(t *testing.T) {
		_, err := read(t, optionsFileMaxBytes+1)
		assert.ErrorIs(t, err, errOptionsFileSize)
	})
	t.Run("Directory", func(t *testing.T) {
		f, err := os.Open(t.TempDir())
		require.NoError(t, err)
		t.Cleanup(func() { _ = f.Close() })
		_, err = readOptionsData(f)
		assert.ErrorIs(t, err, errOptionsFileType)
	})
}

// TestOpenOptionsFile checks which options files are opened.
func TestOpenOptionsFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.yml")
	require.NoError(t, os.WriteFile(target, []byte("DatabaseName: photoprism\n"), 0o600))
	link := filepath.Join(dir, "link.yml")
	require.NoError(t, os.Symlink(target, link))
	pipe := filepath.Join(dir, "pipe.yml")
	require.NoError(t, syscall.Mkfifo(pipe, 0o600))
	t.Run("Success", func(t *testing.T) {
		for _, fileName := range []string{target, link} {
			f, err := openOptionsFile(fileName, os.O_RDONLY, 0)
			require.NoError(t, err, fileName)
			require.NoError(t, f.Close())
		}
		f, err := openOptionsFile(filepath.Join(dir, "new.yml"), os.O_WRONLY|os.O_CREATE, 0o600)
		require.NoError(t, err)
		require.NoError(t, f.Close())
	})
	t.Run("NotRegular", func(t *testing.T) {
		withinTimeout(t, func() {
			_, err := openOptionsFile(pipe, os.O_RDONLY, 0)
			assert.ErrorIs(t, err, errOptionsFileType)
		})
		_, err := openOptionsFile(link, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
		assert.ErrorIs(t, err, errOptionsFileType)
		_, err = openOptionsFile(dir, os.O_RDONLY, 0)
		assert.ErrorIs(t, err, errOptionsFileType)
	})
}
