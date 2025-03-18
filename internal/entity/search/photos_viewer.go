package search

import (
	"encoding/json"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search/viewer"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/thumb"
)

// PhotosViewerResults finds photos based on the search form provided and returns them as viewer.Results.
func PhotosViewerResults(frm form.SearchPhotos, contentUri, apiUri, previewToken, downloadToken string) (viewer.Results, int, error) {
	return UserPhotosViewerResults(frm, nil, contentUri, apiUri, previewToken, downloadToken)
}

// UserPhotosViewerResults finds photos based on the search form and user session and returns them as viewer.Results.
func UserPhotosViewerResults(frm form.SearchPhotos, sess *entity.Session, contentUri, apiUri, previewToken, downloadToken string) (viewer.Results, int, error) {
	if results, count, err := searchPhotos(frm, sess, PhotosColsView); err != nil {
		return viewer.Results{}, count, err
	} else {
		return results.ViewerResults(contentUri, apiUri, previewToken, downloadToken), count, err
	}
}

// ViewerResult returns a new photo viewer result.
func (m *Photo) ViewerResult(contentUri, apiUri, previewToken, downloadToken string) viewer.Result {
	mediaHash, mediaCodec, mediaMime := m.MediaInfo()
	return viewer.Result{
		UID:          m.PhotoUID,
		Type:         m.PhotoType,
		Title:        m.PhotoTitle,
		Caption:      m.PhotoCaption,
		Lat:          m.PhotoLat,
		Lng:          m.PhotoLng,
		TakenAtLocal: m.TakenAtLocal,
		Favorite:     m.PhotoFavorite,
		Playable:     m.IsPlayable(),
		Duration:     m.PhotoDuration,
		Width:        m.FileWidth,
		Height:       m.FileHeight,
		Hash:         mediaHash,
		Codec:        mediaCodec,
		Mime:         mediaMime,
		Thumbs:       thumb.ViewerThumbs(m.FileWidth, m.FileHeight, m.FileHash, contentUri, previewToken),
		DownloadUrl:  viewer.DownloadUrl(m.FileHash, apiUri, downloadToken),
	}
}

// ViewerJSON returns the results as photo viewer JSON.
func (m PhotoResults) ViewerJSON(contentUri, apiUri, previewToken, downloadToken string) ([]byte, error) {
	return json.Marshal(m.ViewerResults(contentUri, apiUri, previewToken, downloadToken))
}

// ViewerResults returns the results photo viewer formatted.
func (m PhotoResults) ViewerResults(contentUri, apiUri, previewToken, downloadToken string) (results viewer.Results) {
	results = make(viewer.Results, 0, len(m))

	for _, p := range m {
		results = append(results, p.ViewerResult(contentUri, apiUri, previewToken, downloadToken))
	}

	return results
}

// ViewerResult creates a new photo viewer result.
func (m GeoResult) ViewerResult(contentUri, apiUri, previewToken, downloadToken string) viewer.Result {
	return viewer.Result{
		UID:          m.PhotoUID,
		Type:         m.PhotoType,
		Title:        m.PhotoTitle,
		Caption:      m.PhotoCaption,
		Lat:          m.PhotoLat,
		Lng:          m.PhotoLng,
		TakenAtLocal: m.TakenAtLocal,
		Favorite:     m.PhotoFavorite,
		Playable:     m.IsPlayable(),
		Duration:     m.PhotoDuration,
		Width:        m.FileWidth,
		Height:       m.FileHeight,
		Hash:         m.FileHash,
		Codec:        m.FileCodec,
		Mime:         m.FileMime,
		Thumbs:       thumb.ViewerThumbs(m.FileWidth, m.FileHeight, m.FileHash, contentUri, previewToken),
		DownloadUrl:  viewer.DownloadUrl(m.FileHash, apiUri, downloadToken),
	}
}

// ViewerJSON returns the results as photo viewer JSON.
func (photos GeoResults) ViewerJSON(contentUri, apiUri, previewToken, downloadToken string) ([]byte, error) {
	results := make(viewer.Results, 0, len(photos))

	for _, p := range photos {
		results = append(results, p.ViewerResult(contentUri, apiUri, previewToken, downloadToken))
	}

	return json.Marshal(results)
}
