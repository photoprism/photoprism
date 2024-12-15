package customize

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
)

// ImportSettings represents import settings.
type ImportSettings struct {
	Path string `json:"path" yaml:"Path"`
	Move bool   `json:"move" yaml:"Move"`
	Dest string `json:"dest" yaml:"Dest,omitempty"`
}

// DefaultImportDest specifies the default import destination file path in the Originals folder.
// The date and time placeholders are described at https://pkg.go.dev/time#Layout.
var DefaultImportDest = "2006/01/20060102_150405_82F63B78.jpg"
var ImportDestRegexp = regexp.MustCompile("^(?P<name>\\D*\\d{2,14}[\\-/_].*\\d{2,14}.*)(?P<checksum>[0-9a-fA-F]{8})(?P<count>\\.\\d{1,6}|\\.COUNT)?(?P<ext>\\.[0-9a-zA-Z]{2,8})$")

// GetPath returns the default import source path, or a custom path if set.
func (s *ImportSettings) GetPath() string {
	if s.Path != "" {
		return s.Path
	}

	return RootPath
}

// GetDest returns the default file import destination, or a custom pattern if set and valid.
func (s *ImportSettings) GetDest() string {
	if dest := strings.Trim(clean.Path(s.Dest), "/."); dest == "" || dest != s.Dest {
		s.Dest = ""
	} else if ImportDestRegexp.MatchString(dest) {
		s.Dest = dest
		return dest
	}

	return DefaultImportDest
}

// GetDestName returns the parsed import destination path and file name patterns.
func (s *ImportSettings) GetDestName() (pathName, fileName string) {
	// Parse import destination pattern.
	m := ImportDestRegexp.FindStringSubmatch(s.GetDest())

	if len(m) < 4 {
		return "", ""
	}

	// Split file path into file and path name.
	pathName, fileName = filepath.Split(m[ImportDestRegexp.SubexpIndex("name")])

	// Make sure path and file name are not empty.
	if pathName == "" || fileName == "" {
		pathName, fileName = filepath.Split(ImportDestRegexp.FindStringSubmatch(DefaultImportDest)[ImportDestRegexp.SubexpIndex("name")])
	}

	// Make sure the file name pattern ends with "_" or "-".
	if end := fileName[len(fileName)-1:]; end != "_" && end != "-" {
		fileName += "_"
	}

	return strings.Trim(pathName, "/. "), fileName
}
