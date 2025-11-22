package batch

import (
	"errors"
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// ApplyAlbums adds/removes the given photo to/from albums according to items action.
func ApplyAlbums(photo *entity.Photo, albums Items) (errs []error) {
	if photo == nil || !photo.HasID() {
		return []error{errors.New("invalid photo")}
	}

	photoUID := photo.GetUID()
	changed := false

	var addTargets []string

	for _, it := range albums.Items {
		switch it.Action {
		case ActionAdd:
			// Validate that we have either value or title
			if it.Value == "" && it.Title == "" {
				errs = append(errs, errors.New("album value or title required for add action"))
				continue
			}

			// Add by UID if provided, otherwise use title to create/find
			if it.Value != "" {
				// If value is provided, validate it's a proper UID format
				if !rnd.IsUID(it.Value, entity.AlbumUID) {
					errs = append(errs, fmt.Errorf("invalid album uid format: %s", it.Value))
					continue
				}

				// Check if album exists when adding by UID
				if _, findErr := entity.CachedAlbumByUID(it.Value); findErr != nil {
					errs = append(errs, fmt.Errorf("album not found: %s", it.Value))
					continue
				}

				addTargets = append(addTargets, it.Value)
			} else if it.Title != "" {
				addTargets = append(addTargets, it.Title)
			}

			changed = true
		case ActionRemove:
			// Validate that we have a value for removal
			if it.Value == "" {
				errs = append(errs, errors.New("album uid required for remove action"))
				continue
			}

			// Remove only if we have a valid album UID
			if !rnd.IsUID(it.Value, entity.AlbumUID) {
				errs = append(errs, fmt.Errorf("invalid album uid format: %s", it.Value))
				continue
			}

			// TODO: Simplify this, so it executes less queries.
			if a, findErr := entity.CachedAlbumByUID(it.Value); findErr != nil {
				errs = append(errs, fmt.Errorf("album not found for removal: %s", it.Value))
				continue
			} else if a.HasID() {
				// TODO: Don't error if photo is not in album, since this is normal in batch edit.
				a.RemovePhotos([]string{photoUID})
				changed = true
			}
		case ActionNone, ActionUpdate:
			// Valid actions that do nothing for albums
			continue
		default:
			errs = append(errs, fmt.Errorf("invalid action: %s", it.Action))
			continue
		}
	}

	if len(addTargets) > 0 {
		if addErr := entity.AddPhotoToAlbums(photoUID, addTargets); addErr != nil {
			errs = append(errs, addErr)
		}
	}

	if changed {
		photo.PreloadAlbums()
	}

	return errs
}
