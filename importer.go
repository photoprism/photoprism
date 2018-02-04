package photoprism

import (
	"path/filepath"
	"os"
	"log"
	"fmt"
	"github.com/pkg/errors"
	"bytes"
	"path"
)

type Importer struct {
	originalsPath string
	converter     *Converter
}

func NewImporter(originalsPath string, converter *Converter) *Importer {
	instance := &Importer{
		originalsPath: originalsPath,
		converter:     converter,
	}

	return instance
}

func (importer *Importer) CreateJpegFromRaw(sourcePath string) {
	err := filepath.Walk(sourcePath, func(filename string, fileInfo os.FileInfo, err error) error {

		if err != nil {
			log.Print(err.Error())
			return nil
		}

		if fileInfo.IsDir() {
			return nil
		}

		mediaFile := NewMediaFile(filename)

		if !mediaFile.Exists() || !mediaFile.IsRaw() {
			return nil
		}

		log.Printf("Converting %s \n", filename)

		if _, err := importer.converter.ConvertToJpeg(mediaFile); err != nil {
			log.Print(err.Error())
		}

		return nil
	})

	if err != nil {
		log.Print(err.Error())
	}
}

func (importer *Importer) ImportJpegFromDirectory(sourcePath string) {
	err := filepath.Walk(sourcePath, func(filename string, fileInfo os.FileInfo, err error) error {

		if err != nil {
			log.Print(err.Error())
			return nil
		}

		if fileInfo.IsDir() {
			return nil
		}

		jpegFile := NewMediaFile(filename)

		if !jpegFile.Exists() || !jpegFile.IsJpeg() {
			return nil
		}

		log.Println(jpegFile.GetFilename() + " -> " + jpegFile.GetCanonicalName())

		log.Println("Getting related files")

		relatedFiles, _ := jpegFile.GetRelatedFiles()

		for _, relatedMediaFile := range relatedFiles {
			log.Println("Processing " + relatedMediaFile.GetFilename())
			if destinationFilename, err := importer.GetDestinationFilename(jpegFile, relatedMediaFile); err == nil {
				log.Println("Creating directories")
				os.MkdirAll(path.Dir(destinationFilename), os.ModePerm)
				log.Println("Moving file " + relatedMediaFile.GetFilename())
				relatedMediaFile.Move(destinationFilename)
				log.Println("Moved file to  " + destinationFilename)
			} else {
				log.Println("File already exists: " + relatedMediaFile.GetFilename() + " -> " + destinationFilename)
			}
		}

		// mediaFile.Move(importer.originalsPath)

		return nil
	})

	if err != nil {
		log.Print(err.Error())
	}
}

func (importer *Importer) GetDestinationFilename(jpegFile *MediaFile, mediaFile *MediaFile) (string, error) {
	canonicalName := jpegFile.GetCanonicalName()
	fileExtension := mediaFile.GetExtension()
	dateCreated := jpegFile.GetDateCreated()

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	path := importer.originalsPath + "/" + dateCreated.UTC().Format("2006/01")

	i := 1

	result := path + "/" + canonicalName + fileExtension

	for FileExists(result) {
		if bytes.Compare(mediaFile.GetHash(), Md5Sum(result)) == 0 {
			return result, errors.New("File already exists")
		}

		i++
		result = path + "/" + canonicalName + "_" + fmt.Sprintf("%02d", i) + fileExtension
//		log.Println(result)
	}

	// os.MkdirAll(folderPath, os.ModePerm)

	return result, nil
}

func (importer *Importer) MoveRelatedFiles() {

}
