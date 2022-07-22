package photoprism

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/photoprism/photoprism/pkg/media"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/djherbis/times"
	"github.com/mandykoh/prism/meta/autometa"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// MediaFile represents a single photo, video or sidecar file.
type MediaFile struct {
	fileName       string
	fileRoot       string
	statErr        error
	modTime        time.Time
	fileSize       int64
	fileType       fs.Type
	mimeType       string
	takenAt        time.Time
	takenAtSrc     string
	hash           string
	checksum       string
	hasJpeg        bool
	noColorProfile bool
	colorProfile   string
	width          int
	height         int
	metaData       meta.Data
	metaOnce       sync.Once
	fileMutex      sync.Mutex
	location       *entity.Cell
	imageConfig    *image.Config
}

// NewMediaFile returns a new media file.
func NewMediaFile(fileName string) (m *MediaFile, err error) {
	// Create struct.
	m = &MediaFile{
		fileName: fileName,
		fileRoot: entity.RootUnknown,
		fileType: fs.UnknownType,
		metaData: meta.New(),
		width:    -1,
		height:   -1,
	}

	// Check if file exists and is not empty.
	if size, _, err := m.Stat(); err != nil {
		return m, fmt.Errorf("%s not found", clean.Log(m.RootRelName()))
	} else if size == 0 {
		log.Infof("media: %s is empty", clean.Log(m.RootRelName()))
	}

	return m, nil
}

// Ok checks if the file has a name, exists and is not empty.
func (m *MediaFile) Ok() bool {
	return m.FileName() != "" && m.statErr == nil && !m.Empty()
}

// Empty checks if the file is empty.
func (m *MediaFile) Empty() bool {
	return m.FileSize() <= 0
}

// Stat returns the media file size and modification time rounded to seconds
func (m *MediaFile) Stat() (size int64, mod time.Time, err error) {
	if m.fileSize > 0 {
		return m.fileSize, m.modTime, m.statErr
	}

	fileName := m.FileName()

	// Resolve symlinks.
	if fileName, err = fs.Resolve(fileName); err != nil {
		m.statErr = err
		m.modTime = time.Time{}
		m.fileSize = -1
	} else if s, err := os.Stat(fileName); err != nil {
		m.statErr = err
		m.modTime = time.Time{}
		m.fileSize = -1
	} else {
		s.Mode()
		m.statErr = nil
		m.modTime = s.ModTime().UTC().Truncate(time.Second)
		m.fileSize = s.Size()
	}

	return m.fileSize, m.modTime, m.statErr
}

// ModTime returns the file modification time.
func (m *MediaFile) ModTime() time.Time {
	_, modTime, _ := m.Stat()

	return modTime
}

// FileSize returns the file size in bytes.
func (m *MediaFile) FileSize() int64 {
	fileSize, _, _ := m.Stat()

	return fileSize
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

		log.Infof("media: %s was taken at %s (%s)", clean.Log(filepath.Base(m.fileName)), m.takenAt.String(), m.takenAtSrc)

		return m.takenAt, m.takenAtSrc
	}

	if nameTime := txt.DateFromFilePath(m.fileName); !nameTime.IsZero() {
		m.takenAt = nameTime
		m.takenAtSrc = entity.SrcName

		log.Infof("media: %s was taken at %s (%s)", clean.Log(filepath.Base(m.fileName)), m.takenAt.String(), m.takenAtSrc)

		return m.takenAt, m.takenAtSrc
	}

	m.takenAtSrc = entity.SrcAuto

	fileInfo, err := times.Stat(m.FileName())

	if err != nil {
		log.Warnf("media: %s (file stat)", err.Error())
		log.Infof("media: %s was taken at %s (now)", clean.Log(filepath.Base(m.fileName)), m.takenAt.String())

		return m.takenAt, m.takenAtSrc
	}

	if fileInfo.HasBirthTime() {
		m.takenAt = fileInfo.BirthTime().UTC()
		log.Infof("media: %s was taken at %s (file birth time)", clean.Log(filepath.Base(m.fileName)), m.takenAt.String())
	} else {
		m.takenAt = fileInfo.ModTime().UTC()
		log.Infof("media: %s was taken at %s (file mod time)", clean.Log(filepath.Base(m.fileName)), m.takenAt.String())
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

// RelatedFiles returns files which are related to this file.
func (m *MediaFile) RelatedFiles(stripSequence bool) (result RelatedFiles, err error) {
	// File path and name without any extensions.
	prefix := m.AbsPrefix(stripSequence)

	// Storage folder path prefixes.
	sidecarPrefix := Config().SidecarPath() + "/"
	originalsPrefix := Config().OriginalsPath() + "/"

	// Ignore RAW images?
	skipRaw := Config().DisableRaw()

	// Replace sidecar with originals path in search prefix.
	if len(sidecarPrefix) > 1 && sidecarPrefix != originalsPrefix && strings.HasPrefix(prefix, sidecarPrefix) {
		prefix = strings.Replace(prefix, sidecarPrefix, originalsPrefix, 1)
		log.Debugf("media: replaced sidecar with originals path in related file matching pattern")
	}

	// Quote path for glob.
	if stripSequence {
		// Strip common name sequences like "copy 2" and escape meta characters.
		prefix = regexp.QuoteMeta(prefix)
	} else {
		// Use strict file name matching and escape meta characters.
		prefix = regexp.QuoteMeta(prefix + ".")
	}

	// Find related files.
	matches, err := filepath.Glob(prefix + "*")

	if err != nil {
		return result, err
	}

	if name := m.EditedName(); name != "" {
		matches = append(matches, name)
	}

	isHEIF := false

	for _, fileName := range matches {
		f, fileErr := NewMediaFile(fileName)

		if fileErr != nil || f.Empty() {
			continue
		}

		// Ignore RAW images?
		if f.IsRaw() && skipRaw {
			log.Debugf("media: skipped related raw file %s", clean.Log(f.RootRelName()))
			continue
		}

		if result.Main == nil && f.IsJpeg() {
			result.Main = f
		} else if f.IsRaw() {
			result.Main = f
		} else if f.IsHEIF() {
			isHEIF = true
			result.Main = f
		} else if f.IsImageOther() {
			result.Main = f
		} else if f.IsVideo() && !isHEIF {
			result.Main = f
		} else if result.Main != nil && f.IsJpeg() {
			if result.Main.IsJpeg() && len(result.Main.FileName()) > len(f.FileName()) {
				result.Main = f
			}
		}

		result.Files = append(result.Files, f)
	}

	if len(result.Files) == 0 || result.Main == nil {
		t := m.MimeType()

		if t == "" {
			t = "unknown type"
		}

		return result, fmt.Errorf("no supported files found for %s (%s)", clean.Log(m.BaseName()), t)
	}

	// Add hidden JPEG if exists.
	if !result.ContainsJpeg() {
		if jpegName := fs.ImageJPEG.FindFirst(result.Main.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), stripSequence); jpegName != "" {
			if resultFile, _ := NewMediaFile(jpegName); resultFile.Ok() {
				result.Files = append(result.Files, resultFile)
			}
		}
	}

	sort.Sort(result.Files)

	return result, nil
}

// PathNameInfo returns file name infos for indexing.
func (m *MediaFile) PathNameInfo(stripSequence bool) (fileRoot, fileBase, relativePath, relativeName string) {
	fileRoot = m.Root()

	var rootPath string

	switch fileRoot {
	case entity.RootSidecar:
		rootPath = Config().SidecarPath()
	case entity.RootImport:
		rootPath = Config().ImportPath()
	case entity.RootExamples:
		rootPath = Config().ExamplesPath()
	case entity.RootOriginals:
		rootPath = Config().OriginalsPath()
	default:
		rootPath = Config().OriginalsPath()
	}

	fileBase = m.BasePrefix(stripSequence)
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
	m.fileRoot = entity.RootUnknown
}

// RootRelName returns the relative filename, and automatically detects the root path.
func (m *MediaFile) RootRelName() string {
	return m.RelName(m.RootPath())
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
	if pathname == "." || pathname == "/" || pathname == "\\" {
		pathname = ""
	}

	return pathname
}

// RootPath returns the file root path based on the configuration.
func (m *MediaFile) RootPath() string {
	switch m.Root() {
	case entity.RootSidecar:
		return Config().SidecarPath()
	case entity.RootImport:
		return Config().ImportPath()
	case entity.RootExamples:
		return Config().ExamplesPath()
	default:
		return Config().OriginalsPath()
	}
}

// RootRelPath returns the relative path and automatically detects the root path.
func (m *MediaFile) RootRelPath() string {
	return m.RelPath(m.RootPath())
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
	if m.fileRoot != entity.RootUnknown {
		return m.fileRoot
	}

	if strings.HasPrefix(m.FileName(), Config().OriginalsPath()) {
		m.fileRoot = entity.RootOriginals
		return m.fileRoot
	}

	importPath := Config().ImportPath()

	if importPath != "" && strings.HasPrefix(m.FileName(), importPath) {
		m.fileRoot = entity.RootImport
		return m.fileRoot
	}

	sidecarPath := Config().SidecarPath()

	if sidecarPath != "" && strings.HasPrefix(m.FileName(), sidecarPath) {
		m.fileRoot = entity.RootSidecar
		return m.fileRoot
	}

	examplesPath := Config().ExamplesPath()

	if examplesPath != "" && strings.HasPrefix(m.FileName(), examplesPath) {
		m.fileRoot = entity.RootExamples
		return m.fileRoot
	}

	return m.fileRoot
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

	var err error
	fileName := m.FileName()

	// Resolve symlinks.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return m.mimeType
	}

	m.mimeType = fs.MimeType(fileName)

	return m.mimeType
}

// openFile opens the file and returns the descriptor.
func (m *MediaFile) openFile() (handle *os.File, err error) {
	fileName := m.FileName()

	// Resolve symlinks.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return nil, fmt.Errorf("%s %s", err, clean.Log(m.RootRelName()))
	}

	handle, err = os.Open(fileName)

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

// Remove permanently removes a media file.
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
		m.SetFileName(dest)

		return nil
	}

	if err := m.Copy(dest); err != nil {
		return err
	}

	if err := os.Remove(m.fileName); err != nil {
		return err
	}

	m.SetFileName(dest)

	return nil
}

// Copy a MediaFile to another file by destinationFilename.
func (m *MediaFile) Copy(dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	m.fileMutex.Lock()
	defer m.fileMutex.Unlock()

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

// IsPng returns true if this is a PNG image.
func (m *MediaFile) IsPng() bool {
	return m.MimeType() == fs.MimeTypePng
}

// IsGif returns true if this is a GIF image.
func (m *MediaFile) IsGif() bool {
	return m.MimeType() == fs.MimeTypeGif
}

// IsTiff returns true if this is a TIFF image.
func (m *MediaFile) IsTiff() bool {
	return m.HasFileType(fs.ImageTIFF) && m.MimeType() == fs.MimeTypeTiff
}

// IsHEIF returns true if this is a High Efficiency Image File Format image.
func (m *MediaFile) IsHEIF() bool {
	return m.MimeType() == fs.MimeTypeHEIF
}

// IsBitmap returns true if this is a bitmap image.
func (m *MediaFile) IsBitmap() bool {
	return m.MimeType() == fs.MimeTypeBitmap
}

// IsWebP returns true if this is a WebP image file.
func (m *MediaFile) IsWebP() bool {
	return m.MimeType() == fs.MimeTypeWebP
}

// IsVideo returns true if this is a video file.
func (m *MediaFile) IsVideo() bool {
	return strings.HasPrefix(m.MimeType(), "video/") || m.Media() == media.Video
}

// IsAnimatedGif returns true if it is an animated GIF.
func (m *MediaFile) IsAnimatedGif() bool {
	return m.IsGif() && m.MetaData().Frames > 1
}

// IsAnimated returns true if it is a video or animated image.
func (m *MediaFile) IsAnimated() bool {
	return m.IsVideo() || m.IsAnimatedGif()
}

// IsJson return true if this media file is a json sidecar file.
func (m *MediaFile) IsJson() bool {
	return m.HasFileType(fs.JsonFile)
}

// FileType returns the file type (jpg, gif, tiff,...).
func (m *MediaFile) FileType() fs.Type {
	switch {
	case m.IsJpeg():
		return fs.ImageJPEG
	case m.IsPng():
		return fs.ImagePNG
	case m.IsGif():
		return fs.ImageGIF
	case m.IsHEIF():
		return fs.ImageHEIF
	case m.IsBitmap():
		return fs.ImageBMP
	default:
		return fs.FileType(m.fileName)
	}
}

// Media returns the media content type (video, image, raw, sidecar,...).
func (m *MediaFile) Media() media.Type {
	return media.FromName(m.fileName)
}

// HasFileType returns true if this is the given type.
func (m *MediaFile) HasFileType(fileType fs.Type) bool {
	if fileType == fs.ImageJPEG {
		return m.IsJpeg()
	}

	return m.FileType() == fileType
}

// IsRaw returns true if this is a RAW file.
func (m *MediaFile) IsRaw() bool {
	return m.HasFileType(fs.RawImage)
}

// IsXMP returns true if this is a XMP sidecar file.
func (m *MediaFile) IsXMP() bool {
	return m.FileType() == fs.XmpFile
}

// InOriginals checks if the file is stored in the 'originals' folder.
func (m *MediaFile) InOriginals() bool {
	return m.Root() == entity.RootOriginals
}

// InSidecar checks if the file is stored in the 'sidecar' folder.
func (m *MediaFile) InSidecar() bool {
	return m.Root() == entity.RootSidecar
}

// IsSidecar checks if the file is a metadata sidecar file, independent of the storage location.
func (m *MediaFile) IsSidecar() bool {
	return m.Media() == media.Sidecar
}

// IsPlayableVideo checks if the file is a video in playable format.
func (m *MediaFile) IsPlayableVideo() bool {
	return m.IsVideo() && (m.HasFileType(fs.VideoMP4) || m.HasFileType(fs.VideoAVC))
}

// IsImageOther returns true if this is a PNG, GIF, BMP, TIFF, or WebP file.
func (m *MediaFile) IsImageOther() bool {
	switch {
	case m.IsPng(), m.IsGif(), m.IsTiff(), m.IsBitmap(), m.IsWebP():
		return true
	default:
		return false
	}
}

// IsImageNative returns true if it is a natively supported image file.
func (m *MediaFile) IsImageNative() bool {
	return m.IsJpeg() || m.IsImageOther()
}

// IsImage checks if the file is an image
func (m *MediaFile) IsImage() bool {
	return m.IsImageNative() || m.IsRaw() || m.IsHEIF()
}

// IsLive checks if the file is a live photo.
func (m *MediaFile) IsLive() bool {
	if m.IsHEIF() {
		return fs.VideoMOV.FindFirst(m.FileName(), []string{}, Config().OriginalsPath(), false) != ""
	}

	if m.IsVideo() {
		return fs.ImageHEIF.FindFirst(m.FileName(), []string{}, Config().OriginalsPath(), false) != ""
	}

	return false
}

// ExifSupported returns true if parsing exif metadata is supported for the media file type.
func (m *MediaFile) ExifSupported() bool {
	return m.IsJpeg() || m.IsRaw() || m.IsHEIF() || m.IsPng() || m.IsTiff()
}

// IsMedia returns true if this is a media file (photo or video, not sidecar or other).
func (m *MediaFile) IsMedia() bool {
	return m.IsJpeg() || m.IsVideo() || m.IsRaw() || m.IsHEIF() || m.IsImageOther()
}

// Jpeg returns the JPEG version of the media file (if exists).
func (m *MediaFile) Jpeg() (*MediaFile, error) {
	if m.IsJpeg() {
		if !fs.FileExists(m.FileName()) {
			return nil, fmt.Errorf("jpeg should exist, but does not: %s", m.RootRelName())
		}

		return m, nil
	} else if m.Empty() {
		return nil, fmt.Errorf("%s is empty", m.RootRelName())
	}

	jpegFilename := fs.ImageJPEG.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), false)

	if jpegFilename == "" {
		return nil, fmt.Errorf("no jpeg found for %s", m.RootRelName())
	}

	return NewMediaFile(jpegFilename)
}

// HasJpeg returns true if the file has or is a jpeg media file.
func (m *MediaFile) HasJpeg() bool {
	if m.hasJpeg {
		return true
	}

	if m.IsJpeg() {
		m.hasJpeg = true
		return true
	}

	jpegName := fs.ImageJPEG.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.HiddenPath}, Config().OriginalsPath(), false)

	if jpegName == "" {
		m.hasJpeg = false
	} else {
		m.hasJpeg = fs.MimeType(jpegName) == fs.MimeTypeJpeg
	}

	return m.hasJpeg
}

func (m *MediaFile) decodeDimensions() error {
	// Media dimensions already known?
	if m.width > 0 && m.height > 0 {
		return nil
	}

	// Valid media file?
	if !m.Ok() || !m.IsMedia() {
		return fmt.Errorf("%s is not a valid media file", clean.Log(m.Extension()))
	}

	// Extract the actual width and height from natively supported formats.
	if m.IsImageNative() {
		cfg, err := m.DecodeConfig()

		if err != nil {
			return err
		}

		orientation := m.Orientation()

		if orientation > 4 && orientation <= 8 {
			m.width = cfg.Height
			m.height = cfg.Width
		} else {
			m.width = cfg.Width
			m.height = cfg.Height
		}

		return nil
	}

	// Extract the width and height from metadata for other formats.
	if data := m.MetaData(); data.Error != nil {
		return data.Error
	} else {
		m.width = data.ActualWidth()
		m.height = data.ActualHeight()

		return nil
	}
}

// DecodeConfig extracts the raw dimensions from the header of natively supported image file formats.
func (m *MediaFile) DecodeConfig() (_ *image.Config, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic %s while decoding %s dimensions\nstack: %s", r, clean.Log(m.Extension()), debug.Stack())
		}
	}()

	if m.imageConfig != nil {
		return m.imageConfig, nil
	}

	if !m.IsImageNative() {
		return nil, fmt.Errorf("%s not supported natively", clean.Log(m.Extension()))
	}

	m.fileMutex.Lock()
	defer m.fileMutex.Unlock()

	fileName := m.FileName()

	// Resolve symlinks.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return nil, fmt.Errorf("%s %s", err, clean.Log(m.RootRelName()))
	}

	file, err := os.Open(fileName)

	if err != nil || file == nil {
		return nil, err
	}

	defer file.Close()

	// Reset file offset.
	// see https://github.com/golang/go/issues/45902#issuecomment-1007953723
	_, err = file.Seek(0, 0)

	if err != nil {
		return nil, fmt.Errorf("%s on seek", err)
	}

	// Decode image config (dimensions).
	cfg, _, err := image.DecodeConfig(file)

	if err != nil {
		return nil, fmt.Errorf("%s while decoding", err)
	}

	m.imageConfig = &cfg

	return m.imageConfig, nil
}

// Width return the width dimension of a MediaFile.
func (m *MediaFile) Width() int {
	// Valid media file?
	if !m.Ok() || !m.IsMedia() {
		return 0
	}

	if m.width < 0 {
		if err := m.decodeDimensions(); err != nil {
			log.Debugf("media: %s", err)
		}
	}

	return m.width
}

// Height returns the height dimension of a MediaFile.
func (m *MediaFile) Height() int {
	// Valid media file?
	if !m.Ok() || !m.IsMedia() {
		return 0
	}

	if m.height < 0 {
		if err := m.decodeDimensions(); err != nil {
			log.Debugf("media: %s", err)
		}
	}

	return m.height
}

// Megapixels returns the resolution in megapixels if possible.
func (m *MediaFile) Megapixels() (resolution int) {
	// Valid media file?
	if !m.Ok() || !m.IsMedia() {
		return 0
	}

	if cfg, err := m.DecodeConfig(); err == nil {
		resolution = int(math.Round(float64(cfg.Width*cfg.Height) / 1000000))
	}

	if resolution <= 0 {
		resolution = m.metaData.Megapixels()
	}

	return resolution
}

// ExceedsFileSize checks if the file exceeds the configured file size limit in MB.
func (m *MediaFile) ExceedsFileSize(limit int) (exceeds bool, actual int) {
	const mega = 1048576

	if limit <= 0 {
		return false, actual
	} else if size := m.FileSize(); size <= 0 {
		return false, actual
	} else {
		actual = int(size / mega)
		return size > int64(limit)*mega, actual
	}
}

// ExceedsResolution checks if an image in a natively supported format exceeds the configured resolution limit in megapixels.
func (m *MediaFile) ExceedsResolution(limit int) (exceeds bool, actual int) {
	if limit <= 0 {
		return false, actual
	} else if !m.IsImage() {
		return false, actual
	} else if actual = m.Megapixels(); actual <= 0 {
		return false, actual
	} else {
		return actual > limit, actual
	}
}

// AspectRatio returns the aspect ratio of a MediaFile.
func (m *MediaFile) AspectRatio() float32 {
	width := float64(m.Width())
	height := float64(m.Height())

	if width <= 0 || height <= 0 {
		return 0
	}

	aspectRatio := float32(math.Round((width/height)*100) / 100)

	return aspectRatio
}

// Portrait tests if the image is a portrait.
func (m *MediaFile) Portrait() bool {
	return m.Width() < m.Height()
}

// Orientation returns the Exif orientation of the media file.
func (m *MediaFile) Orientation() int {
	if data := m.MetaData(); data.Error == nil {
		return data.Orientation
	}

	return 1
}

// RenameSidecarFiles moves related sidecar files.
func (m *MediaFile) RenameSidecarFiles(oldFileName string) (renamed map[string]string, err error) {
	renamed = make(map[string]string)

	sidecarPath := Config().SidecarPath()
	originalsPath := Config().OriginalsPath()

	newName := m.RelPrefix(originalsPath, false)
	oldPrefix := fs.RelPrefix(oldFileName, originalsPath, false)
	globPrefix := filepath.Join(sidecarPath, oldPrefix) + "."

	matches, err := filepath.Glob(regexp.QuoteMeta(globPrefix) + "*")

	if err != nil {
		return renamed, err
	}

	for _, srcName := range matches {
		destName := filepath.Join(sidecarPath, newName+fs.Ext(srcName))

		if fs.FileExists(destName) {
			renamed[fs.RelName(srcName, sidecarPath)] = fs.RelName(destName, sidecarPath)

			if err := os.Remove(srcName); err != nil {
				log.Errorf("files: failed removing sidecar %s", clean.Log(fs.RelName(srcName, sidecarPath)))
			} else {
				log.Infof("files: removed sidecar %s", clean.Log(fs.RelName(srcName, sidecarPath)))
			}

			continue
		}

		if err := fs.Move(srcName, destName); err != nil {
			return renamed, err
		} else {
			log.Infof("files: moved existing sidecar to %s", clean.Log(newName+filepath.Ext(srcName)))
			renamed[fs.RelName(srcName, sidecarPath)] = fs.RelName(destName, sidecarPath)
		}
	}

	return renamed, nil
}

// RemoveSidecarFiles permanently removes related sidecar files.
func (m *MediaFile) RemoveSidecarFiles() (numFiles int, err error) {
	fileName := m.FileName()

	if fileName == "" {
		return numFiles, fmt.Errorf("empty filename")
	}

	sidecarPath := Config().SidecarPath()
	originalsPath := Config().OriginalsPath()

	prefix := fs.RelPrefix(fileName, originalsPath, false)
	globPrefix := filepath.Join(sidecarPath, prefix) + "."

	matches, err := filepath.Glob(regexp.QuoteMeta(globPrefix) + "*")

	if err != nil {
		return numFiles, err
	}

	for _, sidecarName := range matches {
		if err = os.Remove(sidecarName); err != nil {
			log.Errorf("files: failed deleting sidecar %s", clean.Log(fs.RelName(sidecarName, sidecarPath)))
		} else {
			numFiles++
			log.Infof("files: deleted sidecar %s", clean.Log(fs.RelName(sidecarName, sidecarPath)))
		}
	}

	return numFiles, nil
}

// ColorProfile returns the ICC color profile name.
func (m *MediaFile) ColorProfile() string {
	if !m.IsJpeg() || m.colorProfile != "" || m.noColorProfile {
		return m.colorProfile
	}

	start := time.Now()
	logName := clean.Log(m.BaseName())

	m.fileMutex.Lock()
	defer m.fileMutex.Unlock()

	var err error
	fileName := m.FileName()

	// Resolve symlinks.
	if fileName, err = fs.Resolve(fileName); err != nil {
		return m.colorProfile
	}

	// Open file.
	fileReader, err := os.Open(fileName)

	if err != nil {
		m.noColorProfile = true
		return ""
	}

	defer fileReader.Close()

	// Reset file offset.
	// see https://github.com/golang/go/issues/45902#issuecomment-1007953723
	_, err = fileReader.Seek(0, 0)

	if err != nil {
		log.Warnf("media: %s in %s on seek [%s]", err, logName, time.Since(start))
		return ""
	}

	// Read color metadata.
	md, _, err := autometa.Load(fileReader)

	if err != nil || md == nil {
		m.noColorProfile = true
		return ""
	}

	// Read ICC profile and convert colors if possible.
	if iccProfile, err := md.ICCProfile(); err != nil || iccProfile == nil {
		// Do nothing.
	} else if profile, err := iccProfile.Description(); err == nil && profile != "" {
		log.Debugf("media: %s has color profile %s [%s]", logName, clean.Log(profile), time.Since(start))
		m.colorProfile = profile
		return m.colorProfile
	}

	log.Tracef("media: %s has no color profile [%s]", logName, time.Since(start))
	m.noColorProfile = true
	return ""
}
