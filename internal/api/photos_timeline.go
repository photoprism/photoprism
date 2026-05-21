package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/photoprism/get"
	"github.com/photoprism/photoprism/pkg/i18n"
	"github.com/photoprism/photoprism/pkg/log/status"
)

// photosTimelineForm checks authorization and parses the timeline request without requiring a count parameter.
func photosTimelineForm(c *gin.Context) (frm form.SearchPhotos, bucket string, s *entity.Session, err error) {
	s = Auth(c, acl.ResourceCalendar, acl.ActionSearch)

	// Abort if calendar access is not granted.
	if s.Abort(c) {
		return frm, bucket, s, i18n.Error(i18n.ErrForbidden)
	}

	s = AuthAny(c, acl.ResourcePhotos, acl.Permissions{acl.ActionSearch, acl.ActionView, acl.AccessShared})

	// Abort if photo visibility is not granted.
	if s.Abort(c) {
		return frm, bucket, s, i18n.Error(i18n.ErrForbidden)
	}

	values := c.Request.URL.Query()
	bucket = strings.ToLower(strings.TrimSpace(values.Get("bucket")))

	switch bucket {
	case search.PhotoTimelineBucketMonth:
		// Supported.
	default:
		err = fmt.Errorf("invalid timeline bucket")
		event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourceCalendar), "timeline bucket invalid", status.Error(err)}, s.RefID)
		AbortBadRequest(c, err)
		return frm, bucket, s, err
	}

	if values.Has("order") {
		err = fmt.Errorf("unsupported timeline order")
		event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourceCalendar), "timeline order invalid", status.Error(err)}, s.RefID)
		AbortBadRequest(c, err)
		return frm, bucket, s, err
	}

	// Abort if request params are invalid.
	if err = bindPhotoTimelineFilters(values, &frm); err != nil {
		event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourceCalendar), "timeline form invalid", status.Error(err)}, s.RefID)
		AbortBadRequest(c, err)
		return frm, bucket, s, err
	}

	settings := get.Config().Settings()

	if !settings.Features.Calendar {
		AbortFeatureDisabled(c)
		return frm, bucket, s, i18n.Error(i18n.ErrFeatureDisabled)
	}

	// Ignore private flag if feature is disabled.
	if !settings.Features.Private {
		frm.Public = false
	}

	// Apply the same review restriction as /photos search for users without manage permission.
	if frm.Scope == "" &&
		settings.Features.Review &&
		acl.Rules.Deny(acl.ResourcePhotos, s.GetUserRole(), acl.ActionManage) {
		frm.Quality = 3
	}

	// Timeline buckets do not use seek/pagination or file-merge semantics.
	frm.Count = 0
	frm.Offset = 0
	frm.Order = ""
	frm.Reverse = false
	frm.Merged = false

	return frm, bucket, s, nil
}

// bindPhotoTimelineFilters maps search filters without result-shaping params.
func bindPhotoTimelineFilters(values url.Values, frm *form.SearchPhotos) error {
	filters := make(url.Values, len(values))

	for key, value := range values {
		filters[key] = append([]string(nil), value...)
	}

	filters.Del("bucket")
	filters.Del("count")
	filters.Del("offset")
	filters.Del("order")
	filters.Del("reverse")
	filters.Del("merged")

	return binding.MapFormWithTag(frm, filters, "form")
}

// SearchPhotoTimeline returns local photo timeline buckets for matching search filters.
//
//	@Summary		returns local photo timeline buckets
//	@Description	Returns month buckets in descending local-date order.
//	@Description	Result-shaping parameters such as count, offset, order, reverse, and merged are not part of this endpoint.
//	@Id				SearchPhotoTimeline
//	@Tags			Photos
//	@Produce		json
//	@Success		200				{object}	search.PhotoTimeline
//	@Failure		400,401,403,404	{object}	i18n.Response
//	@Param			bucket			query		string	true	"timeline bucket"	Enums(month)
//	@Param			public			query		bool	false	"excludes private pictures"
//	@Param			quality			query		int		false	"minimum quality score (1-7)"	Enums(0, 1, 2, 3, 4, 5, 6, 7)
//	@Param			q				query		string	false	"search query"
//	@Param			s				query		string	false	"album uid"
//	@Param			path			query		string	false	"photo path"
//	@Param			video			query		bool	false	"is type video"
//	@Router			/api/v1/photos/timeline [get]
func SearchPhotoTimeline(router *gin.RouterGroup) {
	router.GET("/photos/timeline", func(c *gin.Context) {
		f, bucket, s, err := photosTimelineForm(c)

		// Abort if authorization or form are invalid.
		if err != nil {
			return
		}

		// Find matching timeline buckets.
		result, err := search.UserPhotoTimeline(f, s, bucket)

		// Ok?
		if err != nil {
			event.AuditWarn([]string{ClientIP(c), "session %s", string(acl.ResourceCalendar), "timeline", status.Error(err)}, s.RefID)
			AbortBadRequest(c, err)
			return
		}

		// Add token headers for clients that need refreshed tokens.
		AddTokenHeaders(c, s)

		// Return as JSON.
		c.JSON(http.StatusOK, result)
	})
}
