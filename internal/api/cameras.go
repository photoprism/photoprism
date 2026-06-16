package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// UpdateCamera updates camera make and model properties.
//
// PUT /api/v1/cameras/:id
//
//	@Summary	updates camera name
//	@Id			UpdateCamera
//	@Tags		Cameras
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	entity.Camera
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		id				path		string		true	"Camera ID"
//	@Param		camera			body		form.Camera	true	"Properties to be updated, only Make and Model supported"
//	@Router		/api/v1/cameras/{id} [put]
func UpdateCamera(router *gin.RouterGroup) {
	router.PUT("/cameras/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourceCameras, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Find camera by ID. A non-numeric id parses to 0 and is rejected as not found below,
		// so the parse error is intentionally ignored.
		cameraId, _ := strconv.ParseUint(clean.Token(c.Param("id")), 10, 32)
		m := query.FindCameraByID(uint(cameraId))

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrCameraNotFound)
			return
		}

		// Create new camera form.
		frm, frmErr := form.NewCamera(m)

		if frmErr != nil {
			Abort(c, http.StatusBadRequest, i18n.ErrBadRequest)
			return
		}

		// Set form values from request.
		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if frmErr = c.BindJSON(frm); frmErr != nil {
			if IsRequestBodyTooLarge(frmErr) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			AbortBadRequest(c, frmErr)
			return
		} else if frmErr = frm.Validate(); frmErr != nil {
			AbortInvalidName(c)
			return
		}

		// Save camera and return new model values if successful.
		if err := m.SaveForm(frm); err != nil {
			log.Errorf("camera: %s", clean.Error(err))
			AbortSaveFailed(c)
			return
		}

		c.JSON(http.StatusOK, m)
	})
}
