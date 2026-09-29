package config

import (
	urlpkg "net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"gopkg.in/yaml.v2"

	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

// optionsModeMask removes group write and all other permissions from an options file that holds a credential.
const optionsModeMask os.FileMode = 0o027

// chmodOptionsFile changes the mode of an open options file and may be replaced in tests.
var chmodOptionsFile = (*os.File).Chmod

// optionsModeBits are the mode bits an options file keeps when access to it is restricted.
const optionsModeBits = os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky

// optionsProcessUid returns the effective user ID of the process and may be replaced in tests.
var optionsProcessUid = os.Geteuid

// optionsFileUid returns the owner of an options file, if known, and may be replaced in tests.
var optionsFileUid = func(info os.FileInfo) (int, bool) {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return int(st.Uid), true
	}

	return 0, false
}

// credentialQueryRegex matches a URL query parameter name that carries a credential, e.g. "api_key".
var credentialQueryRegex = regexp.MustCompile(`(?i)^(?:[a-z0-9]+[_-])*(?:api)?(?:key|token|secret|password|passwd|auth|signature|sig|credential)$`)

// urlHasCredential reports whether s is a URL with a password in its user info, or a query parameter
// whose name indicates a credential.
func urlHasCredential(s string) bool {
	if !strings.Contains(s, "://") {
		return false
	}

	u, err := urlpkg.Parse(s)

	if err != nil {
		return false
	} else if _, set := u.User.Password(); set {
		return true
	}

	for name := range u.Query() {
		if credentialQueryRegex.MatchString(name) {
			return true
		}
	}

	return false
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

// hasCredentialOption reports whether options values contain a credential that is set, including a URL
// that carries one.
func hasCredentialOption(values Values) bool {
	keys := credentialOptionKeys()

	for key, value := range values {
		if value == nil {
			continue
		}

		s, isString := value.(string)

		if keys[key] && (!isString || s != "") || isString && urlHasCredential(s) {
			return true
		}
	}

	return false
}

// restrictOptionsFileWithCredential restricts access to an existing options file that holds a credential,
// so a file written by an earlier version is restricted without waiting for the next write.
func restrictOptionsFileWithCredential(fileName string) {
	if optionsProcessUid() == 0 {
		return
	}

	data, err := os.ReadFile(fileName) //nolint:gosec // path derived from the config directory

	if err != nil {
		return
	}

	values := Values{}

	if yaml.Unmarshal(data, &values) != nil || !hasCredentialOption(values) {
		return
	}

	// A symbolic link, e.g. a mounted Kubernetes ConfigMap, is left as it is.
	f, err := os.OpenFile(fileName, os.O_RDONLY|syscall.O_NOFOLLOW, 0) //nolint:gosec // path derived from the config directory

	if err != nil {
		return
	}

	restrictOptionsFile(f, fileName)
	_ = f.Close()
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
// contain a credential, access is restricted before the content is replaced, unless the process runs as root.
func writeOptionsFile(fileName string, data []byte, credential bool) error {
	restrict := credential && optionsProcessUid() != 0

	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, optionsFileMode(restrict)) //nolint:gosec // path derived from the config directory

	if err != nil {
		return err
	}

	if restrict {
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
// and logs a warning. Root leaves the file as it is, since the server may run as another user.
func restrictOptionsFile(f *os.File, fileName string) {
	if optionsProcessUid() == 0 {
		return
	}

	info, err := f.Stat()

	if err != nil {
		event.SystemWarn([]string{"config", "options", "cannot restrict access to %s", "%s"}, clean.Log(fileName), clean.Error(err))
		return
	}

	mode := info.Mode() & optionsModeBits

	if mode&optionsModeMask == 0 {
		return
	}

	if uid, known := optionsFileUid(info); !known {
		return
	} else if uid != optionsProcessUid() {
		event.SystemWarn([]string{"config", "options", "cannot restrict access to %s", "owned by another user"}, clean.Log(fileName))
		return
	}

	if err = chmodOptionsFile(f, mode&^optionsModeMask); err != nil {
		event.SystemWarn([]string{"config", "options", "cannot restrict access to %s", "%s"}, clean.Log(fileName), clean.Error(err))
	} else if info, err = f.Stat(); err == nil && info.Mode()&optionsModeMask != 0 {
		// Some filesystems, such as CIFS, accept the change without applying it.
		event.SystemWarn([]string{"config", "options", "cannot restrict access to %s", "not supported by the filesystem"}, clean.Log(fileName))
	}
}
