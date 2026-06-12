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

// UpdateLens updates lens make and model properties.
//
// PUT /api/v1/lenses/:id
//
//	@Summary	updates lens name
//	@Id			UpdateLens
//	@Tags		Lenses
//	@Accept		json
//	@Produce	json
//	@Success	200				{object}	entity.Lens
//	@Failure	401,403,404,429	{object}	i18n.Response
//	@Param		id				path		string		true	"Lens ID"
//	@Param		lens			body		form.Lens	true	"Properties to be updated, only Make and Model supported"
//	@Router		/api/v1/lenses/{id} [put]
func UpdateLens(router *gin.RouterGroup) {
	router.PUT("/lenses/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourceLenses, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		// Find lens by ID.
		lensId, err := strconv.ParseUint(clean.Token(c.Param("id")), 10, 32)
		m := query.FindLensByID(uint(lensId))

		if m == nil {
			Abort(c, http.StatusNotFound, i18n.ErrLensNotFound)
			return
		}

		// Create new lens form.
		frm, frmErr := form.NewLens(m)

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

		// Save lens and return new model values if successful.
		if err = m.SaveForm(frm); err != nil {
			log.Errorf("lens: %s", clean.Error(err))
			AbortSaveFailed(c)
			return
		}

		c.JSON(http.StatusOK, m)
	})
}
