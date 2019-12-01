package photoprism

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/djherbis/times"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/util"
)

// MediaFile represents a single file.
type MediaFile struct {
	filename       string
	dateCreated    time.Time
	timeZone       string
	hash           string
	fileType       string
	mimeType       string
	perceptualHash string
	width          int
	height         int
	exifData       *Exif
	location       *models.Location
}

// NewMediaFile returns a new MediaFile.
func NewMediaFile(filename string) (*MediaFile, error) {
	if !util.Exists(filename) {
		return nil, fmt.Errorf("file does not exist: %s", filename)
	}

	instance := &MediaFile{
		filename: filename,
		fileType: FileTypeOther,
	}

	return instance, nil
}

// DateCreated returns the date on which the media file was created.
func (m *MediaFile) DateCreated() time.Time {
	if !m.dateCreated.IsZero() {
		return m.dateCreated
	}

	m.dateCreated = time.Now()

	info, err := m.Exif()

	if err == nil && !info.TakenAt.IsZero() {
		m.dateCreated = info.TakenAt

		return m.dateCreated
	}

	t, err := times.Stat(m.Filename())

	if err != nil {
		log.Debug(err.Error())

		return m.dateCreated
	}

	if t.HasBirthTime() {
		m.dateCreated = t.BirthTime()
	} else {
		m.dateCreated = t.ModTime()
	}

	return m.dateCreated
}

func (m *MediaFile) HasTimeAndPlace() bool {
	exifData, err := m.Exif()

	if err != nil {
		return false
	}

	result := !exifData.TakenAt.IsZero() && exifData.Lat != 0 && exifData.Long != 0

	return result
}

// CameraModel returns the camera model with which the media file was created.
func (m *MediaFile) CameraModel() string {
	info, err := m.Exif()

	var result string

	if err == nil {
		result = info.CameraModel
	}

	return result
}

// CameraMake returns the make of the camera with which the file was created.
func (m *MediaFile) CameraMake() string {
	info, err := m.Exif()

	var result string

	if err == nil {
		result = info.CameraMake
	}

	return result
}

// LensModel returns the lens model of a media file.
func (m *MediaFile) LensModel() string {
	info, err := m.Exif()

	var result string

	if err == nil {
		result = info.LensModel
	}

	return result
}

// LensMake returns the make of the Lens.
func (m *MediaFile) LensMake() string {
	info, err := m.Exif()

	var result string

	if err == nil {
		result = info.LensMake
	}

	return result
}

// FocalLength return the length of the focal for a file.
func (m *MediaFile) FocalLength() int {
	info, err := m.Exif()

	var result int

	if err == nil {
		result = info.FocalLength
	}

	return result
}

// FNumber returns the F number with which the media file was created.
func (m *MediaFile) FNumber() float64 {
	info, err := m.Exif()

	var result float64

	if err == nil {
		result = info.FNumber
	}

	return result
}

// Iso returns the iso rating as int.
func (m *MediaFile) Iso() int {
	info, err := m.Exif()

	var result int

	if err == nil {
		result = info.Iso
	}

	return result
}

// Exposure returns the exposure time as string.
func (m *MediaFile) Exposure() string {
	info, err := m.Exif()

	var result string

	if err == nil {
		result = info.Exposure
	}

	return result
}

// CanonicalName returns the canonical name of a media file.
func (m *MediaFile) CanonicalName() string {
	var postfix string

	dateCreated := m.DateCreated().UTC()

	if fileHash := m.Hash(); len(fileHash) > 12 {
		postfix = strings.ToUpper(fileHash[:12])
	} else {
		postfix = "NOTFOUND"
	}

	result := dateCreated.Format("20060102_150405_") + postfix

	return result
}

// CanonicalNameFromFile returns the canonical name of a file derived from the image name.
func (m *MediaFile) CanonicalNameFromFile() string {
	basename := filepath.Base(m.Filename())

	if end := strings.Index(basename, "."); end != -1 {
		return basename[:end] // Length of canonical name: 16 + 12
	}

	return basename
}

// CanonicalNameFromFileWithDirectory gets the canonical name for a mediafile
// including the directory.
func (m *MediaFile) CanonicalNameFromFileWithDirectory() string {
	return m.Directory() + string(os.PathSeparator) + m.CanonicalNameFromFile()
}

// Hash return a sha1 hash of a mediafile based on the filename.
func (m *MediaFile) Hash() string {
	if len(m.hash) == 0 {
		m.hash = util.Hash(m.Filename())
	}

	return m.hash
}

// EditedFilename When editing photos, iPhones create additional files like IMG_E12345.JPG
func (m *MediaFile) EditedFilename() string {
	basename := filepath.Base(m.filename)

	if strings.ToUpper(basename[:4]) == "IMG_" && strings.ToUpper(basename[:5]) != "IMG_E" {
		if filename := filepath.Dir(m.filename) + string(os.PathSeparator) + basename[:4] + "E" + basename[4:]; util.Exists(filename) {
			return filename
		}
	}

	return ""
}

// RelatedFiles returns files which are related to this file.
func (m *MediaFile) RelatedFiles() (result MediaFiles, mainFile *MediaFile, err error) {
	baseFilename := m.DirectoryBasename()
	// escape any meta characters in the file name
	baseFilename = regexp.QuoteMeta(baseFilename)
	matches, err := filepath.Glob(baseFilename + "*")

	if err != nil {
		return result, nil, err
	}

	if filename := m.EditedFilename(); filename != "" {
		matches = append(matches, filename)
	}

	for _, filename := range matches {
		resultFile, err := NewMediaFile(filename)

		if err != nil {
			continue
		}

		if mainFile == nil && resultFile.IsJpeg() {
			mainFile = resultFile
		} else if resultFile.IsRaw() {
			mainFile = resultFile
		} else if resultFile.IsHEIF() {
			mainFile = resultFile
		} else if resultFile.IsJpeg() && len(mainFile.Filename()) > len(resultFile.Filename()) {
			mainFile = resultFile
		}

		result = append(result, resultFile)
	}

	sort.Sort(result)

	return result, mainFile, nil
}

// Filename returns the filename.
func (m *MediaFile) Filename() string {
	return m.filename
}

// SetFilename sets the filename to the given string.
func (m *MediaFile) SetFilename(filename string) {
	m.filename = filename
}

// RelativeFilename returns the relative filename.
func (m *MediaFile) RelativeFilename(directory string) string {
	if index := strings.Index(m.filename, directory); index == 0 {
		if index := strings.LastIndex(directory, string(os.PathSeparator)); index == len(directory)-1 {
			pos := len(directory)
			return m.filename[pos:]
		} else if index := strings.LastIndex(directory, string(os.PathSeparator)); index != len(directory) {
			pos := len(directory) + 1
			return m.filename[pos:]
		}
	}

	return m.filename
}

// RelativePath returns the relative path without filename.
func (m *MediaFile) RelativePath(directory string) string {
	pathname := m.filename

	if i := strings.Index(pathname, directory); i == 0 {
		if i := strings.LastIndex(directory, string(os.PathSeparator)); i == len(directory)-1 {
			pathname = pathname[len(directory):]
			log.Info(pathname)
		} else if i := strings.LastIndex(directory, string(os.PathSeparator)); i != len(directory) {
			pathname = pathname[len(directory)+1:]
			log.Info(pathname)
		}
	}

	if end := strings.LastIndex(pathname, string(os.PathSeparator)); end != -1 {
		pathname = pathname[:end]
	} else if end := strings.LastIndex(pathname, string(os.PathSeparator)); end == -1 {
		pathname = ""
	}

	return pathname
}

// RelativeBasename returns the relative filename.
func (m *MediaFile) RelativeBasename(directory string) string {
	if relativePath := m.RelativePath(directory); relativePath != "" {
		return relativePath + string(os.PathSeparator) + m.Basename()
	}

	return m.Basename()
}

// Directory returns the directory
func (m *MediaFile) Directory() string {
	return filepath.Dir(m.filename)
}

// Basename returns the filename base without any extensions and path.
func (m *MediaFile) Basename() string {
	basename := filepath.Base(m.Filename())

	if end := strings.Index(basename, "."); end != -1 {
		// ignore everything behind the first dot in the file name
		basename = basename[:end]
	}

	if end := strings.Index(basename, " ("); end != -1 {
		// copies created by Chrome & Windows, example: IMG_1234 (2)
		basename = basename[:end]
	} else if end := strings.Index(basename, " copy"); end != -1 {
		// copies created by OS X, example: IMG_1234 copy 2
		basename = basename[:end]
	}

	return basename
}

// DirectoryBasename returns the directory and base filename without any extensions.
func (m *MediaFile) DirectoryBasename() string {
	return m.Directory() + string(os.PathSeparator) + m.Basename()
}

// MimeType returns the mimetype.
func (m *MediaFile) MimeType() string {
	if m.mimeType != "" {
		return m.mimeType
	}

	handle, err := m.openFile()

	if err != nil {
		log.Errorf("could not read file to determine mime type: %s", m.Filename())
		return ""
	}

	defer handle.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = handle.Read(buffer)

	if err != nil {
		log.Errorf("could not read file to determine mime type: %s", m.Filename())
		return ""
	}

	m.mimeType = http.DetectContentType(buffer)

	return m.mimeType
}

func (m *MediaFile) openFile() (*os.File, error) {
	handle, err := os.Open(m.filename)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return handle, nil
}

// Exists checks if a media file exists by filename.
func (m *MediaFile) Exists() bool {
	return util.Exists(m.Filename())
}

// Remove a media file.
func (m *MediaFile) Remove() error {
	return os.Remove(m.Filename())
}

// HasSameFilename compares a media file with another media file and returns if
// their filenames are matching or not.
func (m *MediaFile) HasSameFilename(other *MediaFile) bool {
	return m.Filename() == other.Filename()
}

// Move file to a new destination with the filename provided in parameter.
func (m *MediaFile) Move(newFilename string) error {
	if err := os.Rename(m.filename, newFilename); err != nil {
		log.Debugf("could not rename file, falling back to copy and delete: %s", err.Error())
 	} else {
		m.filename = newFilename

		return nil
	}

	if err := m.Copy(newFilename); err != nil {
		return err
	}

	if err := os.Remove(m.filename); err != nil {
		return err
	}

	m.filename = newFilename

	return nil
}

// Copy a mediafile to another file by destinationFilename.
func (m *MediaFile) Copy(destinationFilename string) error {
	file, err := m.openFile()

	if err != nil {
		log.Error(err.Error())
		return err
	}

	defer file.Close()

	destination, err := os.OpenFile(destinationFilename, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		log.Error(err.Error())
		return err
	}

	defer destination.Close()

	_, err = io.Copy(destination, file)

	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

// Extension returns the extension of a mediafile.
func (m *MediaFile) Extension() string {
	return strings.ToLower(filepath.Ext(m.filename))
}

// IsJpeg return true if the given mediafile is of mimetype Jpeg.
func (m *MediaFile) IsJpeg() bool {
	// Don't import/use existing thumbnail files (we create our own)
	if m.Extension() == ".thm" {
		return false
	}

	return m.MimeType() == MimeTypeJpeg
}

// Type returns the type of the media file.
func (m *MediaFile) Type() string {
	return FileExtensions[m.Extension()]
}

// HasType checks whether a media file is of a given type.
func (m *MediaFile) HasType(typeString string) bool {
	if typeString == FileTypeJpeg {
		return m.IsJpeg()
	}

	return m.Type() == typeString
}

// IsRaw check whether the given media file a RAW file.
func (m *MediaFile) IsRaw() bool {
	return m.HasType(FileTypeRaw)
}

// IsHEIF check if a given media file is a High Efficiency Image File Format file.
func (m *MediaFile) IsHEIF() bool {
	return m.HasType(FileTypeHEIF)
}

// IsPhoto checks if a media file is a photo / image.
func (m *MediaFile) IsPhoto() bool {
	return m.IsJpeg() || m.IsRaw() || m.IsHEIF()
}

// Jpeg returns a the JPEG version of an image or sidecar file (if exists).
func (m *MediaFile) Jpeg() (*MediaFile, error) {
	if m.IsJpeg() {
		return m, nil
	}

	jpegFilename := fmt.Sprintf("%s.%s", m.DirectoryBasename(), FileTypeJpeg)

	if !util.Exists(jpegFilename) {
		return nil, fmt.Errorf("jpeg file does not exist: %s", jpegFilename)
	}

	return NewMediaFile(jpegFilename)
}

func (m *MediaFile) decodeDimensions() error {
	if !m.IsPhoto() {
		return fmt.Errorf("not a photo: %s", m.Filename())
	}

	var width, height int

	exif, err := m.Exif()

	if err == nil {
		width = exif.Width
		height = exif.Height
	}

	if m.IsJpeg() {
		file, err := os.Open(m.Filename())

		defer file.Close()

		size, _, err := image.DecodeConfig(file)

		if err != nil {
			return err
		}

		width = size.Width
		height = size.Height
	}

	if m.Orientation() > 4 {
		m.width = height
		m.height = width
	} else {
		m.width = width
		m.height = height
	}

	return nil
}

// Width return the width dimension of a mediafile.
func (m *MediaFile) Width() int {
	if !m.IsPhoto() {
		return 0
	}

	if m.width <= 0 {
		if err := m.decodeDimensions(); err != nil {
			log.Error(err)
		}
	}

	return m.width
}

// Height returns the height dimension of a mediafile.
func (m *MediaFile) Height() int {
	if !m.IsPhoto() {
		return 0
	}

	if m.height <= 0 {
		if err := m.decodeDimensions(); err != nil {
			log.Error(err)
		}
	}

	return m.height
}

// AspectRatio returns the aspect ratio of a mediafile.
func (m *MediaFile) AspectRatio() float64 {
	width := float64(m.Width())
	height := float64(m.Height())

	if width <= 0 || height <= 0 {
		return 0
	}

	aspectRatio := width / height

	return aspectRatio
}

// Orientation returns the orientation of a mediafile.
func (m *MediaFile) Orientation() int {
	if exif, err := m.Exif(); err == nil {
		return exif.Orientation
	}

	return 1
}
