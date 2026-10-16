package workers

import (
	"errors"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/photoprism"
	"github.com/photoprism/photoprism/internal/service/webdav"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/txt"
)

// upload transfers eligible local files to a remote account.
func (w *Sync) upload(a entity.Service) (complete bool, err error) {
	if webdav.SkipSyncPath(a.SyncPath) {
		log.Tracef("sync: skipping excluded path %s for service %s (upload)", clean.Log(a.SyncPath), clean.Log(a.AccName))
		return true, nil
	}

	maxResults := 250

	// Files skipped for the length of their remote name below another sync path are checked again.
	w.logAccountErr(entity.ResetFileSyncLongNames(a.ID, a.SyncPath), a)

	// Files skipped for their names are checked again once their names are portable.
	w.logAccountErr(entity.ResetFileSyncUnportable(a.ID), a)

	// Get upload file list from database
	files, err := query.AccountUploads(a, maxResults)

	if err != nil {
		return false, err
	}

	if len(files) == 0 {
		log.Infof("sync: upload complete for %s", clean.Log(a.AccName))
		event.Publish("sync.uploaded", event.Data{"account": a})
		return true, nil
	}

	client, err := webdav.NewClient(a.AccURL, a.AccUser, a.AccPass, webdav.Timeout(a.AccTimeout), w.conf.ServicesCIDR())

	if err != nil {
		return false, err
	}

	// A YAML file refused with 403 disables YAML sync unless the remote also refused another file in this stage.
	stage := uploadStages.Get(a.ID)
	var stalls stallBudget

	// Files whose remote names are too long are reported in one warning per run.
	var skipped skippedNames

	defer skipped.Warn("sync: skipped upload of %s to %s because their remote names are too long (%s)", "file", a.AccName)

	// Files whose names are not portable are not uploaded, and reported in one warning per run.
	var unsupported skippedNames

	defer unsupported.Warn("sync: skipped upload of %s to %s because their names contain characters that are not supported (%s)", "file", a.AccName)

	// Files whose upload failed are set aside until the next sync cycle and reported in one warning per run.
	var failed skippedNames

	defer failed.Warn("sync: failed to upload %s to %s, retrying in the next sync (%s)", "file", a.AccName)

	for _, file := range files {
		if mutex.SyncWorker.Canceled() {
			return false, nil
		}

		if webdav.SkipSyncPath(file.FileName) {
			log.Debugf("sync: skipping excluded path %s", clean.Log(file.FileName))
			w.ignoreUpload(a, path.Join(fs.PPHiddenPathname, "sync"), file.ID)
			continue
		} else if !portableSyncName("/", file.FileName) {
			unsupported.Add(file.FileName)
			w.ignoreUpload(a, entity.FileSyncUnportablePrefix, file.ID)
			continue
		}

		yamlFile := fs.Type(file.FileType) == fs.SidecarYaml

		if yamlFile && (!a.SyncYamlEnabled() || stage.yamlRefused && !stage.otherRefused) {
			continue
		}

		fileName := photoprism.FileName(file.FileRoot, file.FileName)
		remoteName := path.Join(a.SyncPath, file.FileName)
		remoteDir := path.Dir(remoteName)

		// A remote name that is too long is recorded for the sync path, so the file is not selected again.
		if entity.RemoteNameTooLong(remoteName) {
			skipped.Add(file.FileName)
			w.ignoreUpload(a, entity.FileSyncLongNamePrefix(a.SyncPath), file.ID)
			continue
		}

		// Ensure remote folder exists.
		if err = client.MkdirAll(remoteDir); err != nil {
			log.Debugf("sync: %s (create remote folder for %s)", clean.Error(err), clean.Log(a.AccName))
		}

		// A failed upload is set aside until the next sync cycle, so each run moves on to other files.
		if err = client.Upload(fileName, remoteName); err != nil {
			log.Debugf("sync: %s (%s)", clean.Error(err), clean.Log(a.AccName))
			failed.Add(file.FileName)
			w.logAccountErr(entity.FailFileSyncUpload(a.ID, file.ID, clean.ErrorBytes(err, txt.ClipError)), a)

			// Only failures caused by the remote count towards the account's retry limit.
			stage.tried = true
			stalled, forbidden := errors.Is(err, webdav.ErrStalled), errors.Is(err, webdav.ErrForbidden)
			stage.failed = stage.failed || stalled || errors.Is(err, webdav.ErrUnavailable)
			stage.yamlRefused = stage.yamlRefused || forbidden && yamlFile
			stage.otherRefused = stage.otherRefused || forbidden && !yamlFile
			uploadStages.Set(a.ID, stage)

			if stalled && stalls.Add(true) {
				stalls.Warn("sync", a.AccName)
				return false, nil
			}

			continue
		}

		log.Infof("sync: uploaded %s to %s (%s)", clean.Log(file.FileName), clean.Log(remoteName), clean.Log(a.AccName))
		stage.tried, stage.uploaded = true, true
		uploadStages.Set(a.ID, stage)

		fileSync := entity.NewFileSync(a.ID, remoteName)
		fileSync.Status = entity.FileSyncUploaded
		fileSync.RemoteDate = time.Now()
		fileSync.RemoteSize = file.FileSize
		fileSync.FileID = file.ID
		fileSync.Error = ""
		fileSync.Errors = 0

		if mutex.SyncWorker.Canceled() {
			return false, nil
		}

		if err = entity.Db().Save(&fileSync).Error; err != nil {
			w.logFileErr(err, remoteName, a)
		} else {
			w.logAccountErr(entity.ClearFileSyncUpload(a.ID, file.ID), a)
		}
	}

	// The refusal is cleared once acted on, so YAML files are tried again if an admin turns the option back on.
	if stage.yamlRefused && !stage.otherRefused {
		log.Warnf("sync: disabled YAML sidecar files for %s because the remote server refused to store them", clean.Log(a.AccName))
		w.logAccountErr(a.Update("SyncYaml", -1), a)
		stage.yamlRefused = false
		uploadStages.Set(a.ID, stage)
	}

	return false, nil
}

// ignoreUpload records a local file as ignored for an account under a placeholder name in dir, so it is not
// selected for upload again.
func (w *Sync) ignoreUpload(a entity.Service, dir string, fileID uint) {
	ignored := entity.NewFileSync(a.ID, path.Join(dir, strconv.FormatUint(uint64(fileID), 10)))
	ignored.FileID, ignored.Status = fileID, entity.FileSyncIgnore
	w.logAccountErr(entity.Db().Save(ignored).Error, a)
}

// uploadSetAsideSuffix ends the account error that reports uploads set aside until the next sync cycle.
const uploadSetAsideSuffix = " could not be uploaded"

// uploadSetAsideError returns the account error that reports the number of uploads set aside.
func uploadSetAsideError(count int64) string {
	return english.Plural(int(count), "file", "files") + uploadSetAsideSuffix
}

// isUploadSetAsideError reports whether an account error reports uploads set aside.
func isUploadSetAsideError(accError string) bool {
	return strings.HasSuffix(accError, uploadSetAsideSuffix)
}

// uploadStage records what happened in the current upload stage of an account.
type uploadStage struct {
	tried        bool // A file was sent or failed to be sent.
	uploaded     bool // A file was uploaded.
	failed       bool // An upload failed because of the remote rather than the file.
	yamlRefused  bool // A YAML file was refused with 403.
	otherRefused bool // Another file was refused with 403.
	unknown      bool // The stage was started before a restart.
}

// uploadStageMap holds the upload stages of the accounts in this process.
type uploadStageMap struct {
	mu     sync.Mutex
	stages map[uint]uploadStage
}

// uploadStages tracks the upload stages started in this process, from the completed refresh to the end of the stage.
var uploadStages = uploadStageMap{stages: make(map[uint]uploadStage)}

// Start begins a new upload stage for an account.
func (m *uploadStageMap) Start(id uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stages[id] = uploadStage{}
}

// Get returns the current upload stage of an account. A stage started before a restart is unknown, and is
// returned as one in which another file was refused, so that it never turns off YAML sync.
func (m *uploadStageMap) Get(id uint) uploadStage {
	m.mu.Lock()
	defer m.mu.Unlock()

	if stage, ok := m.stages[id]; ok {
		return stage
	}

	return uploadStage{otherRefused: true, unknown: true}
}

// Set stores the current upload stage of an account.
func (m *uploadStageMap) Set(id uint, stage uploadStage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stages[id] = stage
}

// End ends the upload stage of an account and returns it, and whether it was started in this process.
func (m *uploadStageMap) End(id uint) (stage uploadStage, known bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stage, known = m.stages[id]
	delete(m.stages, id)

	return stage, known && !stage.unknown
}
