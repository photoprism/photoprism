package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/ai/classify"
	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// AddPhotoLabel adds a label to a photo.
//
//	@Summary	adds a label to a photo
//	@Id			AddPhotoLabel
//	@Tags		Labels, Photos
//	@Accept		json
//	@Produce	json
//	@Success	200							{object}	entity.Photo
//	@Failure	400,401,403,404,413,429,500	{object}	i18n.Response
//	@Param		label						body		form.Label	true	"label properties"
//	@Param		uid							path		string		true	"photo uid"
//	@Router		/api/v1/photos/{uid}/label [post]
func AddPhotoLabel(router *gin.RouterGroup) {
	router.POST("/photos/:uid/label", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		// Limit by-UID edits to pictures within the session's shared scope, mirroring UpdatePhoto.
		// PhotoSessionSeesEverything is query-free and client and user role aware, so full-access
		// sessions skip the check and restricted sessions stay within their scope.
		if !search.PhotoSessionSeesEverything(s) {
			if visible, vErr := search.PhotoVisibleToSession(uid, s); vErr != nil || !visible {
				AbortForbidden(c)
				return
			}
		}

		m, err := query.PhotoByUID(uid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		frm := &form.Label{}

		// Assign and validate request form values.
		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err = c.BindJSON(frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			AbortBadRequest(c, err)
			return
		} else if err = frm.Validate(); err != nil {
			AbortInvalidName(c)
			return
		}

		labelEntity := entity.FirstOrCreateLabel(entity.NewLabel(frm.LabelName, frm.LabelPriority))

		if labelEntity == nil {
			AbortInvalidName(c)
			return
		}

		if err = labelEntity.Restore(); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not restore label"})
			return
		}

		photoLabel := entity.FirstOrCreatePhotoLabel(entity.NewPhotoLabel(m.ID, labelEntity.ID, frm.Uncertainty, "manual"))

		if photoLabel == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to update photo label"})
			return
		}

		if photoLabel.HasID() && photoLabel.Uncertainty > frm.Uncertainty {
			if updateErr := photoLabel.Updates(entity.Values{
				"Uncertainty": frm.Uncertainty,
				"LabelSrc":    entity.SrcManual,
			}); updateErr != nil {
				log.Errorf("label: %s", updateErr)
			}
		}

		p, err := query.PhotoPreloadByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err = p.SaveLabels(); err != nil {
			logErr("label", err)
			AbortSaveFailed(c)
			return
		}

		PublishPhotoEvent(StatusUpdated, clean.UID(c.Param("uid")))

		p.RedactForSession(s)
		c.JSON(http.StatusOK, p)
	})
}

// RemovePhotoLabel removes a label from a photo.
//
//	@Summary	removes a label from a photo
//	@Id			RemovePhotoLabel
//	@Tags		Labels, Photos
//	@Accept		json
//	@Produce	json
//	@Success	200						{object}	entity.Photo
//	@Failure	400,401,403,404,429,500	{object}	i18n.Response
//	@Param		uid						path		string	true	"photo uid"
//	@Param		id						path		string	true	"label id"
//	@Router		/api/v1/photos/{uid}/label/{id} [delete]
func RemovePhotoLabel(router *gin.RouterGroup) {
	router.DELETE("/photos/:uid/label/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		// Limit by-UID edits to pictures within the session's shared scope, mirroring UpdatePhoto.
		// PhotoSessionSeesEverything is query-free and client and user role aware, so full-access
		// sessions skip the check and restricted sessions stay within their scope.
		if !search.PhotoSessionSeesEverything(s) {
			if visible, vErr := search.PhotoVisibleToSession(uid, s); vErr != nil || !visible {
				AbortForbidden(c)
				return
			}
		}

		m, err := query.PhotoByUID(uid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		labelId, err := strconv.Atoi(clean.Token(c.Param("id")))

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrLabelNotFound)
			return
		}

		if labelId < 0 {
			AbortBadRequest(c, errors.New("invalid label id"))
			return
		}

		label, err := query.PhotoLabel(m.ID, uint(labelId))

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrLabelNotFound)
			return
		}

		switch {
		case (label.LabelSrc == classify.SrcManual || label.LabelSrc == entity.SrcBatch) && label.Uncertainty < 100:
			if err = label.Delete(); err != nil {
				logErr("label", err)
				AbortDeleteFailed(c)
				return
			}
		case label.LabelSrc != classify.SrcManual && label.LabelSrc != entity.SrcBatch:
			if err = label.Updates(entity.Values{"uncertainty": 100, "label_src": entity.SrcManual}); err != nil {
				logErr("label", err)
				AbortSaveFailed(c)
				return
			}
		}

		p, err := query.PhotoPreloadByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if label.Label != nil {
			if err = p.RemoveKeyword(label.Label.LabelName); err != nil {
				logErr("label", err)
				AbortSaveFailed(c)
				return
			}
		}

		if err := p.SaveLabels(); err != nil {
			logErr("label", err)
			AbortSaveFailed(c)
			return
		}

		PublishPhotoEvent(StatusUpdated, clean.UID(c.Param("uid")))

		p.RedactForSession(s)
		c.JSON(http.StatusOK, p)
	})
}

// UpdatePhotoLabel changes a photo label.
//
//	@Summary	changes a photo label
//	@Id			UpdatePhotoLabel
//	@Tags		Labels, Photos
//	@Accept		json
//	@Produce	json
//	@Success	200							{object}	entity.Photo
//	@Failure	400,401,403,404,413,429,500	{object}	i18n.Response
//	@Param		uid							path		string			true	"photo uid"
//	@Param		id							path		string			true	"label id"
//	@Param		label						body		form.PhotoLabel	true	"assignment uncertainty and optional label name"
//	@Router		/api/v1/photos/{uid}/label/{id} [put]
func UpdatePhotoLabel(router *gin.RouterGroup) {
	router.PUT("/photos/:uid/label/:id", func(c *gin.Context) {
		s := Auth(c, acl.ResourcePhotos, acl.ActionUpdate)

		if s.Abort(c) {
			return
		}

		uid := clean.UID(c.Param("uid"))

		// Limit by-UID edits to pictures within the session's shared scope, mirroring UpdatePhoto.
		// PhotoSessionSeesEverything is query-free and client and user role aware, so full-access
		// sessions skip the check and restricted sessions stay within their scope.
		if !search.PhotoSessionSeesEverything(s) {
			if visible, vErr := search.PhotoVisibleToSession(uid, s); vErr != nil || !visible {
				AbortForbidden(c)
				return
			}
		}

		m, err := query.PhotoByUID(uid)

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		labelId, err := strconv.Atoi(clean.Token(c.Param("id")))

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrLabelNotFound)
			return
		}

		if labelId < 0 {
			AbortBadRequest(c, errors.New("invalid label id"))
			return
		}

		label, err := query.PhotoLabel(m.ID, uint(labelId))

		if err != nil {
			Abort(c, http.StatusNotFound, i18n.ErrLabelNotFound)
			return
		}

		frm := form.PhotoLabel{}
		LimitRequestBodyBytes(c, MaxMutationRequestBytes)

		if err = c.BindJSON(&frm); err != nil {
			if IsRequestBodyTooLarge(err) {
				AbortRequestTooLarge(c, i18n.ErrBadRequest)
				return
			}

			AbortBadRequest(c, err)
			return
		}

		if frm.Uncertainty != nil && (*frm.Uncertainty < 0 || *frm.Uncertainty > 100) {
			AbortBadRequest(c, errors.New("uncertainty must be between 0 and 100"))
			return
		}

		if frm.Label != nil && frm.Label.Name != nil {
			if Auth(c, acl.ResourceLabels, acl.ActionUpdate).Abort(c) {
				return
			}
			name := form.Label{LabelName: *frm.Label.Name}
			if err = name.Validate(); err != nil {
				AbortInvalidName(c)
				return
			} else if label.Label == nil {
				Abort(c, http.StatusNotFound, i18n.ErrLabelNotFound)
				return
			} else if err = label.Label.Rename(*frm.Label.Name); err != nil {
				logErr("label", err)
				AbortSaveFailed(c)
				return
			}
			RemoveFromLabelCoverCache(label.Label.LabelUID)
			PublishLabelEvent(StatusUpdated, label.Label.LabelUID)
		}

		values := entity.Values{}
		if frm.Uncertainty != nil {
			values["uncertainty"] = *frm.Uncertainty
			if *frm.Uncertainty == 0 && label.LabelSrc != entity.SrcManual {
				values["label_src"] = entity.SrcManual
			}
		}
		if len(values) > 0 {
			if err = label.Updates(values); err != nil {
				logErr("label", err)
				AbortSaveFailed(c)
				return
			}
		}

		p, err := query.PhotoPreloadByUID(clean.UID(c.Param("uid")))

		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		if err = p.SaveLabels(); err != nil {
			logErr("label", err)
			AbortSaveFailed(c)
			return
		}

		PublishPhotoEvent(StatusUpdated, clean.UID(c.Param("uid")))

		p.RedactForSession(s)
		c.JSON(http.StatusOK, p)
	})
}
