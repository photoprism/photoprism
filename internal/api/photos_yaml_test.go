package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/auth/tokens"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
)

// TestGetPhotoYamlPermissions checks metadata reads against account, client, and scope permissions.
func TestGetPhotoYamlPermissions(t *testing.T) {
	app, router, conf := NewApiTest()
	GetPhotoYaml(router)
	GetPhoto(router)
	GetPhotoDownload(router)

	mode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	conf.Propagate()
	t.Cleanup(func() { conf.SetAuthMode(mode); conf.Propagate() })
	originals := conf.Options().OriginalsPath
	conf.Options().OriginalsPath = t.TempDir()
	t.Cleanup(func() { conf.Options().OriginalsPath = originals })

	photo := entity.NewPhoto(false)
	photo.PhotoTitle = "Metadata Export Control"
	require.NoError(t, photo.Save())
	note := "metadata-export-note-control"
	require.NoError(t, photo.Details.Update("Notes", note))
	t.Cleanup(func() {
		entity.UnscopedDb().Unscoped().Delete(&entity.Details{}, "photo_id = ?", photo.ID)
		entity.UnscopedDb().Unscoped().Delete(&photo)
	})
	file := &entity.File{PhotoID: photo.ID, PhotoUID: photo.PhotoUID,
		FileRoot: entity.RootOriginals, FileName: "metadata-control.jpg", FileType: "jpg",
		FileHash: "a1c771ead6e06ae8f6cba401fe5915779cbbc7bc", MediaType: entity.MediaImage, FilePrimary: true}
	require.NoError(t, file.Create())
	t.Cleanup(func() { entity.UnscopedDb().Unscoped().Delete(file) })
	original := CreateTestOriginal(t, file)

	for _, tc := range []struct {
		name, role, scope, user string
		status                  int
	}{
		{"ReadClient", "client", "read photos", "", http.StatusOK},
		{"AttachedReadClient", "client", "read photos", "alice", http.StatusOK},
		{"FullClient", "client", "photos", "", http.StatusOK},
		{"WriteClient", "client", "write photos", "", http.StatusForbidden},
		{"AttachedWriteClient", "client", "write photos", "alice", http.StatusForbidden},
		{"UnrelatedScope", "client", "metrics", "alice", http.StatusForbidden},
		{"AccountCeiling", "client", "*", "guest", http.StatusForbidden},
		{"ClientCeiling", "instance", "*", "alice", http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var user *entity.User
			if tc.user != "" {
				user = entity.UserFixtures.Pointer(tc.user)
			}
			sess := clientCredentialSession(t, conf, tc.role, tc.scope, user)
			download := PerformRequest(app, http.MethodGet, "/api/v1/photos/"+photo.PhotoUID+"/dl?t="+tokens.DownloadToken(sess.ID))
			r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+photo.PhotoUID+"/yaml", sess.AuthToken())
			json := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+photo.PhotoUID, sess.AuthToken())
			assert.Equal(t, tc.status, r.Code)
			if tc.status == http.StatusOK {
				assert.Contains(t, r.Body.String(), note)
				assert.Contains(t, r.Body.String(), photo.PhotoTitle)
				assert.Equal(t, http.StatusOK, download.Code)
				assert.Equal(t, original, download.Body.Bytes())
				assert.Equal(t, http.StatusOK, json.Code)
			} else {
				assert.NotContains(t, r.Body.String(), note)
			}
			if tc.scope == "write photos" {
				assert.False(t, sess.SeesAnyDetail(acl.ResourcePhotos))
				assert.Equal(t, http.StatusForbidden, json.Code)
				assert.Equal(t, http.StatusForbidden, download.Code)
			}
		})
	}
	for _, scope := range []string{"", "read photos", "write photos"} {
		t.Run(map[string]string{"": "AccountUnrestrictedScope", "read photos": "AccountReadScope", "write photos": "AccountWriteScope"}[scope], func(t *testing.T) {
			sess := entity.NewSession(3600, 0).SetUser(entity.UserFixtures.Pointer("alice"))
			sess.SetScope(scope)
			require.False(t, sess.IsClient())
			require.NoError(t, sess.Create())
			t.Cleanup(func() { require.NoError(t, sess.Delete()) })
			r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+photo.PhotoUID+"/yaml", sess.AuthToken())
			if scope == "write photos" {
				assert.Equal(t, http.StatusForbidden, r.Code)
				assert.NotContains(t, r.Body.String(), note)
			} else {
				assert.Equal(t, http.StatusOK, r.Code)
				assert.Contains(t, r.Body.String(), note)
			}
		})
	}

	stored := entity.FindPhoto(entity.Photo{PhotoUID: photo.PhotoUID})
	require.NotNil(t, stored)
	assert.Equal(t, note, stored.GetDetails().Notes)
}
