package search

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search/viewer"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestPhoto_ViewerResult(t *testing.T) {
	uid := rnd.GenerateUID(entity.PhotoUID)
	imgHash := "img-hash"
	videoHash := "video-hash"
	taken := time.Date(2024, 5, 1, 15, 4, 5, 0, time.UTC)

	photo := Photo{
		PhotoUID:      uid,
		PhotoType:     entity.MediaVideo,
		PhotoTitle:    "Sunset",
		PhotoCaption:  "Golden hour",
		PhotoLat:      12.34,
		PhotoLng:      56.78,
		TakenAtLocal:  taken,
		TimeZone:      "UTC",
		PhotoFavorite: true,
		PhotoDuration: 5 * time.Second,
		FileHash:      imgHash,
		FileWidth:     800,
		FileHeight:    600,
		Files: []entity.File{
			{
				FileVideo:  true,
				MediaType:  entity.MediaVideo,
				FileHash:   videoHash,
				FileCodec:  "avc1",
				FileMime:   header.ContentTypeMp4AvcMain,
				FileWidth:  1920,
				FileHeight: 1080,
			},
		},
	}

	result := photo.ViewerResult("/content", "/api/v1", "preview-token", "download-token")

	assert.Equal(t, uid, result.UID)
	assert.Equal(t, entity.MediaVideo, result.Type)
	assert.Equal(t, "Sunset", result.Title)
	assert.Equal(t, "Golden hour", result.Caption)
	assert.Equal(t, 12.34, result.Lat)
	assert.Equal(t, 56.78, result.Lng)
	assert.Equal(t, taken, result.TakenAtLocal)
	assert.Equal(t, "UTC", result.TimeZone)
	assert.True(t, result.Favorite)
	assert.True(t, result.Playable)
	assert.Equal(t, 5*time.Second, result.Duration)
	assert.Equal(t, videoHash, result.Hash)
	assert.Equal(t, "avc1", result.Codec)
	assert.Equal(t, header.ContentTypeMp4AvcMain, result.Mime)
	assert.Equal(t, 1920, result.Width)
	assert.Equal(t, 1080, result.Height)
	if assert.NotNil(t, result.Thumbs) {
		assert.NotNil(t, result.Thumbs.Fit720)
	}
	assert.Equal(t, "/api/v1/dl/img-hash?t=download-token", result.DownloadUrl)
}

func TestPhotoResults_ViewerFormatting(t *testing.T) {
	uid1 := rnd.GenerateUID(entity.PhotoUID)
	uid2 := rnd.GenerateUID(entity.PhotoUID)

	photos := PhotoResults{
		{PhotoUID: uid1},
		{PhotoUID: uid2},
	}

	results := photos.ViewerResults("/content", "/api", "preview", "download")
	assert.Len(t, results, 2)
	assert.Equal(t, uid1, results[0].UID)
	assert.Equal(t, uid2, results[1].UID)

	data, err := photos.ViewerJSON("/content", "/api", "preview", "download")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed viewer.Results
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal viewer json: %v", err)
	}

	assert.Len(t, parsed, 2)
	assert.Equal(t, uid1, parsed[0].UID)
	assert.Equal(t, uid2, parsed[1].UID)
}

func TestGeoResult_ViewerResult(t *testing.T) {
	uid := rnd.GenerateUID(entity.PhotoUID)
	taken := time.Date(2023, 3, 14, 9, 26, 53, 0, time.UTC)

	geo := GeoResult{
		PhotoUID:      uid,
		PhotoType:     entity.MediaImage,
		PhotoTitle:    "Mountains",
		PhotoCaption:  "Snow peaks",
		PhotoLat:      -12.34,
		PhotoLng:      78.9,
		TakenAtLocal:  taken,
		TimeZone:      "Europe/Berlin",
		PhotoFavorite: false,
		PhotoDuration: 0,
		FileHash:      "img-hash",
		FileCodec:     "jpeg",
		FileMime:      header.ContentTypeJpeg,
		FileWidth:     1024,
		FileHeight:    768,
	}

	result := geo.ViewerResult("/content", "/api", "preview", "download")

	assert.Equal(t, uid, result.UID)
	assert.Equal(t, entity.MediaImage, result.Type)
	assert.Equal(t, "Mountains", result.Title)
	assert.Equal(t, "Snow peaks", result.Caption)
	assert.Equal(t, -12.34, result.Lat)
	assert.Equal(t, 78.9, result.Lng)
	assert.Equal(t, taken, result.TakenAtLocal)
	assert.Equal(t, "Europe/Berlin", result.TimeZone)
	assert.False(t, result.Favorite)
	assert.False(t, result.Playable)
	assert.Equal(t, "img-hash", result.Hash)
	assert.Equal(t, "jpeg", result.Codec)
	assert.Equal(t, header.ContentTypeJpeg, result.Mime)
	assert.Equal(t, 1024, result.Width)
	assert.Equal(t, 768, result.Height)
	if assert.NotNil(t, result.Thumbs) {
		assert.NotNil(t, result.Thumbs.Fit720)
	}
	assert.Equal(t, "/api/dl/img-hash?t=download", result.DownloadUrl)
}

func TestGeoResults_ViewerJSON(t *testing.T) {
	uid1 := rnd.GenerateUID(entity.PhotoUID)
	uid2 := rnd.GenerateUID(entity.PhotoUID)

	items := GeoResults{
		{PhotoUID: uid1, FileHash: "hash1"},
		{PhotoUID: uid2, FileHash: "hash2"},
	}

	data, err := items.ViewerJSON("/content", "/api", "preview", "download")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed viewer.Results
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal viewer json: %v", err)
	}

	assert.Len(t, parsed, 2)
	assert.Equal(t, uid1, parsed[0].UID)
	assert.Equal(t, uid2, parsed[1].UID)
}

func TestPhotosViewerResults(t *testing.T) {
	fixture := entity.PhotoFixtures.Get("19800101_000002_D640C559")
	form := form.SearchPhotos{
		UID:     fixture.PhotoUID,
		Count:   1,
		Primary: true,
	}

	results, count, err := PhotosViewerResults(form, "/content", "/api", "preview", "download")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Greater(t, count, 0)
	if assert.NotEmpty(t, results) {
		assert.Equal(t, fixture.PhotoUID, results[0].UID)
		assert.NotNil(t, results[0].Thumbs)
	}
}
