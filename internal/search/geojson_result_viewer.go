package search

import (
	"encoding/json"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/viewer"
)

// NewViewerResult creates a new photo viewer result.
func NewViewerResult(p GeoResult, contentUri, apiUri, previewToken, downloadToken string) viewer.Result {
	return viewer.Result{
		UID:         p.PhotoUID,
		Title:       p.PhotoTitle,
		Taken:       p.TakenAt,
		Description: p.PhotoDescription,
		Favorite:    p.PhotoFavorite,
		Playable:    p.PhotoType == entity.TypeVideo || p.PhotoType == entity.TypeLive,
		DownloadUrl: viewer.DownloadUrl(p.FileHash, apiUri, downloadToken),
		OriginalW:   p.FileWidth,
		OriginalH:   p.FileHeight,
		Fit720:      viewer.NewThumb(p.FileWidth, p.FileHeight, p.FileHash, thumb.Sizes[thumb.Fit720], contentUri, previewToken),
		Fit1280:     viewer.NewThumb(p.FileWidth, p.FileHeight, p.FileHash, thumb.Sizes[thumb.Fit1280], contentUri, previewToken),
		Fit1920:     viewer.NewThumb(p.FileWidth, p.FileHeight, p.FileHash, thumb.Sizes[thumb.Fit1920], contentUri, previewToken),
		Fit2048:     viewer.NewThumb(p.FileWidth, p.FileHeight, p.FileHash, thumb.Sizes[thumb.Fit2048], contentUri, previewToken),
		Fit2560:     viewer.NewThumb(p.FileWidth, p.FileHeight, p.FileHash, thumb.Sizes[thumb.Fit2560], contentUri, previewToken),
		Fit3840:     viewer.NewThumb(p.FileWidth, p.FileHeight, p.FileHash, thumb.Sizes[thumb.Fit3840], contentUri, previewToken),
		Fit4096:     viewer.NewThumb(p.FileWidth, p.FileHeight, p.FileHash, thumb.Sizes[thumb.Fit4096], contentUri, previewToken),
		Fit7680:     viewer.NewThumb(p.FileWidth, p.FileHeight, p.FileHash, thumb.Sizes[thumb.Fit7680], contentUri, previewToken),
	}
}

// ViewerJSON returns the results as photo viewer JSON.
func (photos GeoResults) ViewerJSON(contentUri, apiUri, previewToken, downloadToken string) ([]byte, error) {
	results := make(viewer.Results, 0, len(photos))

	for _, p := range photos {
		results = append(results, NewViewerResult(p, contentUri, apiUri, previewToken, downloadToken))
	}

	return json.Marshal(results)
}
