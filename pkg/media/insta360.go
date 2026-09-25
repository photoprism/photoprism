package media

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/photoprism/photoprism/pkg/fs"
)

// Supported Insta360 video roles identify the original lens and proxy files in a capture.
const (
	insta360MetadataScanLimit = 64 * 1024
	insta360OneRSModel        = "Insta360 OneRS"
	insta360OneRSImageField   = "\x00Arashi Vision\x00Insta360 OneRS\x00"
)

var insta360OneRSVideoField = append([]byte{0x12, byte(len(insta360OneRSModel))}, []byte(insta360OneRSModel)...)

// Insta360VideoRole identifies a file's purpose within a multi-file Insta360 capture.
type Insta360VideoRole string

const (
	// Insta360VideoUnknown identifies an unsupported or unrecognized capture filename.
	Insta360VideoUnknown Insta360VideoRole = ""
	// Insta360VideoLeft identifies the canonical _00 lens original.
	Insta360VideoLeft Insta360VideoRole = "left"
	// Insta360VideoRight identifies the matching _10 lens original.
	Insta360VideoRight Insta360VideoRole = "right"
	// Insta360VideoProxy identifies the low-resolution _11 proxy.
	Insta360VideoProxy Insta360VideoRole = "proxy"
)

// Insta360VideoName contains the normalized identity of an Insta360 capture file. Photo is set for
// separate-lens photos, whose naming is assumed since no sample contains such a pair.
type Insta360VideoName struct {
	Directory string
	Date      string
	Time      string
	Sequence  string
	Role      Insta360VideoRole
	Photo     bool
}

// ParseInsta360VideoName parses a supported Insta360 multi-file video filename.
func ParseInsta360VideoName(fileName string) (result Insta360VideoName, ok bool) {
	baseName := filepath.Base(fileName)
	matches := fs.Insta360VideoPattern.FindStringSubmatch(baseName)

	if len(matches) != 6 {
		return result, false
	}

	prefix := strings.ToUpper(matches[1])
	lens := matches[4]

	switch {
	case prefix == "VID" && lens == "00":
		result.Role = Insta360VideoLeft
	case prefix == "VID" && lens == "10":
		result.Role = Insta360VideoRight
	case prefix == "LRV" && lens == "11":
		result.Role = Insta360VideoProxy
	default:
		return Insta360VideoName{}, false
	}

	result.Directory = filepath.Dir(fileName)
	result.Date = matches[2]
	result.Time = matches[3]
	result.Sequence = matches[5]

	return result, true
}

// ParseInsta360PhotoName parses the lens file name of an Insta360 separate-lens photo.
func ParseInsta360PhotoName(fileName string) (result Insta360VideoName, ok bool) {
	matches := fs.Insta360PhotoPattern.FindStringSubmatch(filepath.Base(fileName))

	if len(matches) != 6 {
		return result, false
	}

	switch matches[4] {
	case "00":
		result.Role = Insta360VideoLeft
	case "10":
		result.Role = Insta360VideoRight
	default:
		return Insta360VideoName{}, false
	}

	result.Directory = filepath.Dir(fileName)
	result.Date = matches[2]
	result.Time = matches[3]
	result.Sequence = matches[5]
	result.Photo = true

	return result, true
}

// CaptureKey returns a directory-scoped identity shared by all files in the capture.
func (m Insta360VideoName) CaptureKey() string {
	if m.Date == "" || m.Time == "" || m.Sequence == "" {
		return ""
	}

	return filepath.Join(m.Directory, m.Date+"_"+m.Time+"_"+m.Sequence)
}

// FileName returns the expected filename for the specified capture role.
func (m Insta360VideoName) FileName(role Insta360VideoRole) string {
	prefix, ext := "VID", fs.ExtInsv
	var lens string

	if m.Photo {
		prefix, ext = "IMG", fs.ExtInsp
	}

	switch {
	case role == Insta360VideoLeft:
		lens = "00"
	case role == Insta360VideoRight:
		lens = "10"
	case role == Insta360VideoProxy && !m.Photo:
		prefix, lens = "LRV", "11"
	default:
		return ""
	}

	if m.Date == "" || m.Time == "" || m.Sequence == "" {
		return ""
	}

	baseName := fmt.Sprintf("%s_%s_%s_%s_%s%s", prefix, m.Date, m.Time, lens, m.Sequence, ext)

	return filepath.Join(m.Directory, baseName)
}

// Insta360CameraModelFile returns a recognized camera model from an Insta360 original.
func Insta360CameraModelFile(fileName string) (model string, err error) {
	if fileName == "" {
		return "", errors.New("filename missing")
	}

	// #nosec G304 -- the media filename is validated and supplied by the caller.
	file, err := os.Open(fileName)

	if err != nil {
		return "", err
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	fileSize, err := file.Seek(0, io.SeekEnd)

	if err != nil {
		return "", err
	} else if fileSize <= 0 {
		return "", nil
	}

	headSize := min(fileSize, int64(insta360MetadataScanLimit))
	head := make([]byte, headSize)

	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return "", err
	} else if _, err = io.ReadFull(file, head); err != nil {
		return "", fmt.Errorf("read Insta360 metadata header: %w", err)
	} else if bytes.Contains(head, []byte(insta360OneRSImageField)) {
		return insta360OneRSModel, nil
	}

	tailSize := min(fileSize, int64(insta360MetadataScanLimit))
	tail := make([]byte, tailSize)

	if _, err = file.Seek(-tailSize, io.SeekEnd); err != nil {
		return "", err
	} else if _, err = io.ReadFull(file, tail); err != nil {
		return "", fmt.Errorf("read Insta360 metadata trailer: %w", err)
	}

	if bytes.Contains(tail, insta360OneRSVideoField) {
		return insta360OneRSModel, nil
	}

	return "", nil
}
