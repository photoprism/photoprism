package search

import (
	"encoding/json"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search/viewer"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/thumb"
)

// PhotosViewerResults searches public photos using the provided form and returns
// them in the lightweight viewer format that powers the slideshow endpoints.
func PhotosViewerResults(frm form.SearchPhotos, contentUri, apiUri, previewToken, downloadToken string) (viewer.Results, int, error) {
	return UserPhotosViewerResults(frm, nil, contentUri, apiUri, previewToken, downloadToken)
}

// UserPhotosViewerResults behaves like PhotosViewerResults but also applies the
// permissions encoded in the session (for example shared albums and private
// visibility) before returning viewer-formatted results.
func UserPhotosViewerResults(frm form.SearchPhotos, sess *entity.Session, contentUri, apiUri, previewToken, downloadToken string) (viewer.Results, int, error) {
	if results, count, err := searchPhotos(frm, sess, PhotosColsView); err != nil {
		return viewer.Results{}, count, err
	} else {
		return results.ViewerResults(contentUri, apiUri, previewToken, downloadToken), count, err
	}
}

// ViewerResult converts a photo search result into the DTO consumed by the
// frontend viewer, including derived metadata such as thumbnails and download
// URLs.
func (m *Photo) ViewerResult(contentUri, apiUri, previewToken, downloadToken string) viewer.Result {
	mediaHash, mediaCodec, mediaMime, width, height := m.MediaInfo()
	return viewer.Result{
		UID:          m.PhotoUID,
		Type:         m.PhotoType,
		Title:        m.PhotoTitle,
		Caption:      m.PhotoCaption,
		Lat:          m.PhotoLat,
		Lng:          m.PhotoLng,
		TakenAtLocal: m.TakenAtLocal,
		TimeZone:     m.TimeZone,
		Favorite:     m.PhotoFavorite,
		Playable:     m.IsPlayable(),
		Duration:     m.PhotoDuration,
		Width:        width,
		Height:       height,
		Hash:         mediaHash,
		Codec:        mediaCodec,
		Mime:         mediaMime,
		Thumbs:       thumb.ViewerThumbs(m.FileWidth, m.FileHeight, m.FileHash, contentUri, previewToken),
		DownloadUrl:  viewer.DownloadUrl(m.FileHash, apiUri, downloadToken),
	}
}

// ViewerJSON marshals the current result set to the viewer JSON structure.
func (m PhotoResults) ViewerJSON(contentUri, apiUri, previewToken, downloadToken string) ([]byte, error) {
	return json.Marshal(m.ViewerResults(contentUri, apiUri, previewToken, downloadToken))
}

// ViewerResults maps every photo into the viewer DTO while preserving order.
func (m PhotoResults) ViewerResults(contentUri, apiUri, previewToken, downloadToken string) (results viewer.Results) {
	results = make(viewer.Results, 0, len(m))

	for _, p := range m {
		results = append(results, p.ViewerResult(contentUri, apiUri, previewToken, downloadToken))
	}

	return results
}

// ViewerResult converts a geographic search hit into the viewer DTO, reusing
// the thumbnail and download helpers so photos and map results stay aligned.
func (m GeoResult) ViewerResult(contentUri, apiUri, previewToken, downloadToken string) viewer.Result {
	return viewer.Result{
		UID:          m.PhotoUID,
		Type:         m.PhotoType,
		Title:        m.PhotoTitle,
		Caption:      m.PhotoCaption,
		Lat:          m.PhotoLat,
		Lng:          m.PhotoLng,
		TakenAtLocal: m.TakenAtLocal,
		TimeZone:     m.TimeZone,
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

// ViewerJSON marshals geo search hits to the viewer JSON structure.
func (photos GeoResults) ViewerJSON(contentUri, apiUri, previewToken, downloadToken string) ([]byte, error) {
	results := make(viewer.Results, 0, len(photos))

	for _, p := range photos {
		results = append(results, p.ViewerResult(contentUri, apiUri, previewToken, downloadToken))
	}

	return json.Marshal(results)
}
