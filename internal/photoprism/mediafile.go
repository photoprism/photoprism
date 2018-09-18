package photoprism

import (
	"encoding/hex"
	"fmt"
	"github.com/brett-lempereur/ish"
	"github.com/djherbis/times"
	. "github.com/photoprism/photoprism/internal/models"
	"github.com/steakknife/hamming"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	FileTypeOther = "unknown"
	FileTypeYaml  = "yml"
	FileTypeJpeg  = "jpg"
	FileTypeRaw   = "raw"
	FileTypeXmp   = "xmp"
	FileTypeAae   = "aae"
	FileTypeMovie = "mov"
	FileTypeHEIF  = "heif" // High Efficiency Image File Format
)

const (
	MimeTypeJpeg = "image/jpeg"
)

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
	location       *Location
}

func NewMediaFile(filename string) (*MediaFile, error) {
	if !fileExists(filename) {
		return nil, fmt.Errorf("file does not exist: %s", filename)
	}

	instance := &MediaFile{
		filename: filename,
		fileType: FileTypeOther,
	}

	return instance, nil
}

func (m *MediaFile) GetDateCreated() time.Time {
	if !m.dateCreated.IsZero() {
		return m.dateCreated
	}

	m.dateCreated = time.Now()

	info, err := m.GetExifData()

	if err == nil && !info.DateTime.IsZero() {
		m.dateCreated = info.DateTime

		return m.dateCreated
	}

	t, err := times.Stat(m.GetFilename())

	if err != nil {
		log.Println(err.Error())

		return m.dateCreated
	}

	if t.HasBirthTime() {
		m.dateCreated = t.BirthTime()
	} else {
		m.dateCreated = t.ModTime()
	}

	return m.dateCreated
}

func (m *MediaFile) GetCameraModel() string {
	info, err := m.GetExifData()

	var result string

	if err == nil {
		result = info.CameraModel
	}

	return result
}

func (m *MediaFile) GetCanonicalName() string {
	var postfix string

	dateCreated := m.GetDateCreated().UTC()

	if fileHash := m.GetHash(); len(fileHash) > 12 {
		postfix = strings.ToUpper(fileHash[:12])
	} else {
		postfix = "NOTFOUND"
	}

	result := dateCreated.Format("20060102_150405_") + postfix

	return result
}

func (m *MediaFile) GetCanonicalNameFromFile() string {
	basename := filepath.Base(m.GetFilename())

	if end := strings.Index(basename, "."); end != -1 {
		return basename[:end] // Length of canonical name: 16 + 12
	} else {
		return basename
	}
}

func (m *MediaFile) GetCanonicalNameFromFileWithDirectory() string {
	return m.GetDirectory() + string(os.PathSeparator) + m.GetCanonicalNameFromFile()
}

func (m *MediaFile) GetPerceptualHash() (string, error) {
	if m.perceptualHash != "" {
		return m.perceptualHash, nil
	}

	hasher := ish.NewDifferenceHash(8, 8)
	img, _, err := ish.LoadFile(m.GetFilename())

	if err != nil {
		return "", err
	}

	dh, err := hasher.Hash(img)

	if err != nil {
		return "", err
	}

	m.perceptualHash = hex.EncodeToString(dh)

	return m.perceptualHash, nil
}

func (m *MediaFile) GetPerceptualDistance(perceptualHash string) (int, error) {
	var hash1, hash2 []byte

	if imageHash, err := m.GetPerceptualHash(); err != nil {
		return -1, err
	} else {
		if decoded, err := hex.DecodeString(imageHash); err != nil {
			return -1, err
		} else {
			hash1 = decoded
		}
	}

	if decoded, err := hex.DecodeString(perceptualHash); err != nil {
		return -1, err
	} else {
		hash2 = decoded
	}

	result := hamming.Bytes(hash1, hash2)

	return result, nil
}

func (m *MediaFile) GetHash() string {
	if len(m.hash) == 0 {
		m.hash = fileHash(m.GetFilename())
	}

	return m.hash
}

// When editing photos, iPhones create additional files like IMG_E12345.JPG
func (m *MediaFile) GetEditedFilename() (result string) {
	basename := filepath.Base(m.filename)

	if strings.ToUpper(basename[:4]) == "IMG_" && strings.ToUpper(basename[:5]) != "IMG_E" {
		result = filepath.Dir(m.filename) + string(os.PathSeparator) + basename[:4] + "E" + basename[4:]
	}

	return result
}

func (m *MediaFile) GetRelatedFiles() (result MediaFiles, mainFile *MediaFile, err error) {
	baseFilename := m.GetCanonicalNameFromFileWithDirectory()

	matches, err := filepath.Glob(baseFilename + "*")

	if err != nil {
		return result, nil, err
	}

	if editedFilename := m.GetEditedFilename(); editedFilename != "" && fileExists(editedFilename) {
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
		} else if resultFile.IsJpeg() && resultFile.IsJpeg() && len(mainFile.GetFilename()) > len(resultFile.GetFilename()) {
			mainFile = resultFile
		}

		result = append(result, resultFile)
	}

	sort.Sort(result)

	return result, mainFile, nil
}

func (m *MediaFile) GetFilename() string {
	return m.filename
}

func (m *MediaFile) SetFilename(filename string) {
	m.filename = filename
}

func (m *MediaFile) GetRelativeFilename(directory string) string {
	index := strings.Index(m.filename, directory)

	if index == 0 {
		pos := len(directory) + 1
		return m.filename[pos:]
	}

	return m.filename
}

func (m *MediaFile) GetDirectory() string {
	return filepath.Dir(m.filename)
}

func (m *MediaFile) GetBasename() string {
	return filepath.Base(m.filename)
}

func (m *MediaFile) GetMimeType() string {
	if m.mimeType != "" {
		return m.mimeType
	}

	handle, err := m.openFile()

	if err != nil {
		log.Println("Error: Could not open file to determine mime type")
		return ""
	}

	defer handle.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = handle.Read(buffer)

	if err != nil {
		log.Println("Error: Could not read file to determine mime type: " + m.GetFilename())
		return ""
	}

	m.mimeType = http.DetectContentType(buffer)

	return m.mimeType
}

func (m *MediaFile) openFile() (*os.File, error) {
	if handle, err := os.Open(m.filename); err == nil {
		return handle, nil
	} else {
		log.Println(err.Error())
		return nil, err
	}
}

func (m *MediaFile) Exists() bool {
	return fileExists(m.GetFilename())
}

func (m *MediaFile) Remove() error {
	return os.Remove(m.GetFilename())
}

func (m *MediaFile) HasSameFilename(other *MediaFile) bool {
	return m.GetFilename() == other.GetFilename()
}

func (m *MediaFile) Move(newFilename string) error {
	if err := os.Rename(m.filename, newFilename); err != nil {
		return err
	}

	m.filename = newFilename

	return nil
}

func (m *MediaFile) Copy(destinationFilename string) error {
	file, err := m.openFile()

	if err != nil {
		log.Println(err.Error())
		return err
	}

	defer file.Close()

	destination, err := os.OpenFile(destinationFilename, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	defer destination.Close()

	_, err = io.Copy(destination, file)

	if err != nil {
		log.Println(err.Error())
		return err
	}

	return nil
}

func (m *MediaFile) GetExtension() string {
	return strings.ToLower(filepath.Ext(m.filename))
}

func (m *MediaFile) IsJpeg() bool {
	// Don't import/use existing thumbnail files (we create our own)
	if m.GetExtension() == ".thm" {
		return false
	}

	return m.GetMimeType() == MimeTypeJpeg
}

func (m *MediaFile) GetType() string {
	return FileExtensions[m.GetExtension()]
}

func (m *MediaFile) HasType(typeString string) bool {
	if typeString == FileTypeJpeg {
		return m.IsJpeg()
	}

	return m.GetType() == typeString
}

func (m *MediaFile) IsRaw() bool {
	return m.HasType(FileTypeRaw)
}

func (m *MediaFile) IsHighEfficiencyImageFile() bool {
	return m.HasType(FileTypeHEIF)
}

func (m *MediaFile) IsPhoto() bool {
	return m.IsJpeg() || m.IsRaw() || m.IsHighEfficiencyImageFile()
}

func (m *MediaFile) GetJpeg() (*MediaFile, error) {
	if m.IsJpeg() {
		return m, nil
	}

	jpegFilename := m.GetCanonicalNameFromFileWithDirectory() + ".jpg"

	if !fileExists(jpegFilename) {
		return nil, fmt.Errorf("jpeg file does not exist: %s", jpegFilename)
	}

	return NewMediaFile(jpegFilename)
}

func (m *MediaFile) decodeDimensions() error {
	if m.IsJpeg() {
		file, err := os.Open(m.GetFilename())

		defer file.Close()

		if err != nil {
			return err
		}

		size, _, err := image.DecodeConfig(file)

		if err != nil {
			return err
		}

		m.width = size.Width
		m.height = size.Height
	} else {
		if exif, err := m.GetExifData(); err == nil {
			m.width = exif.Width
			m.height = exif.Height
		} else {
			return err
		}
	}

	return nil
}

func (m *MediaFile) GetWidth() int {
	if m.width <= 0 {
		m.decodeDimensions()
	}

	return m.width
}

func (m *MediaFile) GetHeight() int {
	if m.height <= 0 {
		m.decodeDimensions()
	}

	return m.height
}

func (m *MediaFile) GetAspectRatio() float64 {
	width := float64(m.GetWidth())
	height := float64(m.GetHeight())

	if width <= 0 || height <= 0 {
		return 0
	}

	aspectRatio := width / height

	return math.Round(aspectRatio*100) / 100
}

func (m *MediaFile) GetOrientation() int {
	if exif, err := m.GetExifData(); err == nil {
		return exif.Orientation
	} else {
		return 1
	}
}
