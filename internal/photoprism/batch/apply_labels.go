package batch

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// labelRemovalAction enumerates how batch edit should handle a label removal
// request.
type labelRemovalAction int

const (
	labelRemovalKeep labelRemovalAction = iota
	labelRemovalBlock
	labelRemovalDelete
)

// String returns a stable action name so logs and errors stay readable.
func (a labelRemovalAction) String() string {
	switch a {
	case labelRemovalKeep:
		return "keep"
	case labelRemovalBlock:
		return "block"
	case labelRemovalDelete:
		return "delete"
	default:
		return "unknown"
	}
}

// Locking note: testers observed MySQL deadlocks (error 1213) when concurrent
// batch edits inserted / removed rows in photos_labels. The helpers below retry
// a few times with a short backoff so we can surface success whenever InnoDB
// resolves the deadlock after a retry instead of failing the entire request.
const (
	deadlockRetryAttempts = 3
	deadlockRetryDelay    = 25 * time.Millisecond
)

// ApplyLabels adds/removes labels on the given photo according to items action.
func ApplyLabels(photo *entity.Photo, labels Items) (errs []error) {
	if photo == nil || !photo.HasID() {
		return []error{errors.New("invalid photo")}
	}

	var err error

	// Track if we changed anything to call SaveLabels once.
	changed := false
	labelIndex := indexPhotoLabels(photo.Labels)

	for _, it := range labels.Items {
		switch it.Action {
		case ActionAdd:
			// Validate that we have either value or title.
			if it.Value == "" && it.Title == "" {
				errs = append(errs, fmt.Errorf("label value or title required for add action"))
				continue
			}

			// Try by UID first.
			var labelEntity *entity.Label

			if it.Value != "" {
				// If value is provided, validate it's a proper UID format.
				if !rnd.IsUID(it.Value, entity.LabelUID) {
					errs = append(errs, fmt.Errorf("invalid label uid format: %s", it.Value))
					continue
				}

				labelEntity, err = query.LabelByUID(it.Value)
				if err != nil {
					errs = append(errs, fmt.Errorf("label not found: %s", it.Value))
					continue
				}
			}

			if labelEntity == nil && it.Title != "" {
				// Create or find by title.
				if labelEntity, err = entity.FindLabel(it.Title, true); err != nil || labelEntity == nil {
					labelEntity = entity.FirstOrCreateLabel(entity.NewLabel(it.Title, 0))
				}
			}

			if labelEntity == nil {
				errs = append(errs, fmt.Errorf("could not resolve label to add: value=%s title=%s", it.Value, clean.Log(it.Title)))
				continue
			}

			if err = labelEntity.Restore(); err != nil {
				log.Debugf("batch: could not restore label %s: %s", labelEntity.LabelName, err)
			}

			pl := labelIndex[labelEntity.ID]

			if pl == nil {
				pl = entity.FirstOrCreatePhotoLabel(entity.NewPhotoLabel(photo.ID, labelEntity.ID, 0, entity.SrcBatch))
				if pl == nil {
					errs = append(errs, fmt.Errorf("failed creating photo-label for photo %d and label %d", photo.ID, labelEntity.ID))
					continue
				}
				labelIndex[labelEntity.ID] = pl
			}

			// Keep existing label with higher priority than batch updates, even if it wasn't preloaded.
			if entity.SrcPriority[pl.LabelSrc] > entity.SrcPriority[entity.SrcBatch] {
				continue
			}

			// Ensure 100% confidence (uncertainty 0) and source 'batch'.
			if pl.Uncertainty != 0 || pl.LabelSrc != entity.SrcBatch {
				pl.Uncertainty = 0
				pl.LabelSrc = entity.SrcBatch
				if err = updatePhotoLabel(pl, "update label confidence"); err != nil {
					errs = append(errs, fmt.Errorf("update label to 100%% confidence failed: %s", err))
				} else {
					changed = true
				}
			} else {
				changed = true
			}
		case ActionRemove:
			if it.Value == "" {
				errs = append(errs, fmt.Errorf("label uid required for remove action"))
				continue
			}

			// Validate UID format.
			if !rnd.IsUID(it.Value, entity.LabelUID) {
				errs = append(errs, fmt.Errorf("invalid label uid format: %s", clean.Log(it.Value)))
				continue
			}

			pl, labelEntity := findIndexedPhotoLabel(labelIndex, photo, it.Value)

			if pl == nil || labelEntity == nil || !labelEntity.HasID() {
				log.Debugf("batch: label not found for removal label (photo_uid=%s label_uid=%s)", clean.Log(photo.GetUID()), clean.Log(it.Value))
				continue
			}

			labelName := labelEntity.LabelName
			labelChanged := false

			switch determineLabelRemovalAction(pl) {
			case labelRemovalDelete:
				if err = deletePhotoLabel(pl); err != nil {
					errs = append(errs, fmt.Errorf("delete label failed: %s", err))
				} else {
					log.Debugf("batch: deleted label photo=%s label_id=%d", photo.PhotoUID, labelEntity.ID)
					delete(labelIndex, labelEntity.ID)
					labelChanged = true
				}
			case labelRemovalBlock:
				if markLabelBlocked(pl) {
					if err = updatePhotoLabel(pl, "block label"); err != nil {
						errs = append(errs, fmt.Errorf("block label failed: %s", err))
					} else {
						log.Debugf("batch: blocked label: photo=%s label_id=%d", photo.PhotoUID, labelEntity.ID)
						labelChanged = true
					}
				}
			case labelRemovalKeep:
				// Nothing to do.
			}

			if labelChanged {
				changed = true
				if err = photo.DropKeywords([]string{labelName}); err != nil {
					log.Debugf("batch: failed to drop label keyword from photo (%s)", err)
				}
			}
		case ActionNone, ActionUpdate:
			// Valid actions that do nothing for labels.
			continue
		default:
			errs = append(errs, fmt.Errorf("invalid action: %s", it.Action))
			continue
		}
	}

	if changed {
		// Reload labels, but avoid calling Photo.SaveLabels() due to its heavy cost.
		photo.PreloadLabels()
	}

	return errs
}

// indexPhotoLabels builds a PhotoLabel lookup map so ApplyLabels can reuse the associations that
// were already preloaded for the batch selection without re-querying the join table.
func indexPhotoLabels(labels entity.PhotoLabels) map[uint]*entity.PhotoLabel {
	if len(labels) == 0 {
		return map[uint]*entity.PhotoLabel{}
	}

	idx := make(map[uint]*entity.PhotoLabel, len(labels))

	for i := range labels {
		lbl := &labels[i]
		if lbl == nil || lbl.LabelID == 0 {
			continue
		}
		idx[lbl.LabelID] = lbl
	}

	return idx
}

// determineLabelRemovalAction maps a PhotoLabel to the removal outcome defined
// in README rules (keep, update/block, delete).
func determineLabelRemovalAction(pl *entity.PhotoLabel) labelRemovalAction {
	if pl == nil {
		return labelRemovalKeep
	}

	priority := entity.SrcPriority[pl.LabelSrc]
	batchPriority := entity.SrcPriority[entity.SrcBatch]

	if priority > batchPriority {
		return labelRemovalKeep
	}

	if priority == batchPriority && pl.Uncertainty >= 100 {
		return labelRemovalKeep
	}

	if priority < batchPriority {
		return labelRemovalBlock
	}

	if pl.LabelSrc == entity.SrcVision {
		if pl.Uncertainty < 100 {
			return labelRemovalBlock
		}

		return labelRemovalKeep
	}

	return labelRemovalDelete
}

// markLabelBlocked updates the PhotoLabel to represent a user-hidden label by
// setting LabelSrc to `batch` and Uncertainty to 100. It returns true when a
// change was applied so callers can avoid redundant writes.
func markLabelBlocked(pl *entity.PhotoLabel) bool {
	if pl == nil {
		return false
	}

	changed := false

	if pl.LabelSrc != entity.SrcBatch {
		pl.LabelSrc = entity.SrcBatch
		changed = true
	}

	if pl.Uncertainty != 100 {
		pl.Uncertainty = 100
		changed = true
	}

	return changed
}

// findLabelEntityForRemoval tries to resolve the label referenced in a removal
// action using preloaded photo data before falling back to a DB query.
// findIndexedPhotoLabel returns the cached PhotoLabel relation for a removal
// request, optionally falling back to the shared cache when the initial index
// miss occurs.
func findIndexedPhotoLabel(index map[uint]*entity.PhotoLabel, photo *entity.Photo, labelUID string) (*entity.PhotoLabel, *entity.Label) {
	if photo == nil {
		return nil, nil
	}

	tryAttachLabel := func(label *entity.Label) (*entity.PhotoLabel, *entity.Label) {
		if label == nil || !label.HasID() {
			return nil, nil
		}

		if cached := index[label.ID]; cached != nil {
			if cached.Label == nil {
				cached.Label = label
			}
			return cached, label
		}

		if assoc := loadPhotoLabelAssociation(photo.ID, label.ID); assoc != nil {
			assoc.Label = label
			index[label.ID] = assoc
			return assoc, label
		}

		return nil, nil
	}

	if label := lookupLabelOnPhoto(photo, labelUID); label != nil {
		if pl, resolved := tryAttachLabel(label); pl != nil {
			return pl, resolved
		}
	}

	labelEntity, err := query.LabelByUID(labelUID)

	if err != nil || labelEntity == nil || !labelEntity.HasID() {
		return nil, nil
	}

	if pl, resolved := tryAttachLabel(labelEntity); pl != nil {
		return pl, resolved
	}

	return nil, nil
}

// lookupLabelOnPhoto searches the preloaded photo associations for a label UID.
func lookupLabelOnPhoto(photo *entity.Photo, labelUID string) *entity.Label {
	if photo == nil || labelUID == "" {
		return nil
	}

	for i := range photo.Labels {
		pl := &photo.Labels[i]

		if pl == nil || pl.Label == nil {
			continue
		}

		if pl.Label.LabelUID == labelUID {
			return pl.Label
		}
	}

	return nil
}

// loadPhotoLabelAssociation fetches a PhotoLabel using the shared cache helper,
// falling back to a direct lookup when needed.
func loadPhotoLabelAssociation(photoID, labelID uint) *entity.PhotoLabel {
	if cached, err := entity.FindPhotoLabel(photoID, labelID, true); err == nil && cached.HasID() {
		return cached
	}

	return nil
}

// updatePhotoLabel persists the PhotoLabel source and uncertainty with deadlock retry
// so batch edits survive temporary lock conflicts when many updates run in parallel.
func updatePhotoLabel(pl *entity.PhotoLabel, action string) error {
	if pl == nil {
		return fmt.Errorf("photo label is nil")
	}

	return withDeadlockRetry(action, func() error {
		return pl.Updates(entity.Values{"label_src": pl.LabelSrc, "uncertainty": pl.Uncertainty})
	})
}

// deletePhotoLabel removes a PhotoLabel while retrying on transient deadlocks.
func deletePhotoLabel(pl *entity.PhotoLabel) error {
	if pl == nil {
		return fmt.Errorf("photo label is nil")
	}

	return withDeadlockRetry("delete label", func() error {
		return pl.Delete()
	})
}

// withDeadlockRetry executes fn and retries a few times if the database reports
// a deadlock, helping batch edits succeed without surfacing errors to users.
func withDeadlockRetry(action string, fn func() error) (err error) {
	for attempt := 0; attempt < deadlockRetryAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !isDeadlockError(err) {
			return err
		}

		wait := deadlockRetryDelay * time.Duration(attempt+1)
		log.Warnf("batch: %s deadlock (attempt %d/%d): %s", action, attempt+1, deadlockRetryAttempts, err)
		time.Sleep(wait)
	}

	return err
}

// isDeadlockError detects MySQL deadlock errors both via driver codes and
// fallback substring matching so retries trigger reliably across drivers.
func isDeadlockError(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr *mysql.MySQLError

	if errors.As(err, &mysqlErr) {
		if mysqlErr.Number == 1213 {
			return true
		}
	}

	msg := strings.ToLower(err.Error())

	return strings.Contains(msg, "deadlock")
}
