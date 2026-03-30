package search

import (
	"encoding/json"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search/viewer"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/thumb"
)

// documentFileHash returns the file hash of the original document file (e.g.
// PDF) for the given photo UID, or an empty string if none is found. It is used
// so that the viewer download URL points to the real document rather than the
// JPEG sidecar that is stored as the primary file for thumbnail generation.
func documentFileHash(photoUID string) string {
	var f entity.File
	if err := entity.UnscopedDb().
		Select("file_hash").
		Where("photo_uid = ? AND media_type = ? AND file_missing = 0 AND file_error = ''", photoUID, entity.MediaDocument).
		First(&f).Error; err != nil {
		return ""
	}
	return f.FileHash
}

// viewerDownloadHash returns the file hash that should be used for viewer
// downloads. For document media, this resolves to the original document file
// hash instead of the JPEG sidecar hash used by the viewer query.
func viewerDownloadHash(photoType, photoUID, fileHash string) string {
	if photoType != entity.MediaDocument {
		return fileHash
	}

	if h := documentFileHash(photoUID); h != "" {
		return h
	}

	return fileHash
}

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

	downloadHash := viewerDownloadHash(m.PhotoType, m.PhotoUID, m.FileHash)

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
		DownloadUrl:  viewer.DownloadUrl(downloadHash, apiUri, downloadToken),
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
	downloadHash := viewerDownloadHash(m.PhotoType, m.PhotoUID, m.FileHash)
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
		DownloadUrl:  viewer.DownloadUrl(downloadHash, apiUri, downloadToken),
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
