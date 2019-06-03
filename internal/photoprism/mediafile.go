package photoprism

import (
	"fmt"
	"image"
	_ "image/gif" // Import for image.
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/djherbis/times"
	"github.com/photoprism/photoprism/internal/models"
	"github.com/photoprism/photoprism/internal/util"
)

const (
	// FileTypeOther is an unkown file format.
	FileTypeOther = "unknown"
	// FileTypeYaml is a yaml file format.
	FileTypeYaml = "yml"
	// FileTypeJpeg is a jpeg file format.
	FileTypeJpeg = "jpg"
	// FileTypePng is a png file format.
	FileTypePng = "png"
	// FileTypeRaw is a raw file format.
	FileTypeRaw = "raw"
	// FileTypeXmp is an xmp file format.
	FileTypeXmp = "xmp"
	// FileTypeAae is an aae file format.
	FileTypeAae = "aae"
	// FileTypeMovie is a movie file format.
	FileTypeMovie = "mov"
	// FileTypeHEIF High Efficiency Image File Format
	FileTypeHEIF = "heif" // High Efficiency Image File Format
)

const (
	// MimeTypeJpeg is jpeg image type
	MimeTypeJpeg = "image/jpeg"
)

// FileExtensions lists all the available and supported image file formats.
var FileExtensions = map[string]string{
	".crw":  FileTypeRaw,
	".cr2":  FileTypeRaw,
	".nef":  FileTypeRaw,
	".arw":  FileTypeRaw,
	".dng":  FileTypeRaw,
	".mov":  FileTypeMovie,
	".avi":  FileTypeMovie,
	".yml":  FileTypeYaml,
	".jpg":  FileTypeJpeg,
	".thm":  FileTypeJpeg,
	".jpeg": FileTypeJpeg,
	".xmp":  FileTypeXmp,
	".aae":  FileTypeAae,
	".heif": FileTypeHEIF,
	".heic": FileTypeHEIF,
	".3fr":  FileTypeRaw,
	".ari":  FileTypeRaw,
	".bay":  FileTypeRaw,
	".cr3":  FileTypeRaw,
	".cap":  FileTypeRaw,
	".data": FileTypeRaw,
	".dcs":  FileTypeRaw,
	".dcr":  FileTypeRaw,
	".drf":  FileTypeRaw,
	".eip":  FileTypeRaw,
	".erf":  FileTypeRaw,
	".fff":  FileTypeRaw,
	".gpr":  FileTypeRaw,
	".iiq":  FileTypeRaw,
	".k25":  FileTypeRaw,
	".kdc":  FileTypeRaw,
	".mdc":  FileTypeRaw,
	".mef":  FileTypeRaw,
	".mos":  FileTypeRaw,
	".mrw":  FileTypeRaw,
	".nrw":  FileTypeRaw,
	".obm":  FileTypeRaw,
	".orf":  FileTypeRaw,
	".pef":  FileTypeRaw,
	".ptx":  FileTypeRaw,
	".pxn":  FileTypeRaw,
	".r3d":  FileTypeRaw,
	".raf":  FileTypeRaw,
	".raw":  FileTypeRaw,
	".rwl":  FileTypeRaw,
	".rw2":  FileTypeRaw,
	".rwz":  FileTypeRaw,
	".sr2":  FileTypeRaw,
	".srf":  FileTypeRaw,
	".srw":  FileTypeRaw,
	".tif":  FileTypeRaw,
	".x3f":  FileTypeRaw,
}

// MediaFile represents a single file.
type MediaFile struct {
	filename       string
	dateCreated    time.Time
	hash           string
	fileType       string
	mimeType       string
	perceptualHash string
	tags           []string
	width          int
	height         int
	exifData       *ExifData
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

	info, err := m.ExifData()

	if err == nil && !info.DateTime.IsZero() {
		m.dateCreated = info.DateTime

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

// CameraModel returns the camera model with which the mediafile was created.
func (m *MediaFile) CameraModel() string {
	info, err := m.ExifData()

	var result string

	if err == nil {
		result = info.CameraModel
	}

	return result
}

// CameraMake returns the make of the camera with which the file was created.
func (m *MediaFile) CameraMake() string {
	info, err := m.ExifData()

	var result string

	if err == nil {
		result = info.CameraMake
	}

	return result
}

// LensModel returns the lens model of a mediafile.
func (m *MediaFile) LensModel() string {
	info, err := m.ExifData()

	var result string

	if err == nil {
		result = info.LensModel
	}

	return result
}

// LensMake returns the make of the Lens.
func (m *MediaFile) LensMake() string {
	info, err := m.ExifData()

	var result string

	if err == nil {
		result = info.LensMake
	}

	return result
}

// FocalLength return the length of the focal for a file.
func (m *MediaFile) FocalLength() float64 {
	info, err := m.ExifData()

	var result float64

	if err == nil {
		result = info.FocalLength
	}

	return result
}

// Aperture returns the aperture with which the mediafile was created.
func (m *MediaFile) Aperture() float64 {
	info, err := m.ExifData()

	var result float64

	if err == nil {
		result = info.Aperture
	}

	return result
}

// CanonicalName returns the canonical name of a mediafile.
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
func (m *MediaFile) EditedFilename() (result string) {
	basename := filepath.Base(m.filename)

	if strings.ToUpper(basename[:4]) == "IMG_" && strings.ToUpper(basename[:5]) != "IMG_E" {
		result = filepath.Dir(m.filename) + string(os.PathSeparator) + basename[:4] + "E" + basename[4:]
	}

	return result
}

// RelatedFiles returns the mediafiles which are related to a given mediafile.
func (m *MediaFile) RelatedFiles() (result MediaFiles, mainFile *MediaFile, err error) {
	baseFilename := m.CanonicalNameFromFileWithDirectory()

	matches, err := filepath.Glob(baseFilename + "*")

	if err != nil {
		return result, nil, err
	}

	if editedFilename := m.EditedFilename(); editedFilename != "" && util.Exists(editedFilename) {
		matches = append(matches, editedFilename)
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
	index := strings.Index(m.filename, directory)

	if index == 0 {
		pos := len(directory) + 1
		return m.filename[pos:]
	}

	return m.filename
}

// Directory returns the directory
func (m *MediaFile) Directory() string {
	return filepath.Dir(m.filename)
}

// Basename returns the basename.
func (m *MediaFile) Basename() string {
	return filepath.Base(m.filename)
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

// Exists checks if a mediafile exists by filename.
func (m *MediaFile) Exists() bool {
	return util.Exists(m.Filename())
}

// Remove a mediafile.
func (m *MediaFile) Remove() error {
	return os.Remove(m.Filename())
}

// HasSameFilename compares a mediafile with another mediafile and returns if
// their filenames are matching or not.
func (m *MediaFile) HasSameFilename(other *MediaFile) bool {
	return m.Filename() == other.Filename()
}

// Move a mediafile to a new file with the filename provided in parameter.
func (m *MediaFile) Move(newFilename string) error {
	if err := os.Rename(m.filename, newFilename); err != nil {
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

// Type returns the type of the mediafile.
func (m *MediaFile) Type() string {
	return FileExtensions[m.Extension()]
}

// HasType checks whether a mediafile is of a given type.
func (m *MediaFile) HasType(typeString string) bool {
	if typeString == FileTypeJpeg {
		return m.IsJpeg()
	}

	return m.Type() == typeString
}

// IsRaw check whether the given mediafile is of Raw type.
func (m *MediaFile) IsRaw() bool {
	return m.HasType(FileTypeRaw)
}

// IsHighEfficiencyImageFile check if a given mediafile is of HEIF type.
func (m *MediaFile) IsHighEfficiencyImageFile() bool {
	return m.HasType(FileTypeHEIF)
}

// IsPhoto checks if a mediafile is a photo.
func (m *MediaFile) IsPhoto() bool {
	return m.IsJpeg() || m.IsRaw() || m.IsHighEfficiencyImageFile()
}

// Jpeg returns a new mediafile given the current one's canonical name
// plus the extension .jpg.
func (m *MediaFile) Jpeg() (*MediaFile, error) {
	if m.IsJpeg() {
		return m, nil
	}

	jpegFilename := m.CanonicalNameFromFileWithDirectory() + ".jpg"

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

	exif, err := m.ExifData()

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
	if exif, err := m.ExifData(); err == nil {
		return exif.Orientation
	}

	return 1
}
