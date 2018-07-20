package photoprism

import (
	"encoding/hex"
	"github.com/brett-lempereur/ish"
	"github.com/djherbis/times"
	"github.com/steakknife/hamming"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/pkg/errors"
)

const (
	FileTypeOther = ""
	FileTypeYaml  = "yml"
	FileTypeJpeg  = "jpg"
	FileTypeRaw   = "raw"
	FileTypeXmp   = "xmp"
	FileTypeAae   = "aae"
	FileTypeMovie = "mov"
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
}

type MediaFile struct {
	filename       string
	dateCreated    time.Time
	hash           string
	fileType       string
	mimeType       string
	perceptualHash string
	tags           []string
	exifData       *ExifData
}

func NewMediaFile(filename string) *MediaFile {
	instance := &MediaFile{
		filename: filename,
		fileType: FileTypeOther,
	}

	return instance
}

func (m *MediaFile) GetDateCreated() time.Time {
	if !m.dateCreated.IsZero() {
		return m.dateCreated
	}

	info, err := m.GetExifData()

	if err == nil {
		m.dateCreated = info.DateTime
		return info.DateTime
	}

	t, err := times.Stat(m.GetFilename())

	if err != nil {
		log.Fatal(err.Error())
	}

	if t.HasBirthTime() {
		m.dateCreated = t.BirthTime()
		return t.BirthTime()
	}

	m.dateCreated = t.ModTime()
	return t.ModTime()
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
	dateCreated := m.GetDateCreated().UTC()

	result := dateCreated.Format("20060102_150405_") + strings.ToUpper(m.GetHash()[:12])

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

func (m *MediaFile) GetRelatedFiles() (result []*MediaFile, masterFile *MediaFile, err error) {
	extension := m.GetExtension()

	baseFilename := m.filename[0 : len(m.filename)-len(extension)]

	matches, err := filepath.Glob(baseFilename + "*")

	if err != nil {
		return result, nil, err
	}

	for _, filename := range matches {
		resultFile := NewMediaFile(filename)

		if masterFile == nil && resultFile.IsJpeg() {
			masterFile = resultFile
		} else if resultFile.IsRaw() {
			masterFile = resultFile
		}

		result = append(result, resultFile)
	}

	return result, masterFile, nil
}

func (m *MediaFile) GetFilename() string {
	return m.filename
}

func (m *MediaFile) SetFilename(filename string) {
	m.filename = filename
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

func (m *MediaFile) IsPhoto() bool {
	return m.IsJpeg() || m.IsRaw()
}

func (m *MediaFile) GetJpeg() (*MediaFile, error) {
	if m.IsJpeg() {
		return m, nil
	}

	jpegFilename := m.GetFilename()[0:len(m.GetFilename()) - len(filepath.Ext(m.GetFilename()))] + ".jpg"

	if !fileExists(jpegFilename) {
		return nil, errors.New("file does not exist")
	}

	result := NewMediaFile(jpegFilename)

	return result, nil
}
