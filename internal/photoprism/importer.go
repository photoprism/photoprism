package photoprism

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/models"

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
			if err := os.Remove(filename); err != nil {
				log.Errorf("could not remove \"%s\": %s", filename, err.Error())
			}

			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsPhoto() {
			return nil
		}

		relatedFiles, mainFile, err := mediaFile.RelatedFiles()

		if err != nil {
			event.Error(fmt.Sprintf("could not import \"%s\": %s", mediaFile.RelativeFilename(importPath), err.Error()))

			return nil
		}

		event.Publish("import.file", event.Data{
			"fileName": mainFile.Filename(),
			"baseName": filepath.Base(mainFile.Filename()),
		})

		for _, relatedMediaFile := range relatedFiles {
			relativeFilename := relatedMediaFile.RelativeFilename(importPath)

			if destinationFilename, err := i.DestinationFilename(mainFile, relatedMediaFile); err == nil {
				if err := os.MkdirAll(path.Dir(destinationFilename), os.ModePerm); err != nil {
					log.Errorf("could not create directories: %s", err.Error())
				}

				if mainFile.HasSameFilename(relatedMediaFile) {
					destinationMainFilename = destinationFilename
					log.Infof("moving main %s file \"%s\" to \"%s\"", relatedMediaFile.Type(), relativeFilename, destinationFilename)
				} else {
					log.Infof("moving related %s file \"%s\" to \"%s\"", relatedMediaFile.Type(), relativeFilename, destinationFilename)
				}

				if err := relatedMediaFile.Move(destinationFilename); err != nil {
					log.Errorf("could not move file to \"%s\": %s", destinationMainFilename, err.Error())
				}
			} else if i.removeExistingFiles {
				if err := relatedMediaFile.Remove(); err != nil {
					log.Errorf("could not delete file \"%s\": %s", relatedMediaFile.Filename(), err.Error())
				} else {
					log.Infof("deleted \"%s\" (already exists)", relativeFilename)
				}
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
			if importedMainFile.IsHEIF() {
				if _, err := i.converter.ConvertToJpeg(importedMainFile); err != nil {
					log.Errorf("could not create jpeg from heif: %s", err)
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
	fileName := mainFile.CanonicalName()
	fileExtension := mediaFile.Extension()
	dateCreated := mainFile.DateCreated()

	if file, err := models.FindFileByHash(i.conf.Db(), mediaFile.Hash()); err == nil {
		existingFilename := i.conf.OriginalsPath() + string(os.PathSeparator) + file.FileName
		return existingFilename, fmt.Errorf("\"%s\" is identical to \"%s\" (%s)", mediaFile.Filename(), file.FileName, mediaFile.Hash())
	}

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	pathName := i.originalsPath() + string(os.PathSeparator) + dateCreated.UTC().Format("2006/01")

	iteration := 0

	result := pathName + string(os.PathSeparator) + fileName + fileExtension

	for util.Exists(result) {
		if mediaFile.Hash() == util.Hash(result) {
			return result, fmt.Errorf("file already exists: %s", result)
		}

		iteration++

		result = pathName + string(os.PathSeparator) + fileName + "." + fmt.Sprintf("edited_%d", iteration) + fileExtension
	}

	return result, nil
}
