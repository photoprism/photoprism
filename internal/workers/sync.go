package workers

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/dustin/go-humanize/english"
	"gorm.io/gorm"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/service"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// Sync represents a sync worker.
type Sync struct {
	conf *config.Config
}

// NewSync returns a new sync worker.
func NewSync(conf *config.Config) *Sync {
	return &Sync{
		conf: conf,
	}
}

// logErr logs an error message if err is not nil.
func (w *Sync) logErr(err error) {
	if err != nil {
		log.Errorf("sync: %s", clean.Error(err))
	}
}

// logAccountErr logs an error message that names the account if err is not nil.
func (w *Sync) logAccountErr(err error, a entity.Service) {
	if err != nil {
		log.Errorf("sync: %s (%s)", clean.Error(err), clean.Log(a.AccName))
	}
}

// logFileErr logs an error message that names a file and its account if err is not nil.
func (w *Sync) logFileErr(err error, name string, a entity.Service) {
	if err != nil {
		log.Errorf("sync: %s (%s from %s)", clean.Error(err), clean.Log(name), clean.Log(a.AccName))
	}
}

// logWarn logs a warning message if err is not nil.
func (w *Sync) logWarn(err error) {
	if err != nil {
		log.Warnf("sync: %s", clean.Error(err))
	}
}

// Start starts the sync worker.
func (w *Sync) Start() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("sync: %s (worker panic)", clean.Log(fmt.Sprint(r)))
			log.Error(err)
			log.Debugf("sync: stack of the worker panic: %s", debug.Stack())
		}
	}()

	if err := mutex.SyncWorker.Start(); err != nil {
		return err
	}

	defer mutex.SyncWorker.Stop()

	f := form.SearchServices{
		Sync: true,
	}

	accounts, err := search.Accounts(f)

	for _, a := range accounts {
		if a.AccType != service.WebDAV {
			continue
		}

		// Failed too often? The stored row decides, since an admin may have changed it during this run,
		// and the account is skipped either way, so its loaded copy is never written back.
		if a.RetryLimit > 0 && a.AccErrors > a.RetryLimit {
			if res := entity.Db().Model(&entity.Service{}).
				Where("id = ? AND acc_sync = TRUE AND retry_limit > 0 AND acc_errors > retry_limit", a.ID).
				UpdateColumn("acc_sync", false); res.Error != nil {
				w.logAccountErr(res.Error, a)
			} else if res.RowsAffected > 0 {
				log.Warnf("sync: disabled sync, %s failed more than %d times", clean.Log(a.AccName), a.RetryLimit)
			}

			continue
		}

		// Values updated in account: AccError, AccErrors, SyncStatus, SyncDate
		accError := a.AccError
		accErrors := a.AccErrors
		syncStatus := a.SyncStatus
		syncDate := a.SyncDate
		synced, reset := false, false

		switch a.SyncStatus {
		case entity.SyncStatusRefresh:
			if complete, err := w.refresh(a); err != nil {
				accErrors++
				accError = clean.ErrorBytes(err, txt.ClipError)
			} else if complete {
				uploadStages.Start(a.ID)

				// Uploads set aside in the previous cycle keep their count, so a remote that refuses
				// every upload reaches the retry limit.
				if !a.SyncUpload || !isUploadSetAsideError(a.AccError) {
					reset = true
					accErrors = 0
					accError = ""
				}

				switch {
				case a.SyncDownload:
					syncStatus = entity.SyncStatusDownload
				case a.SyncUpload:
					syncStatus = entity.SyncStatusUpload
				default:
					syncStatus = entity.SyncStatusSynced
					syncDate.Time = time.Now()
					syncDate.Valid = true
				}
			}
		case entity.SyncStatusDownload:
			if complete, downloadErr := w.downloadIfEnabled(a); downloadErr != nil {
				accErrors++
				accError = clean.ErrorBytes(downloadErr, txt.ClipError)
				syncStatus = entity.SyncStatusRefresh
			} else if complete {
				if a.SyncUpload {
					syncStatus = entity.SyncStatusUpload
				} else {
					synced = true
					syncStatus = entity.SyncStatusSynced
					syncDate.Time = time.Now()
					syncDate.Valid = true
				}
			}
		case entity.SyncStatusUpload:
			// A stage whose direction was switched off ends without transfers or a report of set-aside uploads.
			if !a.SyncUpload {
				uploadStages.End(a.ID)
				synced, syncStatus = true, entity.SyncStatusSynced
				syncDate.Time, syncDate.Valid = time.Now(), true

				if isUploadSetAsideError(a.AccError) {
					reset, accErrors, accError = true, 0, ""
				}
			} else if complete, uploadErr := w.upload(a); uploadErr != nil {
				accErrors++
				accError = clean.ErrorBytes(uploadErr, txt.ClipError)
				syncStatus = entity.SyncStatusRefresh
			} else if complete {
				syncStatus = entity.SyncStatusSynced
				syncDate.Time = time.Now()
				syncDate.Valid = true

				// Uploads set aside are reported as an account error rather than as synced, and a stage that
				// uploaded nothing because of the remote counts as failed. A stage that uploaded or tried no file
				// resets the count, and others, such as one started before a restart, keep it.
				stage, known := uploadStages.End(a.ID)

				if setAside, countErr := entity.CountFailedFileSyncUploads(a.ID); countErr != nil {
					w.logAccountErr(countErr, a)
				} else if setAside > 0 {
					accError = uploadSetAsideError(setAside)

					switch {
					case known && (stage.uploaded || !stage.tried):
						reset, accErrors = true, 0
					case known && stage.failed:
						accErrors++
					}
				} else {
					synced = true

					if isUploadSetAsideError(a.AccError) {
						reset, accErrors, accError = true, 0, ""
					}
				}
			}
		case entity.SyncStatusSynced:
			if syncDue(a, time.Now()) {
				syncStatus = entity.SyncStatusRefresh
			}
		default:
			syncStatus = entity.SyncStatusRefresh
		}

		if mutex.SyncWorker.Canceled() {
			return nil
		}

		// Errors are counted on the stored value, so a reset or a count saved during the run is kept.
		if accErrors > a.AccErrors {
			w.logAccountErr(entity.Db().Model(&entity.Service{}).Where("id = ?", a.ID).UpdateColumns(entity.Values{
				"acc_errors": gorm.Expr("COALESCE(acc_errors, 0) + 1"), "acc_error": accError}).Error, a)
		} else if reset {
			w.logAccountErr(entity.Db().Model(&entity.Service{}).Where("id = ?", a.ID).UpdateColumns(entity.Values{
				"acc_errors": accErrors, "acc_error": accError}).Error, a)
		} else if accError != a.AccError {
			w.logAccountErr(entity.Db().Model(&entity.Service{}).Where("id = ?", a.ID).
				UpdateColumn("acc_error", accError).Error, a)
		}

		if syncStatus == a.SyncStatus && syncDate.Valid == a.SyncDate.Valid && syncDate.Time.Equal(a.SyncDate.Time) {
			continue
		}

		// The status is written only while the stored status, folder and directions are those loaded, so a run
		// never advances with settings saved meanwhile. Rows of earlier releases may hold NULL instead.
		if res := entity.Db().Model(&entity.Service{}).
			Where("id = ? AND COALESCE(sync_status, '') = ?", a.ID, a.SyncStatus).
			Where("(sync_path = ? OR sync_path IS NULL AND ? = '') AND COALESCE(sync_download, FALSE) = ? AND COALESCE(sync_upload, FALSE) = ?",
				a.SyncPath, a.SyncPath, a.SyncDownload, a.SyncUpload).
			UpdateColumns(entity.Values{"sync_status": syncStatus, "sync_date": syncDate}); res.Error != nil {
			w.logAccountErr(res.Error, a)
		} else if res.RowsAffected == 0 {
			log.Debugf("sync: kept the status saved to %s during the sync", clean.Log(a.AccName))
		} else if synced {
			event.Publish("sync.synced", event.Data{"account": a})
		}
	}

	return err
}

// downloadIfEnabled runs the download stage of an account, which ends at once if downloads are switched off.
func (w *Sync) downloadIfEnabled(a entity.Service) (complete bool, err error) {
	if !a.SyncDownload {
		return true, nil
	}

	return w.download(a)
}

// transferStallLimit is the number of transfers of one account that may stall in a run before the worker stops
// transferring its files until the next run, so that a remote cannot hold the worker, and with it every other
// account, for a whole batch.
const transferStallLimit = 3

// stallBudget counts the transfers of one account that stalled in a run.
type stallBudget int

// Add counts a transfer if it stalled, and reports whether the limit of stalled transfers was reached.
func (s *stallBudget) Add(stalled bool) bool {
	if stalled {
		*s++
	}

	return int(*s) >= transferStallLimit
}

// Warn logs that the worker stopped transferring files of an account in this run because transfers stalled.
func (s *stallBudget) Warn(worker, account string) {
	log.Warnf("%s: stopped transferring files of %s in this run because %d transfers stalled", worker, clean.Log(account), int(*s))
}

// syncNameTooLong reports whether a remote file cannot be downloaded because its name is longer than a file sync
// record holds, or contains a file or folder name longer than local file systems accept.
func syncNameTooLong(name string) bool {
	return entity.RemoteNameTooLong(name) || fs.NameTooLong(name)
}

// skippedNames collects the names a sync run skips, so they are reported in one warning.
type skippedNames struct {
	count int
	names []string
}

// Add counts a skipped name and keeps it for the warning, up to the number of names a log line shows.
func (s *skippedNames) Add(name string) {
	s.count++

	if len(s.names) < clean.LogNamesLimit {
		s.names = append(s.names, name)
	}
}

// Warn logs one warning if names were skipped, with the plural of noun for their number, the account name,
// and the names that were kept, in this order.
func (s *skippedNames) Warn(format, noun, account string) {
	if s.count == 0 {
		return
	}

	names := clean.LogNames(s.names)

	if more := s.count - len(s.names); more > 0 {
		names = fmt.Sprintf("%s and %d more", names, more)
	}

	log.Warnf(format, english.Plural(s.count, noun, noun+"s"), clean.Log(account), names)
}

// syncDue reports whether a synced account is due for a refresh, comparing whole seconds so no stored interval overflows.
func syncDue(a entity.Service, now time.Time) bool {
	return a.SyncInterval > 0 && a.SyncDate.Valid && a.SyncDate.Time.Unix() < now.Unix()-int64(a.SyncInterval)
}
