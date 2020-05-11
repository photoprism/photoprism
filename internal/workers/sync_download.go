package workers

import (
	"fmt"
	"os"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/internal/remote/webdav"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/fs"
)

type Downloads map[string][]entity.FileSync

// downloadPath returns a temporary download path.
func (s *Sync) downloadPath() string {
	return s.conf.TempPath() + "/sync"
}

// relatedDownloads returns files to be downloaded grouped by prefix.
func (s *Sync) relatedDownloads(a entity.Account) (result Downloads, err error) {
	result = make(Downloads)
	maxResults := 1000

	// Get remote files from database
	files, err := query.FileSyncs(a.ID, entity.FileSyncNew, maxResults)

	if err != nil {
		return result, err
	}

	// Group results by directory and base name
	for i, file := range files {
		k := fs.AbsBase(file.RemoteName, s.conf.Settings().Index.Group)

		result[k] = append(result[k], file)

		// Skip last 50 to make sure we see all related files
		if i > (maxResults - 50) {
			return result, nil
		}
	}

	return result, nil
}

// Downloads remote files in batches and imports / indexes them
func (s *Sync) download(a entity.Account) (complete bool, err error) {
	// Set up index worker
	indexJobs := make(chan photoprism.IndexJob)

	go photoprism.IndexWorker(indexJobs)
	defer close(indexJobs)

	// Set up import worker
	importJobs := make(chan photoprism.ImportJob)
	go photoprism.ImportWorker(importJobs)
	defer close(importJobs)

	relatedFiles, err := s.relatedDownloads(a)

	if err != nil {
		log.Errorf("sync: %s", err.Error())
		return false, err
	}

	if len(relatedFiles) == 0 {
		log.Infof("sync: download complete for %s", a.AccName)
		event.Publish("sync.downloaded", event.Data{"account": a})
		return true, nil
	}

	log.Infof("sync: downloading from %s", a.AccName)

	client := webdav.New(a.AccURL, a.AccUser, a.AccPass)

	var baseDir string

	if a.SyncFilenames {
		baseDir = s.conf.OriginalsPath()
	} else {
		baseDir = fmt.Sprintf("%s/%d", s.downloadPath(), a.ID)
	}

	done := make(map[string]bool)

	for _, files := range relatedFiles {
		for i, file := range files {
			if mutex.Sync.Canceled() {
				return false, nil
			}

			if file.Errors > a.RetryLimit {
				log.Debugf("sync: downloading %s failed more than %d times", file.RemoteName, a.RetryLimit)
				continue
			}

			localName := baseDir + file.RemoteName

			if _, err := os.Stat(localName); err == nil {
				log.Warnf("sync: download skipped, %s already exists", localName)
				file.Status = entity.FileSyncExists
			} else {
				if err := client.Download(file.RemoteName, localName, false); err != nil {
					log.Errorf("sync: %s", err.Error())
					file.Errors++
					file.Error = err.Error()
				} else {
					log.Infof("sync: downloaded %s from %s", file.RemoteName, a.AccName)
					file.Status = entity.FileSyncDownloaded
				}

				if mutex.Sync.Canceled() {
					return false, nil
				}
			}

			if err := entity.Db().Save(&file).Error; err != nil {
				log.Errorf("sync: %s", err.Error())
			} else {
				files[i] = file
			}
		}

		for _, file := range files {
			if file.Status != entity.FileSyncDownloaded {
				continue
			}

			mf, err := photoprism.NewMediaFile(baseDir + file.RemoteName)

			if err != nil || !mf.IsMedia() {
				continue
			}

			related, err := mf.RelatedFiles(s.conf.Settings().Index.Group)

			if err != nil {
				log.Warnf("sync: %s", err.Error())
				continue
			}

			var rf photoprism.MediaFiles

			for _, f := range related.Files {
				if done[f.FileName()] {
					continue
				}

				rf = append(rf, f)
				done[f.FileName()] = true
			}

			done[mf.FileName()] = true
			related.Files = rf

			if a.SyncFilenames {
				log.Infof("sync: indexing %s and related files", file.RemoteName)
				indexJobs <- photoprism.IndexJob{
					FileName: mf.FileName(),
					Related:  related,
					IndexOpt: photoprism.IndexOptionsAll(),
					Ind:      service.Index(),
				}
			} else {
				log.Infof("sync: importing %s and related files", file.RemoteName)
				importJobs <- photoprism.ImportJob{
					FileName:  mf.FileName(),
					Related:   related,
					IndexOpt:  photoprism.IndexOptionsAll(),
					ImportOpt: photoprism.ImportOptionsMove(baseDir),
					Imp:       service.Import(),
				}
			}
		}
	}

	if len(done) > 0 {
		if err := entity.UpdatePhotoCounts(); err != nil {
			log.Errorf("sync: %s", err)
		}
	}

	return false, nil
}
