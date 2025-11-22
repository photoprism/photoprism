package batch

import (
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/clean"
)

// resolveBatchItemValues pre-creates album and label references so later per-photo work only
// performs relation changes instead of issuing duplicate lookup/create queries for every photo.
func resolveBatchItemValues(values *PhotosForm) {
	if values == nil {
		return
	}

	resolveAlbumValues(&values.Albums)
	resolveLabelValues(&values.Labels)
}

// resolveAlbumValues pre-creates album UIDs for add actions so ApplyAlbums only performs
// membership changes instead of repeated title lookups.
func resolveAlbumValues(items *Items) {
	if items == nil || items.Action != ActionUpdate {
		return
	}

	items.ResolveValuesByTitle(func(title, action string) string {
		if action != ActionAdd || title == "" {
			return ""
		}

		return ensureAlbumUID(title)
	})
}

// resolveLabelValues pre-creates label UIDs for add actions to avoid recreating the same
// labels per photo when batch updates run.
func resolveLabelValues(items *Items) {
	if items == nil || items.Action != ActionUpdate {
		return
	}

	items.ResolveValuesByTitle(func(title, action string) string {
		if action != ActionAdd || title == "" {
			return ""
		}

		return ensureLabelUID(title)
	})
}

// ensureAlbumUID returns the UID of the album with the provided title, creating or restoring
// the album on demand when it does not already exist. Returns an empty string on failure.
func ensureAlbumUID(title string) string {
	if title == "" {
		return ""
	}

	album := entity.NewUserAlbum(title, entity.AlbumManual, entity.DefaultOrderAlbum, entity.OwnerUnknown)

	if existing := album.Find(); existing != nil && existing.HasID() {
		if existing.Deleted() {
			if err := existing.Restore(); err != nil {
				log.Errorf("batch: failed to restore album %s: %s", clean.Log(title), err)
				return ""
			}
		}
		return existing.AlbumUID
	}

	if err := album.Create(); err != nil {
		log.Errorf("batch: failed to create album %s: %s", clean.Log(title), err)
		return ""
	}

	return album.AlbumUID
}

// ensureLabelUID resolves or creates a label for the given title and returns its UID,
// restoring deleted labels when necessary.
func ensureLabelUID(title string) string {
	if title == "" {
		return ""
	}

	label, err := entity.FindLabel(title, true)

	if err != nil || label == nil || !label.HasUID() {
		label = entity.FirstOrCreateLabel(entity.NewLabel(title, 0))
		err = nil
	}

	if label == nil || !label.HasUID() {
		log.Errorf("batch: failed to resolve label %s", clean.Log(title))
		return ""
	}

	if label.Deleted() {
		if restoreErr := label.Restore(); restoreErr != nil {
			log.Errorf("batch: failed to restore label %s: %s", clean.Log(title), restoreErr)
			return ""
		}
	}

	return label.LabelUID
}
