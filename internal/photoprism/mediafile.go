package photoprism

import (
	"fmt"
	"image"
	"io"
	"os"
	"path"
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
	"github.com/photoprism/photoprism/pkg/txt"
)

// MediaFile represents a single photo, video or sidecar file.
type MediaFile struct {
	fileName     string
	fileType     fs.FileType
	mimeType     string
	takenAt      time.Time
	takenAtSrc   string
	hash         string
	checksum     string
	width        int
	height       int
	metaData     meta.Data
	metaDataOnce sync.Once
	location     *entity.Cell
}

// NewMediaFile returns a new media file.
func NewMediaFile(fileName string) (*MediaFile, error) {
	if !fs.FileExists(fileName) {
		return nil, fmt.Errorf("%s does not exist", filepath.Base(fileName))
	}

	instance := &MediaFile{
		fileName: fileName,
		fileType: fs.TypeOther,
		metaData: meta.NewData(),
	}

	return instance, nil
}

// Stat returns the media file size and modification time.
func (m *MediaFile) Stat() (size int64, mod time.Time) {
	s, err := os.Stat(m.FileName())

	if err != nil {
		log.Errorf("mediafile: unknown size (%s)", err)
		return -1, time.Now()
	}

	return s.Size(), s.ModTime().Round(time.Second)
}

// FileSize returns the media file size.
func (m *MediaFile) FileSize() (size int64) {
	s, err := os.Stat(m.FileName())

	if err != nil {
		log.Errorf("mediafile: unknown size (%s)", err)
		return -1
	}

	return s.Size()
}

// DateCreated returns only the date on which the media file was probably taken in UTC.
func (m *MediaFile) DateCreated() time.Time {
	takenAt, _ := m.TakenAt()

	return takenAt
}

// TakenAt returns the date on which the media file was taken in UTC and the source of this information.
func (m *MediaFile) TakenAt() (time.Time, string) {
	if !m.takenAt.IsZero() {
		return m.takenAt, m.takenAtSrc
	}

	m.takenAt = time.Now().UTC()

	data := m.MetaData()

	if data.Error == nil && !data.TakenAt.IsZero() && data.TakenAt.Year() > 1000 {
		m.takenAt = data.TakenAt.UTC()
		m.takenAtSrc = entity.SrcMeta

		log.Infof("mediafile: %s was taken at %s (%s)", filepath.Base(m.fileName), m.takenAt.String(), m.takenAtSrc)

		return m.takenAt, m.takenAtSrc
	}

	if nameTime := txt.Time(m.fileName); !nameTime.IsZero() {
		m.takenAt = nameTime
		m.takenAtSrc = entity.SrcName

		log.Infof("mediafile: %s was taken at %s (%s)", filepath.Base(m.fileName), m.takenAt.String(), m.takenAtSrc)

		return m.takenAt, m.takenAtSrc
	}

	m.takenAtSrc = entity.SrcAuto

	fileInfo, err := times.Stat(m.FileName())

	if err != nil {
		log.Warnf("mediafile: %s (file stat)", err.Error())
		log.Infof("mediafile: %s was taken at %s (now)", filepath.Base(m.fileName), m.takenAt.String())

		return m.takenAt, m.takenAtSrc
	}

	if fileInfo.HasBirthTime() {
		m.takenAt = fileInfo.BirthTime().UTC()
		log.Infof("mediafile: %s was taken at %s (file birth time)", filepath.Base(m.fileName), m.takenAt.String())
	} else {
		m.takenAt = fileInfo.ModTime().UTC()
		log.Infof("mediafile: %s was taken at %s (file mod time)", filepath.Base(m.fileName), m.takenAt.String())
	}

	return m.takenAt, m.takenAtSrc
}

func (m *MediaFile) HasTimeAndPlace() bool {
	data := m.MetaData()

	result := !data.TakenAt.IsZero() && data.Lat != 0 && data.Lng != 0

	return result
}

// CameraModel returns the camera model with which the media file was created.
func (m *MediaFile) CameraModel() string {
	data := m.MetaData()

	return data.CameraModel
}

// CameraMake returns the make of the camera with which the file was created.
func (m *MediaFile) CameraMake() string {
	data := m.MetaData()

	return data.CameraMake
}

// LensModel returns the lens model of a media file.
func (m *MediaFile) LensModel() string {
	data := m.MetaData()

	return data.LensModel
}

// LensMake returns the make of the Lens.
func (m *MediaFile) LensMake() string {
	data := m.MetaData()

	return data.LensMake
}

// FocalLength return the length of the focal for a file.
func (m *MediaFile) FocalLength() int {
	data := m.MetaData()

	return data.FocalLength
}

// FNumber returns the F number with which the media file was created.
func (m *MediaFile) FNumber() float32 {
	data := m.MetaData()

	return data.FNumber
}

// Iso returns the iso rating as int.
func (m *MediaFile) Iso() int {
	data := m.MetaData()

	return data.Iso
}

// Exposure returns the exposure time as string.
func (m *MediaFile) Exposure() string {
	data := m.MetaData()

	return data.Exposure
}

// CanonicalName returns the canonical name of a media file.
func (m *MediaFile) CanonicalName() string {
	return fs.CanonicalName(m.DateCreated(), m.Checksum())
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
	return m.Dir() + string(os.PathSeparator) + m.CanonicalNameFromFile()
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

// EditedName returns the corresponding edited image file name as used by Apple (e.g. IMG_E12345.JPG).
func (m *MediaFile) EditedName() string {
	basename := filepath.Base(m.fileName)

	if strings.ToUpper(basename[:4]) == "IMG_" && strings.ToUpper(basename[:5]) != "IMG_E" {
		if filename := filepath.Dir(m.fileName) + string(os.PathSeparator) + basename[:4] + "E" + basename[4:]; fs.FileExists(filename) {
			return filename
		}
	}

	return ""
}

// JsonName returns the corresponding JSON sidecar file name as used by Google Photos (and potentially other apps).
func (m *MediaFile) JsonName() string {
	jsonName := m.fileName + ".json"

	if fs.FileExists(jsonName) {
		return jsonName
	}

	return ""
}

// RelatedFiles returns files which are related to this file.
func (m *MediaFile) RelatedFiles(stripSequence bool) (result RelatedFiles, err error) {
	// escape any meta characters in the file name
	matches, err := filepath.Glob(regexp.QuoteMeta(m.AbsPrefix(stripSequence)) + "*")

	if err != nil {
		return result, err
	}

	if name := m.EditedName(); name != "" {
		matches = append(matches, name)
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
		} else if resultFile.IsVideo() {
			result.Main = resultFile
		}

		result.Files = append(result.Files, resultFile)
	}

	if result.Main == nil {
		return result, fmt.Errorf("no main file found for %s", txt.Quote(m.BaseName()))
	}

	// Add hidden JPEG if exists.
	if !result.ContainsJpeg() {
		if jpegName := fs.TypeJpeg.FindFirst(result.Main.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), stripSequence); jpegName != "" {
			if resultFile, err := NewMediaFile(jpegName); err == nil {
				result.Files = append(result.Files, resultFile)
			}
		}
	}

	sort.Sort(result.Files)

	return result, nil
}

// PathNameInfo returns file name infos for indexing.
func (m *MediaFile) PathNameInfo() (fileRoot, fileBase, relativePath, relativeName string) {
	fileRoot = m.Root()
	var rootPath string

	switch fileRoot {
	case entity.RootSidecar:
		rootPath = Config().SidecarPath()
	case entity.RootImport:
		rootPath = Config().ImportPath()
	default:
		rootPath = Config().OriginalsPath()
	}

	fileBase = m.BasePrefix(Config().Settings().Index.Sequences)
	relativePath = m.RelPath(rootPath)
	relativeName = m.RelName(rootPath)

	return fileRoot, fileBase, relativePath, relativeName
}

// FileName returns the filename.
func (m *MediaFile) FileName() string {
	return m.fileName
}

// BaseName returns the filename without path.
func (m *MediaFile) BaseName() string {
	return filepath.Base(m.fileName)
}

// SetFileName sets the filename to the given string.
func (m *MediaFile) SetFileName(fileName string) {
	m.fileName = fileName
}

// RelName returns the relative filename.
func (m *MediaFile) RelName(directory string) string {
	return fs.RelName(m.fileName, directory)
}

// RelPath returns the relative path without filename.
func (m *MediaFile) RelPath(directory string) string {
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

	// Remove hidden sub directory if exists.
	if path.Base(pathname) == fs.HiddenPath {
		pathname = path.Dir(pathname)
	}

	// Use empty string for current / root directory.
	if pathname == "." || pathname == "/" {
		pathname = ""
	}

	return pathname
}

// RelPrefix returns the relative path and file name prefix.
func (m *MediaFile) RelPrefix(directory string, stripSequence bool) string {
	if relativePath := m.RelPath(directory); relativePath != "" {
		return filepath.Join(relativePath, m.BasePrefix(stripSequence))
	}

	return m.BasePrefix(stripSequence)
}

// Dir returns the file path.
func (m *MediaFile) Dir() string {
	return filepath.Dir(m.fileName)
}

// SubDir returns a sub directory name.
func (m *MediaFile) SubDir(dir string) string {
	return filepath.Join(filepath.Dir(m.fileName), dir)
}

// BasePrefix returns the filename base without any extensions and path.
func (m *MediaFile) BasePrefix(stripSequence bool) string {
	return fs.BasePrefix(m.FileName(), stripSequence)
}

// Root returns the file root directory.
func (m *MediaFile) Root() string {
	if strings.HasPrefix(m.FileName(), Config().OriginalsPath()) {
		return entity.RootOriginals
	}

	importPath := Config().ImportPath()

	if importPath != "" && strings.HasPrefix(m.FileName(), importPath) {
		return entity.RootImport
	}

	sidecarPath := Config().SidecarPath()

	if sidecarPath != "" && strings.HasPrefix(m.FileName(), sidecarPath) {
		return entity.RootSidecar
	}

	examplesPath := Config().ExamplesPath()

	if examplesPath != "" && strings.HasPrefix(m.FileName(), examplesPath) {
		return entity.RootExamples
	}

	return ""
}

// AbsPrefix returns the directory and base filename without any extensions.
func (m *MediaFile) AbsPrefix(stripSequence bool) string {
	return fs.AbsPrefix(m.FileName(), stripSequence)
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
func (m *MediaFile) Exists() bool {
	return fs.FileExists(m.FileName())
}

// Remove a media file.
func (m *MediaFile) Remove() error {
	return os.Remove(m.FileName())
}

// HasSameName compares a media file with another media file and returns if
// their filenames are matching or not.
func (m *MediaFile) HasSameName(f *MediaFile) bool {
	if f == nil {
		return false
	}

	return m.FileName() == f.FileName()
}

// Move file to a new destination with the filename provided in parameter.
func (m *MediaFile) Move(dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	if err := os.Rename(m.fileName, dest); err != nil {
		log.Debugf("failed renaming file, fallback to copy and delete: %s", err.Error())
	} else {
		m.fileName = dest

		return nil
	}

	if err := m.Copy(dest); err != nil {
		return err
	}

	if err := os.Remove(m.fileName); err != nil {
		return err
	}

	m.fileName = dest

	return nil
}

// Copy a MediaFile to another file by destinationFilename.
func (m *MediaFile) Copy(dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	thisFile, err := m.openFile()

	if err != nil {
		log.Error(err.Error())
		return err
	}

	defer thisFile.Close()

	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, os.ModePerm)

	if err != nil {
		log.Error(err.Error())
		return err
	}

	defer destFile.Close()

	_, err = io.Copy(destFile, thisFile)

	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

// Extension returns the filename extension of this media file.
func (m *MediaFile) Extension() string {
	return strings.ToLower(filepath.Ext(m.fileName))
}

// IsJpeg return true if this media file is a JPEG image.
func (m *MediaFile) IsJpeg() bool {
	// Don't import/use existing thumbnail files (we create our own)
	if m.Extension() == ".thm" {
		return false
	}

	return m.MimeType() == fs.MimeTypeJpeg
}

// IsJson return true if this media file is a json sidecar file.
func (m *MediaFile) IsJson() bool {
	return m.HasFileType(fs.TypeJson)
}

// FileType returns the file type (jpg, gif, tiff,...).
func (m *MediaFile) FileType() fs.FileType {
	if m.IsJpeg() {
		return fs.TypeJpeg
	}

	return fs.GetFileType(m.fileName)
}

// MediaType returns the media type (video, image, raw, sidecar,...).
func (m *MediaFile) MediaType() fs.MediaType {
	return fs.GetMediaType(m.fileName)
}

// HasFileType returns true if this is the given type.
func (m *MediaFile) HasFileType(fileType fs.FileType) bool {
	if fileType == fs.TypeJpeg {
		return m.IsJpeg()
	}

	return m.FileType() == fileType
}

// IsRaw returns true if this is a RAW file.
func (m *MediaFile) IsRaw() bool {
	return m.HasFileType(fs.TypeRaw)
}

// IsPng returns true if this is a PNG file.
func (m *MediaFile) IsPng() bool {
	return m.HasFileType(fs.TypePng)
}

// IsTiff returns true if this is a TIFF file.
func (m *MediaFile) IsTiff() bool {
	return m.HasFileType(fs.TypeTiff)
}

// IsImageOther returns true if this is a PNG, GIF, BMP or TIFF file.
func (m *MediaFile) IsImageOther() bool {
	switch m.FileType() {
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

// IsHEIF returns true if this is a High Efficiency Image File Format file.
func (m *MediaFile) IsHEIF() bool {
	return m.HasFileType(fs.TypeHEIF)
}

// IsXMP returns true if this is a XMP sidecar file.
func (m *MediaFile) IsXMP() bool {
	return m.FileType() == fs.TypeXMP
}

// IsSidecar returns true if this is a sidecar file (containing metadata).
func (m *MediaFile) IsSidecar() bool {
	return m.MediaType() == fs.MediaSidecar
}

// IsVideo returns true if this is a video file.
func (m *MediaFile) IsVideo() bool {
	return m.MediaType() == fs.MediaVideo
}

// IsPlayableVideo returns true if this is a supported video file format.
func (m *MediaFile) IsPlayableVideo() bool {
	return m.MediaType() == fs.MediaVideo && m.HasFileType(fs.TypeMP4)
}

// IsPhoto returns true if this file is a photo / image.
func (m *MediaFile) IsPhoto() bool {
	return m.IsJpeg() || m.IsRaw() || m.IsHEIF() || m.IsImageOther()
}

// IsMedia returns true if this is a media file (photo or video, not sidecar or other).
func (m *MediaFile) IsMedia() bool {
	return m.IsJpeg() || m.IsVideo() || m.IsRaw() || m.IsHEIF() || m.IsImageOther()
}

// Jpeg returns a the JPEG version of the media file (if exists).
func (m *MediaFile) Jpeg() (*MediaFile, error) {
	if m.IsJpeg() {
		if !fs.FileExists(m.FileName()) {
			return nil, fmt.Errorf("jpeg file should exist, but does not: %s", m.FileName())
		}

		return m, nil
	}

	jpegFilename := fs.TypeJpeg.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), false)

	if jpegFilename == "" {
		return nil, fmt.Errorf("no jpeg found for %s", m.FileName())
	}

	return NewMediaFile(jpegFilename)
}

// ContainsJpeg returns true if this file has or is a jpeg media file.
func (m *MediaFile) HasJpeg() bool {
	if m.IsJpeg() {
		return true
	}

	return fs.TypeJpeg.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), false) != ""
}

// HasJson returns true if this file has or is a json sidecar file.
func (m *MediaFile) HasJson() bool {
	if m.IsJson() {
		return true
	}

	return fs.TypeJson.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), false) != ""
}

func (m *MediaFile) decodeDimensions() error {
	if !m.IsMedia() {
		return fmt.Errorf("not a photo: %s", m.FileName())
	}

	var width, height int

	data := m.MetaData()

	if data.Error == nil {
		width = data.Width
		height = data.Height
	}

	if m.IsJpeg() {
		file, err := os.Open(m.FileName())

		if err != nil || file == nil {
			return err
		}

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
	if !m.IsMedia() {
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
	if !m.IsMedia() {
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
func (m *MediaFile) AspectRatio() float32 {
	width := float64(m.Width())
	height := float64(m.Height())

	if width <= 0 || height <= 0 {
		return 0
	}

	aspectRatio := float32(width / height)

	return aspectRatio
}

// Orientation returns the orientation of a MediaFile.
func (m *MediaFile) Orientation() int {
	if data := m.MetaData(); data.Error == nil {
		return data.Orientation
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
		err = fmt.Errorf("mediafile: failed creating thumbnail for %s (%s)", txt.Quote(m.BaseName()), err)
		log.Error(err)
		return "", err
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
			log.Debug(capture.Time(start, fmt.Sprintf("mediafile: no new thumbnails created for %s", m.BasePrefix(false))))
		case 1:
			log.Info(capture.Time(start, fmt.Sprintf("mediafile: one thumbnail created for %s", m.BasePrefix(false))))
		default:
			log.Info(capture.Time(start, fmt.Sprintf("mediafile: %d thumbnails created for %s", count, m.BasePrefix(false))))
		}
	}()

	hash := m.Hash()

	var originalImg image.Image
	var sourceImg image.Image
	var sourceImgType string

	for _, name := range thumb.DefaultTypes {
		thumbType := thumb.Types[name]

		if thumbType.OnDemand() {
			// Skip, size exceeds limit
			continue
		}

		if fileName, err := thumb.Filename(hash, thumbPath, thumbType.Width, thumbType.Height, thumbType.Options...); err != nil {
			log.Errorf("mediafile: failed creating %s (%s)", txt.Quote(name), err)

			return err
		} else {
			if !force && fs.FileExists(fileName) {
				continue
			}

			if originalImg == nil {
				img, err := imaging.Open(m.FileName(), imaging.AutoOrientation(true))

				if err != nil {
					log.Errorf("mediafile: %s in %s", err.Error(), txt.Quote(m.BaseName()))
					return err
				}

				originalImg = img
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
				log.Errorf("mediafile: failed creating %s (%s)", txt.Quote(name), err)
				return err
			}

			count++
		}
	}

	return nil
}
