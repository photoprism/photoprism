package batch

import (
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
)

// ConvertToPhotoForm converts PhotosForm into a regular form.Photo while
// preserving unchanged fields and marking source fields with SrcBatch where applicable.
// Note: ISO / focal length / f-number / exposure / camera / lens / details keywords
// are present on PhotosForm for future batch UI work. Once inputs exist, extend this
// function to copy those fields (and set Details.KeywordsSrc) before removing this note.
func ConvertToPhotoForm(photo *entity.Photo, v *PhotosForm) (*form.Photo, error) {
	if photo == nil || v == nil {
		return nil, fmt.Errorf("photo or batch values is nil")
	}

	// Start with a form created from the current photo
	photoForm, err := form.NewPhoto(photo)
	if err != nil {
		return nil, fmt.Errorf("failed to create form from photo: %w", err)
	}

	switch v.PhotoTitle.Action {
	case ActionUpdate:
		photoForm.PhotoTitle = v.PhotoTitle.Value
		photoForm.TitleSrc = entity.SrcBatch
	case ActionRemove:
		photoForm.PhotoTitle = ""
		photoForm.TitleSrc = entity.SrcBatch
	}

	switch v.PhotoCaption.Action {
	case ActionUpdate:
		photoForm.PhotoCaption = v.PhotoCaption.Value
		photoForm.CaptionSrc = entity.SrcBatch
	case ActionRemove:
		photoForm.PhotoCaption = ""
		photoForm.CaptionSrc = entity.SrcBatch
	}

	if v.PhotoType.Action == ActionUpdate {
		photoForm.PhotoType = v.PhotoType.Value
		photoForm.TypeSrc = entity.SrcBatch
	}

	// Date/time fields
	timeChanged := false
	if v.PhotoDay.Action == ActionUpdate {
		photoForm.PhotoDay = v.PhotoDay.Value
		timeChanged = true
	}
	if v.PhotoMonth.Action == ActionUpdate {
		photoForm.PhotoMonth = v.PhotoMonth.Value
		timeChanged = true
	}
	if v.PhotoYear.Action == ActionUpdate {
		photoForm.PhotoYear = v.PhotoYear.Value
		timeChanged = true
	}
	if v.TimeZone.Action == ActionUpdate {
		photoForm.TimeZone = v.TimeZone.Value
		timeChanged = true
	}
	if timeChanged {
		photoForm.TakenSrc = entity.SrcBatch
	}

	// If any date field changed, recompute TakenAtLocal per photo via ComputeDateChange.
	if v.PhotoDay.Action == ActionUpdate || v.PhotoMonth.Action == ActionUpdate || v.PhotoYear.Action == ActionUpdate {
		base := photo.TakenAtLocal
		newLocal, outYear, outMonth, outDay := ComputeDateChange(
			base,
			photoForm.PhotoYear, photoForm.PhotoMonth, photoForm.PhotoDay,
			v.PhotoDay.Action, v.PhotoDay.Value,
			v.PhotoMonth.Action, v.PhotoMonth.Value,
			v.PhotoYear.Action, v.PhotoYear.Value,
		)

		photoForm.TakenAtLocal = newLocal
		photoForm.TakenSrc = entity.SrcBatch

		// Apply returned date components respecting unknown rules.
		if v.PhotoYear.Action == ActionUpdate {
			photoForm.PhotoYear = outYear
		}
		if v.PhotoMonth.Action == ActionUpdate {
			photoForm.PhotoMonth = outMonth
		}
		// Always keep PhotoDay consistent with computed result when any date field changed.
		photoForm.PhotoDay = outDay
	}

	// Location fields
	locationChanged := false
	if v.PhotoLat.Action == ActionUpdate {
		photoForm.PhotoLat = v.PhotoLat.Value
		locationChanged = true
	}
	if v.PhotoLng.Action == ActionUpdate {
		photoForm.PhotoLng = v.PhotoLng.Value
		locationChanged = true
	}
	if v.PhotoCountry.Action == ActionUpdate {
		photoForm.PhotoCountry = v.PhotoCountry.Value
		locationChanged = true
	}
	if v.PhotoAltitude.Action == ActionUpdate {
		photoForm.PhotoAltitude = v.PhotoAltitude.Value
		locationChanged = true
	}
	if locationChanged {
		photoForm.PlaceSrc = entity.SrcBatch
	}

	// Boolean flags
	if v.PhotoFavorite.Action == ActionUpdate {
		photoForm.PhotoFavorite = v.PhotoFavorite.Value
	}
	if v.PhotoPrivate.Action == ActionUpdate {
		photoForm.PhotoPrivate = v.PhotoPrivate.Value
	}
	if v.PhotoScan.Action == ActionUpdate {
		photoForm.PhotoScan = v.PhotoScan.Value
	}
	if v.PhotoPanorama.Action == ActionUpdate {
		photoForm.PhotoPanorama = v.PhotoPanorama.Value
	}

	// Details fields - preserve existing values, only update changed ones
	currentDetails := photo.GetDetails()
	if currentDetails != nil {
		// Start with current values to preserve unchanged fields
		photoForm.Details.Subject = currentDetails.Subject
		photoForm.Details.SubjectSrc = currentDetails.SubjectSrc
		photoForm.Details.Artist = currentDetails.Artist
		photoForm.Details.ArtistSrc = currentDetails.ArtistSrc
		photoForm.Details.Copyright = currentDetails.Copyright
		photoForm.Details.CopyrightSrc = currentDetails.CopyrightSrc
		photoForm.Details.License = currentDetails.License
		photoForm.Details.LicenseSrc = currentDetails.LicenseSrc
		photoForm.Details.Keywords = currentDetails.Keywords
		photoForm.Details.KeywordsSrc = currentDetails.KeywordsSrc
		photoForm.Details.Notes = currentDetails.Notes
		photoForm.Details.NotesSrc = currentDetails.NotesSrc
	}

	switch v.DetailsSubject.Action {
	case ActionUpdate:
		photoForm.Details.Subject = v.DetailsSubject.Value
		photoForm.Details.SubjectSrc = entity.SrcBatch
	case ActionRemove:
		photoForm.Details.Subject = ""
		photoForm.Details.SubjectSrc = entity.SrcBatch
	}

	switch v.DetailsArtist.Action {
	case ActionUpdate:
		photoForm.Details.Artist = v.DetailsArtist.Value
		photoForm.Details.ArtistSrc = entity.SrcBatch
	case ActionRemove:
		photoForm.Details.Artist = ""
		photoForm.Details.ArtistSrc = entity.SrcBatch
	}

	switch v.DetailsCopyright.Action {
	case ActionUpdate:
		photoForm.Details.Copyright = v.DetailsCopyright.Value
		photoForm.Details.CopyrightSrc = entity.SrcBatch
	case ActionRemove:
		photoForm.Details.Copyright = ""
		photoForm.Details.CopyrightSrc = entity.SrcBatch
	}

	switch v.DetailsLicense.Action {
	case ActionUpdate:
		photoForm.Details.License = v.DetailsLicense.Value
		photoForm.Details.LicenseSrc = entity.SrcBatch
	case ActionRemove:
		photoForm.Details.License = ""
		photoForm.Details.LicenseSrc = entity.SrcBatch
	}

	// Set the PhotoID for details
	photoForm.Details.PhotoID = photo.ID

	return &photoForm, nil
}
