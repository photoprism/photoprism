package photoprism

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"

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

func (imp *Importer) originalsPath() string {
	return imp.conf.OriginalsPath()
}

// Start imports all the photos from a given directory path.
// This function ignores errors.
func (imp *Importer) Start(importPath string) {
	var directories []string
	done := make(map[string]bool)
	ind := imp.indexer

	if ind.running {
		event.Error("indexer already running")
		return
	}

	ind.running = true
	ind.canceled = false

	defer func() {
		ind.running = false
		ind.canceled = false
	}()

	if err := ind.tensorFlow.Init(); err != nil {
		log.Errorf("import: %s", err.Error())
		return
	}

	jobs := make(chan ImportJob)

	// Start a fixed number of goroutines to read and digest files.
	var wg sync.WaitGroup
	var numWorkers = ind.conf.Workers()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			importerWorker(jobs) // HLc
			wg.Done()
		}()
	}

	options := IndexerOptionsAll()

	err := filepath.Walk(importPath, func(filename string, fileInfo os.FileInfo, err error) error {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("import: %s [panic]", err)
			}
		}()

		if ind.canceled {
			return errors.New("importing canceled")
		}

		if err != nil || done[filename] {
			return nil
		}

		if fileInfo.IsDir() {
			if filename != importPath {
				directories = append(directories, filename)
			}
			return nil
		}

		if imp.removeDotFiles && strings.HasPrefix(filepath.Base(filename), ".") {
			done[filename] = true
			if err := os.Remove(filename); err != nil {
				log.Errorf("import: could not remove \"%s\" (%s)", filename, err.Error())
			}

			return nil
		}

		mediaFile, err := NewMediaFile(filename)

		if err != nil || !mediaFile.IsPhoto() {
			return nil
		}

		related, err := mediaFile.RelatedFiles()

		if err != nil {
			event.Error(fmt.Sprintf("import: %s", err.Error()))

			return nil
		}

		for _, f := range related.files {
			done[f.Filename()] = true
		}

		jobs <- ImportJob{
			related:    related,
			options:    options,
			importPath: importPath,
			imp:        imp,
		}

		return nil
	})

	close(jobs)
	wg.Wait()

	sort.Slice(directories, func(i, j int) bool {
		return len(directories[i]) > len(directories[j])
	})

	if imp.removeEmptyDirectories {
		// Remove empty directories from import path
		for _, directory := range directories {
			if util.DirectoryIsEmpty(directory) {
				if err := os.Remove(directory); err != nil {
					log.Errorf("import: could not deleted empty directory \"%s\" (%s)", directory, err)
				} else {
					log.Infof("import: deleted empty directory \"%s\"", directory)
				}
			}
		}
	}

	if err != nil {
		log.Error(err.Error())
	}
}

// Cancel stops the current import operation.
func (imp *Importer) Cancel() {
	imp.indexer.Cancel()
}

// DestinationFilename get the destination of a media file.
func (imp *Importer) DestinationFilename(mainFile *MediaFile, mediaFile *MediaFile) (string, error) {
	fileName := mainFile.CanonicalName()
	fileExtension := mediaFile.Extension()
	dateCreated := mainFile.DateCreated()

	if file, err := entity.FindFileByHash(imp.conf.Db(), mediaFile.Hash()); err == nil {
		existingFilename := imp.conf.OriginalsPath() + string(os.PathSeparator) + file.FileName
		return existingFilename, fmt.Errorf("\"%s\" is identical to \"%s\" (%s)", mediaFile.Filename(), file.FileName, mediaFile.Hash())
	}

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	pathName := imp.originalsPath() + string(os.PathSeparator) + dateCreated.UTC().Format("2006/01")

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
