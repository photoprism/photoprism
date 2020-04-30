package photoprism

import (
	"errors"
	"fmt"
	"os"
	"path"
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
	conf    *config.Config
	index   *Index
	convert *Convert
}

// NewImport returns a new importer and expects its dependencies as arguments.
func NewImport(conf *config.Config, index *Index, convert *Convert) *Import {
	instance := &Import{
		conf:    conf,
		index:   index,
		convert: convert,
	}

	return instance
}

// originalsPath returns the original media files path as string.
func (imp *Import) originalsPath() string {
	return imp.conf.OriginalsPath()
}

// Start imports media files from a directory and converts/indexes them as needed.
func (imp *Import) Start(opt ImportOptions) {
	var directories []string
	done := make(map[string]bool)
	ind := imp.index
	importPath := opt.Path

	if !fs.PathExists(importPath) {
		event.Error(fmt.Sprintf("import: %s does not exist", importPath))
		return
	}

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
			ImportWorker(jobs)
			wg.Done()
		}()
	}

	indexOpt := IndexOptionsAll()

	err := filepath.Walk(importPath, func(fileName string, fileInfo os.FileInfo, err error) error {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("import: %s [panic]", err)
			}
		}()

		if mutex.Worker.Canceled() {
			return errors.New("import canceled")
		}

		if err != nil || done[fileName] {
			return nil
		}

		if fileInfo.IsDir() {
			if fileName != importPath {
				directories = append(directories, fileName)
			}

			return nil
		}

		if strings.HasPrefix(filepath.Base(fileName), ".") {
			done[fileName] = true

			if !opt.RemoveDotFiles {
				return nil
			}

			if err := os.Remove(fileName); err != nil {
				log.Errorf("import: could not remove \"%s\" (%s)", fileName, err.Error())
			}

			return nil
		}

		mf, err := NewMediaFile(fileName)

		if err != nil || !mf.IsPhoto() {
			return nil
		}

		related, err := mf.RelatedFiles(imp.conf.Settings().Library.GroupRelated)

		if err != nil {
			event.Error(fmt.Sprintf("import: %s", err.Error()))

			return nil
		}

		var files MediaFiles

		for _, f := range related.Files {
			if done[f.FileName()] {
				continue
			}

			files = append(files, f)
			done[f.FileName()] = true
		}

		done[mf.FileName()] = true

		related.Files = files

		jobs <- ImportJob{
			FileName:  fileName,
			Related:   related,
			IndexOpt:  indexOpt,
			ImportOpt: opt,
			Imp:       imp,
		}

		return nil
	})

	close(jobs)
	wg.Wait()

	sort.Slice(directories, func(i, j int) bool {
		return len(directories[i]) > len(directories[j])
	})

	if opt.RemoveEmptyDirectories {
		// Remove empty directories from import path
		for _, directory := range directories {
			if fs.IsEmpty(directory) {
				if err := os.Remove(directory); err != nil {
					log.Errorf("import: could not deleted empty directory %s (%s)", directory, err)
				} else {
					log.Infof("import: deleted empty directory %s", directory)
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

	if !mediaFile.IsSidecar() {
		if f, err := entity.FirstFileByHash(mediaFile.Hash()); err == nil {
			existingFilename := imp.conf.OriginalsPath() + string(os.PathSeparator) + f.FileName
			return existingFilename, fmt.Errorf("\"%s\" is identical to \"%s\" (%s)", mediaFile.FileName(), f.FileName, mediaFile.Hash())
		}
	}

	//	Mon Jan 2 15:04:05 -0700 MST 2006
	pathName := path.Join(imp.originalsPath(), dateCreated.Format("2006/01"))

	iteration := 0

	result := path.Join(pathName, fileName+fileExtension)

	for fs.FileExists(result) {
		if mediaFile.Hash() == fs.Hash(result) {
			return result, fmt.Errorf("file already exists: %s", result)
		}

		iteration++

		result = path.Join(pathName, fileName+"."+fmt.Sprintf("%04d", iteration)+fileExtension)
	}

	return result, nil
}
