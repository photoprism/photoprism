package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/ai/vision"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/media/http/header"
	"github.com/photoprism/photoprism/pkg/media/http/scheme"
)

// PostVisionFace returns the embeddings of a face.
//
//	@Summary	returns the embeddings of a face image
//	@Id			PostVisionFace
//	@Tags		Vision
//	@Produce	json
//	@Success	200				{object}	vision.ApiResponse
//	@Failure	401,403,429,501	{object}	i18n.Response
//	@Param		images			body		vision.ApiRequest	true	"list of image file urls"
//	@Router		/api/v1/vision/face [post]
func PostVisionFace(router *gin.RouterGroup) {
	router.POST("/vision/face", func(c *gin.Context) {
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

		// Return if no thumbnail filenames were given.
		if len(request.Images) == 0 {
			log.Errorf("vision: at least one image required (run face embeddings)")
			c.JSON(http.StatusBadRequest, vision.NewApiError(request.GetId(), http.StatusBadRequest))
			return
		}

		// Run inference to find matching labels.
		results := make([]face.Embeddings, len(request.Images))

		for i := range request.Images {
			if data, err := media.ReadUrl(request.Images[i], scheme.HttpsData); err != nil {
				results[i] = face.Embeddings{}
				log.Errorf("vision: %s (read face embedding from url)", err)
			} else if result, faceErr := vision.Face(data); faceErr != nil {
				results[i] = face.Embeddings{}
				log.Errorf("vision: %s (run face embeddings)", faceErr)
			} else {
				results[i] = result
			}
		}

		// Generate Vision API service response.
		response := vision.ApiResponse{
			Id:     request.GetId(),
			Code:   http.StatusOK,
			Model:  &vision.Model{Type: vision.ModelTypeFace},
			Result: vision.ApiResult{Embeddings: results},
		}

		c.JSON(http.StatusOK, response)
	})
}
