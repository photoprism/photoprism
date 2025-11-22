package api

import (
	"net/http"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/photoprism/batch"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// BatchPhotosEdit returns and updates the metadata of multiple photos.
//
//	@Summary	returns and updates the metadata of multiple photos
//	@Id			BatchPhotosEdit
//	@Tags		Photos
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	batch.PhotosResponse
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		Request					body		batch.PhotosRequest	true	"photos selection and values"
//	@Router		/api/v1/batch/photos/edit [post]
func BatchPhotosEdit(router *gin.RouterGroup) {
	router.Match(MethodsPutPost, "/batch/photos/edit", func(c *gin.Context) {
		// Require access to all photos.
		s := Auth(c, acl.ResourcePhotos, acl.AccessAll)

		if s.Abort(c) {
			return
		}

		// Require update permissions for photos.
		if acl.Rules.Deny(acl.ResourcePhotos, s.GetUserRole(), acl.ActionUpdate) {
			AbortForbidden(c)
			return
		}

		// Check feature flags.
		if !get.Config().Settings().Features.BatchEdit {
			AbortFeatureDisabled(c)
			return
		}

		var frm batch.PhotosRequest

		// Assign and validate request form values.
		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c, err)
			return
		}

		if len(frm.Photos) == 0 {
			Abort(c, http.StatusBadRequest, i18n.ErrNoItemsSelected)
			return
		}

		// Fetch selected photos from database.
		photos, count, err := search.BatchPhotos(frm.Photos, s)

		log.Debugf("batch: %s selected for editing", english.Plural(count, "photo", "photos"))

		// Abort if no photos were found.
		if err != nil {
			log.Errorf("batch: %s (load selection)", clean.Error(err))
			AbortUnexpectedError(c)
			return
		}

		preloadedPhotos := map[string]*entity.Photo{}

		if hydrated, err := query.PhotoPreloadByUIDs(photos.UIDs()); err != nil {
			log.Errorf("batch: failed to preload photo selection: %s", err)
			AbortUnexpectedError(c)
			return
		} else {
			preloadedPhotos = mapPhotosByUID(hydrated)
		}

		var (
			saveRequests []*batch.PhotoSaveRequest
			saveResults  []bool
			savedAny     bool
		)

		if frm.Values != nil {
			outcome, saveErr := batch.PrepareAndSavePhotos(photos, preloadedPhotos, frm.Values)

			if saveErr != nil {
				log.Errorf("batch: failed to persist photo updates: %s", saveErr)
				AbortUnexpectedError(c)
				return
			}

			saveRequests = outcome.Requests
			saveResults = outcome.Results
			preloadedPhotos = outcome.Preloaded
			savedAny = outcome.SavedAny
		}

		// Refresh selected photos from database?
		if !savedAny {
			// Don't refresh.
		} else if photos, count, err = search.BatchPhotos(frm.Photos, s); err != nil {
			log.Errorf("batch: %s (refresh selection)", clean.Error(err))
		}

		// Create batch edit form values form from photo metadata using the refreshed entities so
		// the response reflects persisted album/label edits without issuing per-photo queries.
		batchFrm := batch.NewPhotosFormWithEntities(photos, preloadedPhotos)

		if len(saveResults) > 0 {
			for i, saved := range saveResults {
				if !saved {
					continue
				}

				photo := preloadedPhotos[saveRequests[i].Photo.PhotoUID]

				if photo == nil {
					photo = saveRequests[i].Photo
				}

				// PublishPhotoEvent(StatusUpdated, photo.PhotoUID, c)
				SaveSidecarYaml(photo)
			}

			if savedAny {
				UpdateClientConfig()
				FlushCoverCache()
			}
		}

		// Return models and form values.
		data := batch.PhotosResponse{
			Models: photos,
			Values: batchFrm,
		}

		c.JSON(http.StatusOK, data)
	})
}

// mapPhotosByUID converts the provided list into a UID keyed lookup map so repeated
// selections can reuse already preloaded entities instead of querying again.
func mapPhotosByUID(photos entity.Photos) map[string]*entity.Photo {
	result := make(map[string]*entity.Photo, len(photos))

	for _, e := range photos {
		if e == nil || e.PhotoUID == "" {
			continue
		}
		result[e.PhotoUID] = e
	}

	return result
}
