package mcp

import (
	"strings"
	"time"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
)

// takenAtLocalLayout is the layout of local times, which carry no time zone.
const takenAtLocalLayout = "2006-01-02T15:04:05"

// Media is an item in the library, as returned by the media tools.
type Media struct {
	UID          string `json:"uid"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	TakenAt      string `json:"taken_at"`
	TakenAtLocal string `json:"taken_at_local"`
	Place        *Place `json:"place,omitempty"`
	Favorite     bool   `json:"favorite"`
	Caption      string `json:"caption,omitempty"`
	Camera       string `json:"camera,omitempty"`
	Lens         string `json:"lens,omitempty"`
	Labels       []Ref  `json:"labels,omitempty"`
	People       []Ref  `json:"people,omitempty"`
	Collections  []Ref  `json:"collections,omitempty"`
}

// Place is the place where a media was taken, without coordinates.
type Place struct {
	Name    string `json:"name,omitempty"`
	Country string `json:"country,omitempty"`
}

// Ref refers to another object by its UID, with the fields that name it.
type Ref struct {
	UID   string `json:"uid"`
	Type  string `json:"type,omitempty"`
	Name  string `json:"name,omitempty"`
	Title string `json:"title,omitempty"`
}

// newMedia returns the brief fields of a media from a search result row.
func newMedia(row search.Photo) Media {
	m := Media{
		UID:          row.PhotoUID,
		Type:         row.PhotoType,
		Title:        row.PhotoTitle,
		TakenAt:      row.TakenAt.UTC().Format(time.RFC3339),
		TakenAtLocal: row.TakenAtLocal.Format(takenAtLocalLayout),
		Favorite:     row.PhotoFavorite,
	}

	place := Place{Country: row.PlaceCountry}

	if row.PlaceID != "" && row.PlaceID != entity.UnknownPlace.ID {
		place.Name = row.PlaceLabel
	}

	if place.Country == "" && row.PhotoCountry != entity.UnknownCountry.ID {
		place.Country = row.PhotoCountry
	}

	if place.Country == entity.UnknownCountry.ID {
		place.Country = ""
	}

	if place.Name != "" || place.Country != "" {
		m.Place = &place
	}

	return m
}

// newMediaList returns the brief fields of the media in the search result rows.
func newMediaList(rows search.PhotoResults) []Media {
	result := make([]Media, 0, len(rows))

	for _, row := range rows {
		result = append(result, newMedia(row))
	}

	return result
}

// photoRef is a row that links a photo to a related object.
type photoRef struct {
	PhotoID uint
	UID     string
}

// addMediaDetails adds the full detail fields to the media of the search result rows, in the same order. The
// labels, people, and collections are loaded with one query per kind and filtered by the rules of their own
// resources (search.LabelsPublicOnly, search.UserSubjects, search.UserAlbums), so the session gets only those it may see.
func addMediaDetails(items []Media, rows search.PhotoResults, s *entity.Session) {
	if len(items) == 0 || len(items) != len(rows) {
		return
	}

	ids := make([]uint, len(rows))
	index := make(map[uint]int, len(rows))

	var cameraIDs, lensIDs []uint

	for i, row := range rows {
		ids[i] = row.ID
		index[row.ID] = i
		items[i].Caption = row.PhotoCaption
		cameraIDs = append(cameraIDs, row.CameraID)
		lensIDs = append(lensIDs, row.LensID)
	}

	// Cameras and lenses are named by their display names, without the unknown placeholders.
	cameras := deviceNames("cameras", "camera_name", "camera_slug", cameraIDs)
	lenses := deviceNames("lenses", "lens_name", "lens_slug", lensIDs)

	for i, row := range rows {
		items[i].Camera = cameras[row.CameraID]
		items[i].Lens = lenses[row.LensID]
	}

	// Labels and people are part of the picture details, which require full access to pictures.
	if !s.SeesFullDetail(acl.ResourcePhotos) {
		addMediaCollections(items, index, ids, s)
		return
	}

	// Labels of the pictures, without private labels if the session may see public labels only.
	{
		var refs []struct {
			PhotoID uint
			UID     string
			Name    string
			NSFW    bool
		}

		entity.Log("mcp", "find labels", entity.Db().Table("photos_labels pl").
			Select("pl.photo_id, l.label_uid AS uid, l.label_name AS name, l.label_nsfw AS nsfw").
			Joins("JOIN labels l ON l.id = pl.label_id AND l.deleted_at IS NULL").
			Where("pl.photo_id IN (?) AND pl.uncertainty < 100", ids).
			Order("pl.photo_id, pl.uncertainty, l.label_uid").Scan(&refs).Error)

		publicOnly := search.LabelsPublicOnly(s)

		for _, r := range refs {
			if !publicOnly || !r.NSFW {
				items[index[r.PhotoID]].Labels = append(items[index[r.PhotoID]].Labels, Ref{UID: r.UID, Name: r.Name})
			}
		}
	}

	// People are found through the face markers of the files.
	{
		var refs []photoRef

		entity.Log("mcp", "find people", entity.Db().Table("markers m").
			Select("DISTINCT f.photo_id, m.subj_uid AS uid").
			Joins("JOIN files f ON f.file_uid = m.file_uid AND f.deleted_at IS NULL").
			Where("f.photo_id IN (?) AND m.marker_invalid = FALSE AND m.subj_uid <> ''", ids).
			Order("f.photo_id, m.subj_uid").Scan(&refs).Error)

		if uids := refUIDs(refs); len(uids) > 0 {
			names := make(map[string]Ref)

			if people, err := search.UserSubjects(form.SearchSubjects{UID: strings.Join(uids, "|"), All: true, Count: len(uids)}, s); err == nil {
				for _, p := range people {
					names[p.SubjUID] = Ref{UID: p.SubjUID, Name: p.SubjName}
				}
			}

			for _, r := range refs {
				if ref, ok := names[r.UID]; ok {
					items[index[r.PhotoID]].People = append(items[index[r.PhotoID]].People, ref)
				}
			}
		}
	}

	addMediaCollections(items, index, ids, s)
}

// addMediaCollections adds the collections of the pictures to the media, each type checked by the rules of its own
// resource, if the session may view albums.
func addMediaCollections(items []Media, index map[uint]int, ids []uint, s *entity.Session) {
	if s.SeesAnyDetail(acl.ResourceAlbums) {
		var refs []struct {
			PhotoID uint
			UID     string
			Type    string
		}

		entity.Log("mcp", "find collections", entity.Db().Table("photos_albums pa").
			Select("p.id AS photo_id, a.album_uid AS uid, a.album_type AS type").
			Joins("JOIN photos p ON p.photo_uid = pa.photo_uid").
			Joins("JOIN albums a ON a.album_uid = pa.album_uid AND a.deleted_at IS NULL").
			Where("p.id IN (?) AND pa.hidden = FALSE", ids).
			Order("p.id, a.album_title, a.album_uid").Scan(&refs).Error)

		byType := make(map[string][]string)

		for _, r := range refs {
			byType[r.Type] = appendUnique(byType[r.Type], r.UID)
		}

		titles := make(map[string]Ref)

		for albumType, uids := range byType {
			albums, err := search.UserAlbums(form.SearchAlbums{UID: strings.Join(uids, "|"), Type: albumType, Count: len(uids)}, s)

			if err != nil {
				continue
			}

			for _, a := range albums {
				titles[a.AlbumUID] = Ref{UID: a.AlbumUID, Type: a.AlbumType, Title: a.AlbumTitle}
			}
		}

		for _, r := range refs {
			if ref, ok := titles[r.UID]; ok {
				items[index[r.PhotoID]].Collections = append(items[index[r.PhotoID]].Collections, ref)
			}
		}
	}
}

// deviceNames returns the display names of the cameras or lenses with the specified IDs, except the unknown one.
func deviceNames(table, nameCol, slugCol string, ids []uint) map[uint]string {
	result := make(map[uint]string, len(ids))

	var rows []struct {
		ID   uint
		Name string
	}

	entity.Log("mcp", "find "+table, entity.Db().Table(table).
		Select("id, "+nameCol+" AS name").
		Where("id IN (?) AND "+slugCol+" <> ? AND deleted_at IS NULL", ids, entity.UnknownID).Scan(&rows).Error)

	for _, r := range rows {
		result[r.ID] = r.Name
	}

	return result
}

// refUIDs returns the distinct UIDs of the rows.
func refUIDs(refs []photoRef) (uids []string) {
	for _, r := range refs {
		uids = appendUnique(uids, r.UID)
	}

	return uids
}

// appendUnique appends a value to the list unless it is already in it.
func appendUnique(list []string, value string) []string {
	for _, v := range list {
		if v == value {
			return list
		}
	}

	return append(list, value)
}
