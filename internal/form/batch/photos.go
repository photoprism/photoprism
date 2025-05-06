package batch

import (
	"github.com/photoprism/photoprism/internal/entity/search"
)

// PhotosForm represents photo batch edit form values.
type PhotosForm struct {
	PhotoType        String  `json:"Type"`
	PhotoTitle       String  `json:"Title"`
	PhotoCaption     String  `json:"Caption"`
	TakenAt          Time    `json:"TakenAt"`
	TakenAtLocal     Time    `json:"TakenAtLocal"`
	PhotoDay         Int     `json:"Day"`
	PhotoMonth       Int     `json:"Month"`
	PhotoYear        Int     `json:"Year"`
	TimeZone         String  `json:"TimeZone"`
	PhotoCountry     String  `json:"Country"`
	PhotoAltitude    Int     `json:"Altitude"`
	PhotoLat         Float64 `json:"Lat"`
	PhotoLng         Float64 `json:"Lng"`
	PhotoIso         Int     `json:"Iso"`
	PhotoFocalLength Int     `json:"FocalLength"`
	PhotoFNumber     Float32 `json:"FNumber"`
	PhotoExposure    String  `json:"Exposure"`
	PhotoFavorite    Bool    `json:"Favorite"`
	PhotoPrivate     Bool    `json:"Private"`
	PhotoScan        Bool    `json:"Scan"`
	PhotoPanorama    Bool    `json:"Panorama"`
	CameraID         UInt    `json:"CameraID"`
	LensID           UInt    `json:"LensID"`

	DetailsKeywords  String `json:"DetailsKeywords"`
	DetailsSubject   String `json:"DetailsSubject"`
	DetailsArtist    String `json:"DetailsArtist"`
	DetailsCopyright String `json:"DetailsCopyright"`
	DetailsLicense   String `json:"DetailsLicense"`
}

func NewPhotosForm(photos search.PhotoResults) *PhotosForm {
	frm := &PhotosForm{}

	for i, photo := range photos {
		if i == 0 {
			frm.PhotoType.Value = photo.PhotoType
		} else if photo.PhotoType != frm.PhotoType.Value {
			frm.PhotoType.Mixed = true
			frm.PhotoType.Value = ""
		}

		if i == 0 {
			frm.PhotoTitle.Value = photo.PhotoTitle
		} else if photo.PhotoTitle != frm.PhotoTitle.Value {
			frm.PhotoTitle.Mixed = true
			frm.PhotoTitle.Value = ""
		}

		if i == 0 {
			frm.PhotoCaption.Value = photo.PhotoCaption
		} else if photo.PhotoCaption != frm.PhotoCaption.Value {
			frm.PhotoCaption.Mixed = true
			frm.PhotoCaption.Value = ""
		}

		if i == 0 {
			frm.TakenAt.Value = photo.TakenAt
		} else if photo.TakenAt != frm.TakenAt.Value {
			frm.TakenAt.Mixed = true
		}

		if i == 0 {
			frm.TakenAtLocal.Value = photo.TakenAtLocal
		} else if photo.TakenAtLocal != frm.TakenAtLocal.Value {
			frm.TakenAtLocal.Mixed = true
		}

		if i == 0 {
			frm.PhotoDay.Value = photo.PhotoDay
		} else if photo.PhotoDay != frm.PhotoDay.Value {
			frm.PhotoDay.Mixed = true
			frm.PhotoDay.Value = 0
		}

		if i == 0 {
			frm.PhotoMonth.Value = photo.PhotoMonth
		} else if photo.PhotoMonth != frm.PhotoMonth.Value {
			frm.PhotoMonth.Mixed = true
			frm.PhotoMonth.Value = 0
		}

		if i == 0 {
			frm.PhotoYear.Value = photo.PhotoYear
		} else if photo.PhotoYear != frm.PhotoYear.Value {
			frm.PhotoYear.Mixed = true
			frm.PhotoYear.Value = 0
		}

		if i == 0 {
			frm.TimeZone.Value = photo.TimeZone
		} else if photo.TimeZone != frm.TimeZone.Value {
			frm.TimeZone.Mixed = true
			frm.TimeZone.Value = "Local"
		}

		if i == 0 {
			frm.PhotoCountry.Value = photo.PhotoCountry
		} else if photo.PhotoCountry != frm.PhotoCountry.Value {
			frm.PhotoCountry.Mixed = true
			frm.PhotoCountry.Value = "zz"
		}

		if i == 0 {
			frm.PhotoAltitude.Value = photo.PhotoAltitude
		} else if photo.PhotoAltitude != frm.PhotoAltitude.Value {
			frm.PhotoAltitude.Mixed = true
			frm.PhotoAltitude.Value = 0
		}

		if i == 0 {
			frm.PhotoLat.Value = photo.PhotoLat
		} else if photo.PhotoLat != frm.PhotoLat.Value {
			frm.PhotoLat.Mixed = true
			frm.PhotoLat.Value = 0.0
		}

		if i == 0 {
			frm.PhotoLng.Value = photo.PhotoLng
		} else if photo.PhotoLng != frm.PhotoLng.Value {
			frm.PhotoLng.Mixed = false
			frm.PhotoLng.Value = 0.0
		}

		if i == 0 {
			frm.PhotoIso.Value = photo.PhotoIso
		} else if photo.PhotoIso != frm.PhotoIso.Value {
			frm.PhotoIso.Mixed = true
			frm.PhotoIso.Value = 0
		}

		if i == 0 {
			frm.PhotoFocalLength.Value = photo.PhotoFocalLength
		} else if photo.PhotoFocalLength != frm.PhotoFocalLength.Value {
			frm.PhotoFocalLength.Mixed = true
			frm.PhotoFocalLength.Value = 0
		}

		if i == 0 {
			frm.PhotoFNumber.Value = photo.PhotoFNumber
		} else if photo.PhotoFNumber != frm.PhotoFNumber.Value {
			frm.PhotoFNumber.Mixed = true
			frm.PhotoFNumber.Value = 0
		}

		if i == 0 {
			frm.PhotoExposure.Value = photo.PhotoExposure
		} else if photo.PhotoExposure != frm.PhotoExposure.Value {
			frm.PhotoExposure.Mixed = true
			frm.PhotoExposure.Value = ""
		}

		if i == 0 {
			frm.PhotoFavorite.Value = photo.PhotoFavorite
		} else if photo.PhotoFavorite != frm.PhotoFavorite.Value {
			frm.PhotoFavorite.Mixed = true
			frm.PhotoFavorite.Value = false
		}

		if i == 0 {
			frm.PhotoPrivate.Value = photo.PhotoPrivate
		} else if photo.PhotoPrivate != frm.PhotoPrivate.Value {
			frm.PhotoPrivate.Mixed = true
			frm.PhotoPrivate.Value = false
		}

		if i == 0 {
			frm.PhotoScan.Value = photo.PhotoScan
		} else if photo.PhotoScan != frm.PhotoScan.Value {
			frm.PhotoScan.Mixed = true
			frm.PhotoScan.Value = false
		}

		if i == 0 {
			frm.PhotoPanorama.Value = photo.PhotoPanorama
		} else if photo.PhotoPanorama != frm.PhotoPanorama.Value {
			frm.PhotoPanorama.Mixed = true
			frm.PhotoPanorama.Value = false
		}

		if i == 0 {
			frm.CameraID.Value = photo.CameraID
		} else if photo.CameraID != frm.CameraID.Value {
			frm.CameraID.Mixed = true
			frm.CameraID.Value = 1
		}

		if i == 0 {
			frm.LensID.Value = photo.LensID
		} else if photo.LensID != frm.LensID.Value {
			frm.LensID.Mixed = true
			frm.LensID.Value = 1
		}

		if i == 0 {
			frm.DetailsKeywords.Value = photo.DetailsKeywords
		} else if photo.DetailsKeywords != frm.DetailsKeywords.Value {
			frm.DetailsKeywords.Mixed = true
			frm.DetailsKeywords.Value = ""
		}

		if i == 0 {
			frm.DetailsSubject.Value = photo.DetailsSubject
		} else if photo.DetailsSubject != frm.DetailsSubject.Value {
			frm.DetailsSubject.Mixed = true
			frm.DetailsSubject.Value = ""
		}

		if i == 0 {
			frm.DetailsArtist.Value = photo.DetailsArtist
		} else if photo.DetailsArtist != frm.DetailsArtist.Value {
			frm.DetailsArtist.Mixed = true
			frm.DetailsArtist.Value = ""
		}

		if i == 0 {
			frm.DetailsCopyright.Value = photo.DetailsCopyright
		} else if photo.DetailsCopyright != frm.DetailsCopyright.Value {
			frm.DetailsCopyright.Mixed = true
			frm.DetailsCopyright.Value = ""
		}

		if i == 0 {
			frm.DetailsLicense.Value = photo.DetailsLicense
		} else if photo.DetailsLicense != frm.DetailsLicense.Value {
			frm.DetailsLicense.Mixed = true
			frm.DetailsLicense.Value = ""
		}
	}

	// Use defaults for the following values if they are empty:
	if frm.PhotoCountry.Value == "" {
		frm.PhotoCountry.Value = "zz"
	}

	if frm.CameraID.Value < 1 {
		frm.CameraID.Value = 1
	}

	if frm.LensID.Value < 1 {
		frm.LensID.Value = 1
	}

	return frm
}
