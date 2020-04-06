package photoprism

import (
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/djherbis/times"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/pkg/capture"
	"github.com/photoprism/photoprism/pkg/fs"
)

// MediaFile represents a single photo, video or sidecar file.
type MediaFile struct {
	fileName    string
	fileType    fs.Type
	mimeType    string
	dateCreated time.Time
	hash        string
	checksum    string
	width       int
	height      int
	once        sync.Once
	metaData    meta.Data
	location    *entity.Location
}

// NewMediaFile returns a new media file.
func NewMediaFile(fileName string) (*MediaFile, error) {
	if !fs.FileExists(fileName) {
		return nil, fmt.Errorf("file does not exist: %s", fileName)
	}

	instance := &MediaFile{
		fileName: fileName,
		fileType: fs.TypeOther,
	}

	return instance, nil
}

// Stat returns the media file size and modification time.
func (m MediaFile) Stat() (size int64, mod time.Time) {
	s, err := os.Stat(m.FileName())

	if err != nil {
		log.Errorf("mediafile: unknown size (%s)", err)
		return -1, time.Now()
	}

	return s.Size(), s.ModTime()
}

// DateCreated returns the date on which the media file was created in UTC.
func (m *MediaFile) DateCreated() time.Time {
	if !m.dateCreated.IsZero() {
		return m.dateCreated
	}

	m.dateCreated = time.Now().UTC()

	info, err := m.MetaData()

	if err == nil && !info.TakenAt.IsZero() && info.TakenAt.Year() > 1000 {
		m.dateCreated = info.TakenAt.UTC()

		log.Infof("exif: taken at %s", m.dateCreated.String())

		return m.dateCreated
	}

	t, err := times.Stat(m.FileName())

	if err != nil {
		log.Debug(err.Error())

		return m.dateCreated
	}

	if t.HasBirthTime() {
		m.dateCreated = t.BirthTime().UTC()
	} else {
		m.dateCreated = t.ModTime().UTC()
	}

	log.Infof("mediafile: taken at %s", m.dateCreated.String())

	return m.dateCreated
}

func (m *MediaFile) HasTimeAndPlace() bool {
	exifData, err := m.MetaData()

	if err != nil {
		return false
	}

	result := !exifData.TakenAt.IsZero() && exifData.Lat != 0 && exifData.Lng != 0

	return result
}

// CameraModel returns the camera model with which the media file was created.
func (m *MediaFile) CameraModel() string {
	info, err := m.MetaData()

	var result string

	if err == nil {
		result = info.CameraModel
	}

	return result
}

// CameraMake returns the make of the camera with which the file was created.
func (m *MediaFile) CameraMake() string {
	info, err := m.MetaData()

	var result string

	if err == nil {
		result = info.CameraMake
	}

	return result
}

// LensModel returns the lens model of a media file.
func (m *MediaFile) LensModel() string {
	info, err := m.MetaData()

	var result string

	if err == nil {
		result = info.LensModel
	}

	return result
}

// LensMake returns the make of the Lens.
func (m *MediaFile) LensMake() string {
	info, err := m.MetaData()

	var result string

	if err == nil {
		result = info.LensMake
	}

	return result
}

// FocalLength return the length of the focal for a file.
func (m *MediaFile) FocalLength() int {
	info, err := m.MetaData()

	var result int

	if err == nil {
		result = info.FocalLength
	}

	return result
}

// FNumber returns the F number with which the media file was created.
func (m *MediaFile) FNumber() float64 {
	info, err := m.MetaData()

	var result float64

	if err == nil {
		result = info.FNumber
	}

	return result
}

// Iso returns the iso rating as int.
func (m *MediaFile) Iso() int {
	info, err := m.MetaData()

	var result int

	if err == nil {
		result = info.Iso
	}

	return result
}

// Exposure returns the exposure time as string.
func (m *MediaFile) Exposure() string {
	info, err := m.MetaData()

	var result string

	if err == nil {
		result = info.Exposure
	}

	return result
}

// CanonicalName returns the canonical name of a media file.
func (m *MediaFile) CanonicalName() string {
	return CanonicalName(m.DateCreated(), m.Checksum())
}

// CanonicalNameFromFile returns the canonical name of a file derived from the image name.
func (m *MediaFile) CanonicalNameFromFile() string {
	basename := filepath.Base(m.FileName())

	if end := strings.Index(basename, "."); end != -1 {
		return basename[:end] // Length of canonical name: 16 + 12
	}

	return basename
}

// CanonicalNameFromFileWithDirectory gets the canonical name for a MediaFile
// including the directory.
func (m *MediaFile) CanonicalNameFromFileWithDirectory() string {
	return m.Directory() + string(os.PathSeparator) + m.CanonicalNameFromFile()
}

// Hash returns the SHA1 hash of a media file.
func (m *MediaFile) Hash() string {
	if len(m.hash) == 0 {
		m.hash = fs.Hash(m.FileName())
	}

	return m.hash
}

// Checksum returns the CRC32 checksum of a media file.
func (m *MediaFile) Checksum() string {
	if len(m.checksum) == 0 {
		m.checksum = fs.Checksum(m.FileName())
	}

	return m.checksum
}

// EditedName When editing photos, iPhones create additional files like IMG_E12345.JPG
func (m *MediaFile) EditedName() string {
	basename := filepath.Base(m.fileName)

	if strings.ToUpper(basename[:4]) == "IMG_" && strings.ToUpper(basename[:5]) != "IMG_E" {
		if filename := filepath.Dir(m.fileName) + string(os.PathSeparator) + basename[:4] + "E" + basename[4:]; fs.FileExists(filename) {
			return filename
		}
	}

	return ""
}

// RelatedFiles returns files which are related to this file.
func (m *MediaFile) RelatedFiles() (result RelatedFiles, err error) {
	baseFilename := m.AbsBase()
	// escape any meta characters in the file name
	baseFilename = regexp.QuoteMeta(baseFilename)
	matches, err := filepath.Glob(baseFilename + "*")

	if err != nil {
		return result, err
	}

	if filename := m.EditedName(); filename != "" {
		matches = append(matches, filename)
	}

	for _, filename := range matches {
		resultFile, err := NewMediaFile(filename)

		if err != nil {
			continue
		}

		if result.Main == nil && resultFile.IsJpeg() {
			result.Main = resultFile
		} else if resultFile.IsRaw() {
			result.Main = resultFile
		} else if resultFile.IsHEIF() {
			result.Main = resultFile
		} else if resultFile.IsJpeg() && len(result.Main.FileName()) > len(resultFile.FileName()) {
			result.Main = resultFile
		} else if resultFile.IsImageOther() {
			result.Main = resultFile
		}

		result.Files = append(result.Files, resultFile)
	}

	sort.Sort(result.Files)

	return result, nil
}

// FileName returns the filename.
func (m MediaFile) FileName() string {
	return m.fileName
}

// SetFileName sets the filename to the given string.
func (m *MediaFile) SetFileName(fileName string) {
	m.fileName = fileName
}

// RelativeName returns the relative filename.
func (m MediaFile) RelativeName(directory string) string {
	if index := strings.Index(m.fileName, directory); index == 0 {
		if index := strings.LastIndex(directory, string(os.PathSeparator)); index == len(directory)-1 {
			pos := len(directory)
			return m.fileName[pos:]
		} else if index := strings.LastIndex(directory, string(os.PathSeparator)); index != len(directory) {
			pos := len(directory) + 1
			return m.fileName[pos:]
		}
	}

	return m.fileName
}

// RelativePath returns the relative path without filename.
func (m MediaFile) RelativePath(directory string) string {
	pathname := m.fileName

	if i := strings.Index(pathname, directory); i == 0 {
		if i := strings.LastIndex(directory, string(os.PathSeparator)); i == len(directory)-1 {
			pathname = pathname[len(directory):]
		} else if i := strings.LastIndex(directory, string(os.PathSeparator)); i != len(directory) {
			pathname = pathname[len(directory)+1:]
		}
	}

	if end := strings.LastIndex(pathname, string(os.PathSeparator)); end != -1 {
		pathname = pathname[:end]
	} else if end := strings.LastIndex(pathname, string(os.PathSeparator)); end == -1 {
		pathname = ""
	}

	return pathname
}

// RelativeBase returns the relative filename.
func (m MediaFile) RelativeBase(directory string) string {
	if relativePath := m.RelativePath(directory); relativePath != "" {
		return relativePath + string(os.PathSeparator) + m.Base()
	}

	return m.Base()
}

// Directory returns the directory
func (m MediaFile) Directory() string {
	return filepath.Dir(m.fileName)
}

// Base returns the filename base without any extensions and path.
func (m MediaFile) Base() string {
	return fs.Base(m.FileName())
}

// AbsBase returns the directory and base filename without any extensions.
func (m MediaFile) AbsBase() string {
	return m.Directory() + string(os.PathSeparator) + m.Base()
}

// MimeType returns the mime type.
func (m *MediaFile) MimeType() string {
	if m.mimeType != "" {
		return m.mimeType
	}

	m.mimeType = fs.MimeType(m.FileName())

	return m.mimeType
}

func (m *MediaFile) openFile() (*os.File, error) {
	handle, err := os.Open(m.fileName)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return handle, nil
}

// Exists checks if a media file exists by filename.
func (m MediaFile) Exists() bool {
	return fs.FileExists(m.FileName())
}

// Remove a media file.
func (m MediaFile) Remove() error {
	return os.Remove(m.FileName())
}

// HasSameName compares a media file with another media file and returns if
// their filenames are matching or not.
func (m MediaFile) HasSameName(f *MediaFile) bool {
	if f == nil {
		return false
	}

	return m.FileName() == f.FileName()
}

// Move file to a new destination with the filename provided in parameter.
func (m *MediaFile) Move(newFilename string) error {
	if err := os.Rename(m.fileName, newFilename); err != nil {
		log.Debugf("could not rename file, falling back to copy and delete: %s", err.Error())
	} else {
		m.fileName = newFilename

		return nil
	}

	if err := m.Copy(newFilename); err != nil {
		return err
	}

	if err := os.Remove(m.fileName); err != nil {
		return err
	}

	m.fileName = newFilename

	return nil
}

// Copy a MediaFile to another file by destinationFilename.
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

// Extension returns the filename extension of this media file.
func (m MediaFile) Extension() string {
	return strings.ToLower(filepath.Ext(m.fileName))
}

// IsJpeg return true if this media file is a JPEG image.
func (m MediaFile) IsJpeg() bool {
	// Don't import/use existing thumbnail files (we create our own)
	if m.Extension() == ".thm" {
		return false
	}

	return m.MimeType() == fs.MimeTypeJpeg
}

// Type returns the type of the media file.
func (m MediaFile) Type() fs.Type {
	return fs.Ext[m.Extension()]
}

// HasType returns true if this media file is of a given type.
func (m MediaFile) HasType(t fs.Type) bool {
	if t == fs.TypeJpeg {
		return m.IsJpeg()
	}

	return m.Type() == t
}

// IsRaw returns true if this media file a RAW file.
func (m MediaFile) IsRaw() bool {
	return m.HasType(fs.TypeRaw)
}

// IsPng returns true if this media file a PNG file.
func (m MediaFile) IsPng() bool {
	return m.HasType(fs.TypePng)
}

// IsTiff returns true if this media file a TIFF file.
func (m MediaFile) IsTiff() bool {
	return m.HasType(fs.TypeTiff)
}

// IsImageOther returns true this media file a PNG, GIF, BMP or TIFF file.
func (m MediaFile) IsImageOther() bool {
	switch m.Type() {
	case fs.TypeBitmap:
		return true
	case fs.TypeGif:
		return true
	case fs.TypePng:
		return true
	case fs.TypeTiff:
		return true
	default:
		return false
	}
}

// IsHEIF returns true if this media file is a High Efficiency Image File Format file.
func (m MediaFile) IsHEIF() bool {
	return m.HasType(fs.TypeHEIF)
}

// IsXMP returns true if this file is a XMP sidecar file.
func (m MediaFile) IsXMP() bool {
	return m.Type() == fs.TypeXMP
}

// IsSidecar returns true if this media file is a sidecar file (containing metadata).
func (m MediaFile) IsSidecar() bool {
	switch m.Type() {
	case fs.TypeXMP:
		return true
	case fs.TypeAAE:
		return true
	case fs.TypeXML:
		return true
	case fs.TypeYaml:
		return true
	case fs.TypeJson:
		return true
	case fs.TypeText:
		return true
	case fs.TypeMarkdown:
		return true
	default:
		return false
	}
}

// IsVideo returns true if this media file is a video file.
func (m MediaFile) IsVideo() bool {
	switch m.Type() {
	case fs.TypeMovie:
		return true
	}

	return false
}

// IsPhoto checks if this media file is a photo / image.
func (m MediaFile) IsPhoto() bool {
	return m.IsJpeg() || m.IsRaw() || m.IsHEIF() || m.IsImageOther()
}

// Jpeg returns a the JPEG version of an image or sidecar file (if exists).
func (m *MediaFile) Jpeg() (*MediaFile, error) {
	if m.IsJpeg() {
		if !fs.FileExists(m.FileName()) {
			return nil, fmt.Errorf("jpeg file should exist, but does not: %s", m.FileName())
		}

		return m, nil
	}

	jpegFilename := fmt.Sprintf("%s.%s", m.AbsBase(), fs.TypeJpeg)

	if !fs.FileExists(jpegFilename) {
		return nil, fmt.Errorf("jpeg file does not exist: %s", jpegFilename)
	}

	return NewMediaFile(jpegFilename)
}

func (m *MediaFile) decodeDimensions() error {
	if !m.IsPhoto() {
		return fmt.Errorf("not a photo: %s", m.FileName())
	}

	var width, height int

	exif, err := m.MetaData()

	if err == nil {
		width = exif.Width
		height = exif.Height
	}

	if m.IsJpeg() {
		file, err := os.Open(m.FileName())

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

// Width return the width dimension of a MediaFile.
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

// Height returns the height dimension of a MediaFile.
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

// AspectRatio returns the aspect ratio of a MediaFile.
func (m *MediaFile) AspectRatio() float64 {
	width := float64(m.Width())
	height := float64(m.Height())

	if width <= 0 || height <= 0 {
		return 0
	}

	aspectRatio := width / height

	return aspectRatio
}

// Orientation returns the orientation of a MediaFile.
func (m *MediaFile) Orientation() int {
	if exif, err := m.MetaData(); err == nil {
		return exif.Orientation
	}

	return 1
}

// Thumbnail returns a thumbnail filename.
func (m *MediaFile) Thumbnail(path string, typeName string) (filename string, err error) {
	thumbType, ok := thumb.Types[typeName]

	if !ok {
		log.Errorf("mediafile: invalid type %s", typeName)
		return "", fmt.Errorf("mediafile: invalid type %s", typeName)
	}

	thumbnail, err := thumb.FromFile(m.FileName(), m.Hash(), path, thumbType.Width, thumbType.Height, thumbType.Options...)

	if err != nil {
		log.Errorf("mediafile: could not create thumbnail (%s)", err)
		return "", fmt.Errorf("mediafile: could not create thumbnail (%s)", err)
	}

	return thumbnail, nil
}

// Thumbnail returns a resampled image of the file.
func (m *MediaFile) Resample(path string, typeName string) (img image.Image, err error) {
	filename, err := m.Thumbnail(path, typeName)

	if err != nil {
		return nil, err
	}

	return imaging.Open(filename, imaging.AutoOrientation(true))
}

func (m *MediaFile) ResampleDefault(thumbPath string, force bool) (err error) {
	count := 0
	start := time.Now()

	defer func() {
		switch count {
		case 0:
			log.Info(capture.Time(start, fmt.Sprintf("mediafile: no new thumbnails created for %s", m.Base())))
		case 1:
			log.Info(capture.Time(start, fmt.Sprintf("mediafile: one thumbnail created for %s", m.Base())))
		default:
			log.Info(capture.Time(start, fmt.Sprintf("mediafile: %d thumbnails created for %s", count, m.Base())))
		}
	}()

	hash := m.Hash()

	var originalImg *image.Image
	var sourceImg *image.Image
	var sourceImgType string

	for _, name := range thumb.DefaultTypes {
		thumbType := thumb.Types[name]

		if thumbType.SkipPreRender() {
			// Skip, size exceeds limit
			continue
		}

		if fileName, err := thumb.Filename(hash, thumbPath, thumbType.Width, thumbType.Height, thumbType.Options...); err != nil {
			log.Errorf("mediafile: could not create \"%s\" (%s)", name, err)

			return err
		} else {
			if !force && fs.FileExists(fileName) {
				continue
			}

			if originalImg == nil {
				img, err := imaging.Open(m.FileName(), imaging.AutoOrientation(true))

				if err != nil {
					log.Errorf("mediafile: can't open \"%s\" (%s)", m.FileName(), err.Error())
					return err
				}

				originalImg = &img
			}

			if thumbType.Source != "" {
				if thumbType.Source == sourceImgType && sourceImg != nil {
					_, err = thumb.Create(sourceImg, fileName, thumbType.Width, thumbType.Height, thumbType.Options...)
				} else {
					_, err = thumb.Create(originalImg, fileName, thumbType.Width, thumbType.Height, thumbType.Options...)
				}
			} else {
				sourceImg, err = thumb.Create(originalImg, fileName, thumbType.Width, thumbType.Height, thumbType.Options...)
				sourceImgType = name
			}

			if err != nil {
				log.Errorf("mediafile: could not create \"%s\" (%s)", name, err)
				return err
			}

			count++
		}
	}

	return nil
}
