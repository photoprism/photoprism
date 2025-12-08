package fs

import (
	"os"
	"path/filepath"
)

// ConfigFilePath builds an absolute path for a configuration file using the
// provided directory, base name, and preferred extension. If a file with the
// preferred extension already exists, that path is returned. Otherwise the
// helper searches for known sibling extensions (for example `.yaml` vs
// `.yml`) so callers transparently reuse whichever variant an admin created.
// When no matching file exists, the preferred extension is appended.
func ConfigFilePath(configPath, baseName, defaultExt string) string {
	// Return empty file path is no file name was specified.
	if baseName == "" {
		return ""
	}

	// Search file in current directory if configPath is empty.
	if configPath == "" {
		if dir, err := os.Getwd(); err == nil && dir != "" {
			configPath = dir
		}
	}

	defaultPath := filepath.Join(configPath, baseName+defaultExt)

	// If the default file exists, return its file path and look no further.
	if FileExists(defaultPath) {
		return defaultPath
	}

	// If the default file does not exist, check for a file
	// with an alternative extension that already exists.
	switch defaultExt {
	case ExtNone:
		if altPath := filepath.Join(configPath, baseName+ExtLocal); FileExists(altPath) {
			return altPath
		}
	case ExtYml:
		if altPath := filepath.Join(configPath, baseName+ExtYaml); FileExists(altPath) {
			return altPath
		}
	case ExtYaml:
		if altPath := filepath.Join(configPath, baseName+ExtYml); FileExists(altPath) {
			return altPath
		}
	case ExtGeoJson:
		if altPath := filepath.Join(configPath, baseName+ExtJson); FileExists(altPath) {
			return altPath
		}
	case ExtTml:
		if altPath := filepath.Join(configPath, baseName+ExtToml); FileExists(altPath) {
			return altPath
		}
	case ExtToml:
		if altPath := filepath.Join(configPath, baseName+ExtTml); FileExists(altPath) {
			return altPath
		}
	case ExtMd:
		if altPath := filepath.Join(configPath, baseName+ExtMarkdown); FileExists(altPath) {
			return altPath
		}
	case ExtMarkdown:
		if altPath := filepath.Join(configPath, baseName+ExtMd); FileExists(altPath) {
			return altPath
		}
	case ExtHTML:
		if altPath := filepath.Join(configPath, baseName+ExtHTM); FileExists(altPath) {
			return altPath
		} else if altPath = filepath.Join(configPath, baseName+ExtXHTML); FileExists(altPath) {
			return altPath
		}
	case ExtHTM:
		if altPath := filepath.Join(configPath, baseName+ExtHTML); FileExists(altPath) {
			return altPath
		} else if altPath = filepath.Join(configPath, baseName+ExtXHTML); FileExists(altPath) {
			return altPath
		}
	case ExtPb:
		if altPath := filepath.Join(configPath, baseName+ExtProto); FileExists(altPath) {
			return altPath
		}
	case ExtProto:
		if altPath := filepath.Join(configPath, baseName+ExtPb); FileExists(altPath) {
			return altPath
		}
	}

	// Return default config file path.
	return defaultPath
}
