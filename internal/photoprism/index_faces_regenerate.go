package photoprism

import (
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/dustin/go-humanize/english"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/thumb/crop"
	"github.com/photoprism/photoprism/pkg/clean"
)

// FaceRegeneration counts what regenerating the face markers changed in the files of a run. The
// index workers update it concurrently.
type FaceRegeneration struct {
	processed   sync.Map
	failed      sync.Map
	skipped     sync.Map
	skippedDirs sync.Map
	Files       atomic.Int64
	FailedFiles atomic.Int64
	Updated     atomic.Int64
	Added       atomic.Int64
	Removed     atomic.Int64
	Kept        atomic.Int64
	Failed      atomic.Int64
}

// add adds the changes made to one file.
func (s *FaceRegeneration) add(fileUID string, r faceRegenerationResult) {
	if s == nil {
		return
	}

	s.processed.Store(fileUID, true)

	s.Files.Add(1)
	s.Updated.Add(int64(r.Updated))
	s.Added.Add(int64(r.Added))
	s.Removed.Add(int64(r.Removed))
	s.Kept.Add(int64(r.Kept))
	s.Failed.Add(int64(r.Failed))
}

// Processed reports whether the markers of the file were regenerated.
func (s *FaceRegeneration) Processed(fileUID string) bool {
	if s == nil {
		return false
	}

	_, ok := s.processed.Load(fileUID)

	return ok
}

// addSkipped records a file the index was set to skip, such as an ignored file or a disabled RAW.
func (s *FaceRegeneration) addSkipped(fileName string) {
	if s != nil {
		s.skipped.Store(fileName, true)
	}
}

// addSkippedDir records a folder the index was set to skip, such as an ignored or hidden folder.
func (s *FaceRegeneration) addSkippedDir(dir string) {
	if s != nil {
		s.skippedDirs.Store(filepath.Clean(dir), true)
	}
}

// Skipped reports whether the index was set to skip the file, or a folder that holds it. The parent
// folders are looked up one by one, since a library can hold a skipped folder in every folder.
func (s *FaceRegeneration) Skipped(fileName string) bool {
	if s == nil {
		return false
	} else if _, ok := s.skipped.Load(fileName); ok {
		return true
	}

	for dir := filepath.Dir(fileName); ; dir = filepath.Dir(dir) {
		if _, ok := s.skippedDirs.Load(dir); ok {
			return true
		} else if parent := filepath.Dir(dir); parent == dir {
			return false
		}
	}
}

// HasSkipped reports whether the index was set to skip any file or folder.
func (s *FaceRegeneration) HasSkipped() (found bool) {
	if s == nil {
		return false
	}

	for _, m := range []*sync.Map{&s.skipped, &s.skippedDirs} {
		m.Range(func(_, _ any) bool {
			found = true
			return false
		})
	}

	return found
}

// addError counts a file whose markers could not be regenerated and were left unchanged.
func (s *FaceRegeneration) addError(fileUID string) {
	if s != nil {
		s.failed.Store(fileUID, true)
		s.FailedFiles.Add(1)
	}
}

// FileFailed reports whether regenerating the markers of the file failed.
func (s *FaceRegeneration) FileFailed(fileUID string) bool {
	if s == nil {
		return false
	}

	_, ok := s.failed.Load(fileUID)

	return ok
}

// String returns a summary of the changes for logging.
func (s *FaceRegeneration) String() string {
	if s == nil {
		return ""
	}

	result := fmt.Sprintf("%d updated, %d added, %d removed, %d kept unmatched, %d failed",
		s.Updated.Load(), s.Added.Load(), s.Removed.Load(), s.Kept.Load(), s.Failed.Load())

	if n := s.FailedFiles.Load(); n > 0 {
		result += fmt.Sprintf(", %s could not be processed", english.Plural(int(n), "file", "files"))
	}

	return result
}

// faceRegenerationResult reports what regenerating the face markers of one file changed.
type faceRegenerationResult struct {
	Updated int
	Added   int
	Removed int
	Kept    int
	Failed  int
}

// Changed reports whether markers were updated, added, or removed.
func (r faceRegenerationResult) Changed() bool {
	return r.Updated+r.Added+r.Removed > 0
}

// regenerateFaces detects the faces in a primary image again and applies them to its markers, see
// the "Resetting Face Recognition" section of the internal/ai/face README. The caller lowers the
// detector's score floor to the migration floor, so markers a previous detector placed are found
// again, while a new face must still clear FACE_SIZE and FACE_SCORE.
func (ind *Index) regenerateFaces(jpeg *MediaFile, file *entity.File, importFaceTags bool) (result faceRegenerationResult, err error) {
	if jpeg == nil || file == nil {
		return result, fmt.Errorf("faces: file required for regenerating markers")
	}

	conf := Config()

	thumbName, err := jpeg.Thumbnail(conf.ThumbCachePath(), thumb.Fit720)

	if err != nil {
		return result, err
	} else if thumbName == "" {
		return result, fmt.Errorf("thumbnail %s not found", thumb.Fit720)
	}

	detected, err := face.Detect(thumbName, conf.FaceMigrateSize())

	if err != nil {
		return result, err
	}

	// Loaded again rather than taken from the file, which reads a failed lookup as no markers and
	// would add every face a second time.
	stored, err := entity.FindMarkers(file.FileUID)

	if err != nil {
		return result, fmt.Errorf("faces: %s (load markers)", err)
	}

	markers := file.Markers()
	*markers = stored

	var valid, rejected entity.Markers

	for i := range *markers {
		if m := (*markers)[i]; m.MarkerType != entity.MarkerFace || m.MarkerUID == "" {
			continue
		} else if m.MarkerInvalid {
			rejected = append(rejected, m)
		} else {
			valid = append(valid, m)
		}
	}

	assignments, claimed := matchRegeneratedFaces(valid, rejected, detected)

	anchored := xmpAnchoredMarkers(jpeg, valid, importFaceTags)

	added := newRegeneratedFaces(detected, claimed, conf.FaceScoreEffective(), conf.FaceSize(), conf.FaceSizeRetry())

	// Only the matched and the new faces are embedded: the lower floors find more candidates than
	// an ordinary index would keep, and the smallest face decides the rendition for all of them.
	embed := make(face.Faces, 0, len(assignments)+len(added))
	matched := make(map[string]int, len(assignments))

	for markerUID, i := range assignments {
		matched[markerUID] = len(embed)
		embed = append(embed, detected[i])
	}

	for _, i := range added {
		embed = append(embed, detected[i])
	}

	renderCropSource := func(faces face.Faces) {
		if _, renderErr := cacheFaceCropSource(conf, jpeg, faces); renderErr != nil {
			log.Debugf("vision: %s in %s (render face crop source)", renderErr, clean.Log(jpeg.BaseName()))
		}
	}

	if err = vision.EmbedFaces(thumbName, embed, true, renderCropSource); err != nil {
		return result, err
	}

	kept := make(entity.Markers, 0, len(*markers))

	for i := range *markers {
		marker := &(*markers)[i]

		if marker.MarkerType != entity.MarkerFace || marker.MarkerUID == "" {
			kept = append(kept, *marker)
			continue
		}

		if j, ok := matched[marker.MarkerUID]; ok {
			if changed, updateErr := marker.Redetect(embed[j], *file, anchored[marker.MarkerUID]); updateErr != nil {
				log.Warnf("faces: %s", updateErr)
				result.Failed++
			} else if changed {
				result.Updated++
			}

			kept = append(kept, *marker)
		} else if marker.MarkerInvalid || !marker.DetectedFace() || marker.SubjUID != "" || marker.MarkerName != "" || anchored[marker.MarkerUID] {
			result.Kept++
			kept = append(kept, *marker)
		} else if deleteErr := marker.Delete(); deleteErr != nil {
			log.Warnf("faces: %s while removing marker %s", deleteErr, clean.Log(marker.MarkerUID))
			result.Failed++
			kept = append(kept, *marker)
		} else {
			result.Removed++
		}
	}

	*markers = kept

	// Added after the removals, so a detection is not refused for overlapping a marker that no
	// detection matched.
	before := len(*markers)

	file.AddFaces(embed[len(assignments):])

	result.Added = len(*markers) - before

	return result, nil
}

// matchRegeneratedFaces pairs detections with face markers, and returns the detection each marker got
// and those no new marker may be created for. Valid markers claim first, by best fit, but not what a
// rejected marker would keep an ordinary index from adding; a named marker whose own face is not
// detected may claim a neighboring face without a marker, which is accepted.
func matchRegeneratedFaces(valid, rejected entity.Markers, detected face.Faces) (assignments map[string]int, claimed map[int]bool) {
	blocked := make(map[int]bool)

	for i := range detected {
		for j := range rejected {
			if rejected[j].CropArea().OverlapPercent(detected[i].CropArea()) > face.OverlapThreshold {
				blocked[i] = true
			}
		}
	}

	claimed = make(map[int]bool, len(detected))

	for i := range blocked {
		claimed[i] = true
	}

	assignments = valid.MatchFacesBestFit(detected, claimed)

	for _, i := range assignments {
		claimed[i] = true
	}

	// Rejected markers only take the detections they block, so they cannot keep a face from being
	// added that they merely touch.
	skip := make(map[int]bool, len(detected))

	for i := range detected {
		if !blocked[i] {
			skip[i] = true
		}
	}

	for markerUID, i := range rejected.MatchFacesBestFit(detected, skip) {
		for j := range rejected {
			if rejected[j].MarkerUID == markerUID && rejected[j].CropArea().OverlapPercent(detected[i].CropArea()) > face.OverlapThreshold {
				assignments[markerUID] = i
			}
		}
	}

	return assignments, claimed
}

// xmpAnchoredMarkers returns the uids of the markers a sidecar region is matched to when face tags
// are imported. They keep their area, so the region still matches them after a new detection.
func xmpAnchoredMarkers(jpeg *MediaFile, markers entity.Markers, importFaceTags bool) map[string]bool {
	result := make(map[string]bool)

	if !importFaceTags || len(markers) == 0 {
		return result
	}

	regions, err := collectXmpFaces(jpeg)

	if err != nil {
		log.Debugf("faces: %s while reading xmp face regions", clean.Error(err))
		return result
	}

	for _, region := range regions.Faces {
		area := crop.NewArea("face", region.X, region.Y, region.W, region.H)

		for i := range markers {
			if markers[i].CropArea().OverlapPercent(area) > face.OverlapThreshold {
				result[markers[i].MarkerUID] = true
			}
		}
	}

	return result
}

// newRegeneratedFaces returns the unclaimed detections that clear the thresholds an ordinary index
// applies, including its retry at retrySize for a picture in which none clears minSize. The score is
// compared with the cutoff as face.Detect compares it.
func newRegeneratedFaces(detected face.Faces, claimed map[int]bool, minScore float64, minSize, retrySize int) (result []int) {
	qualifies := func(f face.Face, size int) bool {
		return (minScore < 0 || float64(f.Score) >= minScore) && f.Size() >= size
	}

	found := false

	for i := range detected {
		if !qualifies(detected[i], minSize) {
			continue
		}

		found = true

		if !claimed[i] {
			result = append(result, i)
		}
	}

	if found || retrySize <= 0 {
		return result
	}

	for i := range detected {
		if !claimed[i] && qualifies(detected[i], retrySize) {
			result = append(result, i)
		}
	}

	return result
}
