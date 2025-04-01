package fs

import (
	"sort"
	"strings"
)

// ExtLists represents multiple extension lists.
type ExtLists map[string]ExtList

// NewExtLists creates and initializes file extension list.
func NewExtLists() ExtLists {
	return make(ExtLists)
}

// ExtList represents a file extension list.
type ExtList map[string]bool

// NewExtList creates and initializes a file extension list.
func NewExtList(extensions string) ExtList {
	list := make(ExtList)
	if extensions != "" {
		list.Set(extensions)
	}
	return list
}

// Contains tests if a file extension is listed.
func (b ExtList) Contains(ext string) bool {
	// Skip check if list is empty.
	if len(b) == 0 {
		return false
	}

	// Remove unwanted characters from file extension and make it lowercase for comparison.
	ext = TrimExt(ext)

	// Skip check if extension is empty.
	if ext == "" {
		return false
	}

	if _, ok := b[ext]; ok {
		return true
	}

	return false
}

// Excludes tests if the extension is not included, or returns false if the list is empty.
func (b ExtList) Excludes(ext string) bool {
	if len(b) == 0 {
		return false
	}

	return !b.Contains(ext)
}

// Allow tests if a file extension is NOT listed.
func (b ExtList) Allow(ext string) bool {
	return !b.Contains(ext)
}

// Set initializes the list with a list of file extensions.
func (b ExtList) Set(extensions string) {
	if extensions == "" {
		return
	}

	ext := strings.Split(extensions, ",")

	for i := range ext {
		b.Add(ext[i])
	}
}

// Add adds a file extension to the list.
func (b ExtList) Add(ext string) {
	// Remove unwanted characters from file extension and make it lowercase for comparison.
	ext = TrimExt(ext)

	if ext == "" {
		return
	}

	b[ext] = true
}

// String returns the list as a comma-separated list in alphabetical order.
func (b ExtList) String() string {
	if len(b) == 0 {
		return ""
	}

	list := make([]string, 0, len(b))

	for s := range b {
		list = append(list, s)
	}

	sort.Strings(list)

	return strings.Join(list, ", ")
}

// Accept returns a comma-separated list in alphabetical order for use as an input accept attribute.
func (b ExtList) Accept() string {
	if len(b) == 0 {
		return ""
	}

	list := make([]string, 0, len(b))

	for s := range b {
		if s = TrimExt(s); s != "" {
			list = append(list, "."+s)
		}
	}

	sort.Strings(list)

	return strings.Join(list, ",")
}
