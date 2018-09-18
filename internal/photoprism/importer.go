package photoprism

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type Importer struct {
	originalsPath          string
	indexer                *Indexer
	converter              *Converter
	removeDotFiles         bool
	removeExistingFiles    bool
	removeEmptyDirectories bool
}

func NewImporter(originalsPath string, indexer *Indexer, converter *Converter) *Importer {
	instance := &Importer{
		originalsPath:          originalsPath,
		indexer:                indexer,
		converter:              converter,
		removeDotFiles:         true,
		removeExistingFiles:    true,
		removeEmptyDirectories: true,
	}

	return instance
}

func (i *Importer) ImportPhotosFromDirectory(importPath string) {
	var directories []string

	err := filepath.Walk(importPath, func(filename string, fileInfo os.FileInfo, err error) error {
		var destinationMainFilename string

		if err != nil {
			// log.Print(err.Error())
			return nil
		}

		if fileInfo.IsDir() {
			if filename != importPath {
				directories = append(directories, filename)
			}
			return nil
		}

		if i.removeDotFiles && strings.HasPrefix(filepath.Base(filename), ".") {
			os.Remove(filename)

			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsPhoto() {
			return nil
		}

		relatedFiles, mainFile, err := mediaFile.GetRelatedFiles()

		if err != nil {
			log.Printf("Could not import \"%s\": %s", mediaFile.GetRelativeFilename(importPath), err.Error())

			return nil
		}

		for _, relatedMediaFile := range relatedFiles {
			if destinationFilename, err := i.GetDestinationFilename(mainFile, relatedMediaFile); err == nil {
				os.MkdirAll(path.Dir(destinationFilename), os.ModePerm)

				if mainFile.HasSameFilename(relatedMediaFile) {
					destinationMainFilename = destinationFilename
					log.Printf("Moving main %s file \"%s\" to \"%s\"", relatedMediaFile.GetType(), relatedMediaFile.GetRelativeFilename(importPath), destinationFilename)
				} else {
					log.Printf("Moving related %s file \"%s\" to \"%s\"", relatedMediaFile.GetType(), relatedMediaFile.GetRelativeFilename(importPath), destinationFilename)
				}

				relatedMediaFile.Move(destinationFilename)
			} else if i.removeExistingFiles {
				relatedMediaFile.Remove()
				log.Printf("Deleted \"%s\" (already exists)", relatedMediaFile.GetRelativeFilename(importPath))
			}
		}

		if destinationMainFilename != "" {
			importedMainFile, err := NewMediaFile(destinationMainFilename)

			if err != nil {
				log.Printf("Could not index \"%s\" after import: %s", destinationMainFilename, err.Error())

				return nil
			}

			if importedMainFile.IsRaw() {
				i.converter.ConvertToJpeg(importedMainFile)
			}

			i.indexer.IndexRelated(importedMainFile)
		}

		return nil
	})

	sort.Slice(directories, func(i, j int) bool {
		return len(directories[i]) > len(directories[j])
	})

	if i.removeEmptyDirectories {
		// Remove empty directories from import path
		for _, directory := range directories {
			if directoryIsEmpty(directory) {
				os.Remove(directory)
				log.Printf("Deleted empty directory \"%s\"", directory)
			}
		}
	}

	if err != nil {
		log.Print(err.Error())
	}
}

func (i *Importer) GetDestinationFilename(mainFile *MediaFile, mediaFile *MediaFile) (string, error) {
	canonicalName := mainFile.GetCanonicalName()
	fileExtension := mediaFile.GetExtension()
	dateCreated := mainFile.GetDateCreated()

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	pathName := i.originalsPath + "/" + dateCreated.UTC().Format("2006/01")

	iteration := 0

	result := pathName + "/" + canonicalName + fileExtension

	for fileExists(result) {
		if mediaFile.GetHash() == fileHash(result) {
			return result, errors.New("File already exists")
		}

		iteration++

		result = pathName + "/" + canonicalName + "." + fmt.Sprintf("edit%d", iteration) + fileExtension
	}

	return result, nil
}
