package mcp

import (
	"context"
	"path"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

// visiblePhotoUID returns the UID of a picture with files, coordinates, and a primary file that matches the condition.
func visiblePhotoUID(t *testing.T, cond string, args ...any) string {
	t.Helper()

	var uid string
	require.NoError(t, entity.Db().Table("photos").Select("photos.photo_uid").
		Joins("JOIN files ON files.photo_id = photos.id AND files.file_primary = ? AND files.deleted_at IS NULL", true).
		Where("photos.deleted_at IS NULL AND photos.photo_quality > -1").Where(cond, args...).
		Order("photos.id").Limit(1).Row().Scan(&uid))

	return uid
}

func TestGetMedia(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")

	t.Run("Admin", func(t *testing.T) {
		uid := visiblePhotoUID(t, "photos.photo_lat <> 0")
		_, out, err := getMedia(context.Background(), sessionRequest(admin), GetMediaInput{UID: uid})
		require.NoError(t, err)
		assert.Equal(t, uid, out.UID)
		assert.NotEmpty(t, out.Files)
		assert.Positive(t, out.Width)
		assert.Positive(t, out.Height)
		require.NotNil(t, out.Location)

		for _, f := range out.Files {
			assert.NotContains(t, f.Name, "/")
		}
	})
	t.Run("UpperCase", func(t *testing.T) {
		uid := visiblePhotoUID(t, "1 = 1")
		_, out, err := getMedia(context.Background(), sessionRequest(admin), GetMediaInput{UID: " " + string(rune(uid[0]-32)) + uid[1:]})
		require.NoError(t, err)
		assert.Equal(t, uid, out.UID)
	})
	t.Run("NotFound", func(t *testing.T) {
		_, _, err := getMedia(context.Background(), sessionRequest(admin), GetMediaInput{UID: "ps6sg6be2lvl0yzz"})
		require.Error(t, err)
		assert.Equal(t, "media ps6sg6be2lvl0yzz not found; call search_media to find media", err.Error())
	})
	t.Run("Invisible", func(t *testing.T) {
		// A picture the session cannot see reads exactly like one that does not exist.
		uid := visiblePhotoUID(t, "photos.photo_private = ?", true)
		guest := &entity.Session{}
		guest.SetUser(entity.UserFixtures.Pointer("guest"))
		_, _, err := getMedia(context.Background(), sessionRequest(guest), GetMediaInput{UID: uid})
		require.Error(t, err)
		assert.Equal(t, "media "+uid+" not found; call search_media to find media", err.Error())
	})
	t.Run("InvalidUID", func(t *testing.T) {
		_, _, err := getMedia(context.Background(), sessionRequest(admin), GetMediaInput{UID: "as6sg6bxpogaaba7"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "uid must be a media UID")
	})
	t.Run("NoSession", func(t *testing.T) {
		_, _, err := getMedia(context.Background(), &sdkmcp.CallToolRequest{}, GetMediaInput{UID: "ps6sg6be2lvl0yzz"})
		assert.Error(t, err)
	})
}

func TestNewMediaDetails(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")
	uid := visiblePhotoUID(t, "photos.photo_lat <> 0")
	p, err := query.PhotoPreloadByUID(uid)
	require.NoError(t, err)

	t.Run("Admin", func(t *testing.T) {
		out := newMediaDetails(&p, admin)
		assert.Equal(t, p.PhotoQuality, out.Quality)
		assert.Equal(t, p.PhotoQuality >= 0 && p.PhotoQuality < 3, out.Review)
		assert.Equal(t, p.PhotoPrivate, out.Private)
		require.NotNil(t, out.Location)
		assert.Equal(t, p.PhotoLat, out.Location.Lat)
	})
	t.Run("WithoutPlaces", func(t *testing.T) {
		// Coordinates require view access to places, which this token's scope does not grant.
		client := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "mcp photos"}
		require.False(t, client.SeesAnyDetail(acl.ResourcePlaces))
		out := newMediaDetails(&p, client)
		assert.Nil(t, out.Location)
		assert.NotEmpty(t, out.Files)
	})
	t.Run("WithoutFullAccess", func(t *testing.T) {
		client := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "*"}
		client.SetClient(&entity.Client{ClientUID: "cs5gfen1bgxz7s9i", ClientName: "test", ClientRole: acl.RoleInstance.String(),
			AuthProvider: authn.ProviderClient.String(), AuthScope: "*"})
		assert.Empty(t, newMediaDetails(&p, client).Keywords, "without redaction")
		out := mediaDetailsFor(t, uid, client)
		assert.Empty(t, out.Keywords)
		assert.Empty(t, out.Labels)
		assert.Empty(t, out.People)
	})
}

func TestRegisterGetMedia(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test"}, nil)
	registerGetMedia(server)

	access, ok := ToolDeclaration("get_media")
	assert.True(t, ok)
	assert.Equal(t, ToolAccess{Resource: acl.ResourcePhotos, Action: acl.ActionView}, access)
	assert.Equal(t, []string{"get_media"}, toolNames(t, connectServer(t, server)))
}

func TestErrMediaNotFound(t *testing.T) {
	assert.Equal(t, "media ps6sg6be2lvl0yzz not found; call search_media to find media", errMediaNotFound("ps6sg6be2lvl0yzz").Error())
}

// mediaDetailsFor returns the details of a picture as get_media builds them for the session.
func mediaDetailsFor(t *testing.T, uid string, s *entity.Session) MediaDetails {
	t.Helper()

	p, err := query.PhotoPreloadByUID(uid)
	require.NoError(t, err)
	p.RedactForSession(s)

	return newMediaDetails(&p, s)
}

func TestNewMediaDetails_Fields(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")

	t.Run("SameAsSearch", func(t *testing.T) {
		// The media fields equal the full detail of search_media for the same session.
		rows, _, err := search.UserPhotos(form.SearchPhotos{Order: "newest", Count: search.MaxResults}, admin)
		require.NoError(t, err)

		instance := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "*"}
		instance.SetClient(&entity.Client{ClientUID: "cs5gfen1bgxz7s9i", ClientName: "test", ClientRole: acl.RoleInstance.String(),
			AuthProvider: authn.ProviderClient.String(), AuthScope: "*"})

		for _, s := range []*entity.Session{admin, {AuthProvider: authn.ProviderClient.String(), AuthScope: "mcp photos"}, instance} {
			seen := make(map[uint]bool)

			for _, row := range rows {
				if seen[row.ID] {
					continue
				}

				seen[row.ID] = true
				items := []Media{newMedia(row)}
				addMediaDetails(items, search.PhotoResults{row}, s)
				assert.Equal(t, items[0], mediaDetailsFor(t, row.PhotoUID, s).Media, row.PhotoUID)
			}
		}
	})
	t.Run("Files", func(t *testing.T) {
		uid := visiblePhotoUID(t, "photos.photo_lat <> 0")
		p, err := query.PhotoPreloadByUID(uid)
		require.NoError(t, err)
		out := mediaDetailsFor(t, uid, admin)

		primary, err := p.PrimaryFile()
		require.NoError(t, err)
		assert.Equal(t, primary.FileWidth, out.Width)
		assert.Equal(t, primary.FileHeight, out.Height)
		assert.Equal(t, p.PhotoDuration.Seconds(), out.Duration)
		assert.Equal(t, &Location{Lat: p.PhotoLat, Lng: p.PhotoLng, Altitude: p.PhotoAltitude}, out.Location)

		var primaries int
		for _, f := range out.Files {
			if f.Primary {
				primaries++
				assert.Equal(t, MediaFile{Name: path.Base(primary.FileName), Type: primary.FileType, Size: primary.FileSize, Primary: true}, f)
			}
		}
		assert.Equal(t, 1, primaries)
	})
	t.Run("MissingFile", func(t *testing.T) {
		var uid string
		require.NoError(t, entity.Db().Table("files").Select("photo_uid").Where("file_missing = ? AND deleted_at IS NULL", true).Limit(1).Row().Scan(&uid))
		p, err := query.PhotoPreloadByUID(uid)
		require.NoError(t, err)

		var missing string
		for _, f := range p.Files {
			if f.FileMissing {
				missing = path.Base(f.FileName)
			}
		}
		require.NotEmpty(t, missing)

		for _, f := range mediaDetailsFor(t, uid, admin).Files {
			assert.NotEqual(t, missing, f.Name)
		}
	})
	t.Run("Archived", func(t *testing.T) {
		var uid string
		require.NoError(t, entity.UnscopedDb().Table("photos").Select("photo_uid").Where("deleted_at IS NOT NULL").Limit(1).Row().Scan(&uid))
		assert.True(t, mediaDetailsFor(t, uid, admin).Archived)
		assert.False(t, mediaDetailsFor(t, visiblePhotoUID(t, "1 = 1"), admin).Archived)
	})
	t.Run("Review", func(t *testing.T) {
		assert.True(t, mediaDetailsFor(t, visiblePhotoUID(t, "photos.photo_quality BETWEEN 0 AND 2"), admin).Review)
		assert.False(t, mediaDetailsFor(t, visiblePhotoUID(t, "photos.photo_quality >= 3"), admin).Review)

		var hidden string
		require.NoError(t, entity.UnscopedDb().Table("photos").Select("photo_uid").Where("photo_quality < 0").Limit(1).Row().Scan(&hidden))
		p, err := query.PhotoPreloadByUID(hidden)
		require.NoError(t, err)
		assert.False(t, newMediaDetails(&p, admin).Review)
	})
	t.Run("Keywords", func(t *testing.T) {
		var uid string
		require.NoError(t, entity.Db().Table("details").Select("photos.photo_uid").Joins("JOIN photos ON photos.id = details.photo_id").
			Where("details.keywords <> ''").Limit(1).Row().Scan(&uid))
		p, err := query.PhotoPreloadByUID(uid)
		require.NoError(t, err)
		require.NotNil(t, p.Details)

		out := mediaDetailsFor(t, uid, admin)
		require.NotEmpty(t, out.Keywords)
		for _, k := range out.Keywords {
			assert.Contains(t, p.Details.Keywords, k)
		}

		instance := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "*"}
		instance.SetClient(&entity.Client{ClientUID: "cs5gfen1bgxz7s9i", ClientName: "test", ClientRole: acl.RoleInstance.String(),
			AuthProvider: authn.ProviderClient.String(), AuthScope: "*"})
		assert.Empty(t, mediaDetailsFor(t, uid, instance).Keywords)
	})
	t.Run("People", func(t *testing.T) {
		var uid string
		require.NoError(t, entity.Db().Table("markers m").Select("p.photo_uid").Joins("JOIN files f ON f.file_uid = m.file_uid").
			Joins("JOIN photos p ON p.id = f.photo_id").Joins("JOIN subjects s ON s.subj_uid = m.subj_uid AND s.deleted_at IS NULL").
			Where("m.subj_uid <> '' AND m.marker_invalid = ? AND f.deleted_at IS NULL", false).Limit(1).Row().Scan(&uid))
		assert.NotEmpty(t, mediaDetailsFor(t, uid, admin).People)
	})
}
