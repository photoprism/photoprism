package batch

import (
	"fmt"
	"time"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/txt"
)

// PhotoSaveRequest bundles the context required to persist a batch update for a single photo.
type PhotoSaveRequest struct {
	Photo  *entity.Photo
	Form   *form.Photo
	Values *PhotosForm
}

// SaveBatchResult describes the outcome of PrepareAndSavePhotos so callers can publish events
// and build responses without re-walking the selection.
type SaveBatchResult struct {
	Requests     []*PhotoSaveRequest
	Results      []bool
	Preloaded    map[string]*entity.Photo
	UpdatedCount int
	SavedAny     bool
	Stats        MutationStats
}

// MutationStats captures how many photos received supporting mutations and
// how many errors occurred while applying them.
type MutationStats struct {
	AlbumMutations int
	LabelMutations int
	AlbumErrors    int
	LabelErrors    int
}

// NewPhotoSaveRequest converts the batch values into a form.Photo and bundles it with the
// target entity so callers outside this package do not have to depend on ConvertToPhotoForm.
func NewPhotoSaveRequest(photo *entity.Photo, values *PhotosForm) (*PhotoSaveRequest, error) {
	if values == nil {
		return nil, fmt.Errorf("batch: values are required")
	}

	frm, err := ConvertToPhotoForm(photo, values)
	if err != nil {
		return nil, err
	}

	return &PhotoSaveRequest{Photo: photo, Form: frm, Values: values}, nil
}

// PreparePhotoSaveRequests converts the given selection into save requests, ensuring each photo
// entity is hydrated and album/label actions are applied once per selection. The returned map is
// the (possibly newly populated) preloaded set so callers can reuse it for response rendering.
func PreparePhotoSaveRequests(photos search.PhotoResults, preloaded map[string]*entity.Photo, values *PhotosForm) ([]*PhotoSaveRequest, map[string]*entity.Photo, MutationStats) {
	if values == nil {
		return nil, preloaded, MutationStats{}
	}

	// Pre-create any album/label add targets so Apply* can reuse resolved UIDs per photo.
	resolveBatchItemValues(values)

	if preloaded == nil {
		preloaded = map[string]*entity.Photo{}
	}

	log.Infof("batch: updating %d photos", len(photos))

	saveRequests := make([]*PhotoSaveRequest, 0, len(photos))
	stats := MutationStats{}

	for _, result := range photos {
		photoID := result.PhotoUID

		if photoID == "" {
			continue
		}

		fullPhoto := preloaded[photoID]

		if fullPhoto == nil {
			loaded, err := query.PhotoPreloadByUID(photoID)
			if err != nil {
				log.Errorf("batch: failed to load photo %s: %s", photoID, err)
				continue
			}
			fullPhoto = &loaded
			preloaded[photoID] = fullPhoto
		}

		saveReq, err := NewPhotoSaveRequest(fullPhoto, values)

		if err != nil {
			log.Errorf("batch: failed to build save request for photo %s: %s", photoID, err)
			continue
		}

		saveRequests = append(saveRequests, saveReq)

		if values.Albums.Action == ActionUpdate {
			if errs := ApplyAlbums(fullPhoto, values.Albums); errs != nil {
				log.Errorf("batch: failed to update albums for photo %s: (%s)", photoID, errs)
				stats.AlbumErrors += len(errs)
			}

			if len(values.Albums.Items) > 0 {
				stats.AlbumMutations++
			}
		}

		if values.Labels.Action == ActionUpdate {
			if errs := ApplyLabels(fullPhoto, values.Labels); errs != nil {
				log.Errorf("batch: failed to update labels for photo %s (%s)", photoID, errs)
				stats.LabelErrors += len(errs)
			}

			if len(values.Labels.Items) > 0 {
				stats.LabelMutations++
			}
		}
	}

	return saveRequests, preloaded, stats
}

// PrepareAndSavePhotos hydrates the photo selection, applies album/label mutations, builds save
// requests, and persists the changes in one step. It returns a SaveBatchResult so callers can run
// follow-up work (events, cache flushes) without re-querying state.
func PrepareAndSavePhotos(photos search.PhotoResults, preloaded map[string]*entity.Photo, values *PhotosForm) (*SaveBatchResult, error) {
	start := time.Now()
	result := &SaveBatchResult{Preloaded: preloaded}

	if values == nil {
		if result.Preloaded == nil {
			result.Preloaded = map[string]*entity.Photo{}
		}
		return result, nil
	}

	requests, preloaded, mutationStats := PreparePhotoSaveRequests(photos, result.Preloaded, values)

	result.Requests = requests
	result.Preloaded = preloaded
	result.Stats = mutationStats

	if len(requests) == 0 {
		return result, nil
	}

	saveResults, err := SavePhotos(requests)

	if err != nil {
		return nil, err
	}

	result.Results = saveResults

	for i, saved := range saveResults {
		if saved {
			result.UpdatedCount++
			result.SavedAny = true
			log.Debugf("batch: successfully updated photo %s", requests[i].Photo.PhotoUID)
		}
	}

	var logFields []string

	if result.UpdatedCount > 0 {
		logFields = append(logFields, fmt.Sprintf("metadata (%d/%d)", result.UpdatedCount, len(photos)))
	}

	if result.Stats.LabelMutations > 0 {
		entry := fmt.Sprintf("labels (%d/%d)", result.Stats.LabelMutations, len(photos))
		if result.Stats.LabelErrors > 0 {
			entry = fmt.Sprintf("%s with %d errors", entry, result.Stats.LabelErrors)
		}
		logFields = append(logFields, entry)
	}

	if result.Stats.AlbumMutations > 0 {
		entry := fmt.Sprintf("albums (%d/%d)", result.Stats.AlbumMutations, len(photos))
		if result.Stats.AlbumErrors > 0 {
			entry = fmt.Sprintf("%s with %d errors", entry, result.Stats.AlbumErrors)
		}
		logFields = append(logFields, entry)
	}

	if len(logFields) > 0 {
		log.Infof("batch: updated photo %s [%s]", txt.JoinAnd(logFields), time.Since(start))
	} else {
		log.Infof("batch: no photos have been updated [%s]", time.Since(start))
	}

	// Update YAML backups for all albums referenced in the current batch request.
	if result.Stats.AlbumMutations > 0 {
		updateAlbumBackups(values)
	}

	return result, nil
}

// SavePhotos persists the batch updates described by the provided requests while skipping the
// heavyweight per-photo maintenance performed by entity.SavePhotoForm. It only updates the
// columns that actually changed, flags the photo for background metadata refresh by clearing
// CheckedAt, and updates the shared counts once per batch instead of once per photo.
func SavePhotos(requests []*PhotoSaveRequest) ([]bool, error) {
	results := make([]bool, len(requests))
	anySaved := false

	for i, req := range requests {
		saved, err := savePhoto(req)

		if err != nil {
			return results, err
		}

		results[i] = saved

		if saved {
			anySaved = true
		}
	}

	if anySaved {
		entity.UpdateCountsAsync()
	}

	return results, nil
}
