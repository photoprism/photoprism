package fs

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type IgnoreLogFunc func(fileName string)

// IgnorePattern represents a name pattern to be ignored.
type IgnorePattern struct {
	Dir     string
	Pattern string
}

// NewIgnorePattern returns a new IgnorePattern.
func NewIgnorePattern(dir, pattern string, caseSensitive bool) IgnorePattern {
	if caseSensitive {
		return IgnorePattern{Dir: dir + PathSeparator, Pattern: pattern}
	} else {
		return IgnorePattern{Dir: strings.ToLower(dir) + PathSeparator, Pattern: strings.ToLower(pattern)}
	}
}

// Ignore checks if the specified file or path name in the directory should be ignored.
func (i IgnorePattern) Ignore(dir, name string) bool {
	if i.Pattern == "*" {
		// Ignore directories in which all files are ignored.
		if pathName := filepath.Join(dir, name) + PathSeparator; pathName == i.Dir {
			return true
		}
	}

	if !strings.HasPrefix(dir+PathSeparator, i.Dir) {
		// Skip check if directory does not match.
		return false
	} else if i.Pattern == name {
		// Ignore if the name exactly matches.
		return true
	}

	if strings.ContainsRune(i.Pattern, filepath.Separator) {
		name = filepath.Join(RelName(dir, i.Dir), name)
	}

	if ignore, err := filepath.Match(i.Pattern, name); ignore && err == nil {
		return true
	}

	return false
}

// IgnoreList checks a list of configurable name patterns to see whether they should be ignored.
type IgnoreList struct {
	Log           IgnoreLogFunc
	ignore        []IgnorePattern
	ignored       []string
	ignoredMutex  sync.Mutex
	ignoreHidden  bool
	hidden        []string
	caseSensitive bool
	configFile    string
	configFiles   map[string][]string
	configMutex   sync.Mutex
}

// NewIgnoreList returns a pointer to a new IgnoreList instance.
func NewIgnoreList(configFile string, ignoreHidden bool, caseSensitive bool) *IgnoreList {
	return &IgnoreList{
		Log:           func(fileName string) {},
		configFile:    configFile,
		ignored:       make([]string, 0, 256),
		ignoredMutex:  sync.Mutex{},
		hidden:        make([]string, 0, 256),
		ignoreHidden:  ignoreHidden,
		caseSensitive: caseSensitive,
		configFiles:   make(map[string][]string),
		configMutex:   sync.Mutex{},
	}
}

// Hidden returns hidden files that were ignored.
func (l *IgnoreList) Hidden() []string {
	return l.hidden
}

// Ignored returns files that were ignored in addition to hidden files.
func (l *IgnoreList) Ignored() []string {
	return l.ignored
}

// AddPatterns adds items to the list of ignored items.
func (l *IgnoreList) AddPatterns(dir string, patterns []string) error {
	if dir == "" {
		return errors.New("empty directory name")
	} else if len(patterns) == 0 {
		// Nothing to add.
		return nil
	}

	for _, pattern := range patterns {
		// Trim slashes and null bytes from the pattern.
		pattern = strings.Trim(pattern, "/\x00\t\r\n")

		// Skip empty patterns and comments that begin with "# ".
		if strings.TrimSpace(pattern) != "" && !strings.HasPrefix(pattern, "# ") {
			l.ignore = append(l.ignore, NewIgnorePattern(dir, pattern, l.caseSensitive))
		}
	}

	return nil
}

// File adds ignore patterns from the specified config file if it exists and is not empty.
func (l *IgnoreList) File(fileName string) error {
	// Return if no config file name was provided.
	if fileName == "" {
		return errors.New("empty config file name")
	}

	l.configMutex.Lock()
	defer l.configMutex.Unlock()

	// Return if config file was already added.
	if _, done := l.configFiles[fileName]; done {
		return nil
	}

	// Check if file exists and is not empty,
	// return otherwise.
	info, err := os.Stat(fileName)

	if err != nil || info.IsDir() || info.Size() == 0 {
		l.configFiles[fileName] = []string{}
		return nil
	}

	// Parse ignore config file.
	items, err := ReadLines(fileName)

	// Return error if unsuccessful,
	if err != nil {
		return err
	}

	// Add ignore patterns.
	l.configFiles[fileName] = items

	return l.AddPatterns(filepath.Dir(fileName), items)
}

// Path adds the ignore patterns found in the ignore config file of the specified directory, if any.
func (l *IgnoreList) Path(dir string) error {
	if dir == "" {
		return errors.New("missing directory name")
	} else if l.configFile == "" {
		return errors.New("missing config file name")
	}

	return l.File(filepath.Join(dir, l.configFile))
}

// Ignore checks if the file or folder name should be ignored.
func (l *IgnoreList) Ignore(name string) bool {
	// Determine the parent directory path
	// and the base name without the path.
	dir := filepath.Dir(name)
	baseName := filepath.Base(name)

	// Change name to lowercase for case-insensitive comparison.
	if l.caseSensitive == false {
		dir = strings.ToLower(dir)
		baseName = strings.ToLower(baseName)
	}

	// Ignore if name matches the config file name.
	if l.configFile != "" && baseName == l.configFile {
		_ = l.File(name)
		return true
	}

	// Use mutex for thread safety.
	l.ignoredMutex.Lock()
	defer l.ignoredMutex.Unlock()

	// Iterate through configured patterns to determine if the name should be ignored.
	for _, pattern := range l.ignore {
		if pattern.Ignore(dir, baseName) {
			l.ignored = append(l.ignored, name)

			if l.Log != nil {
				l.Log(name)
			}

			return true
		}
	}

	// Ignore hidden files and folders whose name e.g. starts with a "."?
	if l.ignoreHidden && FileNameHidden(name) {
		l.hidden = append(l.hidden, name)

		return true
	}

	return false
}

// Reset resets ignored and hidden files.
func (l *IgnoreList) Reset() {
	l.ignore = []IgnorePattern{}
	l.ignored = make([]string, 0, 256)
	l.ignoredMutex = sync.Mutex{}
	l.hidden = make([]string, 0, 256)
	l.configFiles = make(map[string][]string)
	l.configMutex = sync.Mutex{}
}
