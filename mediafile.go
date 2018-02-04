package photoprism

import (
	"path/filepath"
	"encoding/hex"
	"github.com/brett-lempereur/ish"
	"net/http"
	"os"
	"log"
	"strings"
	"time"
	"github.com/djherbis/times"
	"fmt"
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

var FileExtensions = map[string]string {
	".crw": FileTypeRaw,
	".cr2": FileTypeRaw,
	".nef": FileTypeRaw,
	".arw": FileTypeRaw,
	".dng": FileTypeRaw,
	".mov": FileTypeMovie,
	".avi": FileTypeMovie,
	".yml": FileTypeYaml,
	".jpg": FileTypeJpeg,
	".jpeg": FileTypeJpeg,
	".xmp": FileTypeXmp,
	".aae": FileTypeAae,
}

type MediaFile struct {
	filename string
	dateCreated time.Time
	hash     []byte
	fileType string
	mimeType string
	tags     []string
	exifData *ExifData
}

func NewMediaFile(filename string) *MediaFile {
	instance := &MediaFile{filename: filename}

	return instance
}

func (mediaFile *MediaFile) GetDateCreated() time.Time {
	if !mediaFile.dateCreated.IsZero() {
		return mediaFile.dateCreated
	}

	info, err := mediaFile.GetExifData()

	if err == nil {
		mediaFile.dateCreated = info.DateTime
		return info.DateTime
	}

	t, err := times.Stat(mediaFile.GetFilename())

	if err != nil {
		log.Fatal(err.Error())
	}

	if t.HasBirthTime() {
		mediaFile.dateCreated = t.BirthTime()
		return t.BirthTime()
	}

	mediaFile.dateCreated = t.ModTime()
	return t.ModTime()
}

func (mediaFile *MediaFile) GetCameraModel () string {
	info, err := mediaFile.GetExifData()

	var result string

	if err == nil {
		result = info.CameraModel
	}

	return result
}

func (mediaFile *MediaFile) GetCanonicalName() string {
	dateCreated := mediaFile.GetDateCreated().UTC()
	cameraModel := strings.Replace(mediaFile.GetCameraModel(), " ", "_", -1)

	result := dateCreated.Format("20060102_150405_") + strings.ToUpper(mediaFile.GetHashString()[:8])

	if cameraModel != "" {
		result = result + "_" + cameraModel
	}

	return result
}

func (mediaFile *MediaFile) GetPerceptiveHash() (string, error) {
	hasher := ish.NewDifferenceHash(8, 8)
	img, _, err := ish.LoadFile(mediaFile.GetFilename())

	if err != nil {
		return "", err
	}

	dh, err := hasher.Hash(img)

	if err != nil {
		return "", err
	}

	dhs := hex.EncodeToString(dh)

	return dhs, nil
}

func (mediaFile *MediaFile) GetHash() []byte  {
	if len(mediaFile.hash) == 0 {
		mediaFile.hash = Md5Sum(mediaFile.GetFilename())
	}

	return mediaFile.hash
}

func (mediaFile *MediaFile) GetHashString() string {
	return fmt.Sprintf("%x", mediaFile.GetHash())
}

func (mediaFile *MediaFile) GetRelatedFiles() (result []*MediaFile, err error) {
	extension := mediaFile.GetExtension()

	baseFilename := mediaFile.filename[0:len(mediaFile.filename)-len(extension)]

	matches, err := filepath.Glob(baseFilename + "*")

	if err != nil {
		return result, err
	}

	for _, filename := range matches {
		result = append(result, NewMediaFile(filename))
	}

	return result, nil
}


func (mediaFile *MediaFile) GetFilename() string {
	return mediaFile.filename
}

func (mediaFile *MediaFile) SetFilename(filename string)  {
	mediaFile.filename = filename
}

func (mediaFile *MediaFile) GetMimeType() string {
	if mediaFile.mimeType != "" {
		return mediaFile.mimeType
	}

	handle, err := mediaFile.openFile()

	if err != nil {
		log.Println("Error: Could not open file to determine mime type")
		return ""
	}

	defer handle.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err = handle.Read(buffer)

	if err != nil {
		log.Println("Error: Could not read file to determine mime type: " + mediaFile.GetFilename())
		return ""
	}

	mediaFile.mimeType = http.DetectContentType(buffer)

	return mediaFile.mimeType
}

func (mediaFile *MediaFile) openFile() (*os.File, error) {
	if handle, err := os.Open(mediaFile.filename); err == nil {
		return handle, nil
	} else {
		log.Println(err.Error())
		return nil, err
	}
}

func (mediaFile *MediaFile) Exists() bool {
	return FileExists(mediaFile.GetFilename())
}

func (mediaFile *MediaFile) Move(newFilename string) error {
	if err := os.Rename(mediaFile.filename, newFilename); err != nil {
		return err
	}

	mediaFile.filename = newFilename

	return nil
}

func (mediaFile *MediaFile) GetExtension() string {
	return strings.ToLower(filepath.Ext(mediaFile.filename))
}

func (mediaFile *MediaFile) IsJpeg() bool {
	return mediaFile.GetMimeType() == MimeTypeJpeg
}

func (mediaFile *MediaFile) HasType(typeString string) bool {
	if typeString == FileTypeJpeg {
		return mediaFile.IsJpeg()
	}

	return FileExtensions[mediaFile.GetExtension()] == typeString
}

func (mediaFile *MediaFile) IsRaw() bool {
	return mediaFile.HasType(FileTypeRaw)
}

func (mediaFile *MediaFile) IsPhoto() bool {
	return mediaFile.IsJpeg() || mediaFile.IsRaw()
}