package photoprism

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/photoprism/photoprism/internal/fsutil"
)

// Importer todo: Fill me.
type Importer struct {
	originalsPath          string
	indexer                *Indexer
	converter              *Converter
	removeDotFiles         bool
	removeExistingFiles    bool
	removeEmptyDirectories bool
}

// NewImporter returns a new importer.
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

// ImportPhotosFromDirectory imports all the photos from a given directory path.
// This function ignores errors.
func (i *Importer) ImportPhotosFromDirectory(importPath string) {
	var directories []string

	err := filepath.Walk(importPath, func(filename string, fileInfo os.FileInfo, err error) error {
		var destinationMainFilename string

		if err != nil {
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
			log.Errorf("could not import \"%s\": %s", mediaFile.GetRelativeFilename(importPath), err.Error())

			return nil
		}

		for _, relatedMediaFile := range relatedFiles {
			if destinationFilename, err := i.GetDestinationFilename(mainFile, relatedMediaFile); err == nil {
				os.MkdirAll(path.Dir(destinationFilename), os.ModePerm)

				if mainFile.HasSameFilename(relatedMediaFile) {
					destinationMainFilename = destinationFilename
					log.Infof("moving main %s file \"%s\" to \"%s\"", relatedMediaFile.GetType(), relatedMediaFile.GetRelativeFilename(importPath), destinationFilename)
				} else {
					log.Infof("moving related %s file \"%s\" to \"%s\"", relatedMediaFile.GetType(), relatedMediaFile.GetRelativeFilename(importPath), destinationFilename)
				}

				relatedMediaFile.Move(destinationFilename)
			} else if i.removeExistingFiles {
				relatedMediaFile.Remove()
				log.Infof("deleted \"%s\" (already exists)", relatedMediaFile.GetRelativeFilename(importPath))
			}
		}

		if destinationMainFilename != "" {
			importedMainFile, err := NewMediaFile(destinationMainFilename)

			if err != nil {
				log.Errorf("could not index \"%s\" after import: %s", destinationMainFilename, err.Error())

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
				if err := os.Remove(directory); err != nil {
					log.Errorf("could not deleted empty directory \"%s\": %s", directory, err)
				} else {
					log.Infof("deleted empty directory \"%s\"", directory)
				}
			}
		}
	}

	if err != nil {
		log.Error(err.Error())
	}
}

// GetDestinationFilename get the destination of a media file.
func (i *Importer) GetDestinationFilename(mainFile *MediaFile, mediaFile *MediaFile) (string, error) {
	canonicalName := mainFile.GetCanonicalName()
	fileExtension := mediaFile.GetExtension()
	dateCreated := mainFile.GetDateCreated()

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	pathName := i.originalsPath + "/" + dateCreated.UTC().Format("2006/01")

	iteration := 0

	result := pathName + "/" + canonicalName + fileExtension

	for fsutil.Exists(result) {
		if mediaFile.GetHash() == fsutil.Hash(result) {
			return result, fmt.Errorf("file already exists: %s", result)
		}

		iteration++

		result = pathName + "/" + canonicalName + "." + fmt.Sprintf("edited_%d", iteration) + fileExtension
	}

	return result, nil
}

func directoryIsEmpty(path string) bool {
	f, err := os.Open(path)

	if err != nil {
		return false
	}

	defer f.Close()

	_, err = f.Readdirnames(1)

	if err == io.EOF {
		return true
	}

	return false
}
