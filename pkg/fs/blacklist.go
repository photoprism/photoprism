package fs

import (
	"strings"
)

// Blacklists represents multiple blacklists.
type Blacklists map[string]Blacklist

// NewBlacklists creates and initializes file extension blacklists.
func NewBlacklists() Blacklists {
	return make(Blacklists)
}

// Blacklist represents a file extension blacklist.
type Blacklist map[string]bool

// NewBlacklist creates and initializes a file extension blacklist.
func NewBlacklist(extensions string) Blacklist {
	list := make(Blacklist)
	if extensions != "" {
		list.Set(extensions)
	}
	return list
}

// Contains tests if a file extension is blacklisted.
func (b Blacklist) Contains(ext string) bool {
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

// Allow tests if a file extension is NOT blacklisted.
func (b Blacklist) Allow(ext string) bool {
	return !b.Contains(ext)
}

// Set initializes the blacklist with a list of file extensions.
func (b Blacklist) Set(extensions string) {
	if extensions == "" {
		return
	}

	ext := strings.Split(extensions, ",")

	for i := range ext {
		b.Add(ext[i])
	}
}

// Add adds a file extension to the blacklist.
func (b Blacklist) Add(ext string) {
	// Remove unwanted characters from file extension and make it lowercase for comparison.
	ext = TrimExt(ext)

	if ext == "" {
		return
	}

	b[ext] = true
}
