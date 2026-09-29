package config

import (
	"os"
	"reflect"
	"strings"
	"sync"
	"syscall"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// optionsModeMask removes group write and all other permissions from an options file that holds a credential.
const optionsModeMask os.FileMode = 0o027

// chmodOptionsFile changes the mode of an open options file and may be replaced in tests.
var chmodOptionsFile = (*os.File).Chmod

// optionsFileUid returns the owner of an options file, if known, and may be replaced in tests.
var optionsFileUid = func(info os.FileInfo) (int, bool) {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return int(st.Uid), true
	}

	return 0, false
}

var (
	credentialOptionKeysOnce sync.Once
	credentialOptionKeysMap  map[string]bool
)

// credentialOptionKeys returns the options.yml keys of the options marked as secret, and of the DSNs,
// which embed a database password.
func credentialOptionKeys() map[string]bool {
	credentialOptionKeysOnce.Do(func() {
		secret := make(map[string]bool)

		for _, f := range Flags {
			if f.Secret {
				secret[f.Name()] = true
			}
		}

		keys := map[string]bool{
			"DatabaseDSN":               true,
			"DatabaseDsn":               true,
			"DatabaseProvisionDSN":      true,
			"DatabaseProvisionProxyDSN": true,
		}
		t := reflect.TypeFor[Options]()

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			if key, _, _ := strings.Cut(field.Tag.Get("yaml"), ","); key != "" && key != "-" && secret[field.Tag.Get("flag")] {
				keys[key] = true
			}
		}

		credentialOptionKeysMap = keys
	})

	return credentialOptionKeysMap
}

// hasCredentialOption reports whether options values contain a credential that is set.
func hasCredentialOption(values Values) bool {
	keys := credentialOptionKeys()

	for key, value := range values {
		if !keys[key] || value == nil {
			continue
		}

		if s, ok := value.(string); !ok || s != "" {
			return true
		}
	}

	return false
}

// optionsFileMode returns the mode an options file is created with, which excludes group write and
// other permissions if it holds a credential.
func optionsFileMode(credential bool) os.FileMode {
	if credential {
		return fs.ModeConfigFile &^ optionsModeMask
	}

	return fs.ModeConfigFile
}

// writeOptionsFile writes options in place, so an existing file keeps its owner, group, and inode. If they
// contain a credential, access is restricted before the content is replaced.
func writeOptionsFile(fileName string, data []byte, credential bool) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, optionsFileMode(credential)) //nolint:gosec // path derived from the config directory

	if err != nil {
		return err
	}

	if credential {
		restrictOptionsFile(f, fileName)
	}

	if err = f.Truncate(0); err == nil {
		_, err = f.Write(data)
	}

	if closeErr := f.Close(); err == nil {
		err = closeErr
	}

	return err
}

// restrictOptionsFile removes group write and other permissions from an open options file that holds a
// credential, if it belongs to the process user. Otherwise, or if that fails, it keeps the file as it is
// and logs a warning, so the options that are written still apply.
func restrictOptionsFile(f *os.File, fileName string) {
	info, err := f.Stat()

	if err != nil {
		event.SystemWarn([]string{"config", "options", "cannot restrict access to %s", "%s"}, clean.Log(fileName), clean.Error(err))
		return
	}

	mode := info.Mode().Perm()

	if mode&optionsModeMask == 0 {
		return
	}

	if uid, known := optionsFileUid(info); !known {
		return
	} else if uid != os.Geteuid() {
		event.SystemWarn([]string{"config", "options", "cannot restrict access to %s", "owned by another user"}, clean.Log(fileName))
		return
	}

	if err = chmodOptionsFile(f, mode&^optionsModeMask); err != nil {
		event.SystemWarn([]string{"config", "options", "cannot restrict access to %s", "%s"}, clean.Log(fileName), clean.Error(err))
	}
}
