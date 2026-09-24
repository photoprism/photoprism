package photoprism

import (
	"fmt"
	"strings"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/fs/disk"
)

// runFacesReindex delegates face-only indexing to the supplied Index instance; tests may override it.
var runFacesReindex = func(index *Index, opt IndexOptions) (fs.Done, int, error) {
	if index == nil {
		return nil, 0, fmt.Errorf("faces: index service unavailable")
	}

	found, updated := index.Start(opt)
	return found, updated, nil
}

// Reset removes automatically added face clusters, marker matches, and dangling subjects.
func (w *Faces) Reset() (err error) {
	return w.reset(false, false)
}

// ResetAll additionally removes the clusters and matches a person or an XMP sidecar created, so a
// library returns to the state it had before any face was recognized. Markers and their embeddings
// are kept, which makes it far cheaper than detecting again; a name survives only where the person
// is flagged subjects.verified, which keeps their row.
func (w *Faces) ResetAll() (err error) {
	return w.reset(true, true)
}

// reset clears face recognition state, including what a person asserted when all is set, and the
// clusters a person created when clusters is set.
func (w *Faces) reset(all, clusters bool) (err error) {
	var removedMarkers int64
	var removedFaces int

	// Remove subject and face references from the markers table.
	if all {
		removedMarkers, err = query.ResetAllFaceMarkerMatches()
	} else {
		removedMarkers, err = query.ResetFaceMarkerMatches()
	}

	if err != nil {
		return fmt.Errorf("faces: %s (reset markers)", err)
	}

	log.Infof("faces: removed %d face matches", removedMarkers)

	// Remove face clusters from the index.
	if all || clusters {
		removedFaces, err = query.RemoveAllFaceClusters()
	} else {
		removedFaces, err = query.RemoveAutoFaceClusters()
	}

	if err != nil {
		return fmt.Errorf("faces: %s (reset faces)", err)
	}

	log.Infof("faces: removed %d face clusters", removedFaces)

	// Clear references to the clusters just deleted.
	//
	// The reset above clears a marker's face only where the subject was assigned automatically, so
	// a hand-named marker sitting on an automatic cluster keeps pointing at a row that no longer
	// exists. Measured on a real library, that was 11 of 12 hand-named markers.
	if removed, faceErr := query.RemoveNonExistentMarkerFaces(); faceErr != nil {
		return fmt.Errorf("faces: %s (reset marker faces)", faceErr)
	} else if removed > 0 {
		log.Infof("faces: cleared %d references to removed clusters", removed)
	}

	// Remove dangling marker subjects.
	if removed, subjErr := query.RemoveOrphanSubjects(); subjErr != nil {
		return fmt.Errorf("faces: %s (reset subjects)", subjErr)
	} else {
		log.Infof("faces: removed %d dangling subjects", removed)
	}

	// Remove the unverified people whose names were just cleared, whatever their source; a
	// soft-deleted person is restored when the same name is assigned again.
	if all {
		if removed, peopleErr := entity.DeleteOrphanPeople(); peopleErr != nil {
			return fmt.Errorf("faces: %s (reset people)", peopleErr)
		} else if removed > 0 {
			log.Infof("faces: removed %s", english.Plural(removed, "orphan person", "orphan people"))
		}
	}

	return nil
}

// ResetAndReindex resets face data and regenerates the markers with the specified detector, or
// resets only when none is named. Regenerating removes every cluster, and keeps the names on the
// markers unless all is set.
func (w *Faces) ResetAndReindex(detector string, index *Index, all bool) error {
	_, err := w.resetAndReindex(detector, index, all, "/")
	return err
}

// resetAndReindex resets face data and regenerates the markers of the files in path, and returns
// what the regeneration changed. Everything that can refuse the request is checked before anything
// is removed, so a request it cannot meet leaves the index as it was.
func (w *Faces) resetAndReindex(detector string, index *Index, all bool, path string) (*FaceRegeneration, error) {
	name := strings.TrimSpace(detector)

	if name != "" && !face.KnownDetectorName(name) {
		return nil, fmt.Errorf("faces: unsupported face detector %q", detector)
	}

	// The detector is what a caller names, because every one of them runs on the same runtime:
	// naming the runtime would not say which model places the landmarks, and those decide the crop.
	regenerate := name != "" && face.ParseDetectorName(name) != face.DetectorNone

	if !regenerate {
		return nil, w.reset(all, false)
	} else if w.conf == nil {
		return nil, fmt.Errorf("faces: configuration not available")
	}

	configured := w.conf.FaceDetector()

	// Auto keeps the configured detector rather than the one a blank setting would derive.
	if parsed := face.ParseDetectorName(name); parsed != face.DetectorAuto {
		w.conf.Options().FaceDetector = parsed
	}

	opt := IndexOptionsFacesOnly(w.conf)
	opt.Path = path
	opt.Convert = w.conf.Settings().Index.Convert && w.conf.SidecarWritable()
	opt.SkipArchived = false
	opt.RegenerateFaces = true
	opt.FaceRegeneration = &FaceRegeneration{}

	if err := w.regenerateRefused(index, opt); err != nil {
		return nil, err
	}

	// Detects at the floors a migration uses, so the markers an earlier detector placed are found
	// again. Loaded before the reset, so a detector that fails to load removes nothing.
	restoreDetector, err := w.useMigrationDetector()

	if err != nil {
		return nil, err
	}

	defer restoreDetector()

	if err = w.reset(all, true); err != nil {
		return nil, err
	}

	found, _, err := runFacesReindex(index, opt)

	if err != nil {
		return opt.FaceRegeneration, err
	}

	log.Infof("faces: regenerated markers in %s with detector %s: %s (%s scanned)",
		english.Plural(int(opt.FaceRegeneration.Files.Load()), "file", "files"), clean.Log(w.conf.FaceDetector()),
		opt.FaceRegeneration, english.Plural(len(found), "file", "files"))

	if w.conf.FaceDetector() != configured {
		log.Warnf("faces: set the face detector to %s in the configuration, so that new pictures and migrations use it as well",
			clean.Log(w.conf.FaceDetector()))
	}

	// The index skips files it cannot process without an error, and may stop early, so what it did
	// not regenerate is counted from the markers table, and split by whether the walk found the file.
	unreached, err := unregeneratedMarkers(w.conf, path, found, opt.FaceRegeneration)

	if err != nil {
		return opt.FaceRegeneration, err
	}

	if unreached.SkippedFiles > 0 {
		log.Warnf("faces: skipped %s in %s the index is set to skip, such as ignored or hidden files and disabled RAW files",
			english.Plural(unreached.SkippedMarkers, "marker", "markers"), english.Plural(unreached.SkippedFiles, "file", "files"))
	}

	if unreached.FailedFiles > 0 {
		log.Warnf("faces: skipped %s in %s because of errors, see the warnings above",
			english.Plural(unreached.FailedMarkers, "marker", "markers"), english.Plural(unreached.FailedFiles, "file", "files"))
	}

	if unreached.Files > 0 {
		return opt.FaceRegeneration, fmt.Errorf("faces: could not regenerate %s in %s the index did not reach, "+
			"so index or purge the library if originals were moved or removed, and run it again",
			english.Plural(unreached.Markers, "marker", "markers"), english.Plural(unreached.Files, "file", "files"))
	}

	return opt.FaceRegeneration, nil
}

// regenerateRefused returns the reason markers cannot be regenerated with the passed options, or
// nil if they can. It covers every reason the faces-only index would skip the run without an error.
func (w *Faces) regenerateRefused(index *Index, opt IndexOptions) error {
	switch {
	case w.conf.FaceDetector() == face.DetectorNone:
		return fmt.Errorf("faces: face detector %s cannot be used, so markers cannot be regenerated", clean.Log(string(w.conf.Options().FaceDetector)))
	case !opt.DetectFaces:
		return fmt.Errorf("faces: face detection is disabled, so markers cannot be regenerated")
	case face.EmbeddingsDisabled():
		return fmt.Errorf("faces: face embeddings are disabled, so markers cannot be regenerated")
	case face.EmbeddingsBlocked():
		return fmt.Errorf("faces: %s, so markers cannot be regenerated until the library is migrated", face.EmbeddingsBlockedReason())
	case w.conf.FacesLocked() != "":
		return fmt.Errorf("faces: waiting for the %s to complete", w.conf.FacesLocked())
	case mutex.IndexWorker.Running():
		return fmt.Errorf("faces: indexing is already running")
	case index == nil:
		return fmt.Errorf("faces: index service unavailable")
	}

	if originalsPath := index.originalsPath(); fs.DirIsEmpty(originalsPath) {
		return fmt.Errorf("faces: originals folder %s is empty or cannot be read", clean.Log(originalsPath))
	} else if indexPath, err := ResolveIndexPath(originalsPath, opt.Path); err != nil {
		return fmt.Errorf("faces: %s", clean.Error(err))
	} else if !fs.PathExists(indexPath) {
		return fmt.Errorf("faces: folder %s not found", clean.Log(indexPath))
	}

	disk.FlushFree()

	if index.storageLow() {
		return fmt.Errorf("faces: storage is low, so markers cannot be regenerated")
	}

	return nil
}

// unregeneratedFaceMarkers counts the face markers, and the files holding them, a regeneration did not
// reach, those in files the index was set to skip, and those in files it found but failed on.
type unregeneratedFaceMarkers struct {
	Markers        int
	Files          int
	SkippedMarkers int
	SkippedFiles   int
	FailedMarkers  int
	FailedFiles    int
}

// unregeneratedMarkers returns the face markers a regeneration of the folder did not regenerate: a
// file that failed or that the walk found was tried and the log names it, an existing file the index
// was set to skip is skipped, and any other was not reached.
func unregeneratedMarkers(conf *config.Config, dir string, found fs.Done, stats *FaceRegeneration) (result unregeneratedFaceMarkers, err error) {
	files, err := query.FaceMarkerFiles(dir)

	if err != nil {
		return result, err
	}

	unreached := make(map[string]query.FaceMarkerFile)

	for fileUID, f := range files {
		fileName := ConfigFileName(conf, f.FileRoot, f.FileName)
		_, walked := found[fileName]

		switch {
		case stats.Processed(fileUID):
		case stats.FileFailed(fileUID):
			result.FailedMarkers += f.Markers
			result.FailedFiles++
		case f.FileRoot == entity.RootOriginals && stats.Skipped(fileName) && fs.FileExists(fileName):
			result.SkippedMarkers += f.Markers
			result.SkippedFiles++
		case walked:
			result.FailedMarkers += f.Markers
			result.FailedFiles++
		default:
			unreached[fileUID] = f
		}
	}

	// A primary JPEG in the sidecar folder is reached only through the file it was created from, so
	// it counts as skipped when that file was, such as a RAW while RAW files are disabled.
	skippedFiles, err := skippedPhotoFiles(conf, unreached, stats)

	if err != nil {
		return result, err
	}

	for _, f := range unreached {
		if f.FileRoot == entity.RootSidecar && convertedFromAny(f.FileName, skippedFiles[f.PhotoID]) &&
			fs.FileExists(ConfigFileName(conf, f.FileRoot, f.FileName)) {
			result.SkippedMarkers += f.Markers
			result.SkippedFiles++
		} else {
			result.Markers += f.Markers
			result.Files++
		}
	}

	return result, nil
}

// skippedPhotoFiles returns, by picture, the names of the files in the originals folder the index was
// set to skip, for the pictures of the sidecar files among the files.
func skippedPhotoFiles(conf *config.Config, files map[string]query.FaceMarkerFile, stats *FaceRegeneration) (map[uint][]string, error) {
	result := make(map[uint][]string)

	if !stats.HasSkipped() {
		return result, nil
	}

	ids := make([]uint, 0, len(files))

	for _, f := range files {
		if f.FileRoot == entity.RootSidecar {
			ids = append(ids, f.PhotoID)
		}
	}

	siblings, err := query.FilesByPhotoIDs(ids)

	if err != nil {
		return nil, err
	}

	for _, f := range siblings {
		if f.FileRoot == entity.RootOriginals && stats.Skipped(ConfigFileName(conf, f.FileRoot, f.FileName)) {
			result[f.PhotoID] = append(result[f.PhotoID], f.FileName)
		}
	}

	return result, nil
}

// convertedFromAny reports whether the sidecar file was created from one of the original files, which
// its name then starts with, as in "IMG_1.CR2.jpg" for "IMG_1.CR2".
func convertedFromAny(sidecarName string, originals []string) bool {
	for _, name := range originals {
		if strings.HasPrefix(sidecarName, name+".") {
			return true
		}
	}

	return false
}
