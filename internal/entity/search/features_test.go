package search

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestApplyPhotoFeatures(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")
	client := &entity.Session{AuthProvider: authn.ProviderClient.String()}
	guest := &entity.Session{}
	guest.SetUser(entity.UserFixtures.Pointer("guest"))

	t.Run("PrivateDisabled", func(t *testing.T) {
		frm := form.SearchPhotos{Public: true}
		ApplyPhotoFeatures(&frm, admin, false, false)
		assert.False(t, frm.Public)
	})
	t.Run("PrivateEnabled", func(t *testing.T) {
		frm := form.SearchPhotos{Public: true}
		ApplyPhotoFeatures(&frm, admin, true, false)
		assert.True(t, frm.Public)
	})
	t.Run("ReviewManager", func(t *testing.T) {
		frm := form.SearchPhotos{}
		ApplyPhotoFeatures(&frm, admin, true, true)
		assert.Equal(t, 0, frm.Quality)
		ApplyPhotoFeatures(&frm, client, true, true)
		assert.Equal(t, 0, frm.Quality)
	})
	t.Run("ReviewOthers", func(t *testing.T) {
		frm := form.SearchPhotos{}
		ApplyPhotoFeatures(&frm, guest, true, true)
		assert.Equal(t, 3, frm.Quality)
	})
	t.Run("ReviewScoped", func(t *testing.T) {
		frm := form.SearchPhotos{Scope: "as6sg6bxpogaaba7"}
		ApplyPhotoFeatures(&frm, guest, true, true)
		assert.Equal(t, 0, frm.Quality)
	})
	t.Run("ReviewDisabled", func(t *testing.T) {
		frm := form.SearchPhotos{}
		ApplyPhotoFeatures(&frm, guest, true, false)
		assert.Equal(t, 0, frm.Quality)
	})
	t.Run("Nil", func(t *testing.T) {
		ApplyPhotoFeatures(nil, guest, false, true)
	})
}

func TestApplyAlbumFeatures(t *testing.T) {
	t.Run("PrivateDisabled", func(t *testing.T) {
		frm := form.SearchAlbums{Public: true}
		ApplyAlbumFeatures(&frm, false)
		assert.False(t, frm.Public)
	})
	t.Run("PrivateEnabled", func(t *testing.T) {
		frm := form.SearchAlbums{Public: true}
		ApplyAlbumFeatures(&frm, true)
		assert.True(t, frm.Public)
	})
	t.Run("Nil", func(t *testing.T) {
		ApplyAlbumFeatures(nil, false)
	})
}

func TestUserLabels(t *testing.T) {
	var label entity.Label

	if err := entity.Db().Where("photo_count > 0 AND label_nsfw = ? AND deleted_at IS NULL", false).Order("id").First(&label).Error; err != nil {
		t.Fatal(err)
	}

	if err := entity.Db().Model(&entity.Label{}).Where("id = ?", label.ID).UpdateColumn("label_nsfw", true).Error; err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = entity.Db().Model(&entity.Label{}).Where("id = ?", label.ID).UpdateColumn("label_nsfw", false).Error
	})

	frm := form.SearchLabels{UID: label.LabelUID, All: true, Count: 10}

	t.Run("Admin", func(t *testing.T) {
		results, err := UserLabels(frm, entity.SessionFixtures.Pointer("alice"))
		assert.NoError(t, err)
		assert.Len(t, results, 1)
	})
	t.Run("ClientWithoutUser", func(t *testing.T) {
		results, err := UserLabels(frm, &entity.Session{AuthProvider: authn.ProviderClient.String()})
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
	t.Run("Guest", func(t *testing.T) {
		guest := &entity.Session{}
		guest.SetUser(entity.UserFixtures.Pointer("guest"))
		results, err := UserLabels(frm, guest)
		assert.NoError(t, err)
		assert.Empty(t, results)
	})
	t.Run("NoSession", func(t *testing.T) {
		results, err := UserLabels(frm, nil)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestLabelsPublicOnly(t *testing.T) {
	guest := &entity.Session{}
	guest.SetUser(entity.UserFixtures.Pointer("guest"))

	assert.False(t, LabelsPublicOnly(nil))
	assert.False(t, LabelsPublicOnly(entity.SessionFixtures.Pointer("alice")))
	assert.True(t, LabelsPublicOnly(guest))
	assert.True(t, LabelsPublicOnly(&entity.Session{AuthProvider: authn.ProviderClient.String()}))
}

func TestAlbumTypeResource(t *testing.T) {
	for albumType, resource := range map[string]acl.Resource{
		entity.AlbumManual: acl.ResourceAlbums,
		entity.AlbumFolder: acl.ResourceFolders,
		entity.AlbumMoment: acl.ResourceMoments,
		entity.AlbumMonth:  acl.ResourceCalendar,
		entity.AlbumState:  acl.ResourcePlaces,
		"":                 acl.ResourceDefault,
		"album|folder":     acl.ResourceDefault,
	} {
		assert.Equal(t, resource, AlbumTypeResource(albumType), albumType)
	}
}
