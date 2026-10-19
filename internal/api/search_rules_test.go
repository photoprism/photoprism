package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/authn"
)

// privateLabelFixture marks a label with pictures as private and restores it after the test.
func privateLabelFixture(t *testing.T) entity.Label {
	t.Helper()

	var label entity.Label
	require.NoError(t, entity.Db().Where("photo_count > 0 AND label_nsfw = ? AND deleted_at IS NULL", false).Order("id").First(&label).Error)
	require.NoError(t, entity.Db().Model(&entity.Label{}).Where("id = ?", label.ID).UpdateColumn("label_nsfw", true).Error)
	t.Cleanup(func() {
		_ = entity.Db().Model(&entity.Label{}).Where("id = ?", label.ID).UpdateColumn("label_nsfw", false).Error
	})

	return label
}

// searchablePhotoUID returns the UID of a picture that the photo search finds and that matches the condition.
func searchablePhotoUID(t *testing.T, cond string, args ...any) string {
	t.Helper()

	var uid string
	require.NoError(t, entity.Db().Table("photos").Select("photos.photo_uid").
		Joins("JOIN files ON files.photo_id = photos.id AND files.media_id IS NOT NULL AND files.file_sidecar = ? AND files.file_error = ''", false).
		Where("photos.deleted_at IS NULL AND photos.photo_quality > -1").Where(cond, args...).
		Order("photos.id").Limit(1).Row().Scan(&uid))

	return uid
}

// TestSearchLabels_PrivateLabels checks that the labels endpoint returns private labels only to sessions
// whose user role may see private content.
func TestSearchLabels_PrivateLabels(t *testing.T) {
	label := privateLabelFixture(t)
	url := "/api/v1/labels?count=100&all=true&uid=" + label.LabelUID

	t.Run("Admin", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		SearchLabels(router)

		r := AuthenticatedRequest(app, http.MethodGet, url, AuthenticateAdmin(app, router))
		require.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), label.LabelUID)
	})
	t.Run("ClientWithoutUser", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		defer conf.SetAuthMode(config.AuthModePublic)
		SearchLabels(router)

		sess, err := entity.AddClientSession("labels-rule-client", 3600, "labels", authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })

		r := AuthenticatedRequest(app, http.MethodGet, url, sess.AuthToken())
		require.Equal(t, http.StatusOK, r.Code)
		assert.NotContains(t, r.Body.String(), label.LabelUID)

		// The label is returned once it is public.
		require.NoError(t, entity.Db().Model(&entity.Label{}).Where("id = ?", label.ID).UpdateColumn("label_nsfw", false).Error)
		r = AuthenticatedRequest(app, http.MethodGet, url, sess.AuthToken())
		require.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), label.LabelUID)
		require.NoError(t, entity.Db().Model(&entity.Label{}).Where("id = ?", label.ID).UpdateColumn("label_nsfw", true).Error)
	})
}

// TestSearchAlbums_PrivateFeature checks that the public filter of the albums endpoint is ignored while the
// Private feature is disabled.
func TestSearchAlbums_PrivateFeature(t *testing.T) {
	const albumUID = "as6sg6bxpogaaba7"

	require.NoError(t, entity.Db().Model(&entity.Album{}).Where("album_uid = ?", albumUID).UpdateColumn("album_private", true).Error)
	t.Cleanup(func() {
		_ = entity.Db().Model(&entity.Album{}).Where("album_uid = ?", albumUID).UpdateColumn("album_private", false).Error
	})

	app, router, conf := NewApiTest()
	SearchAlbums(router)

	features := conf.Settings().Features
	t.Cleanup(func() { conf.Settings().Features = features })

	t.Run("Enabled", func(t *testing.T) {
		conf.Settings().Features.Private = true
		r := PerformRequest(app, http.MethodGet, "/api/v1/albums?count=10&public=true&uid="+albumUID)
		require.Equal(t, http.StatusOK, r.Code)
		assert.NotContains(t, r.Body.String(), albumUID)

		r = PerformRequest(app, http.MethodGet, "/api/v1/albums?count=10&uid="+albumUID)
		require.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), albumUID)
	})
	t.Run("Disabled", func(t *testing.T) {
		conf.Settings().Features.Private = false
		r := PerformRequest(app, http.MethodGet, "/api/v1/albums?count=10&public=true&uid="+albumUID)
		require.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), albumUID)
	})
}

// TestSearchPhotos_PrivateFeature checks that the public filter of the photos endpoint is ignored while the
// Private feature is disabled.
func TestSearchPhotos_PrivateFeature(t *testing.T) {
	app, router, conf := NewApiTest()
	photoUID := searchablePhotoUID(t, "photos.photo_private = ?", true)

	SearchPhotos(router)

	features := conf.Settings().Features
	t.Cleanup(func() { conf.Settings().Features = features })

	t.Run("Enabled", func(t *testing.T) {
		conf.Settings().Features.Private = true
		r := PerformRequest(app, http.MethodGet, "/api/v1/photos?count=10&merged=true&public=true&uid="+photoUID)
		require.Equal(t, http.StatusOK, r.Code)
		assert.NotContains(t, r.Body.String(), photoUID)

		r = PerformRequest(app, http.MethodGet, "/api/v1/photos?count=10&merged=true&uid="+photoUID)
		require.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), photoUID)
	})
	t.Run("Disabled", func(t *testing.T) {
		conf.Settings().Features.Private = false
		r := PerformRequest(app, http.MethodGet, "/api/v1/photos?count=10&merged=true&public=true&uid="+photoUID)
		require.Equal(t, http.StatusOK, r.Code)
		assert.Contains(t, r.Body.String(), photoUID)
	})
}
