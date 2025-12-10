package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/media"
)

// PostVisionCaption returns a suitable caption for an image.
//
//	@Summary	returns a suitable caption for an image
//	@Id			PostVisionCaption
//	@Tags		Vision
//	@Produce	json
//	@Success	200					{object}	vision.ApiResponse
//	@Failure	401,403,404,429,501	{object}	i18n.Response
//	@Param		images				body		vision.ApiRequest	true	"list of image file urls"
//	@Router		/api/v1/vision/caption [post]
func PostVisionCaption(router *gin.RouterGroup) {
	router.POST("/vision/caption", func(c *gin.Context) {
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

		// Run inference to generate a caption.
		result, model, err := vision.GenerateCaption(request.Images, media.SrcRemote)

		switch {
		case err != nil:
			log.Errorf("vision: %s (caption)", err)
			c.JSON(http.StatusBadRequest, vision.NewApiError(request.GetId(), http.StatusBadRequest))
			return
		case model == nil:
			log.Errorf("vision: no model specified (caption)")
			c.JSON(http.StatusInternalServerError, vision.NewApiError(request.GetId(), http.StatusInternalServerError))
			return
		case result == nil:
			log.Errorf("vision: no result (caption)")
			c.JSON(http.StatusInternalServerError, vision.NewApiError(request.GetId(), http.StatusInternalServerError))
			return
		}

		// Generate Vision API service response.
		response := vision.NewCaptionResponse(
			request.GetId(),
			&vision.Model{Type: model.Type, Name: model.Name, Version: model.Version},
			result,
		)

		c.JSON(http.StatusOK, response)
	})
}
