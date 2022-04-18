package search

import (
	"encoding/json"

	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/thumb"
	"github.com/photoprism/photoprism/internal/viewer"
)

// PhotosViewerResults finds photos based on the search form provided and returns them as viewer.Results.
func PhotosViewerResults(f form.SearchPhotos, contentUri, apiUri, previewToken, downloadToken string) (viewer.Results, int, error) {
	if results, count, err := searchPhotos(f, PhotosColsView); err != nil {
		return viewer.Results{}, count, err
	} else {
		return results.ViewerResults(contentUri, apiUri, previewToken, downloadToken), count, err
	}
}

// ViewerResult returns a new photo viewer result.
func (photo Photo) ViewerResult(contentUri, apiUri, previewToken, downloadToken string) viewer.Result {
	return viewer.Result{
		UID:          photo.PhotoUID,
		Title:        photo.PhotoTitle,
		TakenAtLocal: photo.TakenAtLocal,
		Description:  photo.PhotoDescription,
		Favorite:     photo.PhotoFavorite,
		Playable:     photo.IsPlayable(),
		DownloadUrl:  viewer.DownloadUrl(photo.FileHash, apiUri, downloadToken),
		Width:        photo.FileWidth,
		Height:       photo.FileHeight,
		Thumbs: viewer.Thumbs{
			Fit720:  viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit720], contentUri, previewToken),
			Fit1280: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit1280], contentUri, previewToken),
			Fit1920: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit1920], contentUri, previewToken),
			Fit2048: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit2048], contentUri, previewToken),
			Fit2560: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit2560], contentUri, previewToken),
			Fit3840: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit3840], contentUri, previewToken),
			Fit4096: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit4096], contentUri, previewToken),
			Fit7680: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit7680], contentUri, previewToken),
		},
	}
}

// ViewerJSON returns the results as photo viewer JSON.
func (photos PhotoResults) ViewerJSON(contentUri, apiUri, previewToken, downloadToken string) ([]byte, error) {
	return json.Marshal(photos.ViewerResults(contentUri, apiUri, previewToken, downloadToken))
}

// ViewerResults returns the results photo viewer formatted.
func (photos PhotoResults) ViewerResults(contentUri, apiUri, previewToken, downloadToken string) (results viewer.Results) {
	results = make(viewer.Results, 0, len(photos))

	for _, p := range photos {
		results = append(results, p.ViewerResult(contentUri, apiUri, previewToken, downloadToken))
	}

	return results
}

// ViewerResult creates a new photo viewer result.
func (photo GeoResult) ViewerResult(contentUri, apiUri, previewToken, downloadToken string) viewer.Result {
	return viewer.Result{
		UID:          photo.PhotoUID,
		Title:        photo.PhotoTitle,
		TakenAtLocal: photo.TakenAtLocal,
		Description:  photo.PhotoDescription,
		Favorite:     photo.PhotoFavorite,
		Playable:     photo.IsPlayable(),
		DownloadUrl:  viewer.DownloadUrl(photo.FileHash, apiUri, downloadToken),
		Width:        photo.FileWidth,
		Height:       photo.FileHeight,
		Thumbs: viewer.Thumbs{
			Fit720:  viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit720], contentUri, previewToken),
			Fit1280: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit1280], contentUri, previewToken),
			Fit1920: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit1920], contentUri, previewToken),
			Fit2048: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit2048], contentUri, previewToken),
			Fit2560: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit2560], contentUri, previewToken),
			Fit3840: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit3840], contentUri, previewToken),
			Fit4096: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit4096], contentUri, previewToken),
			Fit7680: viewer.NewThumb(photo.FileWidth, photo.FileHeight, photo.FileHash, thumb.Sizes[thumb.Fit7680], contentUri, previewToken),
		},
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
