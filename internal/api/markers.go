package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dustin/go-humanize/english"
	"github.com/gin-gonic/gin"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/crop"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/internal/mutex"
	"github.com/photoprism/photoprism/internal/query"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/i18n"
)

// Checks if background worker runs less than once per hour.
func wakeupIntervalTooHigh(c *gin.Context) bool {
	if conf := get.Config(); conf.Unsafe() {
		return false
	} else if i := conf.WakeupInterval(); i > time.Hour {
		Abort(c, http.StatusForbidden, i18n.ErrWakeupInterval, i.String())
		return true
	} else {
		return false
	}
}

// findFileMarker returns a file and marker entity matching the api request.
func findFileMarker(c *gin.Context) (file *entity.File, marker *entity.Marker, err error) {
	// Check authorization.
	s := Auth(c, acl.ResourceFiles, acl.ActionUpdate)

	if s.Invalid() {
		AbortForbidden(c)
		return nil, nil, fmt.Errorf("unauthorized")
	}

	// Check feature flags.
	conf := get.Config()
	if !conf.Settings().Features.People {
		AbortFeatureDisabled(c)
		return nil, nil, fmt.Errorf("feature disabled")
	}

	// Find marker.
	if uid := c.Param("marker_uid"); uid == "" {
		AbortBadRequest(c)
		return nil, nil, fmt.Errorf("bad request")
	} else if marker, err = query.MarkerByUID(uid); err != nil {
		AbortEntityNotFound(c)
		return nil, nil, fmt.Errorf("uid %s %s", uid, err)
	} else if marker.FileUID == "" {
		AbortEntityNotFound(c)
		return nil, marker, fmt.Errorf("marker file missing")
	}

	// Find file.
	if file, err = query.FileByUID(marker.FileUID); err != nil {
		AbortEntityNotFound(c)
		return file, marker, fmt.Errorf("file %s %s", marker.FileUID, err)
	}

	return file, marker, nil
}

// CreateMarker adds a new file area marker to assign faces or other subjects.
//
// See internal/form/marker.go for the values required to create a new marker.
//
// POST /api/v1/markers
func CreateMarker(router *gin.RouterGroup) {
	router.POST("/markers", func(c *gin.Context) {
		s := Auth(c, acl.ResourceFiles, acl.ActionUpdate)

		// Abort if permission was not granted.
		if s.Abort(c) {
			return
		}

		// Initialize form.
		frm := form.Marker{
			FileUID:       "",
			MarkerType:    entity.MarkerFace,
			MarkerSrc:     entity.SrcManual,
			MarkerReview:  false,
			MarkerInvalid: false,
		}

		// Initialize form.
		if err := c.BindJSON(&frm); err != nil {
			log.Errorf("faces: %s (bind marker form)", err)
			AbortBadRequest(c)
			return
		}

		// Find related file.
		file, err := query.FileByUID(frm.FileUID)

		// Abort if not found.
		if err != nil {
			AbortEntityNotFound(c)
			return
		}

		// Validate form values.
		if err = frm.Validate(); err != nil {
			log.Errorf("faces: %s (validate new marker)", err)
			AbortBadRequest(c)
			return
		} else if frm.W <= 0 || frm.H <= 0 {
			log.Errorf("faces: width and height must be greater than zero")
			AbortBadRequest(c)
			return
		}

		// Create new face marker area.
		area := crop.NewArea("face", frm.X, frm.Y, frm.W, frm.H)

		// Create new marker entity.
		marker := entity.NewMarker(*file, area, "", frm.MarkerSrc, frm.MarkerType, int(frm.W*float32(file.FileWidth)), 100)

		// Update marker from form values.
		if err = marker.Create(); err != nil {
			log.Errorf("faces: %s (create marker)", err)
			AbortBadRequest(c)
			return
		}

		// Update marker subject if a name was provided.
		if strings.TrimSpace(frm.MarkerName) == "" {
			log.Infof("faces: added new %s marker", clean.Log(marker.MarkerType))
		} else if changed, saveErr := marker.SaveForm(frm); err != nil {
			log.Errorf("faces: %s (update marker)", saveErr)
			AbortSaveFailed(c)
			return
		} else if changed {
			if updateErr := query.UpdateSubjectCovers(); updateErr != nil {
				log.Errorf("faces: %s (update covers)", updateErr)
			}

			if updateErr := entity.UpdateSubjectCounts(); updateErr != nil {
				log.Errorf("faces: %s (update counts)", updateErr)
			}
		}

		// Update photo metadata.
		if !file.FilePrimary {
			log.Infof("faces: skipped updating photo for non-primary file")
		} else if p, err := query.PhotoByUID(file.PhotoUID); err != nil {
			log.Errorf("faces: %s (find photo))", err)
		} else if err := p.UpdateAndSaveTitle(); err != nil {
			log.Errorf("faces: %s (update photo title)", err)
		} else {
			// Publish updated photo entity.
			PublishPhotoEvent(StatusUpdated, file.PhotoUID, c)
		}

		// Display success message.
		event.SuccessMsg(i18n.MsgChangesSaved)

		// Return new marker.
		c.JSON(http.StatusOK, marker)
	})
}

// UpdateMarker updates an existing file area marker to assign faces or other subjects.
//
// The request parameters are:
//
//   - marker_uid: string Marker UID as returned by the API
//
// PUT /api/v1/markers/:marker_uid
func UpdateMarker(router *gin.RouterGroup) {
	router.PUT("/markers/:marker_uid", func(c *gin.Context) {
		// Abort if workers runs less than once per hour.
		if wakeupIntervalTooHigh(c) {
			return
		}

		// Abort if another update is running.
		if err := mutex.UpdatePeople.Start(); err != nil {
			AbortBusy(c)
			return
		}

		defer mutex.UpdatePeople.Stop()

		file, marker, err := findFileMarker(c)

		if err != nil {
			log.Debugf("faces: %s (find marker to update)", err)
			return
		}

		// Initialize form.
		frm, err := form.NewMarker(*marker)

		if err != nil {
			log.Errorf("faces: %s (create marker form)", err)
			AbortSaveFailed(c)
			return
		} else if err = c.BindJSON(&frm); err != nil {
			log.Errorf("faces: %s (bind marker form)", err)
			AbortBadRequest(c)
			return
		}

		// Validate form values.
		if err = frm.Validate(); err != nil {
			log.Errorf("faces: %s (validate updated marker)", err)
			AbortBadRequest(c)
			return
		}

		// Update marker from form values.
		if changed, saveErr := marker.SaveForm(frm); saveErr != nil {
			log.Errorf("faces: %s (update marker)", saveErr)
			AbortSaveFailed(c)
			return
		} else if changed {
			if marker.FaceID != "" && marker.SubjUID != "" && marker.SubjSrc == entity.SrcManual {
				if res, err := get.Faces().Optimize(); err != nil {
					log.Errorf("faces: %s (optimize)", err)
				} else if res.Merged > 0 {
					log.Infof("faces: merged %s", english.Plural(res.Merged, "cluster", "clusters"))
				}
			}

			if updateErr := query.UpdateSubjectCovers(); updateErr != nil {
				log.Errorf("faces: %s (update covers)", updateErr)
			}

			if updateErr := entity.UpdateSubjectCounts(); updateErr != nil {
				log.Errorf("faces: %s (update counts)", updateErr)
			}
		}

		// Update photo metadata.
		if !file.FilePrimary {
			log.Infof("faces: skipped updating photo for non-primary file")
		} else if p, err := query.PhotoByUID(file.PhotoUID); err != nil {
			log.Errorf("faces: %s (find photo))", err)
		} else if err := p.UpdateAndSaveTitle(); err != nil {
			log.Errorf("faces: %s (update photo title)", err)
		} else {
			// Notify clients.
			PublishPhotoEvent(StatusUpdated, file.PhotoUID, c)
		}

		// Display success message.
		event.SuccessMsg(i18n.MsgChangesSaved)

		// Return updated marker.
		c.JSON(http.StatusOK, marker)
	})
}

// ClearMarkerSubject removes an existing marker subject association.
//
// The request parameters are:
//
//   - uid: string Photo UID as returned by the API
//   - file_uid: string File UID as returned by the API
//   - id: int Marker ID as returned by the API
//
// DELETE /api/v1/markers/:marker_uid/subject
func ClearMarkerSubject(router *gin.RouterGroup) {
	router.DELETE("/markers/:marker_uid/subject", func(c *gin.Context) {
		// Abort if workers runs less than once per hour.
		if wakeupIntervalTooHigh(c) {
			return
		}

		// Abort if another update is running.
		if err := mutex.UpdatePeople.Start(); err != nil {
			AbortBusy(c)
			return
		}

		defer mutex.UpdatePeople.Stop()

		file, marker, err := findFileMarker(c)

		if err != nil {
			log.Debugf("faces: %s (find marker to clear subject)", err)
			return
		}

		if err := marker.ClearSubject(entity.SrcManual); err != nil {
			log.Errorf("faces: %s (clear marker subject)", err)
			AbortSaveFailed(c)
			return
		} else if err := query.UpdateSubjectCovers(); err != nil {
			log.Errorf("faces: %s (update covers)", err)
		} else if err := entity.UpdateSubjectCounts(); err != nil {
			log.Errorf("faces: %s (update counts)", err)
		}

		// Update photo metadata.
		if !file.FilePrimary {
			log.Infof("faces: skipped updating photo for non-primary file")
		} else if p, err := query.PhotoByUID(file.PhotoUID); err != nil {
			log.Errorf("faces: %s (find photo))", err)
		} else if err := p.UpdateAndSaveTitle(); err != nil {
			log.Errorf("faces: %s (update photo title)", err)
		} else {
			// Notify clients.
			PublishPhotoEvent(StatusUpdated, file.PhotoUID, c)
		}

		event.SuccessMsg(i18n.MsgChangesSaved)

		c.JSON(http.StatusOK, marker)
	})
}
