package workers

import (
	"fmt"
	"os"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/internal/service/webdav"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
)

type Downloads map[string][]entity.FileSync

// downloadPath returns a temporary download path.
func (w *Sync) downloadPath() string {
	return w.conf.TempPath() + "/sync"
}

// relatedDownloads returns files to be downloaded grouped by prefix.
func (w *Sync) relatedDownloads(a entity.Service) (result Downloads, err error) {
	result = make(Downloads)
	maxResults := 1000

	// Get remote files from database
	files, err := query.FileSyncs(a.ID, entity.FileSyncNew, maxResults)

	if err != nil {
		return result, err
	}

	// Group results by directory and base name
	for i, file := range files {
		k := fs.AbsPrefix(file.RemoteName, w.conf.Settings().StackSequences())

		result[k] = append(result[k], file)

		// Skip last 50 to make sure we see all related files
		if i > (maxResults - 50) {
			return result, nil
		}
	}

	return result, nil
}

// Downloads remote files in batches and imports / indexes them
func (w *Sync) download(a entity.Service) (complete bool, err error) {
	// Set up index worker
	indexJobs := make(chan photoprism.IndexJob)

	go photoprism.IndexWorker(indexJobs)
	defer close(indexJobs)

	// Set up import worker
	importJobs := make(chan photoprism.ImportJob)
	go photoprism.ImportWorker(importJobs)
	defer close(importJobs)

	relatedFiles, err := w.relatedDownloads(a)

	if err != nil {
		w.logErr(err)
		return false, err
	}

	// Check if files must and can be downloaded.
	if l := len(relatedFiles); l == 0 {
		log.Infof("sync: no files to download from %s", clean.Log(a.AccName))
		event.Publish("sync.downloaded", event.Data{"account": a})
		return true, nil
	} else if w.conf.ReadOnly() {
		err = fmt.Errorf("failed to download %s from %s because read-only mode is enabled",
			english.Plural(l, "file", "files"),
			clean.Log(a.AccName))
		log.Errorf("sync: %s", err)
		event.Publish("sync.downloaded", event.Data{"account": a, "error": err.Error()})
		return true, nil
	}

	// Display log message.
	log.Infof("sync: downloading from %s", a.AccName)

	client, err := webdav.NewClient(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout))

	if err != nil {
		return false, err
	}

	var baseDir string

	if a.SyncFilenames {
		baseDir = w.conf.OriginalsPath()
	} else {
		baseDir = fmt.Sprintf("%s/%d", w.downloadPath(), a.ID)
	}

	done := make(map[string]bool)

	for _, files := range relatedFiles {
		for i, file := range files {
			if mutex.SyncWorker.Canceled() {
				return false, nil
			}

			// Failed too often?
			if a.RetryLimit > 0 && file.Errors > a.RetryLimit {
				log.Debugf("sync: downloading %s failed more than %d times", file.RemoteName, a.RetryLimit)
				continue
			}

			localName := baseDir + file.RemoteName

			if _, err = os.Stat(localName); err == nil {
				log.Warnf("sync: download skipped, %s already exists", localName)
				file.Status = entity.FileSyncExists
				file.Error = ""
				file.Errors = 0
			} else {
				if err = client.Download(file.RemoteName, localName, false); err != nil {
					file.Errors++
					file.Error = err.Error()
				} else {
					log.Infof("sync: downloaded %s from %s", file.RemoteName, a.AccName)
					file.Status = entity.FileSyncDownloaded
					file.Error = ""
					file.Errors = 0
				}

				if mutex.SyncWorker.Canceled() {
					return false, nil
				}
			}

			if err = entity.Db().Save(&file).Error; err != nil {
				w.logErr(err)
			} else {
				files[i] = file
			}
		}

		for _, file := range files {
			if file.Status != entity.FileSyncDownloaded {
				continue
			}

			mf, err := photoprism.NewMediaFile(baseDir + file.RemoteName)

			if err != nil || !mf.IsMedia() || mf.Empty() {
				continue
			}

			related, err := mf.RelatedFiles(w.conf.Settings().StackSequences())

			if err != nil {
				w.logWarn(err)
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
					Ind:      get.Index(),
				}
			} else {
				log.Infof("sync: importing %s and related files", file.RemoteName)
				importJobs <- photoprism.ImportJob{
					FileName:  mf.FileName(),
					Related:   related,
					IndexOpt:  photoprism.IndexOptionsAll(),
					ImportOpt: photoprism.ImportOptionsMove(baseDir, w.conf.ImportDest()),
					Imp:       get.Import(),
				}
			}
		}
	}

	// Any files downloaded?
	if len(done) > 0 {
		// Update precalculated photo and file counts.
		w.logWarn(entity.UpdateCounts())

		// Update album, subject, and label cover thumbs.
		w.logWarn(query.UpdateCovers())
	}

	return false, nil
}
