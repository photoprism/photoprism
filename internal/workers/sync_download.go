package workers

import (
	"errors"
	"fmt"
	"os"
	"strings"

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
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/txt"
)

// downloadLimit is the number of queued files a download run reads, of which the last 48 are left to the next
// run, so that related files are not split. Tests may lower it to no less than 51.
var downloadLimit = 1000

// syncIndexWorker and syncImportWorker process the downloaded files; tests replace them to observe the jobs.
var (
	syncIndexWorker  = photoprism.IndexWorker
	syncImportWorker = photoprism.ImportWorker
)

// Downloads groups files to sync by their directory/prefix.
type Downloads map[string][]entity.FileSync

// downloadPath returns a temporary download path.
func (w *Sync) downloadPath() string {
	return w.conf.TempPath() + "/sync"
}

// relatedDownloads returns files to be downloaded grouped by prefix.
func (w *Sync) relatedDownloads(a entity.Service) (result Downloads, err error) {
	result = make(Downloads)
	maxResults := downloadLimit

	// Get list of remote files from database.
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

// download transfers eligible remote files in batches for import and indexing.
func (w *Sync) download(a entity.Service) (complete bool, err error) {
	if webdav.SkipSyncPath(a.SyncPath) {
		log.Tracef("sync: skipping excluded path %s for service %s (download)", clean.Log(a.SyncPath), clean.Log(a.AccName))
		return true, nil
	}
	// Set up index worker
	indexJobs := make(chan photoprism.IndexJob)

	go syncIndexWorker(indexJobs)
	defer close(indexJobs)

	// Set up import worker
	importJobs := make(chan photoprism.ImportJob)
	go syncImportWorker(importJobs)
	defer close(importJobs)

	relatedFiles, err := w.relatedDownloads(a)

	if err != nil {
		w.logAccountErr(err, a)
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

	// The stage completes once a run makes no progress, see below.
	before, err := query.CountQueuedFileSyncs(a.ID, a.RetryLimit)

	if err != nil {
		return false, err
	}

	// Display log message.
	log.Infof("sync: downloading from %s", clean.Log(a.AccName))

	client, err := webdav.NewClient(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout), w.conf.ServicesCIDR())

	if err != nil {
		return false, err
	}

	// Bound each download to the originals size limit so a remote sync target
	// cannot overrun local storage with a single oversized file.
	client.SetDownloadLimit(w.conf.OriginalsLimitBytes())

	var baseDir string

	if a.SyncFilenames {
		baseDir = w.conf.OriginalsPath()
	} else {
		baseDir = fmt.Sprintf("%s/%d", w.downloadPath(), a.ID)
	}

	done := make(map[string]bool)

	var stalls stallBudget

	stopped := false

	for _, files := range relatedFiles {
		if w.conf.InsufficientStorage() {
			log.Warnf("sync: skipped downloading files from %s due to insufficient storage", clean.Log(a.AccName))
			return false, nil
		}

		for i, file := range files {
			if mutex.SyncWorker.Canceled() {
				return false, nil
			}

			if webdav.SkipSyncPath(file.RemoteName) {
				log.Debugf("sync: skipping excluded path %s", clean.Log(file.RemoteName))
				file.Status, file.Error, file.Errors = entity.FileSyncIgnore, "", 0
				w.logFileErr(entity.Db().Save(&file).Error, file.RemoteName, a)
				files[i] = file
				continue
			}

			if syncNameTooLong(file.RemoteName) {
				log.Debugf("sync: skipping %s from %s because its name is too long", clean.Log(file.RemoteName), clean.Log(a.AccName))
				file.Status, file.Error, file.Errors = entity.FileSyncIgnore, "", 0
				w.logFileErr(entity.Db().Save(&file).Error, file.RemoteName, a)
				files[i] = file
				continue
			}

			if !a.SyncYamlEnabled() && fs.FileType(file.RemoteName) == fs.SidecarYaml {
				file.Status, file.Error, file.Errors = entity.FileSyncIgnore, "", 0
				w.logFileErr(entity.Db().Save(&file).Error, file.RemoteName, a)
				files[i] = file
				continue
			}

			// Raw images and videos queued while the option was on are not downloaded once it is off.
			if content := media.FromName(file.RemoteName); !a.SyncRaw && (content == media.Raw || content == media.Video) {
				file.Status, file.Error, file.Errors = entity.FileSyncIgnore, "", 0
				w.logFileErr(entity.Db().Save(&file).Error, file.RemoteName, a)
				files[i] = file
				continue
			}

			// Failed too often?
			if a.RetryLimit > 0 && file.Errors > a.RetryLimit {
				log.Debugf("sync: downloading %s from %s failed more than %d times", clean.Log(file.RemoteName), clean.Log(a.AccName), a.RetryLimit)
				continue
			}

			// Resolve the local destination safely so a remote sync target cannot
			// escape the sync base directory via a crafted parent-directory path.
			localName, joinErr := fs.SafeJoin(baseDir, strings.TrimPrefix(file.RemoteName, "/"))

			if joinErr != nil {
				log.Warnf("sync: skipped download of %s from %s because the remote path is invalid", clean.Log(file.RemoteName), clean.Log(a.AccName))
				file.Status = entity.FileSyncFailed
				file.Error = "invalid remote path"
				file.Errors++

				if err = entity.Db().Save(&file).Error; err != nil {
					w.logFileErr(err, file.RemoteName, a)
				} else {
					files[i] = file
				}

				continue
			}

			stalled := false

			if _, err = os.Stat(localName); err == nil {
				log.Warnf("sync: skipped download of %s from %s because local file %s already exists", clean.Log(file.RemoteName), clean.Log(a.AccName), clean.Log(localName))
				file.Status = entity.FileSyncExists
				file.Error = ""
				file.Errors = 0
			} else {
				// A listed size above the limit fails as the download would, without transferring the file.
				if err = client.CheckDownloadSize(localName, file.RemoteSize); err == nil {
					err = client.Download(file.RemoteName, localName, false)
				}

				if errors.Is(err, os.ErrExist) {
					log.Infof("sync: skipped download of %s from %s because a local file was created meanwhile", clean.Log(file.RemoteName), clean.Log(a.AccName))
					file.Status = entity.FileSyncExists
					file.Error = ""
					file.Errors = 0
				} else if err != nil {
					stalled = errors.Is(err, webdav.ErrStalled)
					file.Errors++
					file.Error = clean.ErrorBytes(err, txt.ClipError)

					if a.RetryLimit > 0 && file.Errors > a.RetryLimit {
						file.Status = entity.FileSyncFailed
					}
				} else {
					log.Infof("sync: downloaded %s from %s", clean.Log(file.RemoteName), clean.Log(a.AccName))
					file.Status = entity.FileSyncDownloaded
					file.Error = ""
					file.Errors = 0
				}

				if mutex.SyncWorker.Canceled() {
					return false, nil
				}
			}

			if err = entity.Db().Save(&file).Error; err != nil {
				w.logFileErr(err, file.RemoteName, a)
			} else {
				files[i] = file
			}

			// Transfers stop after too many stalled, while the files of this group already downloaded are still imported.
			if stalls.Add(stalled) {
				stalls.Warn("sync", a.AccName)
				stopped = true
				break
			}
		}

		for _, file := range files {
			if file.Status != entity.FileSyncDownloaded {
				continue
			}

			localName, joinErr := fs.SafeJoin(baseDir, strings.TrimPrefix(file.RemoteName, "/"))

			if joinErr != nil {
				continue
			}

			mf, err := photoprism.NewMediaFile(localName)

			if err != nil || !mf.IsMedia() || mf.Empty() {
				continue
			}

			related, err := mf.RelatedFiles(w.conf.Settings().StackSequences())

			if err != nil {
				log.Warnf("sync: %s (%s from %s)", clean.Error(err), clean.Log(file.RemoteName), clean.Log(a.AccName))
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
				log.Infof("sync: indexing %s and related files", clean.Log(file.RemoteName))
				indexJobs <- photoprism.IndexJob{
					FileName: mf.FileName(),
					Related:  related,
					IndexOpt: photoprism.IndexOptionsAll(w.conf),
					Ind:      get.Index(),
				}
			} else {
				log.Infof("sync: importing %s and related files", clean.Log(file.RemoteName))
				importJobs <- photoprism.ImportJob{
					FileName:  mf.FileName(),
					Related:   related,
					IndexOpt:  photoprism.IndexOptionsAll(w.conf),
					ImportOpt: photoprism.ImportOptionsMove(baseDir, w.conf.ImportDest()),
					Imp:       get.Import(),
				}
			}
		}

		if stopped {
			break
		}
	}

	// Any files downloaded?
	if len(done) > 0 {
		// Update precalculated photo and file counts.
		w.logWarn(entity.UpdateCounts())

		// Update album, subject, and label cover thumbs.
		w.logWarn(query.UpdateCovers())
	}

	after, err := query.CountQueuedFileSyncs(a.ID, a.RetryLimit)

	if err != nil {
		return false, err
	}

	// The stage stays open while a run changes the status of a queued file, or tries files that may be tried again:
	// without a retry limit, failed files stay new and are tried again in the next sync, so only files not tried yet
	// count, which the next run reads first; with one, files are tried in each run until they exceed it. A file
	// whose outcome cannot be saved therefore keeps the stage open only as long as others make progress.
	if a.RetryLimit <= 0 {
		if after.Queued < before.Queued || after.Untried > 0 && after.Untried < before.Untried {
			return false, nil
		}

		log.Warnf("sync: failed to download %s from %s, retrying in the next sync", english.Plural(int(after.Failed), "file", "files"), clean.Log(a.AccName))

		return true, nil
	} else if after.Retryable > 0 && (after.Queued < before.Queued || after.Errors > before.Errors) {
		return false, nil
	} else if after.Retryable > 0 {
		log.Warnf("sync: %s from %s still queued, retrying in the next sync", english.Plural(int(after.Retryable), "file", "files"), clean.Log(a.AccName))
	}

	return true, nil
}
