package fs

import (
	"errors"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// ErrReservedPath identifies an entry omitted by the filesystem path policy.
var ErrReservedPath = errors.New("reserved path")

// reservedPathNames contains administrative names excluded at transfer boundaries.
var reservedPathNames = map[string]struct{}{
	".aws":                         {},
	".azure":                       {},
	".bash_logout":                 {},
	".bashrc":                      {},
	".cache":                       {},
	".cargo":                       {},
	".claude":                      {},
	".claude.json":                 {},
	".codex":                       {},
	".composer":                    {},
	".config":                      {},
	".docker":                      {},
	".emulator_console_auth_token": {},
	EnvFileName:                    {},
	".env-keys":                    {},
	".forgejo":                     {},
	".git":                         {},
	".git-credentials":             {},
	".gitconfig":                   {},
	".github":                      {},
	".gnupg":                       {},
	".hg":                          {},
	".htaccess":                    {},
	".httpie":                      {},
	".kube":                        {},
	".local":                       {},
	".m2":                          {},
	".mozilla":                     {},
	".netrc":                       {},
	".npmrc":                       {},
	PPHiddenPathname:               {},
	PPStorageFilename:              {},
	".profile":                     {},
	".pypirc":                      {},
	".rsync-filter":                {},
	".secrets":                     {},
	".ssh":                         {},
	".svn":                         {},
	".temp":                        {},
	".thunderbird":                 {},
	".var":                         {},
	".xauthority":                  {},
	".zshenv":                      {},
	".zshrc":                       {},
	"_netrc":                       {},
	ClientSecretFile:               {},
	JoinTokenFile:                  {},
	SigningKeyFile:                 {},
}

// reservedPathPatterns contains dot-prefixed component patterns excluded at transfer boundaries.
var reservedPathPatterns = []string{EnvFileName + ".*", IgnoreFilePattern, ".*_history", ".bash_history-*.tmp", ".*.cnf"}

// ReservedPathNames returns a sorted copy of the reserved administrative names.
func ReservedPathNames() []string {
	return slices.Sorted(maps.Keys(reservedPathNames))
}

// ReservedPathPatterns returns a copy of the reserved component patterns.
func ReservedPathPatterns() []string {
	return slices.Clone(reservedPathPatterns)
}

// ReservedPathPolicy controls admission of managed ignore names at a filesystem boundary.
type ReservedPathPolicy struct {
	AllowIgnoreNames bool
}

// IsIgnoreFileName reports whether the basename matches the ignore-configuration pattern.
func IsIgnoreFileName(name string) bool {
	base := path.Base(strings.ReplaceAll(name, "\\", "/"))
	matched, _ := path.Match(IgnoreFilePattern, strings.ToLower(base))
	return matched
}

// HasReservedComponent reports whether a relative path contains a reserved administrative name.
func HasReservedComponent(name string) bool {
	return (ReservedPathPolicy{}).HasReservedComponent(name)
}

// HasReservedComponent checks relative path components using the boundary's ignore-name policy.
func (p ReservedPathPolicy) HasReservedComponent(name string) bool {
	for _, component := range strings.Split(strings.ReplaceAll(name, "\\", "/"), "/") {
		component = strings.ToLower(component)

		if _, reserved := reservedPathNames[component]; reserved {
			return true
		}

		if !strings.HasPrefix(component, ".") {
			continue
		}

		for _, pattern := range reservedPathPatterns {
			if p.AllowIgnoreNames && pattern == IgnoreFilePattern {
				continue
			}

			if matched, _ := path.Match(pattern, component); matched {
				return true
			}
		}
	}

	return false
}

// HasReservedTarget checks logical and resolved components relative to an operator-selected root.
func HasReservedTarget(root, name string) (bool, error) {
	return (ReservedPathPolicy{}).HasReservedTarget(root, name)
}

// HasReservedTarget checks logical names and resolved targets using the boundary's path policy.
func (p ReservedPathPolicy) HasReservedTarget(root, name string) (bool, error) {
	if p.HasReservedComponent(name) {
		return true, nil
	}

	absolute, err := filepath.Abs(root)

	if err != nil {
		return false, err
	}

	canonical, err := resolvePathAncestors(absolute)

	if err != nil {
		return false, err
	}

	target, err := resolvePathAncestors(filepath.Join(canonical, filepath.FromSlash(name)))

	if err != nil {
		return false, err
	}

	relative, err := filepath.Rel(canonical, target)

	if err != nil {
		return false, err
	}

	return p.HasReservedComponent(relative), nil
}

// resolvePathAncestors resolves existing ancestors while retaining an absent destination suffix.
func resolvePathAncestors(name string) (string, error) {
	var suffix []string

	for {
		resolved, err := filepath.EvalSymlinks(name)

		if err == nil {
			for _, s := range slices.Backward(suffix) {
				resolved = filepath.Join(resolved, s)
			}

			return resolved, nil
		}

		if !os.IsNotExist(err) {
			return "", err
		}

		if info, statErr := os.Lstat(name); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
			return "", err
		}

		parent := filepath.Dir(name)

		if parent == name {
			return "", err
		}

		suffix = append(suffix, filepath.Base(name))
		name = parent
	}
}
