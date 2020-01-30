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
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/fs"
)

// Import represents an importer that can copy/move MediaFiles to the originals directory.
type Import struct {
	conf                   *config.Config
	index                  *Index
	convert                *Convert
	removeDotFiles         bool
	removeExistingFiles    bool
	removeEmptyDirectories bool
}

// NewImport returns a new importer and expects its dependencies as arguments.
func NewImport(conf *config.Config, index *Index, convert *Convert) *Import {
	instance := &Import{
		conf:                   conf,
		index:                  index,
		convert:                convert,
		removeDotFiles:         true,
		removeExistingFiles:    true,
		removeEmptyDirectories: true,
	}

	return instance
}

func (imp *Import) originalsPath() string {
	return imp.conf.OriginalsPath()
}

// Start imports MediaFiles from a directory and converts/indexes them as needed.
func (imp *Import) Start(importPath string) {
	var directories []string
	done := make(map[string]bool)
	ind := imp.index

	if err := mutex.Worker.Start(); err != nil {
		event.Error(fmt.Sprintf("import: %s", err.Error()))
		return
	}

	defer mutex.Worker.Stop()

	if err := ind.tensorFlow.Init(); err != nil {
		log.Errorf("import: %s", err.Error())
		return
	}

	jobs := make(chan ImportJob)

	// Start a fixed number of goroutines to import files.
	var wg sync.WaitGroup
	var numWorkers = ind.conf.Workers()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			importWorker(jobs)
			wg.Done()
		}()
	}

	options := IndexOptionsAll()

	err := filepath.Walk(importPath, func(filename string, fileInfo os.FileInfo, err error) error {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("import: %s [panic]", err)
			}
		}()

		if mutex.Worker.Canceled() {
			return errors.New("import canceled")
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

		mf, err := NewMediaFile(filename)

		if err != nil || !mf.IsPhoto() {
			return nil
		}

		related, err := mf.RelatedFiles()

		if err != nil {
			event.Error(fmt.Sprintf("import: %s", err.Error()))

			return nil
		}

		var files MediaFiles

		for _, f := range related.files {
			if done[f.Filename()] {
				continue
			}

			files = append(files, f)
			done[f.Filename()] = true
		}

		done[mf.Filename()] = true

		related.files = files

		jobs <- ImportJob{
			filename: filename,
			related: related,
			opt:     options,
			path:    importPath,
			imp:     imp,
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
			if fs.IsEmpty(directory) {
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
func (imp *Import) Cancel() {
	mutex.Worker.Cancel()
}

// DestinationFilename returns the destination filename of a MediaFile to be imported.
func (imp *Import) DestinationFilename(mainFile *MediaFile, mediaFile *MediaFile) (string, error) {
	fileName := mainFile.CanonicalName()
	fileExtension := mediaFile.Extension()
	dateCreated := mainFile.DateCreated()

	if f, err := entity.FindFileByHash(imp.conf.Db(), mediaFile.Hash()); err == nil {
		existingFilename := imp.conf.OriginalsPath() + string(os.PathSeparator) + f.FileName
		return existingFilename, fmt.Errorf("\"%s\" is identical to \"%s\" (%s)", mediaFile.Filename(), f.FileName, mediaFile.Hash())
	}

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	pathName := imp.originalsPath() + string(os.PathSeparator) + dateCreated.UTC().Format("2006/01")

	iteration := 0

	result := pathName + string(os.PathSeparator) + fileName + fileExtension

	for fs.FileExists(result) {
		if mediaFile.Hash() == fs.Hash(result) {
			return result, fmt.Errorf("file already exists: %s", result)
		}

		iteration++

		result = pathName + string(os.PathSeparator) + fileName + "." + fmt.Sprintf("edited_%d", iteration) + fileExtension
	}

	return result, nil
}
