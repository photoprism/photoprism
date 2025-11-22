package batch

import (
	"sort"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
)

// PhotosForm represents photo batch edit form values.
// Several EXIF / details fields (ISO, focal length, f-number, exposure,
// camera/lens IDs, DetailsKeywords) are included here so their mixed state
// survives round-trips even though the batch dialog doesn’t expose inputs yet.
// Once the frontend adds those controls, extend ConvertToPhotoForm and AddLabels
// to persist the values before removing this note.
type PhotosForm struct {
	PhotoType        String  `json:"Type,omitempty"`
	PhotoTitle       String  `json:"Title,omitempty"`
	PhotoCaption     String  `json:"Caption,omitempty"`
	TakenAt          Time    `json:"TakenAt,omitempty"`
	TakenAtLocal     Time    `json:"TakenAtLocal,omitempty"`
	PhotoDay         Int     `json:"Day,omitempty"`
	PhotoMonth       Int     `json:"Month,omitempty"`
	PhotoYear        Int     `json:"Year,omitempty"`
	TimeZone         String  `json:"TimeZone,omitempty"`
	PhotoCountry     String  `json:"Country,omitempty"`
	PhotoAltitude    Int     `json:"Altitude,omitempty"`
	PhotoLat         Float64 `json:"Lat,omitempty"`
	PhotoLng         Float64 `json:"Lng,omitempty"`
	PhotoIso         Int     `json:"Iso,omitempty"`
	PhotoFocalLength Int     `json:"FocalLength,omitempty"`
	PhotoFNumber     Float32 `json:"FNumber,omitempty"`
	PhotoExposure    String  `json:"Exposure,omitempty"`
	PhotoFavorite    Bool    `json:"Favorite,omitempty"`
	PhotoPrivate     Bool    `json:"Private,omitempty"`
	PhotoScan        Bool    `json:"Scan,omitempty"`
	PhotoPanorama    Bool    `json:"Panorama,omitempty"`
	CameraID         Int     `json:"CameraID,omitempty"`
	LensID           Int     `json:"LensID,omitempty"`
	Albums           Items   `json:"Albums,omitempty"`
	Labels           Items   `json:"Labels,omitempty"`

	DetailsKeywords  String `json:"DetailsKeywords,omitempty"`
	DetailsSubject   String `json:"DetailsSubject,omitempty"`
	DetailsArtist    String `json:"DetailsArtist,omitempty"`
	DetailsCopyright String `json:"DetailsCopyright,omitempty"`
	DetailsLicense   String `json:"DetailsLicense,omitempty"`
}

// NewPhotosForm returns a new batch edit form instance initialized with values from the
// selected photos. It falls back to reloading each photo individually.
func NewPhotosForm(photos search.PhotoResults) *PhotosForm {
	return NewPhotosFormWithEntities(photos, nil)
}

// NewPhotosFormWithEntities builds the batch edit form using the supplied photo selection and
// an optional map of preloaded entity.Photo instances that should be reused instead of issuing
// additional queries. When the map is nil or a photo is missing, the helper falls back to
// `query.PhotoPreloadByUID` so legacy call sites keep working.
func NewPhotosFormWithEntities(photos search.PhotoResults, preloaded map[string]*entity.Photo) *PhotosForm {
	// Create a new batch edit form and initialize it
	// with the values from the selected photos.
	frm := &PhotosForm{
		Albums: Items{Items: []Item{}, Mixed: false, Action: ActionNone},
		Labels: Items{Items: []Item{}, Mixed: false, Action: ActionNone},
	}

	getPhoto := func(uid string) *entity.Photo {
		if uid == "" {
			return nil
		}

		if preloaded != nil {
			if p := preloaded[uid]; p != nil {
				return p
			}
		}

		if p, err := query.PhotoPreloadByUID(uid); err == nil && p.HasID() {
			return &p
		}

		return nil
	}

	// Populate Albums and Labels from selected photos (no raw SQL; use preload helpers).
	total := len(photos)
	if total > 0 {
		type albumAgg struct {
			title string
			cnt   int
		}

		type labelAgg struct {
			name string
			cnt  int
		}

		albumCount := map[string]albumAgg{}
		labelCount := map[string]labelAgg{}

		for _, sp := range photos {
			if sp.PhotoUID == "" {
				continue
			}
			p := getPhoto(sp.PhotoUID)
			if p == nil || !p.HasID() {
				continue
			}

			// Albums on this photo
			for _, a := range p.Albums {
				if a.AlbumUID == "" || a.Deleted() {
					continue
				}
				v := albumCount[a.AlbumUID]
				v.title = a.AlbumTitle
				v.cnt++
				albumCount[a.AlbumUID] = v
			}

			// Labels on this photo (only visible ones: uncertainty < 100).
			for _, pl := range p.Labels {
				if pl.Uncertainty >= 100 || pl.Label == nil || !pl.Label.HasID() {
					continue
				}
				uid := pl.Label.LabelUID
				if uid == "" {
					continue
				}
				v := labelCount[uid]
				v.name = pl.Label.LabelName
				v.cnt++
				labelCount[uid] = v
			}
		}

		// Build Albums items.
		frm.Albums.Items = make([]Item, 0, len(albumCount))
		anyAlbumMixed := false

		for uid, agg := range albumCount {
			mixed := agg.cnt > 0 && agg.cnt < total
			if mixed {
				anyAlbumMixed = true
			}
			frm.Albums.Items = append(frm.Albums.Items, Item{Value: uid, Title: agg.title, Mixed: mixed, Action: ActionNone})
		}

		// Sort shared-first (Mixed=false), then by Title alphabetically.
		sort.Slice(frm.Albums.Items, func(i, j int) bool {
			if frm.Albums.Items[i].Mixed != frm.Albums.Items[j].Mixed {
				return !frm.Albums.Items[i].Mixed && frm.Albums.Items[j].Mixed
			}
			return frm.Albums.Items[i].Title < frm.Albums.Items[j].Title
		})

		frm.Albums.Mixed = anyAlbumMixed
		frm.Albums.Action = ActionNone

		// Build Labels items.
		frm.Labels.Items = make([]Item, 0, len(labelCount))
		anyLabelMixed := false
		for uid, agg := range labelCount {
			mixed := agg.cnt > 0 && agg.cnt < total
			if mixed {
				anyLabelMixed = true
			}
			frm.Labels.Items = append(frm.Labels.Items, Item{Value: uid, Title: agg.name, Mixed: mixed, Action: ActionNone})
		}

		// Sort shared-first (Mixed=false), then by Title alphabetically.
		sort.Slice(frm.Labels.Items, func(i, j int) bool {
			if frm.Labels.Items[i].Mixed != frm.Labels.Items[j].Mixed {
				return !frm.Labels.Items[i].Mixed && frm.Labels.Items[j].Mixed
			}
			return frm.Labels.Items[i].Title < frm.Labels.Items[j].Title
		})

		frm.Labels.Mixed = anyLabelMixed
		frm.Labels.Action = ActionNone
	}

	// TODO: Verify that all required PhotosForm values are present and
	//       properly initialized or use in the frontend form component.
	for i, photo := range photos {
		if i == 0 {
			frm.PhotoType.Value = photo.PhotoType
			frm.PhotoType.Action = ActionNone
		} else if photo.PhotoType != frm.PhotoType.Value {
			frm.PhotoType.Mixed = true
			frm.PhotoType.Value = ""
		}

		if i == 0 {
			frm.PhotoTitle.Value = photo.PhotoTitle
			frm.PhotoTitle.Action = ActionNone
		} else if photo.PhotoTitle != frm.PhotoTitle.Value {
			frm.PhotoTitle.Mixed = true
			frm.PhotoTitle.Value = ""
		}

		if i == 0 {
			frm.PhotoCaption.Value = photo.PhotoCaption
			frm.PhotoCaption.Action = ActionNone
		} else if photo.PhotoCaption != frm.PhotoCaption.Value {
			frm.PhotoCaption.Mixed = true
			frm.PhotoCaption.Value = ""
		}

		if i == 0 {
			frm.TakenAt.Value = photo.TakenAt
			frm.TakenAt.Action = ActionNone
		} else if photo.TakenAt != frm.TakenAt.Value {
			frm.TakenAt.Mixed = true
		}

		if i == 0 {
			frm.TakenAtLocal.Value = photo.TakenAtLocal
			frm.TakenAtLocal.Action = ActionNone
		} else if photo.TakenAtLocal != frm.TakenAtLocal.Value {
			frm.TakenAtLocal.Mixed = true
		}

		if i == 0 {
			frm.PhotoDay.Value = photo.PhotoDay
			frm.PhotoDay.Action = ActionNone
		} else if photo.PhotoDay != frm.PhotoDay.Value {
			frm.PhotoDay.Mixed = true
			frm.PhotoDay.Value = -2
		}

		if i == 0 {
			frm.PhotoMonth.Value = photo.PhotoMonth
			frm.PhotoMonth.Action = ActionNone
		} else if photo.PhotoMonth != frm.PhotoMonth.Value {
			frm.PhotoMonth.Mixed = true
			frm.PhotoMonth.Value = -2
		}

		if i == 0 {
			frm.PhotoYear.Value = photo.PhotoYear
			frm.PhotoYear.Action = ActionNone
		} else if photo.PhotoYear != frm.PhotoYear.Value {
			frm.PhotoYear.Mixed = true
			frm.PhotoYear.Value = -2
		}

		if i == 0 {
			frm.TimeZone.Value = photo.TimeZone
			frm.TimeZone.Action = ActionNone
		} else if photo.TimeZone != frm.TimeZone.Value {
			frm.TimeZone.Mixed = true
			frm.TimeZone.Value = ""
		}

		if i == 0 {
			frm.PhotoCountry.Value = photo.PhotoCountry
			frm.PhotoCountry.Action = ActionNone
		} else if photo.PhotoCountry != frm.PhotoCountry.Value {
			frm.PhotoCountry.Mixed = true
			frm.PhotoCountry.Value = ""
		}

		if i == 0 {
			frm.PhotoAltitude.Value = photo.PhotoAltitude
			frm.PhotoAltitude.Action = ActionNone
		} else if photo.PhotoAltitude != frm.PhotoAltitude.Value {
			frm.PhotoAltitude.Mixed = true
			frm.PhotoAltitude.Value = 0
		}

		if i == 0 {
			frm.PhotoLat.Value = photo.PhotoLat
			frm.PhotoLat.Action = ActionNone
		} else if photo.PhotoLat != frm.PhotoLat.Value {
			frm.PhotoLat.Mixed = true
			frm.PhotoLat.Value = 0.0
		}

		if i == 0 {
			frm.PhotoLng.Value = photo.PhotoLng
			frm.PhotoLng.Action = ActionNone
		} else if photo.PhotoLng != frm.PhotoLng.Value {
			frm.PhotoLng.Mixed = true
			frm.PhotoLng.Value = 0.0
		}

		if i == 0 {
			frm.PhotoIso.Value = photo.PhotoIso
			frm.PhotoIso.Action = ActionNone
		} else if photo.PhotoIso != frm.PhotoIso.Value {
			frm.PhotoIso.Mixed = true
			frm.PhotoIso.Value = 0
		}

		if i == 0 {
			frm.PhotoFocalLength.Value = photo.PhotoFocalLength
			frm.PhotoFocalLength.Action = ActionNone
		} else if photo.PhotoFocalLength != frm.PhotoFocalLength.Value {
			frm.PhotoFocalLength.Mixed = true
			frm.PhotoFocalLength.Value = 0
		}

		if i == 0 {
			frm.PhotoFNumber.Value = photo.PhotoFNumber
			frm.PhotoFNumber.Action = ActionNone
		} else if photo.PhotoFNumber != frm.PhotoFNumber.Value {
			frm.PhotoFNumber.Mixed = true
			frm.PhotoFNumber.Value = 0
		}

		if i == 0 {
			frm.PhotoExposure.Value = photo.PhotoExposure
			frm.PhotoExposure.Action = ActionNone
		} else if photo.PhotoExposure != frm.PhotoExposure.Value {
			frm.PhotoExposure.Mixed = true
			frm.PhotoExposure.Value = ""
		}

		if i == 0 {
			frm.PhotoFavorite.Value = photo.PhotoFavorite
			frm.PhotoFavorite.Action = ActionNone
		} else if photo.PhotoFavorite != frm.PhotoFavorite.Value {
			frm.PhotoFavorite.Mixed = true
			frm.PhotoFavorite.Value = false
		}

		if i == 0 {
			frm.PhotoPrivate.Value = photo.PhotoPrivate
			frm.PhotoPrivate.Action = ActionNone
		} else if photo.PhotoPrivate != frm.PhotoPrivate.Value {
			frm.PhotoPrivate.Mixed = true
			frm.PhotoPrivate.Value = false
		}

		if i == 0 {
			frm.PhotoScan.Value = photo.PhotoScan
			frm.PhotoScan.Action = ActionNone
		} else if photo.PhotoScan != frm.PhotoScan.Value {
			frm.PhotoScan.Mixed = true
			frm.PhotoScan.Value = false
		}

		if i == 0 {
			frm.PhotoPanorama.Value = photo.PhotoPanorama
			frm.PhotoPanorama.Action = ActionNone
		} else if photo.PhotoPanorama != frm.PhotoPanorama.Value {
			frm.PhotoPanorama.Mixed = true
			frm.PhotoPanorama.Value = false
		}

		if i == 0 {
			frm.CameraID.Value = int(photo.CameraID)
			frm.CameraID.Action = ActionNone
		} else if photo.CameraID != uint(frm.CameraID.Value) {
			frm.CameraID.Mixed = true
			frm.CameraID.Value = -2
		}

		if i == 0 {
			frm.LensID.Value = int(photo.LensID)
			frm.LensID.Action = ActionNone
		} else if photo.LensID != uint(frm.LensID.Value) {
			frm.LensID.Mixed = true
			frm.LensID.Value = -2
		}

		if i == 0 {
			frm.DetailsKeywords.Value = photo.DetailsKeywords
			frm.DetailsKeywords.Action = ActionNone
		} else if photo.DetailsKeywords != frm.DetailsKeywords.Value {
			frm.DetailsKeywords.Mixed = true
			frm.DetailsKeywords.Value = ""
		}

		if i == 0 {
			frm.DetailsSubject.Value = photo.DetailsSubject
			frm.DetailsSubject.Action = ActionNone
		} else if photo.DetailsSubject != frm.DetailsSubject.Value {
			frm.DetailsSubject.Mixed = true
			frm.DetailsSubject.Value = ""
		}

		if i == 0 {
			frm.DetailsArtist.Value = photo.DetailsArtist
			frm.DetailsArtist.Action = ActionNone
		} else if photo.DetailsArtist != frm.DetailsArtist.Value {
			frm.DetailsArtist.Mixed = true
			frm.DetailsArtist.Value = ""
		}

		if i == 0 {
			frm.DetailsCopyright.Value = photo.DetailsCopyright
			frm.DetailsCopyright.Action = ActionNone
		} else if photo.DetailsCopyright != frm.DetailsCopyright.Value {
			frm.DetailsCopyright.Mixed = true
			frm.DetailsCopyright.Value = ""
		}

		if i == 0 {
			frm.DetailsLicense.Value = photo.DetailsLicense
			frm.DetailsLicense.Action = ActionNone
		} else if photo.DetailsLicense != frm.DetailsLicense.Value {
			frm.DetailsLicense.Mixed = true
			frm.DetailsLicense.Value = ""
		}
	}

	// Return initialized PhotosForm.
	return frm
}
