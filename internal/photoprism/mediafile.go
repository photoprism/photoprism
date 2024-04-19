package photoprism

import (
	"errors"
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
	"strings"
	"sync"
	"time"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/djherbis/times"
	"github.com/dustin/go-humanize"
	"github.com/mandykoh/prism/meta/autometa"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/meta"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/txt"
	"github.com/photoprism/photoprism/pkg/video"
)

// MediaFile represents a single photo, video, sidecar, or other supported media file.
type MediaFile struct {
	fileName         string
	fileNameResolved string
	fileRoot         string
	statErr          error
	modTime          time.Time
	fileSize         int64
	fileType         fs.Type
	mimeType         string
	takenAt          time.Time
	takenAtSrc       string
	hash             string
	checksum         string
	hasPreviewImage  bool
	noColorProfile   bool
	colorProfile     string
	width            int
	height           int
	metaData         meta.Data
	metaOnce         sync.Once
	videoInfo        video.Info
	videoOnce        sync.Once
	fileMutex        sync.Mutex
	location         *entity.Cell
	imageConfig      *image.Config
}

// NewMediaFile returns a new media file and automatically resolves any symlinks.
func NewMediaFile(fileName string) (*MediaFile, error) {
	if fileNameResolved, err := fs.Resolve(fileName); err != nil {
		// Don't return nil on error, as this would change the previous behavior.
		return &MediaFile{}, err
	} else {
		return NewMediaFileSkipResolve(fileName, fileNameResolved)
	}
}

// NewMediaFileSkipResolve returns a new media file without resolving symlinks.
// This is useful because if it is known that the filename is fully resolved, it is much faster.
func NewMediaFileSkipResolve(fileName string, fileNameResolved string) (*MediaFile, error) {
	// Create and initialize the new media file.
	m := &MediaFile{
		fileName:         fileName,
		fileNameResolved: fileNameResolved,
		fileRoot:         entity.RootUnknown,
		fileType:         fs.TypeUnknown,
		metaData:         meta.NewData(),
		videoInfo:        video.NewInfo(),
		width:            -1,
		height:           -1,
	}

	// Check if the file exists and is not empty.
	if size, _, err := m.Stat(); err != nil {
		// Return error if os.Stat() failed.
		return m, fmt.Errorf("%s not found", clean.Log(m.RootRelName()))
	} else if size == 0 {
		// Notify the user that the file is empty.
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

// Stat calls os.Stat() to return the file size and modification time,
// or an error if this failed.
func (m *MediaFile) Stat() (size int64, mod time.Time, err error) {
	if m.fileSize > 0 {
		return m.fileSize, m.modTime, m.statErr
	}

	if stat, statErr := os.Stat(m.fileNameResolved); statErr != nil {
		m.statErr = statErr
		m.modTime = time.Time{}
		m.fileSize = -1
	} else {
		stat.Mode()
		m.statErr = nil
		m.modTime = stat.ModTime().UTC().Truncate(time.Second)
		m.fileSize = stat.Size()
	}

	return m.fileSize, m.modTime, m.statErr
}

// ModTime returns the file modification time.
func (m *MediaFile) ModTime() time.Time {
	_, modTime, _ := m.Stat()

	return modTime
}

// SetModTime sets the file modification time.
func (m *MediaFile) SetModTime(modTime time.Time) *MediaFile {
	modTime = modTime.UTC()

	if err := os.Chtimes(m.FileName(), time.Time{}, modTime); err != nil {
		log.Debugf("media: failed to set mtime for %s (%s)", clean.Log(m.RootRelName()), clean.Error(err))
	} else {
		m.modTime = modTime
	}

	return m
}

// FileSize returns the file size in bytes.
func (m *MediaFile) FileSize() int64 {
	fileSize, _, _ := m.Stat()

	return fileSize
}

// DateCreated returns the media creation time in UTC.
func (m *MediaFile) DateCreated() time.Time {
	takenAt, _ := m.TakenAt()

	return takenAt
}

// TakenAt returns the media creation time in UTC and the source from which it originates.
func (m *MediaFile) TakenAt() (time.Time, string) {
	// Check if creation time has been cached.
	if !m.takenAt.IsZero() {
		return m.takenAt, m.takenAtSrc
	}

	m.takenAt = time.Now().UTC()

	// First try to extract the creation time from the file metadata,
	data := m.MetaData()

	if data.Error == nil && !data.TakenAt.IsZero() && data.TakenAt.Year() > 1000 {
		m.takenAt = data.TakenAt.UTC()
		m.takenAtSrc = entity.SrcMeta

		log.Infof("media: %s was taken at %s (%s)", clean.Log(filepath.Base(m.fileName)), m.takenAt.String(), m.takenAtSrc)

		return m.takenAt, m.takenAtSrc
	}

	// Otherwiese, try to determine creation time from file name and path.
	if nameTime := txt.DateFromFilePath(m.fileName); !nameTime.IsZero() {
		m.takenAt = nameTime
		m.takenAtSrc = entity.SrcName

		log.Infof("media: %s was taken at %s (%s)", clean.Log(filepath.Base(m.fileName)), m.takenAt.String(), m.takenAtSrc)

		return m.takenAt, m.takenAtSrc
	}

	m.takenAtSrc = entity.SrcAuto

	fileInfo, err := times.Stat(m.FileName())

	if err != nil {
		log.Warnf("media: %s (stat call failed)", err.Error())
		log.Infof("media: %s was taken at %s (unknown mod time)", clean.Log(filepath.Base(m.fileName)), m.takenAt.String())

		return m.takenAt, m.takenAtSrc
	}

	// Use file modification time as fallback.
	m.takenAt = fileInfo.ModTime().UTC()

	log.Infof("media: %s was taken at %s (file mod time)", clean.Log(filepath.Base(m.fileName)), m.takenAt.String())

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
	if m == nil {
		log.Errorf("media: file %s is nil - you may have found a bug", clean.Log(fileName))
		return
	}

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
	if path.Base(pathname) == fs.PPHiddenPathname {
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

// AbsPrefix returns the directory and base filename without any extensions.
func (m *MediaFile) AbsPrefix(stripSequence bool) string {
	return fs.AbsPrefix(m.FileName(), stripSequence)
}

// BasePrefix returns the filename base without any extensions and path.
func (m *MediaFile) BasePrefix(stripSequence bool) string {
	return fs.BasePrefix(m.FileName(), stripSequence)
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
	destName := filepath.Base(dest)
	destDir := filepath.Dir(dest)

	// Check destination filename and create path if it does not exist yet.
	if destName == "" {
		return errors.New("move: invalid destination filename")
	} else if destDir == "" {
		return errors.New("move: invalid destination path")
	} else if err := fs.MkdirAll(destDir); err != nil {
		return err
	}

	// Remember file modification time.
	modTime := m.ModTime()

	// First try to rename existing file as that's faster than copying it and then deleting the original.
	if err := os.Rename(m.fileName, dest); err != nil {
		log.Tracef("move: cannot rename %s, fallback to copy and delete (%s)", clean.Log(destName), clean.Error(err))
	} else {
		m.SetFileName(dest)
		m.SetModTime(modTime)

		return nil
	}

	// If renaming is not possible, copy the file and then delete the original.
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
	destName := filepath.Base(dest)
	destDir := filepath.Dir(dest)

	// Check destination filename and create path if it does not exist yet.
	if destName == "" {
		return errors.New("copy: invalid destination filename")
	} else if destDir == "" {
		return errors.New("copy: invalid destination path")
	} else if err := fs.MkdirAll(destDir); err != nil {
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

	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, fs.ModeFile)

	if err != nil {
		log.Error(err.Error())
		return err
	}

	defer func() {
		if err = destFile.Close(); err != nil {
			log.Debugf("copy: failed to close %s (%s)", clean.Log(destName), clean.Error(err))
		} else if err = os.Chtimes(dest, time.Time{}, m.ModTime()); err != nil {
			log.Debugf("copy: failed to set mtime for %s (%s)", clean.Log(destName), clean.Error(err))
		}
	}()

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

// IsPreviewImage return true if this media file is a JPEG or PNG image.
func (m *MediaFile) IsPreviewImage() bool {
	return m.IsJpeg() || m.IsPNG()
}

// IsJpeg checks if the file is a JPEG image with a supported file type extension.
func (m *MediaFile) IsJpeg() bool {
	if m.Extension() == fs.ExtTHM {
		// Ignore .thm files, as some cameras automatically
		// create them as thumbnails.
		return false
	} else if fs.FileType(m.fileName) != fs.ImageJPEG {
		// Files with an incorrect file extension are no longer
		// recognized as JPEG to improve indexing performance.
		return false
	}

	// Check the mime type after other tests have passed to improve performance.
	return m.MimeType() == fs.MimeTypeJPEG
}

// IsJpegXL checks if the file is a JPEG XL image with a supported file type extension.
func (m *MediaFile) IsJpegXL() bool {
	if fs.FileType(m.fileName) != fs.ImageJPEGXL {
		return false
	}

	// Check the mime type after other tests have passed to improve performance.
	return m.MimeType() == fs.MimeTypeJPEGXL
}

// IsPNG checks if the file is a PNG image with a supported file type extension.
func (m *MediaFile) IsPNG() bool {
	if fs.FileType(m.fileName) != fs.ImagePNG {
		// Files with an incorrect file extension are no longer
		// recognized as PNG to improve indexing performance.
		return false
	}

	// Check the mime type after other tests have passed to improve performance.
	mimeType := m.MimeType()
	return mimeType == fs.MimeTypePNG || mimeType == fs.MimeTypeAPNG
}

// IsGIF checks if the file is a GIF image with a supported file type extension.
func (m *MediaFile) IsGIF() bool {
	if fs.FileType(m.fileName) != fs.ImageGIF {
		return false
	}

	// Check the mime type after other tests have passed to improve performance.
	return m.MimeType() == fs.MimeTypeGIF
}

// IsTIFF checks if the file is a TIFF image with a supported file type extension.
func (m *MediaFile) IsTIFF() bool {
	if fs.FileType(m.fileName) != fs.ImageTIFF {
		return false
	}

	// Check the mime type after other tests have passed to improve performance.
	return m.MimeType() == fs.MimeTypeTIFF
}

// IsDNG checks if the file is a Adobe Digital Negative (DNG) image with a supported file type extension.
func (m *MediaFile) IsDNG() bool {
	if fs.FileType(m.fileName) != fs.ImageDNG {
		return false
	}

	return m.MimeType() == fs.MimeTypeDNG
}

// IsHEIF checks if the file is a High Efficiency Image File Format (HEIF) container with a supported file type extension.
func (m *MediaFile) IsHEIF() bool {
	return m.IsHEIC() || m.IsHEICS() || m.IsAVIF() || m.IsAVIFS()
}

// IsHEIC checks if the file is a High Efficiency Image Container (HEIC) image with a supported file type extension.
func (m *MediaFile) IsHEIC() bool {
	if t := fs.FileType(m.fileName); t != fs.ImageHEIF && t != fs.ImageHEIC {
		return false
	}

	// Check the mime type after other tests have passed to improve performance.
	mimeType := m.MimeType()
	return mimeType == fs.MimeTypeHEIC || mimeType == fs.MimeTypeHEICS
}

// IsHEICS checks if the file is a HEIC image sequence with a supported file type extension.
func (m *MediaFile) IsHEICS() bool {
	return m.HasFileType(fs.ImageHEICS)
}

// IsAVIF checks if the file is an AV1 Image File Format image with a supported file type extension.
func (m *MediaFile) IsAVIF() bool {
	if t := fs.FileType(m.fileName); t != fs.ImageAVIF {
		return false
	}

	return m.MimeType() == fs.MimeTypeAVIF
}

// IsAVIFS checks if the file is an AVIF image sequence with a supported file type extension.
func (m *MediaFile) IsAVIFS() bool {
	return m.HasFileType(fs.ImageAVIFS)
}

// IsBMP checks if the file is a bitmap image with a supported file type extension.
func (m *MediaFile) IsBMP() bool {
	if fs.FileType(m.fileName) != fs.ImageBMP {
		return false
	}

	// Check the mime type after other tests have passed to improve performance.
	return m.MimeType() == fs.MimeTypeBMP
}

// IsWebP checks if the file is a WebP image file with a supported file type extension.
func (m *MediaFile) IsWebP() bool {
	if fs.FileType(m.fileName) != fs.ImageWebP {
		return false
	}

	return m.MimeType() == fs.MimeTypeWebP
}

// Duration returns the duration is the media content is playable.
func (m *MediaFile) Duration() time.Duration {
	return m.MetaData().Duration
}

// IsAnimatedImage checks if the file is an animated image.
func (m *MediaFile) IsAnimatedImage() bool {
	return fs.IsAnimatedImage(m.fileName) && (m.MetaData().Frames > 1 || m.MetaData().Duration > 0)
}

// IsJSON checks if the file is a JSON sidecar file with a supported file type extension.
func (m *MediaFile) IsJSON() bool {
	return m.HasFileType(fs.SidecarJSON)
}

// FileType returns the file type (jpg, gif, tiff,...).
func (m *MediaFile) FileType() fs.Type {
	switch {
	case m.IsJpeg():
		return fs.ImageJPEG
	case m.IsPNG():
		return fs.ImagePNG
	case m.IsGIF():
		return fs.ImageGIF
	case m.IsBMP():
		return fs.ImageBMP
	case m.IsDNG():
		return fs.ImageDNG
	case m.IsAVIF():
		return fs.ImageAVIF
	case m.IsHEIC():
		return fs.ImageHEIC
	default:
		return fs.FileType(m.fileName)
	}
}

// CheckType returns an error if the file extension is missing or invalid,
// see https://github.com/photoprism/photoprism/issues/3518 for details.
func (m *MediaFile) CheckType() error {
	// Get extension and return error if missing.
	extension := m.Extension()

	if extension == "" {
		return fmt.Errorf("missing extension")
	}

	// Detect file type and return error if unknown.
	fileType := fs.FileType(m.fileName)

	if fileType == fs.TypeUnknown {
		return fmt.Errorf("unknown file type")
	}

	// Detect media type (formerly known as a MIME type),
	// see https://en.wikipedia.org/wiki/Media_type
	mimeType := m.MimeType()

	// Perform mime type checks for selected file types.
	var valid bool
	switch fileType {
	case fs.ImageJPEG:
		valid = mimeType == fs.MimeTypeJPEG
	case fs.ImagePNG:
		valid = mimeType == fs.MimeTypePNG || mimeType == fs.MimeTypeAPNG
	case fs.ImageGIF:
		valid = mimeType == fs.MimeTypeGIF
	case fs.ImageTIFF:
		valid = mimeType == fs.MimeTypeTIFF
	case fs.ImageHEIC, fs.ImageHEIF:
		valid = mimeType == fs.MimeTypeHEIC || mimeType == fs.MimeTypeHEICS
	default:
		// Skip mime type check. Note: Checks for additional formats and/or generic
		// checks based on the media content type can be added over time as needed.
		return nil
	}

	// Ok?
	if valid {
		return nil
	}

	// Exclude mime type from the error message if it could not be detected.
	if mimeType == fs.MimeTypeUnknown {
		return fmt.Errorf("invalid extension (unknown media type)")
	}

	return fmt.Errorf("invalid extension for media type %s", clean.LogQuote(mimeType))
}

// Media returns the media content type (video, image, raw, sidecar,...).
func (m *MediaFile) Media() media.Type {
	return media.FromName(m.fileName)
}

// HasMediaType checks if the file has is the given media type.
func (m *MediaFile) HasMediaType(mediaType media.Type) bool {
	return m.Media() == mediaType
}

// HasFileType checks if the file has the given file type.
func (m *MediaFile) HasFileType(fileType fs.Type) bool {
	if fileType == fs.ImageJPEG {
		return m.IsJpeg()
	}

	return m.FileType() == fileType
}

// IsImage checks if the file is an image.
func (m *MediaFile) IsImage() bool {
	return m.HasMediaType(media.Image)
}

// IsRaw returns true if this is a RAW file.
func (m *MediaFile) IsRaw() bool {
	return m.HasFileType(fs.ImageRaw) || m.HasMediaType(media.Raw) || m.IsDNG()
}

// IsAnimated returns true if it is a video or animated image.
func (m *MediaFile) IsAnimated() bool {
	return m.IsVideo() || m.IsAnimatedImage()
}

// NotAnimated checks if the file is not a video or an animated image.
func (m *MediaFile) NotAnimated() bool {
	return !m.IsAnimated()
}

// IsVideo returns true if this is a video file.
func (m *MediaFile) IsVideo() bool {
	return m.HasMediaType(media.Video)
}

// IsVector returns true if this is a vector graphics.
func (m *MediaFile) IsVector() bool {
	return m.HasMediaType(media.Vector) || m.IsSVG()
}

// IsSidecar checks if the file is a metadata sidecar file, independent of the storage location.
func (m *MediaFile) IsSidecar() bool {
	return !m.Media().Main()
}

// IsSVG returns true if this is a SVG vector graphics.
func (m *MediaFile) IsSVG() bool {
	return m.FileType() == fs.VectorSVG
}

// IsXMP returns true if this is a XMP sidecar file.
func (m *MediaFile) IsXMP() bool {
	return m.FileType() == fs.SidecarXMP
}

// InOriginals checks if the file is stored in the 'originals' folder.
func (m *MediaFile) InOriginals() bool {
	return m.Root() == entity.RootOriginals
}

// InSidecar checks if the file is stored in the 'sidecar' folder.
func (m *MediaFile) InSidecar() bool {
	return m.Root() == entity.RootSidecar
}

// NeedsTranscoding checks whether the media file is a video or an animated image and should be transcoded to a playable format.
func (m *MediaFile) NeedsTranscoding() bool {
	if m.NotAnimated() {
		return false
	} else if m.HasFileType(fs.VideoAVC) || m.HasFileType(fs.VideoMP4) && m.MetaData().CodecAvc() {
		return false
	}

	if m.IsAnimatedImage() {
		return fs.VideoMP4.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false) == ""
	}

	return fs.VideoAVC.FindFirst(m.FileName(), []string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false) == ""
}

// SkipTranscoding checks if the media file is not animated or has already been transcoded to a playable format.
func (m *MediaFile) SkipTranscoding() bool {
	return !m.NeedsTranscoding()
}

// IsImageOther returns true if this is a PNG, GIF, BMP, TIFF, or WebP file.
func (m *MediaFile) IsImageOther() bool {
	switch {
	case m.IsPNG(), m.IsGIF(), m.IsTIFF(), m.IsBMP(), m.IsWebP():
		return true
	default:
		return false
	}
}

// IsImageNative returns true if it is a natively supported image file.
func (m *MediaFile) IsImageNative() bool {
	return m.IsJpeg() || m.IsImageOther()
}

// IsLive checks if the file is a live photo.
func (m *MediaFile) IsLive() bool {
	if m.IsHEIC() {
		return fs.VideoMOV.FindFirst(m.FileName(), []string{}, Config().OriginalsPath(), false) != ""
	}

	if m.IsVideo() {
		return fs.ImageHEIC.FindFirst(m.FileName(), []string{}, Config().OriginalsPath(), false) != ""
	}

	return m.MetaData().MediaType == media.Live && m.VideoInfo().Compatible
}

// ExifSupported returns true if parsing exif metadata is supported for the media file type.
func (m *MediaFile) ExifSupported() bool {
	return m.IsJpeg() || m.IsRaw() || m.IsHEIF() || m.IsPNG() || m.IsTIFF()
}

// IsMedia returns true if this is a media file (photo or video, not sidecar or other).
func (m *MediaFile) IsMedia() bool {
	return m.IsImage() || m.IsRaw() || m.IsVideo() || m.IsVector()
}

// PreviewImage returns a PNG or JPEG version of the media file, if exists.
func (m *MediaFile) PreviewImage() (*MediaFile, error) {
	if m.IsJpeg() {
		if !fs.FileExists(m.FileName()) {
			return nil, fmt.Errorf("jpeg should exist, but does not: %s", m.RootRelName())
		}

		return m, nil
	} else if m.Empty() {
		return nil, fmt.Errorf("%s is empty", m.RootRelName())
	}

	jpegName := fs.ImageJPEG.FindFirst(m.FileName(),
		[]string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false)

	if jpegName != "" {
		return NewMediaFile(jpegName)
	}

	pngName := fs.ImagePNG.FindFirst(m.FileName(),
		[]string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false)

	if pngName != "" {
		return NewMediaFile(pngName)
	}

	return nil, fmt.Errorf("no preview image found for %s", m.RootRelName())
}

// HasPreviewImage returns true if the file has or is a JPEG or PNG image.
func (m *MediaFile) HasPreviewImage() bool {
	if m.hasPreviewImage {
		return true
	}

	if m.IsPreviewImage() {
		m.hasPreviewImage = true
		return true
	}

	jpegName := fs.ImageJPEG.FindFirst(m.FileName(),
		[]string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false)

	if m.hasPreviewImage = fs.MimeType(jpegName) == fs.MimeTypeJPEG; m.hasPreviewImage {
		return true
	}

	pngName := fs.ImagePNG.FindFirst(m.FileName(),
		[]string{Config().SidecarPath(), fs.PPHiddenPathname}, Config().OriginalsPath(), false)

	if m.hasPreviewImage = fs.MimeType(pngName) == fs.MimeTypePNG; m.hasPreviewImage {
		return true
	}

	return false
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

// ExceedsBytes checks if the file exceeds the specified size limit in bytes.
func (m *MediaFile) ExceedsBytes(limit int64) (err error, fileSize int64) {
	if fileSize = m.FileSize(); limit <= 0 {
		return nil, fileSize
	} else if fileSize <= 0 || fileSize <= limit {
		return nil, fileSize
	} else {
		return fmt.Errorf("%s exceeds file size limit (%s / %s)", clean.Log(m.RootRelName()), humanize.Bytes(uint64(fileSize)), humanize.Bytes(uint64(limit))), fileSize
	}
}

// ExceedsResolution checks if an image in a natively supported format exceeds the configured resolution limit in megapixels.
func (m *MediaFile) ExceedsResolution(limit int) (err error, resolution int) {
	if limit <= 0 {
		return nil, resolution
	} else if !m.IsImage() {
		return nil, resolution
	} else if resolution = m.Megapixels(); resolution <= 0 || resolution <= limit {
		return nil, resolution
	} else {
		return fmt.Errorf("%s exceeds resolution limit (%d / %d MP)", clean.Log(m.RootRelName()), resolution, limit), resolution
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
