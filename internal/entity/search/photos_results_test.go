package search

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/media/projection"
	"github.com/photoprism/photoprism/pkg/media/video"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestPhoto_Ids(t *testing.T) {
	r := Photo{
		ID:           1111198,
		CreatedAt:    time.Time{},
		UpdatedAt:    time.Time{},
		DeletedAt:    &time.Time{},
		TakenAt:      time.Time{},
		TakenAtLocal: time.Time{},
		PhotoUID:     "ps6sg6be2lvl0o98",
	}

	assert.Equal(t, uint(1111198), r.GetID())
	assert.True(t, r.HasID())
	assert.Equal(t, "ps6sg6be2lvl0o98", r.GetUID())
}

func TestPhoto_String(t *testing.T) {
	testcases := []struct {
		name  string
		photo *Photo
		want  string
	}{
		{
			name:  "Nil",
			photo: nil,
			want:  "Photo<nil>",
		},
		{
			name: "PhotoName",
			photo: &Photo{
				PhotoPath: "albums/test",
				PhotoName: "my photo.jpg",
			},
			want: "'albums/test/my photo.jpg'",
		},
		{
			name: "OriginalName",
			photo: &Photo{
				OriginalName: "orig name.dng",
			},
			want: "'orig name.dng'",
		},
		{
			name: "UID",
			photo: &Photo{
				PhotoUID: "ps123",
			},
			want: "uid ps123",
		},
		{
			name: "ID",
			photo: &Photo{
				ID: 42,
			},
			want: "id 42",
		},
		{
			name:  "Fallback",
			photo: &Photo{},
			want:  "*Photo",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.photo == nil {
				var p *Photo
				assert.Equal(t, tc.want, p.String())
			} else {
				assert.Equal(t, tc.want, tc.photo.String())
			}
		})
	}
}

// TestPhoto_IdentifyingScalars checks that the XMP DocumentID and camera serial are neither selected
// nor serialized, so re-adding either field fails here instead of silently reopening the gap between
// this path and Photo.RedactForSession.
func TestPhoto_IdentifyingScalars(t *testing.T) {
	t.Run("NotSelected", func(t *testing.T) {
		for _, cols := range []struct{ name, sel string }{
			{"PhotosColsAll", PhotosColsAll},
			{"PhotosColsView", PhotosColsView},
			{"BatchCols", BatchCols},
		} {
			assert.NotContains(t, cols.sel, "photos.uuid", cols.name)
			assert.NotContains(t, cols.sel, "photos.camera_serial", cols.name)
		}
	})
	t.Run("NotSerialized", func(t *testing.T) {
		b, err := json.Marshal(Photo{PhotoUID: "ps6sg6be2lvl0o98"})
		assert.NoError(t, err)
		assert.NotContains(t, string(b), "DocumentID")
		assert.NotContains(t, string(b), "CameraSerial")
	})
}

func TestPhoto_Approve(t *testing.T) {
	t.Run("EmptyPhoto", func(t *testing.T) {
		r := Photo{}
		err := r.Approve()

		assert.Error(t, err)
	})
	t.Run("PhotoNotInReview", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
			PhotoQuality: 4,
		}

		err := r.Approve()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 4, r.PhotoQuality)
	})
	t.Run("Approve", func(t *testing.T) {
		r := Photo{
			ID:           100028476,
			CreatedAt:    time.Time{},
			UpdatedAt:    time.Time{},
			DeletedAt:    &time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0j76",
			PhotoQuality: 2,
		}

		err := r.Approve()

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 3, r.PhotoQuality)
		assert.Nil(t, r.DeletedAt)
		assert.NotNil(t, r.EditedAt)
	})
}

func TestPhoto_Restore(t *testing.T) {
	t.Run("EmptyPhoto", func(t *testing.T) {
		r := Photo{}

		err := r.Restore()

		assert.Error(t, err)
	})
	t.Run("PhotoNotInArchive", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
		}

		err := r.Restore()

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, r.DeletedAt)
	})
	t.Run("Restore", func(t *testing.T) {
		r := Photo{
			ID:           100028476,
			CreatedAt:    time.Time{},
			UpdatedAt:    time.Time{},
			DeletedAt:    &time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0j76",
			PhotoQuality: 2,
		}

		assert.NotNil(t, r.DeletedAt)

		err := r.Restore()

		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, r.DeletedAt)
	})
}

func TestPhoto_IsPlayable(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
			PhotoType:    "live",
		}

		assert.True(t, r.IsPlayable())
	})
	t.Run("False", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
			PhotoType:    "image",
		}

		assert.False(t, r.IsPlayable())
	})
}

func TestPhoto_MediaProjection(t *testing.T) {
	t.Run("VideoUsesVideoFileProjection", func(t *testing.T) {
		r := Photo{
			PhotoType:      "video",
			FileProjection: "",
			Files: []entity.File{
				{FileVideo: true, MediaType: media.Video.String(), FileHash: "v", FileProjection: "equirectangular"},
			},
		}
		assert.Equal(t, "equirectangular", r.MediaProjection())
	})
	t.Run("VideoFallsBackToPrimaryProjection", func(t *testing.T) {
		r := Photo{
			PhotoType:      "video",
			FileProjection: "cubemap",
			Files: []entity.File{
				{FileVideo: true, MediaType: media.Video.String(), FileHash: "v", FileProjection: ""},
			},
		}
		assert.Equal(t, "cubemap", r.MediaProjection())
	})
	t.Run("ImageUsesPrimaryProjection", func(t *testing.T) {
		r := Photo{
			PhotoType:      "image",
			FileProjection: "equirectangular",
		}
		assert.Equal(t, "equirectangular", r.MediaProjection())
	})
	t.Run("ImagePrefersEquirectDerivative", func(t *testing.T) {
		r := Photo{
			PhotoType:      "image",
			FileProjection: "dual-fisheye",
			Files: []entity.File{
				{MediaType: media.Image.String(), FileHash: "e", FileProjection: "equirectangular"},
			},
		}
		assert.Equal(t, "equirectangular", r.MediaProjection())
	})
	t.Run("ImageDualFisheyeWithoutDerivativeRedacted", func(t *testing.T) {
		r := Photo{
			PhotoType:      "image",
			FileProjection: "dual-fisheye",
		}
		assert.Equal(t, "", r.MediaProjection())
	})
	t.Run("VideoPrefersEquirectDerivative", func(t *testing.T) {
		r := Photo{
			PhotoType:      "video",
			FileProjection: "",
			Files: []entity.File{
				{FileVideo: true, MediaType: media.Video.String(), FileHash: "v", FileProjection: "dual-fisheye"},
				{MediaType: media.Image.String(), FileHash: "e", FileProjection: "equirectangular"},
			},
		}
		assert.Equal(t, "equirectangular", r.MediaProjection())
	})
	t.Run("VideoDualFisheyeWithoutDerivativeRedacted", func(t *testing.T) {
		r := Photo{
			PhotoType:      "video",
			FileProjection: "dual-fisheye",
			Files: []entity.File{
				{FileVideo: true, MediaType: media.Video.String(), FileHash: "v", FileProjection: "dual-fisheye"},
			},
		}
		assert.Equal(t, "", r.MediaProjection())
	})
}

func TestSphereProjection(t *testing.T) {
	assert.Equal(t, "", sphereProjection("fisheye"))
	assert.Equal(t, "", sphereProjection("dual-fisheye"))
	assert.Equal(t, "equirectangular", sphereProjection("equirectangular"))
	assert.Equal(t, "cubestrip", sphereProjection("cubestrip"))
	assert.Equal(t, "", sphereProjection(""))
}

func TestPhoto_MediaInfo(t *testing.T) {
	t.Run("EquirectangularDerivativePreferred", func(t *testing.T) {
		r := Photo{
			PhotoType: media.Video.String(),
			// The indexer tags every .insv original dual-fisheye, so the lens file carries that value.
			Files: []entity.File{
				{FileVideo: true, FileHash: "square-original", FileCodec: video.CodecHvc1, FileWidth: 3072, FileHeight: 3072, FileProjection: projection.DualFisheye.String()},
				{FileVideo: true, FileHash: "sphere-avc", FileCodec: video.CodecAvc1, FileMime: header.ContentTypeMp4AvcMain, FileWidth: 1920, FileHeight: 960, FileProjection: projection.Equirectangular.String()},
			},
		}

		mediaHash, mediaCodec, _, width, height := r.MediaInfo()
		assert.Equal(t, "sphere-avc", mediaHash)
		assert.Equal(t, video.CodecAvc1, mediaCodec)
		assert.Equal(t, 1920, width)
		assert.Equal(t, 960, height)
	})
	t.Run("StackedEquirectangularNotPreferred", func(t *testing.T) {
		// Without a fisheye original there is nothing to substitute, so an unrelated 360° file
		// stacked on a normal video must not replace its dimensions and codec.
		r := Photo{
			PhotoType: media.Video.String(),
			Files: []entity.File{
				{FileVideo: true, FileHash: "flat-video", FileCodec: video.CodecAvc1, FileWidth: 1920, FileHeight: 1080},
				{FileVideo: true, FileHash: "sphere-avc", FileCodec: video.CodecAvc1, FileWidth: 1920, FileHeight: 960, FileProjection: projection.Equirectangular.String()},
			},
		}

		mediaHash, _, _, width, height := r.MediaInfo()
		assert.Equal(t, "flat-video", mediaHash)
		assert.Equal(t, 1920, width)
		assert.Equal(t, 1080, height)
	})
	t.Run("LiveCodecAVC", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
			PhotoType:    "live",
			FileWidth:    800,
			FileHeight:   600,
			FileHash:     "e22a06fb5b63dae7f3d08ab95fb958935b744e51",
			Files: []entity.File{
				{
					FileVideo:  true,
					MediaType:  media.Video.String(),
					FileMime:   header.ContentTypeMp4AvcMain,
					FileCodec:  video.CodecAvc1,
					FileWidth:  1920,
					FileHeight: 1080,
					FileHash:   "53c89dcfa006c9e592dd9e6db4b31cd57be64b81",
				},
			},
		}

		assert.True(t, r.IsPlayable())

		mediaHash, mediaCodec, mediaMime, width, height := r.MediaInfo()
		assert.Equal(t, "53c89dcfa006c9e592dd9e6db4b31cd57be64b81", mediaHash)
		assert.Equal(t, video.CodecAvc1, mediaCodec)
		assert.Equal(t, header.ContentTypeMp4AvcMain, mediaMime)
		assert.Equal(t, 1920, width)
		assert.Equal(t, 1080, height)
	})
	t.Run("Raw", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0abc",
			PhotoType:    "raw",
			FileWidth:    800,
			FileHeight:   600,
			FileMime:     "image/jpeg",
			FileHash:     "e22a06fb5b63dae7f3d08ab95fb958935b744e51",
			Files: []entity.File{
				{
					FileVideo:  false,
					MediaType:  media.Raw.String(),
					FileMime:   "image/x-raw",
					FileCodec:  "raw",
					FileWidth:  1920,
					FileHeight: 1080,
					FileHash:   "53c89dcfa006c9e592dd9e6db4b31cd57be64b81",
				},
			},
		}

		assert.False(t, r.IsPlayable())

		mediaHash, mediaCodec, mediaMime, width, height := r.MediaInfo()
		assert.Equal(t, "e22a06fb5b63dae7f3d08ab95fb958935b744e51", mediaHash)
		assert.Equal(t, "raw", mediaCodec)
		assert.Equal(t, "image/x-raw", mediaMime)
		assert.Equal(t, 1920, width)
		assert.Equal(t, 1080, height)
	})
	t.Run("RawFisheyeEquirectangularDerivativePreferred", func(t *testing.T) {
		// The dewarp derivative holds the pixels the sphere viewer shows, so its 2:1 frame is
		// reported instead of the fisheye original's portrait frame.
		r := Photo{
			PhotoType:  media.Raw.String(),
			FileHash:   "primary-jpeg",
			FileWidth:  5760,
			FileHeight: 2880,
			Files: []entity.File{
				{MediaType: media.Raw.String(), FileHash: "fisheye-dng", FileMime: "image/x-raw", FileCodec: "raw", FileWidth: 3264, FileHeight: 6528, FileProjection: projection.DualFisheye.String()},
				{MediaType: media.Image.String(), FileHash: "sphere-jpeg", FileMime: "image/jpeg", FileCodec: "jpeg", FileWidth: 5760, FileHeight: 2880, FileProjection: projection.Equirectangular.String()},
			},
		}

		assert.Equal(t, projection.Equirectangular.String(), r.MediaProjection())

		mediaHash, mediaCodec, mediaMime, width, height := r.MediaInfo()
		assert.Equal(t, "primary-jpeg", mediaHash)
		assert.Equal(t, "jpeg", mediaCodec)
		assert.Equal(t, "image/jpeg", mediaMime)
		assert.Equal(t, 5760, width)
		assert.Equal(t, 2880, height)
	})
	t.Run("RawStackedEquirectangularNotPreferred", func(t *testing.T) {
		// Without a fisheye original there is nothing to substitute, so an unrelated 360° file
		// stacked on a normal RAW must not replace its dimensions.
		r := Photo{
			PhotoType: media.Raw.String(),
			FileHash:  "primary-jpeg",
			Files: []entity.File{
				{MediaType: media.Raw.String(), FileHash: "flat-dng", FileMime: "image/x-raw", FileCodec: "raw", FileWidth: 6000, FileHeight: 4000},
				{MediaType: media.Image.String(), FileHash: "sphere-jpeg", FileMime: "image/jpeg", FileCodec: "jpeg", FileWidth: 5760, FileHeight: 2880, FileProjection: projection.Equirectangular.String()},
			},
		}

		mediaHash, mediaCodec, mediaMime, width, height := r.MediaInfo()
		assert.Equal(t, "primary-jpeg", mediaHash)
		assert.Equal(t, "raw", mediaCodec)
		assert.Equal(t, "image/x-raw", mediaMime)
		assert.Equal(t, 6000, width)
		assert.Equal(t, 4000, height)
	})
	t.Run("RawFisheyeWithoutDerivative", func(t *testing.T) {
		// A failed or disabled dewarp leaves no derivative, so the original's frame is reported.
		r := Photo{
			PhotoType: media.Raw.String(),
			FileHash:  "primary-jpeg",
			Files: []entity.File{
				{MediaType: media.Raw.String(), FileHash: "fisheye-dng", FileMime: "image/x-raw", FileCodec: "raw", FileWidth: 3264, FileHeight: 6528, FileProjection: projection.DualFisheye.String()},
			},
		}

		mediaHash, mediaCodec, mediaMime, width, height := r.MediaInfo()
		assert.Equal(t, "primary-jpeg", mediaHash)
		assert.Equal(t, "raw", mediaCodec)
		assert.Equal(t, "image/x-raw", mediaMime)
		assert.Equal(t, 3264, width)
		assert.Equal(t, 6528, height)
	})
	t.Run("Animated", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0abc",
			PhotoType:    "animated",
			FileWidth:    800,
			FileHeight:   600,
			FileMime:     "image/gif",
			FileHash:     "e22a06fb5b63dae7f3d08ab95fb958935b744e51",
			Files: []entity.File{
				{
					FileVideo:    false,
					MediaType:    media.Image.String(),
					FileMime:     "image/gif",
					FileCodec:    "gif",
					FileDuration: 1000,
					FileFrames:   100,
					FileWidth:    1920,
					FileHeight:   1080,
					FileHash:     "53c89dcfa006c9e592dd9e6db4b31cd57be64b81",
				},
			},
		}

		assert.True(t, r.IsPlayable())

		mediaHash, mediaCodec, mediaMime, width, height := r.MediaInfo()
		assert.Equal(t, "53c89dcfa006c9e592dd9e6db4b31cd57be64b81", mediaHash)
		assert.Equal(t, "gif", mediaCodec)
		assert.Equal(t, "image/gif", mediaMime)
		assert.Equal(t, 1920, width)
		assert.Equal(t, 1080, height)
	})
	t.Run("VideoCodecHVC", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
			PhotoType:    "video",
			FileWidth:    800,
			FileHeight:   600,
			FileHash:     "e22a06fb5b63dae7f3d08ab95fb958935b744e51",
			Files: []entity.File{
				{
					FileVideo: false,
					MediaType: media.Image.String(),
					FileMime:  header.ContentTypeJpeg,
					FileCodec: "jpeg",
				},
				{
					FileVideo: true,
					MediaType: media.Video.String(),
					FileMime:  header.ContentTypeMp4AvcMain,
					FileCodec: "xyz",
					FileHash:  "",
				},
				{
					FileVideo:  true,
					MediaType:  media.Video.String(),
					FileCodec:  video.CodecHvc1,
					FileMime:   header.ContentTypeMp4HvcMain10,
					FileWidth:  1920,
					FileHeight: 1080,
					FileHash:   "057258b0c88c2e017ec171cc8799a5df7badbadf",
				},
				{
					FileVideo: true,
					MediaType: media.Video.String(),
					FileCodec: video.CodecAvc1,
					FileMime:  header.ContentTypeMp4AvcMain,
					FileHash:  "ddb3f44eb500d7669cbe0a95e66d5a63f642487d",
				},
			},
		}

		assert.True(t, r.IsPlayable())

		mediaHash, mediaCodec, mediaMime, width, height := r.MediaInfo()
		assert.Equal(t, "057258b0c88c2e017ec171cc8799a5df7badbadf", mediaHash)
		assert.Equal(t, video.CodecHvc1, mediaCodec)
		assert.Equal(t, header.ContentTypeMp4HvcMain10, mediaMime)
		assert.Equal(t, 1920, width)
		assert.Equal(t, 1080, height)
	})
	t.Run("NoVideoHash", func(t *testing.T) {
		r := Photo{
			ID:           1111154,
			CreatedAt:    time.Time{},
			TakenAt:      time.Time{},
			TakenAtLocal: time.Time{},
			TakenSrc:     "",
			TimeZone:     "",
			PhotoUID:     "ps6sg6be2lvl0r41",
			PhotoType:    "live",
			FileWidth:    800,
			FileHeight:   600,
			FileHash:     "e22a06fb5b63dae7f3d08ab95fb958935b744e51",
			Files: []entity.File{
				{
					FileVideo:  true,
					MediaType:  media.Video.String(),
					FileMime:   header.ContentTypeMp4AvcMain,
					FileWidth:  1024,
					FileHeight: 512,
					FileHash:   "",
				},
			},
		}

		assert.True(t, r.IsPlayable())

		mediaHash, mediaCodec, mediaMime, width, height := r.MediaInfo()
		assert.Equal(t, "e22a06fb5b63dae7f3d08ab95fb958935b744e51", mediaHash)
		assert.Equal(t, "", mediaCodec)
		assert.Equal(t, "", mediaMime)
		assert.Equal(t, 800, width)
		assert.Equal(t, 600, height)
	})
}

func TestPhotoResults_Photos(t *testing.T) {
	photo1 := Photo{
		ID:           1111154,
		CreatedAt:    time.Time{},
		TakenAt:      time.Time{},
		TakenAtLocal: time.Time{},
		TakenSrc:     "",
		TimeZone:     "",
		PhotoUID:     "ps6sg6be2lvl0r41",
		PhotoType:    "live",
	}

	photo2 := Photo{
		ID:           1111155,
		CreatedAt:    time.Time{},
		TakenAt:      time.Time{},
		TakenAtLocal: time.Time{},
		TakenSrc:     "",
		TimeZone:     "",
		PhotoUID:     "ps6sg6be2lvl0986",
		PhotoType:    "image",
	}

	r := PhotoResults{photo1, photo2}

	assert.Len(t, r.Photos(), 2)
}

func TestPhotosResults_Merged(t *testing.T) {
	fileUIDA := rnd.GenerateUID(entity.FileUID)
	fileUIDB := rnd.GenerateUID(entity.FileUID)
	fileUIDC := rnd.GenerateUID(entity.FileUID)

	results := PhotoResults{
		{ID: 1, FileID: 10, FileUID: fileUIDA, FileName: "a.jpg", FileError: "unsupported codec"},
		{ID: 1, FileID: 11, FileUID: fileUIDB, FileName: "b.jpg", FileError: "metadata read failed"},
		{ID: 2, FileID: 20, FileUID: fileUIDC, FileName: "c.jpg"},
	}

	merged, count, err := results.Merge()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	assert.Len(t, merged, 2)

	first := merged[0]
	assert.Equal(t, "1-10", first.CompositeID)
	assert.True(t, first.Merged)
	assert.Len(t, first.Files, 2)
	assert.Equal(t, uint(10), first.Files[0].ID)
	assert.Equal(t, uint(11), first.Files[1].ID)
	assert.Equal(t, "unsupported codec", first.Files[0].FileError)
	assert.Equal(t, "metadata read failed", first.Files[1].FileError)

	second := merged[1]
	assert.Equal(t, "2-20", second.CompositeID)
	assert.False(t, second.Merged)
	assert.Len(t, second.Files, 1)
	assert.Equal(t, uint(20), second.Files[0].ID)
}
func TestPhotosResults_UIDs(t *testing.T) {
	uid1 := rnd.GenerateUID(entity.PhotoUID)
	uid2 := rnd.GenerateUID(entity.PhotoUID)

	result1 := Photo{
		ID:               111111,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
		DeletedAt:        &time.Time{},
		TakenAt:          time.Time{},
		TakenAtLocal:     time.Time{},
		TakenSrc:         "",
		TimeZone:         "Local",
		PhotoUID:         uid1,
		PhotoPath:        "",
		PhotoName:        "",
		PhotoTitle:       "Photo1",
		PhotoYear:        0,
		PhotoMonth:       0,
		PhotoCountry:     "",
		PhotoFavorite:    false,
		PhotoPrivate:     false,
		PhotoLat:         0,
		PhotoLng:         0,
		PhotoAltitude:    0,
		PhotoIso:         0,
		PhotoFocalLength: 0,
		PhotoFNumber:     0,
		PhotoExposure:    "",
		PhotoQuality:     0,
		PhotoResolution:  0,
		Merged:           false,
		CameraID:         0,
		CameraModel:      "",
		CameraMake:       "",
		CameraType:       "",
		LensID:           0,
		LensModel:        "",
		LensMake:         "",
		CellID:           "",
		PlaceID:          "",
		PlaceLabel:       "",
		PlaceCity:        "",
		PlaceState:       "",
		PlaceCountry:     "",
		FileID:           0,
		FileUID:          "",
		FilePrimary:      false,
		FileMissing:      false,
		FileName:         "",
		FileHash:         "",
		FileType:         "",
		FileMime:         "",
		FileWidth:        0,
		FileHeight:       0,
		FileOrientation:  0,
		FileAspectRatio:  0,
		FileColors:       "",
		FileChroma:       0,
		FileLuminance:    "",
		FileDiff:         0,
		Files:            nil,
	}

	result2 := Photo{
		ID:               22222,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
		DeletedAt:        &time.Time{},
		TakenAt:          time.Time{},
		TakenAtLocal:     time.Time{},
		TakenSrc:         "",
		TimeZone:         "Local",
		PhotoUID:         uid2,
		PhotoPath:        "",
		PhotoName:        "",
		PhotoTitle:       "Photo2",
		PhotoYear:        0,
		PhotoMonth:       0,
		PhotoCountry:     "",
		PhotoFavorite:    false,
		PhotoPrivate:     false,
		PhotoLat:         0,
		PhotoLng:         0,
		PhotoAltitude:    0,
		PhotoIso:         0,
		PhotoFocalLength: 0,
		PhotoFNumber:     0,
		PhotoExposure:    "",
		PhotoQuality:     0,
		PhotoResolution:  0,
		Merged:           false,
		CameraID:         0,
		CameraModel:      "",
		CameraMake:       "",
		CameraType:       "",
		LensID:           0,
		LensModel:        "",
		LensMake:         "",
		CellID:           "",
		PlaceID:          "",
		PlaceLabel:       "",
		PlaceCity:        "",
		PlaceState:       "",
		PlaceCountry:     "",
		FileID:           0,
		FileUID:          "",
		FilePrimary:      false,
		FileMissing:      false,
		FileName:         "",
		FileHash:         "",
		FileType:         "",
		FileMime:         "",
		FileWidth:        0,
		FileHeight:       0,
		FileOrientation:  0,
		FileAspectRatio:  0,
		FileColors:       "",
		FileChroma:       0,
		FileLuminance:    "",
		FileDiff:         0,
		Files:            nil,
	}

	results := PhotoResults{result1, result2}

	result := results.UIDs()
	assert.Equal(t, []string{uid1, uid2}, result)
}

func TestPhotosResult_ShareFileName(t *testing.T) {
	t.Run("WithTitle", func(t *testing.T) {
		uid := rnd.GenerateUID(entity.PhotoUID)
		result1 := Photo{
			ID:               111111,
			CreatedAt:        time.Time{},
			UpdatedAt:        time.Time{},
			DeletedAt:        &time.Time{},
			TakenAt:          time.Date(2015, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenAtLocal:     time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenSrc:         "",
			TimeZone:         "Local",
			PhotoUID:         uid,
			PhotoPath:        "",
			PhotoName:        "",
			PhotoTitle:       "PhotoTitle123",
			PhotoYear:        0,
			PhotoMonth:       0,
			PhotoCountry:     "",
			PhotoFavorite:    false,
			PhotoPrivate:     false,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoAltitude:    0,
			PhotoIso:         0,
			PhotoFocalLength: 0,
			PhotoFNumber:     0,
			PhotoExposure:    "",
			PhotoQuality:     0,
			PhotoResolution:  0,
			Merged:           false,
			CameraID:         0,
			CameraModel:      "",
			CameraMake:       "",
			CameraType:       "",
			LensID:           0,
			LensModel:        "",
			LensMake:         "",
			CellID:           "",
			PlaceID:          "",
			PlaceLabel:       "",
			PlaceCity:        "",
			PlaceState:       "",
			PlaceCountry:     "",
			FileID:           0,
			FileUID:          "",
			FilePrimary:      false,
			FileMissing:      false,
			FileName:         "",
			FileHash:         "",
			FileType:         "",
			FileMime:         "",
			FileWidth:        0,
			FileHeight:       0,
			FileOrientation:  0,
			FileAspectRatio:  0,
			FileColors:       "",
			FileChroma:       0,
			FileLuminance:    "",
			FileDiff:         0,
			Files:            nil,
		}

		r := result1.ShareBase(0)
		assert.Contains(t, r, "20131111-090718-Phototitle123")
	})
	t.Run("NoTitle", func(t *testing.T) {
		uid := rnd.GenerateUID(entity.PhotoUID)
		result1 := Photo{
			ID:               111111,
			CreatedAt:        time.Time{},
			UpdatedAt:        time.Time{},
			DeletedAt:        &time.Time{},
			TakenAt:          time.Date(2013, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenAtLocal:     time.Date(2015, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenSrc:         "",
			TimeZone:         "Local",
			PhotoUID:         uid,
			PhotoPath:        "",
			PhotoName:        "",
			PhotoTitle:       "",
			PhotoYear:        0,
			PhotoMonth:       0,
			PhotoCountry:     "",
			PhotoFavorite:    false,
			PhotoPrivate:     false,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoAltitude:    0,
			PhotoIso:         0,
			PhotoFocalLength: 0,
			PhotoFNumber:     0,
			PhotoExposure:    "",
			PhotoQuality:     0,
			PhotoResolution:  0,
			Merged:           false,
			CameraID:         0,
			CameraModel:      "",
			CameraMake:       "",
			CameraType:       "",
			LensID:           0,
			LensModel:        "",
			LensMake:         "",
			CellID:           "",
			PlaceID:          "",
			PlaceLabel:       "",
			PlaceCity:        "",
			PlaceState:       "",
			PlaceCountry:     "",
			FileID:           0,
			FileUID:          "",
			FilePrimary:      false,
			FileMissing:      false,
			FileName:         "",
			FileHash:         "",
			FileType:         "",
			FileMime:         "",
			FileWidth:        0,
			FileHeight:       0,
			FileOrientation:  0,
			FileAspectRatio:  0,
			FileColors:       "",
			FileChroma:       0,
			FileLuminance:    "",
			FileDiff:         0,
			Files:            nil,
		}

		r := result1.ShareBase(0)
		assert.Contains(t, r, fmt.Sprintf("20151111-090718-%s", uid))
	})
	t.Run("SeqGreater0", func(t *testing.T) {
		uid := rnd.GenerateUID(entity.PhotoUID)
		result1 := Photo{
			ID:               111111,
			CreatedAt:        time.Time{},
			UpdatedAt:        time.Time{},
			DeletedAt:        &time.Time{},
			TakenAt:          time.Date(2022, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenAtLocal:     time.Date(2022, 11, 11, 9, 7, 18, 0, time.UTC),
			TakenSrc:         "",
			TimeZone:         "Local",
			PhotoUID:         uid,
			PhotoPath:        "",
			PhotoName:        "",
			PhotoTitle:       "PhotoTitle123",
			PhotoYear:        0,
			PhotoMonth:       0,
			PhotoCountry:     "",
			PhotoFavorite:    false,
			PhotoPrivate:     false,
			PhotoLat:         0,
			PhotoLng:         0,
			PhotoAltitude:    0,
			PhotoIso:         0,
			PhotoFocalLength: 0,
			PhotoFNumber:     0,
			PhotoExposure:    "",
			PhotoQuality:     0,
			PhotoResolution:  0,
			Merged:           false,
			CameraID:         0,
			CameraModel:      "",
			CameraMake:       "",
			CameraType:       "",
			LensID:           0,
			LensModel:        "",
			LensMake:         "",
			CellID:           "",
			PlaceID:          "",
			PlaceLabel:       "",
			PlaceCity:        "",
			PlaceState:       "",
			PlaceCountry:     "",
			FileID:           0,
			FileUID:          "",
			FilePrimary:      false,
			FileMissing:      false,
			FileName:         "",
			FileHash:         "",
			FileType:         "",
			FileMime:         "",
			FileWidth:        0,
			FileHeight:       0,
			FileOrientation:  0,
			FileAspectRatio:  0,
			FileColors:       "",
			FileChroma:       0,
			FileLuminance:    "",
			FileDiff:         0,
			Files:            nil,
		}

		r := result1.ShareBase(3)
		assert.Contains(t, r, "20221111-090718-Phototitle123 (3)")
	})
}

func TestPhoto_HasFisheyeOriginal(t *testing.T) {
	t.Run("PrimaryFisheye", func(t *testing.T) {
		r := Photo{FileProjection: "dual-fisheye"}
		assert.True(t, r.HasFisheyeOriginal())
	})
	t.Run("MergedFileFisheye", func(t *testing.T) {
		r := Photo{Files: []entity.File{{FileProjection: "fisheye"}}}
		assert.True(t, r.HasFisheyeOriginal())
	})
	t.Run("EquirectangularOnly", func(t *testing.T) {
		r := Photo{FileProjection: "equirectangular", Files: []entity.File{{FileProjection: "equirectangular"}}}
		assert.False(t, r.HasFisheyeOriginal())
	})
	t.Run("None", func(t *testing.T) {
		r := Photo{}
		assert.False(t, r.HasFisheyeOriginal())
	})
}

// TestPhoto_MediaProjection_StackedEquirectangular verifies that an unrelated 360° file stacked on
// a flat picture does not promote that picture to a sphere.
func TestPhoto_MediaProjection_StackedEquirectangular(t *testing.T) {
	r := Photo{
		PhotoType:      "image",
		FileProjection: "",
		Files: []entity.File{
			{MediaType: media.Image.String(), FileHash: "flat", FileProjection: ""},
			{MediaType: media.Image.String(), FileHash: "sphere", FileProjection: "equirectangular"},
		},
	}

	assert.Equal(t, "", r.MediaProjection())
}
