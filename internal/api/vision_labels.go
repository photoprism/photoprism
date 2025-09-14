package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/service/http/header"
)

// PostVisionLabels returns suitable labels for an image.
//
//	@Summary	returns suitable labels for an image
//	@Id			PostVisionLabels
//	@Tags		Vision
//	@Accept		json
//	@Produce	json
//	@Success	200			{object}	vision.ApiResponse
//	@Failure	401,403,429	{object}	i18n.Response
//	@Param		images		body		vision.ApiRequest	true	"list of image file urls"
//	@Router		/api/v1/vision/labels [post]
func PostVisionLabels(router *gin.RouterGroup) {
	router.POST("/vision/labels", func(c *gin.Context) {
		s := Auth(c, acl.ResourceVision, acl.ActionUse)

		// Abort if permission is not granted.
		if s.Abort(c) {
			return
		}

		var request vision.ApiRequest

		// File uploads are not currently supported for this API endpoint.
		if header.HasContentType(&c.Request.Header, header.ContentTypeMultipart) {
			c.JSON(http.StatusBadRequest, vision.NewApiError(request.GetId(), http.StatusBadRequest))
			return
		}

		// Assign and validate request form values.
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, vision.NewApiError(request.GetId(), http.StatusBadRequest))
			return
		}

		// Check if the Computer Vision API is enabled, otherwise abort with an error.
		if !get.Config().VisionApi() {
			AbortFeatureDisabled(c)
			c.JSON(http.StatusForbidden, vision.NewApiError(request.GetId(), http.StatusForbidden))
			return
		}

		// Run inference to find matching labels.
		labels, err := vision.Labels(request.Images, media.SrcRemote, entity.SrcAuto)

		if err != nil {
			log.Errorf("vision: %s (run labels)", err)
			c.JSON(http.StatusBadRequest, vision.NewApiError(request.GetId(), http.StatusBadRequest))
			return
		}

		// Generate Vision API service response.
		response := vision.NewLabelsResponse(
			request.GetId(),
			&vision.Model{Type: vision.ModelTypeLabels},
			labels,
		)

		c.JSON(http.StatusOK, response)
	})
}
