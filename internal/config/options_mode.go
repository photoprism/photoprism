package config

import (
	"errors"
	"io"
	iofs "io/fs"
	urlpkg "net/url"
	"os"
	"reflect"
	"regexp"
	"slices"
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

// optionsFileMaxBytes is the maximum size of an options file.
const optionsFileMaxBytes = 1 << 20

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

// credentialQueryRegex matches a URL query parameter name that carries a credential, e.g. "api_key" or
// "accessToken". It matches any name ending like one, so a false positive only restricts the file.
var credentialQueryRegex = regexp.MustCompile(`(?i)(?:key|token|secret|password|passwd|pwd|pass|pw|auth|authorization|bearer|jwt|signature|sig|credential)s?$`)

// schemelessUserRegex matches a URL without a scheme that begins with a user and password, e.g. "user:pass@proxy:3128".
var schemelessUserRegex = regexp.MustCompile(`^[^\s/@:]+:[^\s/@]*@[^\s/@]+`)

// schemelessNoUserRegex matches a URI scheme whose value contains "@" without user info, e.g. "mailto:".
var schemelessNoUserRegex = regexp.MustCompile(`(?i)^(?:mailto|sips?|tel|xmpp):`)

// urlHasCredential reports whether s is a URL with user info, including one without a scheme, or with a
// query or fragment parameter whose name indicates a credential. A URL that cannot be parsed counts if it
// contains "@".
func urlHasCredential(s string) bool {
	if !strings.Contains(s, "://") {
		return schemelessUserRegex.MatchString(s) && !schemelessNoUserRegex.MatchString(s)
	}

	u, err := urlpkg.Parse(s)

	if err != nil {
		return strings.Contains(s, "@")
	} else if u.User != nil && u.User.String() != "" {
		return true
	}

	return paramsHaveCredential(u.RawQuery) || paramsHaveCredential(u.Fragment)
}

// paramsHaveCredential reports whether URL parameters, separated by "&" or ";", include one with a value
// whose name indicates a credential.
func paramsHaveCredential(params string) bool {
	for _, param := range strings.FieldsFunc(params, func(r rune) bool { return r == '&' || r == ';' }) {
		name, value, _ := strings.Cut(param, "=")

		if unescaped, err := urlpkg.QueryUnescape(name); err == nil {
			name = unescaped
		}

		if value != "" && credentialQueryRegex.MatchString(name) {
			return true
		}
	}

	return false
}

// valueHasCredential reports whether an options value is, or contains in a list or map, a URL that
// carries a credential.
func valueHasCredential(value any) bool {
	switch v := value.(type) {
	case string:
		return urlHasCredential(v)
	case []any:
		return slices.ContainsFunc(v, valueHasCredential)
	case map[any]any:
		for key, item := range v {
			if valueHasCredential(key) || valueHasCredential(item) {
				return true
			}
		}
	case map[string]any:
		for key, item := range v {
			if urlHasCredential(key) || valueHasCredential(item) {
				return true
			}
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

		if s, isString := value.(string); keys[key] && (!isString || s != "") || valueHasCredential(value) {
			return true
		}
	}

	return false
}

// restrictOptionsFileWithCredential restricts access to an existing options file that holds a credential,
// so a file written by an earlier version is restricted without waiting for the next write. Only a regular
// file is read, so a symbolic link, e.g. a mounted Kubernetes ConfigMap, or a named pipe is left as it is.
func restrictOptionsFileWithCredential(fileName string) {
	if optionsProcessUid() == 0 {
		return
	}

	f, err := openOptionsFile(fileName, os.O_RDONLY|syscall.O_NOFOLLOW, 0)

	if err != nil {
		return
	}

	defer f.Close()

	data, err := readOptionsData(f)

	// A file that is too large or cannot be parsed may hold a credential, so it is restricted as well.
	if errors.Is(err, errOptionsFileType) {
		return
	} else if values := (Values{}); err == nil && yaml.Unmarshal(data, &values) == nil && !hasCredentialOption(values) {
		return
	}

	restrictOptionsFile(f, fileName)
}

var (
	// errOptionsFileType is returned for an options file that is not a regular file.
	errOptionsFileType = errors.New("not a regular file")
	// errOptionsFileSize is returned for an options file larger than optionsFileMaxBytes.
	errOptionsFileSize = errors.New("file too large")
)

// openOptionsFile opens an options file that is missing or a regular file, so a special file such as a
// named pipe is never opened; a symbolic link is refused if flag contains O_NOFOLLOW.
func openOptionsFile(fileName string, flag int, perm os.FileMode) (*os.File, error) {
	stat := os.Stat

	if flag&syscall.O_NOFOLLOW != 0 {
		stat = os.Lstat
	}

	if info, err := stat(fileName); err == nil && !info.Mode().IsRegular() {
		return nil, &iofs.PathError{Op: "open", Path: fileName, Err: errOptionsFileType}
	}

	return os.OpenFile(fileName, flag, perm) //nolint:gosec // path derived from the config directory
}

// readOptionsFile reads an options file, following a symbolic link, if it is a regular file.
func readOptionsFile(fileName string) ([]byte, error) {
	f, err := openOptionsFile(fileName, os.O_RDONLY, 0)

	if err != nil {
		return nil, err
	}

	defer f.Close()

	return readOptionsData(f)
}

// readOptionsData reads an open options file of up to optionsFileMaxBytes, if it is a regular file.
func readOptionsData(f *os.File) ([]byte, error) {
	if info, err := f.Stat(); err != nil {
		return nil, err
	} else if !info.Mode().IsRegular() {
		return nil, &iofs.PathError{Op: "read", Path: f.Name(), Err: errOptionsFileType}
	}

	data, err := io.ReadAll(io.LimitReader(f, optionsFileMaxBytes+1))

	if err == nil && len(data) > optionsFileMaxBytes {
		return nil, &iofs.PathError{Op: "read", Path: f.Name(), Err: errOptionsFileSize}
	}

	return data, err
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

	f, err := openOptionsFile(fileName, os.O_WRONLY|os.O_CREATE, optionsFileMode(restrict))

	if err != nil {
		return err
	} else if info, statErr := f.Stat(); statErr != nil {
		_ = f.Close()
		return statErr
	} else if !info.Mode().IsRegular() {
		_ = f.Close()
		return &iofs.PathError{Op: "write", Path: fileName, Err: errOptionsFileType}
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
		warnOptionsFile(fileName, clean.Error(err))
		return
	}

	mode := info.Mode() & optionsModeBits

	if mode&optionsModeMask == 0 {
		return
	}

	if uid, known := optionsFileUid(info); !known {
		return
	} else if uid != optionsProcessUid() {
		warnOptionsFile(fileName, "owned by another user")
		return
	}

	if err = chmodOptionsFile(f, mode&^optionsModeMask); err != nil {
		warnOptionsFile(fileName, clean.Error(err))
	} else if info, err = f.Stat(); err == nil && info.Mode()&optionsModeMask != 0 {
		// Some filesystems, such as CIFS, accept the change without applying it.
		warnOptionsFile(fileName, "not supported by the filesystem")
	}
}

// optionsFileWarned records the options file warnings that were logged.
var optionsFileWarned sync.Map

// warnOptionsFile logs that access to an options file cannot be restricted, once per file and reason.
func warnOptionsFile(fileName, reason string) {
	if _, warned := optionsFileWarned.LoadOrStore(fileName+"\x00"+reason, true); warned {
		return
	}

	event.SystemWarn([]string{"config", "options", "cannot restrict access to %s", "%s"}, clean.Log(fileName), reason)
}
