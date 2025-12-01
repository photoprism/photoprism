package batch

import (
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/txt"
)

// savePhoto applies the batched form values to a single entity.Photo and writes
// only the changed columns back to the database.
func savePhoto(req *PhotoSaveRequest) (bool, error) {
	if req == nil || req.Photo == nil || req.Form == nil || req.Values == nil {
		return false, fmt.Errorf("batch: invalid save request")
	}

	p := req.Photo
	formValues := req.Values
	formData := req.Form

	if !p.HasID() {
		return false, fmt.Errorf("batch: photo has no id")
	}

	original := *p
	details := p.GetDetails()
	origDetails := *details

	if shouldUpdateString(formValues.PhotoTitle) {
		p.PhotoTitle = formData.PhotoTitle
		p.TitleSrc = formData.TitleSrc
	}

	if shouldUpdateString(formValues.PhotoCaption) {
		p.PhotoCaption = formData.PhotoCaption
		p.CaptionSrc = formData.CaptionSrc
	}

	if formValues.PhotoType.Action == ActionUpdate {
		p.PhotoType = formData.PhotoType
		p.TypeSrc = formData.TypeSrc
	}

	if formValues.PhotoFavorite.Action == ActionUpdate {
		p.PhotoFavorite = formValues.PhotoFavorite.Value
	}

	if formValues.PhotoPrivate.Action == ActionUpdate {
		p.PhotoPrivate = formValues.PhotoPrivate.Value
	}

	if formValues.PhotoScan.Action == ActionUpdate {
		p.PhotoScan = formValues.PhotoScan.Value
	}

	if formValues.PhotoPanorama.Action == ActionUpdate {
		p.PhotoPanorama = formValues.PhotoPanorama.Value
	}

	timeChanged := formValues.PhotoDay.Action == ActionUpdate ||
		formValues.PhotoMonth.Action == ActionUpdate ||
		formValues.PhotoYear.Action == ActionUpdate ||
		formValues.TimeZone.Action == ActionUpdate

	if timeChanged {
		p.PhotoYear = formData.PhotoYear
		p.PhotoMonth = formData.PhotoMonth
		p.PhotoDay = formData.PhotoDay
		if formValues.TimeZone.Action == ActionUpdate {
			p.TimeZone = formData.TimeZone
		}
		p.TakenAtLocal = formData.TakenAtLocal
		p.TakenSrc = formData.TakenSrc

		p.NormalizeValues()

		if p.TimeZoneUTC() {
			p.TakenAt = p.TakenAtLocal
		} else {
			p.TakenAt = p.GetTakenAt()
		}

		p.UpdateDateFields()
	}

	locationChanged := formValues.PhotoLat.Action == ActionUpdate ||
		formValues.PhotoLng.Action == ActionUpdate ||
		formValues.PhotoCountry.Action == ActionUpdate ||
		formValues.PhotoAltitude.Action == ActionUpdate

	if formValues.PhotoLat.Action == ActionUpdate {
		p.PhotoLat = formValues.PhotoLat.Value
	}

	if formValues.PhotoLng.Action == ActionUpdate {
		p.PhotoLng = formValues.PhotoLng.Value
	}

	if formValues.PhotoAltitude.Action == ActionUpdate {
		p.PhotoAltitude = formValues.PhotoAltitude.Value
	}

	if formValues.PhotoCountry.Action == ActionUpdate {
		p.PhotoCountry = formValues.PhotoCountry.Value
	}

	if locationChanged {
		p.PlaceSrc = entity.SrcBatch
		locKeywords, locLabels := p.UpdateLocation()
		if len(locLabels) > 0 {
			p.AddLabels(locLabels)
		}
		if len(locKeywords) > 0 {
			words := txt.UniqueWords(txt.Words(details.Keywords))
			words = append(words, locKeywords...)
			details.Keywords = strings.Join(txt.UniqueWords(words), ", ")
		}
	}

	if formValues.DetailsSubject.Action == ActionUpdate || formValues.DetailsSubject.Action == ActionRemove {
		details.Subject = formData.Details.Subject
		details.SubjectSrc = formData.Details.SubjectSrc
	}

	if formValues.DetailsArtist.Action == ActionUpdate || formValues.DetailsArtist.Action == ActionRemove {
		details.Artist = formData.Details.Artist
		details.ArtistSrc = formData.Details.ArtistSrc
	}

	if formValues.DetailsCopyright.Action == ActionUpdate || formValues.DetailsCopyright.Action == ActionRemove {
		details.Copyright = formData.Details.Copyright
		details.CopyrightSrc = formData.Details.CopyrightSrc
	}

	if formValues.DetailsLicense.Action == ActionUpdate || formValues.DetailsLicense.Action == ActionRemove {
		details.License = formData.Details.License
		details.LicenseSrc = formData.Details.LicenseSrc
	}

	updates := entity.Values{}
	addUpdate := func(column string, changed bool, value interface{}) {
		if changed {
			updates[column] = value
		}
	}

	addUpdate("photo_title", p.PhotoTitle != original.PhotoTitle, p.PhotoTitle)
	addUpdate("title_src", p.TitleSrc != original.TitleSrc, p.TitleSrc)
	addUpdate("photo_caption", p.PhotoCaption != original.PhotoCaption, p.PhotoCaption)
	addUpdate("caption_src", p.CaptionSrc != original.CaptionSrc, p.CaptionSrc)
	addUpdate("photo_type", p.PhotoType != original.PhotoType, p.PhotoType)
	addUpdate("type_src", p.TypeSrc != original.TypeSrc, p.TypeSrc)
	addUpdate("photo_favorite", p.PhotoFavorite != original.PhotoFavorite, p.PhotoFavorite)
	addUpdate("photo_private", p.PhotoPrivate != original.PhotoPrivate, p.PhotoPrivate)
	addUpdate("photo_scan", p.PhotoScan != original.PhotoScan, p.PhotoScan)
	addUpdate("photo_panorama", p.PhotoPanorama != original.PhotoPanorama, p.PhotoPanorama)
	addUpdate("photo_year", p.PhotoYear != original.PhotoYear, p.PhotoYear)
	addUpdate("photo_month", p.PhotoMonth != original.PhotoMonth, p.PhotoMonth)
	addUpdate("photo_day", p.PhotoDay != original.PhotoDay, p.PhotoDay)
	addUpdate("time_zone", p.TimeZone != original.TimeZone, p.TimeZone)
	addUpdate("taken_src", p.TakenSrc != original.TakenSrc, p.TakenSrc)
	addUpdate("taken_at", !p.TakenAt.Equal(original.TakenAt), p.TakenAt)
	addUpdate("taken_at_local", !p.TakenAtLocal.Equal(original.TakenAtLocal), p.TakenAtLocal)
	addUpdate("photo_lat", p.PhotoLat != original.PhotoLat, p.PhotoLat)
	addUpdate("photo_lng", p.PhotoLng != original.PhotoLng, p.PhotoLng)
	addUpdate("photo_altitude", p.PhotoAltitude != original.PhotoAltitude, p.PhotoAltitude)
	addUpdate("photo_country", p.PhotoCountry != original.PhotoCountry, p.PhotoCountry)
	addUpdate("place_id", p.PlaceID != original.PlaceID, p.PlaceID)
	addUpdate("cell_id", p.CellID != original.CellID, p.CellID)
	addUpdate("place_src", p.PlaceSrc != original.PlaceSrc, p.PlaceSrc)

	detailUpdates := entity.Values{}
	if details.Subject != origDetails.Subject {
		detailUpdates["subject"] = details.Subject
	}
	if details.SubjectSrc != origDetails.SubjectSrc {
		detailUpdates["subject_src"] = details.SubjectSrc
	}
	if details.Artist != origDetails.Artist {
		detailUpdates["artist"] = details.Artist
	}
	if details.ArtistSrc != origDetails.ArtistSrc {
		detailUpdates["artist_src"] = details.ArtistSrc
	}
	if details.Copyright != origDetails.Copyright {
		detailUpdates["copyright"] = details.Copyright
	}
	if details.CopyrightSrc != origDetails.CopyrightSrc {
		detailUpdates["copyright_src"] = details.CopyrightSrc
	}
	if details.License != origDetails.License {
		detailUpdates["license"] = details.License
	}
	if details.LicenseSrc != origDetails.LicenseSrc {
		detailUpdates["license_src"] = details.LicenseSrc
	}
	if details.Keywords != origDetails.Keywords {
		detailUpdates["keywords"] = details.Keywords
	}

	if len(updates) == 0 && len(detailUpdates) == 0 {
		return false, nil
	}

	log.Debugf("batch: saving photo %s with updates=%v details=%v", p.PhotoUID, updates, detailUpdates)

	edited := entity.Now()
	updates["edited_at"] = edited
	p.EditedAt = &edited
	updates["updated_at"] = edited
	p.UpdatedAt = edited
	updates["checked_at"] = nil
	p.CheckedAt = nil

	if err := p.Updates(updates); err != nil {
		return false, err
	}

	if len(detailUpdates) > 0 {
		if err := details.Updates(detailUpdates); err != nil {
			return false, err
		}
	}

	return true, nil
}

// shouldUpdateString reports whether the provided field requests an update or removal.
func shouldUpdateString(v String) bool {
	return v.Action == ActionUpdate || v.Action == ActionRemove
}
