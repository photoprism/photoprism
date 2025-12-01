package batch

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
)

func TestConvertToPhotoForm(t *testing.T) {
	t.Run("NilInputs", func(t *testing.T) {
		form, err := ConvertToPhotoForm(nil, nil)
		assert.Nil(t, form)
		assert.Error(t, err)
	})
	t.Run("UpdateTitleCaptionTypeAndBooleans", func(t *testing.T) {
		photo := &entity.Photo{
			PhotoTitle:   "Old Title",
			PhotoCaption: "Old Caption",
			PhotoType:    "image",
		}

		v := &PhotosForm{}
		v.PhotoTitle.Action = ActionUpdate
		v.PhotoTitle.Value = "New Title"

		v.PhotoCaption.Action = ActionRemove // remove caption
		v.PhotoCaption.Value = "whatever"

		v.PhotoType.Action = ActionUpdate
		v.PhotoType.Value = "live"

		v.PhotoFavorite.Action = ActionUpdate
		v.PhotoFavorite.Value = true

		v.PhotoPrivate.Action = ActionUpdate
		v.PhotoPrivate.Value = true

		v.PhotoScan.Action = ActionUpdate
		v.PhotoScan.Value = true

		v.PhotoPanorama.Action = ActionUpdate
		v.PhotoPanorama.Value = true

		form, err := ConvertToPhotoForm(photo, v)
		assert.NoError(t, err)

		assert.Equal(t, "New Title", form.PhotoTitle)
		assert.Equal(t, entity.SrcBatch, form.TitleSrc)

		assert.Equal(t, "", form.PhotoCaption)
		assert.Equal(t, entity.SrcBatch, form.CaptionSrc)

		assert.Equal(t, "live", form.PhotoType)
		assert.Equal(t, entity.SrcBatch, form.TypeSrc)

		assert.True(t, form.PhotoFavorite)
		assert.True(t, form.PhotoPrivate)
		assert.True(t, form.PhotoScan)
		assert.True(t, form.PhotoPanorama)
	})
	t.Run("UpdateLocationSetsPlaceSrc", func(t *testing.T) {
		photo := &entity.Photo{}

		v := &PhotosForm{}
		v.PhotoLat.Action = ActionUpdate
		v.PhotoLat.Value = 37.5
		v.PhotoLng.Action = ActionUpdate
		v.PhotoLng.Value = -122.4
		v.PhotoCountry.Action = ActionUpdate
		v.PhotoCountry.Value = "us"
		v.PhotoAltitude.Action = ActionUpdate
		v.PhotoAltitude.Value = 15

		form, err := ConvertToPhotoForm(photo, v)
		assert.NoError(t, err)

		assert.Equal(t, 37.5, form.PhotoLat)
		assert.Equal(t, -122.4, form.PhotoLng)
		assert.Equal(t, "us", form.PhotoCountry)
		assert.Equal(t, 15, form.PhotoAltitude)
		assert.Equal(t, entity.SrcBatch, form.PlaceSrc)
	})
	t.Run("TitleRemove_CaptionUpdate", func(t *testing.T) {
		photo := &entity.Photo{PhotoTitle: "Old", PhotoCaption: "OldCap"}

		v := &PhotosForm{}
		v.PhotoTitle.Action = ActionRemove
		v.PhotoCaption.Action = ActionUpdate
		v.PhotoCaption.Value = "NewCap"

		form, err := ConvertToPhotoForm(photo, v)
		assert.NoError(t, err)
		assert.Equal(t, "", form.PhotoTitle)
		assert.Equal(t, entity.SrcBatch, form.TitleSrc)
		assert.Equal(t, "NewCap", form.PhotoCaption)
		assert.Equal(t, entity.SrcBatch, form.CaptionSrc)
	})
	t.Run("TimeZoneUpdateSetsTakenSrc", func(t *testing.T) {
		photo := &entity.Photo{}

		v := &PhotosForm{}
		v.TimeZone.Action = ActionUpdate
		v.TimeZone.Value = "Europe/Berlin"

		form, err := ConvertToPhotoForm(photo, v)
		assert.NoError(t, err)
		assert.Equal(t, "Europe/Berlin", form.TimeZone)
		assert.Equal(t, entity.SrcBatch, form.TakenSrc)
	})
	t.Run("YearUpdate_RecomputesTakenAtLocalAndOutputs", func(t *testing.T) {
		photo := &entity.Photo{}
		// Assume current date fields are set via NewPhoto(photo)

		v := &PhotosForm{}
		v.PhotoYear.Action = ActionUpdate
		v.PhotoYear.Value = 2021

		form, err := ConvertToPhotoForm(photo, v)
		assert.NoError(t, err)
		// Must set TakenSrc and keep PhotoDay consistent
		assert.Equal(t, entity.SrcBatch, form.TakenSrc)
		assert.Equal(t, 2021, form.PhotoYear)
		// PhotoDay is set via ComputeDateChange; must be non-zero or -1 depending on base
		assert.NotZero(t, form.PhotoDay)
	})
	t.Run("UpdateDetails", func(t *testing.T) {
		photo := &entity.Photo{}

		v := &PhotosForm{}
		v.DetailsSubject.Action = ActionUpdate
		v.DetailsSubject.Value = "Subject"
		v.DetailsArtist.Action = ActionUpdate
		v.DetailsArtist.Value = "Artist"
		v.DetailsCopyright.Action = ActionUpdate
		v.DetailsCopyright.Value = "Copyright"
		v.DetailsLicense.Action = ActionUpdate
		v.DetailsLicense.Value = "License"

		form, err := ConvertToPhotoForm(photo, v)
		assert.NoError(t, err)

		assert.Equal(t, "Subject", form.Details.Subject)
		assert.Equal(t, entity.SrcBatch, form.Details.SubjectSrc)
		assert.Equal(t, "Artist", form.Details.Artist)
		assert.Equal(t, entity.SrcBatch, form.Details.ArtistSrc)
		assert.Equal(t, "Copyright", form.Details.Copyright)
		assert.Equal(t, entity.SrcBatch, form.Details.CopyrightSrc)
		assert.Equal(t, "License", form.Details.License)
		assert.Equal(t, entity.SrcBatch, form.Details.LicenseSrc)
	})
	t.Run("RemoveDetailsFieldsAndCarryPhotoID", func(t *testing.T) {
		photo := &entity.Photo{ID: 42}

		v := &PhotosForm{}
		v.DetailsSubject.Action = ActionRemove
		v.DetailsArtist.Action = ActionRemove
		v.DetailsCopyright.Action = ActionRemove
		v.DetailsLicense.Action = ActionRemove

		form, err := ConvertToPhotoForm(photo, v)
		assert.NoError(t, err)
		assert.Equal(t, uint(42), form.Details.PhotoID)
		assert.Equal(t, "", form.Details.Subject)
		assert.Equal(t, entity.SrcBatch, form.Details.SubjectSrc)
		assert.Equal(t, "", form.Details.Artist)
		assert.Equal(t, entity.SrcBatch, form.Details.ArtistSrc)
		assert.Equal(t, "", form.Details.Copyright)
		assert.Equal(t, entity.SrcBatch, form.Details.CopyrightSrc)
		assert.Equal(t, "", form.Details.License)
		assert.Equal(t, entity.SrcBatch, form.Details.LicenseSrc)
	})
}
