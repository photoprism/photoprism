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
	removeDotFiles         bool
	removeExistingFiles    bool
	removeEmptyDirectories bool
}

func NewImporter(originalsPath string, indexer *Indexer) *Importer {
	instance := &Importer{
		originalsPath:          originalsPath,
		indexer:                indexer,
		removeDotFiles:         true,
		removeExistingFiles:    true,
		removeEmptyDirectories: true,
	}

	return instance
}

func (i *Importer) ImportPhotosFromDirectory(importPath string) {
	var directories []string

	err := filepath.Walk(importPath, func(filename string, fileInfo os.FileInfo, err error) error {
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

		mediaFile := NewMediaFile(filename)

		if !mediaFile.Exists() || !mediaFile.IsPhoto() {
			return nil
		}

		relatedFiles, masterFile, _ := mediaFile.GetRelatedFiles()

		for _, relatedMediaFile := range relatedFiles {
			if destinationFilename, err := i.GetDestinationFilename(masterFile, relatedMediaFile); err == nil {
				os.MkdirAll(path.Dir(destinationFilename), os.ModePerm)
				log.Printf("Moving file %s to %s", relatedMediaFile.GetFilename(), destinationFilename)
				relatedMediaFile.Move(destinationFilename)
				i.indexer.IndexMediaFile(relatedMediaFile)
			} else if i.removeExistingFiles {
				relatedMediaFile.Remove()
				log.Printf("Deleted %s (already exists)", relatedMediaFile.GetFilename())
			}
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
				log.Printf("Deleted empty directory %s", directory)
			}
		}
	}

	if err != nil {
		log.Print(err.Error())
	}
}

func (i *Importer) GetDestinationFilename(masterFile *MediaFile, mediaFile *MediaFile) (string, error) {
	canonicalName := masterFile.GetCanonicalName()
	fileExtension := mediaFile.GetExtension()
	dateCreated := masterFile.GetDateCreated()

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	pathName := i.originalsPath + "/" + dateCreated.UTC().Format("2006/01")

	iteration := 1

	result := pathName + "/" + canonicalName + fileExtension

	for fileExists(result) {
		if mediaFile.GetHash() == fileHash(result) {
			return result, errors.New("File already exists")
		}

		iteration++

		result = pathName + "/" + canonicalName + "." + fmt.Sprintf("v%d", iteration) + fileExtension
	}

	return result, nil
}
