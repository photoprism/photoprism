package api

import (
	"net/http"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form/batch"
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
//	@Success	200						{object}	batch.PhotosForm
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		photos					body		batch.PhotosRequest	true	"photos selection and values"
//	@Router		/api/v1/batch/photos/edit [post]
func BatchPhotosEdit(router *gin.RouterGroup) {
	router.POST("/batch/photos/edit", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		conf := get.Config()

		if !conf.Develop() && !conf.Experimental() {
			AbortNotImplemented(c)
			return
		}

		var frm batch.PhotosRequest

		// Assign and validate request form values.
		if err := c.BindJSON(&frm); err != nil {
			AbortBadRequest(c)
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
			log.Errorf("batch: %s", clean.Error(err))
			AbortUnexpectedError(c)
			return
		}

		// TODO: Implement photo metadata update based on submitted form values.
		if frm.Values != nil {
			log.Debugf("batch: updating photo metadata %#v (not yet implemented)", frm.Values)
			for _, photo := range photos {
				log.Debugf("batch: updating metadata of photo %s (not yet implemented)", photo.PhotoUID)
			}
		}

		// Create batch edit form values form from photo metadata.
		batchFrm := batch.NewPhotosForm(photos)

		var data gin.H

		if frm.Return {
			data = gin.H{
				"photos": photos,
				"values": batchFrm,
			}
		} else {
			data = gin.H{
				"values": batchFrm,
			}
		}

		c.JSON(http.StatusOK, data)
	})
}
