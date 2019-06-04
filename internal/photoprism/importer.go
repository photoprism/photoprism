package photoprism

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/config"

	"github.com/photoprism/photoprism/internal/util"
)

// Importer is responsible for importing new files to originals.
type Importer struct {
	conf                   *config.Config
	indexer                *Indexer
	converter              *Converter
	removeDotFiles         bool
	removeExistingFiles    bool
	removeEmptyDirectories bool
}

// NewImporter returns a new importer.
func NewImporter(conf *config.Config, indexer *Indexer, converter *Converter) *Importer {
	instance := &Importer{
		conf:                   conf,
		indexer:                indexer,
		converter:              converter,
		removeDotFiles:         true,
		removeExistingFiles:    true,
		removeEmptyDirectories: true,
	}

	return instance
}

func (i *Importer) originalsPath() string {
	return i.conf.OriginalsPath()
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

		relatedFiles, mainFile, err := mediaFile.RelatedFiles()

		if err != nil {
			log.Errorf("could not import \"%s\": %s", mediaFile.RelativeFilename(importPath), err.Error())

			return nil
		}

		for _, relatedMediaFile := range relatedFiles {
			if destinationFilename, err := i.DestinationFilename(mainFile, relatedMediaFile); err == nil {
				os.MkdirAll(path.Dir(destinationFilename), os.ModePerm)

				if mainFile.HasSameFilename(relatedMediaFile) {
					destinationMainFilename = destinationFilename
					log.Infof("moving main %s file \"%s\" to \"%s\"", relatedMediaFile.Type(), relatedMediaFile.RelativeFilename(importPath), destinationFilename)
				} else {
					log.Infof("moving related %s file \"%s\" to \"%s\"", relatedMediaFile.Type(), relatedMediaFile.RelativeFilename(importPath), destinationFilename)
				}

				relatedMediaFile.Move(destinationFilename)
			} else if i.removeExistingFiles {
				relatedMediaFile.Remove()
				log.Infof("deleted \"%s\" (already exists)", relatedMediaFile.RelativeFilename(importPath))
			}
		}

		if destinationMainFilename != "" {
			importedMainFile, err := NewMediaFile(destinationMainFilename)

			if err != nil {
				log.Errorf("could not index \"%s\" after import: %s", destinationMainFilename, err.Error())

				return nil
			}

			if importedMainFile.IsRaw() {
				if _, err := i.converter.ConvertToJpeg(importedMainFile); err != nil {
					log.Errorf("could not create jpeg from raw: %s", err)
				}
			}

			if jpg, err := importedMainFile.Jpeg(); err != nil {
				log.Error(err)
			} else {
				if err := jpg.CreateDefaultThumbnails(i.conf.ThumbnailsPath(), false); err != nil {
					log.Errorf("could not create default thumbnails: %s", err)
				}
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
			if util.DirectoryIsEmpty(directory) {
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

// DestinationFilename get the destination of a media file.
func (i *Importer) DestinationFilename(mainFile *MediaFile, mediaFile *MediaFile) (string, error) {
	canonicalName := mainFile.CanonicalName()
	fileExtension := mediaFile.Extension()
	dateCreated := mainFile.DateCreated()

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	pathName := i.originalsPath() + "/" + dateCreated.UTC().Format("2006/01")

	iteration := 0

	result := pathName + "/" + canonicalName + fileExtension

	for util.Exists(result) {
		if mediaFile.Hash() == util.Hash(result) {
			return result, fmt.Errorf("file already exists: %s", result)
		}

		iteration++

		result = pathName + "/" + canonicalName + "." + fmt.Sprintf("edited_%d", iteration) + fileExtension
	}

	return result, nil
}
