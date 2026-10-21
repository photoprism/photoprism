package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/auth"
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

// sessionRequest returns a tool call request whose token info carries the session.
func sessionRequest(s *entity.Session) *sdkmcp.CallToolRequest {
	return &sdkmcp.CallToolRequest{Extra: &sdkmcp.RequestExtra{TokenInfo: &auth.TokenInfo{Extra: map[string]any{SessionKey: s}}}}
}

func TestNewMedia(t *testing.T) {
	takenAt := time.Date(2019, 7, 1, 10, 0, 0, 0, time.UTC)

	t.Run("Place", func(t *testing.T) {
		m := newMedia(search.Photo{PhotoUID: "ps6sg6be2lvl0y11", PhotoType: "image", PhotoTitle: "Beach", TakenAt: takenAt,
			TakenAtLocal: takenAt.Add(2 * time.Hour), PhotoFavorite: true, PlaceID: "de:abc", PlaceLabel: "Berlin, Germany",
			PlaceCountry: "de", PhotoCountry: "de", PhotoLat: 52.5, PhotoLng: 13.4})
		assert.Equal(t, "ps6sg6be2lvl0y11", m.UID)
		assert.Equal(t, "image", m.Type)
		assert.Equal(t, "Beach", m.Title)
		assert.Equal(t, "2019-07-01T10:00:00Z", m.TakenAt)
		assert.Equal(t, "2019-07-01T12:00:00", m.TakenAtLocal)
		assert.True(t, m.Favorite)
		assert.Equal(t, &Place{Name: "Berlin, Germany", Country: "de"}, m.Place)
	})
	t.Run("UnknownPlace", func(t *testing.T) {
		m := newMedia(search.Photo{PhotoUID: "ps6sg6be2lvl0y11", TakenAt: takenAt, PlaceID: entity.UnknownPlace.ID,
			PlaceLabel: "Unknown", PhotoCountry: entity.UnknownCountry.ID})
		assert.Nil(t, m.Place)
	})
	t.Run("CountryOnly", func(t *testing.T) {
		m := newMedia(search.Photo{PhotoUID: "ps6sg6be2lvl0y11", TakenAt: takenAt, PlaceID: entity.UnknownPlace.ID, PhotoCountry: "fr"})
		assert.Equal(t, &Place{Country: "fr"}, m.Place)
	})
}

func TestNewMediaList(t *testing.T) {
	assert.Empty(t, newMediaList(nil))
	assert.Len(t, newMediaList(search.PhotoResults{{PhotoUID: "a"}, {PhotoUID: "b"}}), 2)
}

func TestAddMediaDetails(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")
	rows, _, err := search.UserPhotos(form.SearchPhotos{Order: "newest", Count: search.MaxResults}, admin)
	require.NoError(t, err)

	items := newMediaList(rows)
	addMediaDetails(items, rows, admin)

	var labels, people, collections, cameras int

	for _, m := range items {
		labels += len(m.Labels)
		people += len(m.People)
		collections += len(m.Collections)

		if m.Camera != "" {
			cameras++
			assert.NotEqual(t, "Unknown", m.Camera)
		}
	}

	assert.Positive(t, labels)
	assert.Positive(t, people)
	assert.Positive(t, collections)
	assert.Positive(t, cameras)

	t.Run("WithoutFullAccess", func(t *testing.T) {
		// A client that may search pictures, but has no full access to them, gets no labels and no people.
		client := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "*"}
		client.SetClient(&entity.Client{ClientUID: "cs5gfen1bgxz7s9i", ClientName: "test", ClientRole: acl.RoleInstance.String(),
			AuthProvider: authn.ProviderClient.String(), AuthScope: "*"})
		require.True(t, client.Grants(acl.ResourcePhotos, acl.ActionSearch))
		require.False(t, client.SeesFullDetail(acl.ResourcePhotos))

		items := newMediaList(rows)
		addMediaDetails(items, rows, client)

		for _, m := range items {
			assert.Empty(t, m.Labels, m.UID)
			assert.Empty(t, m.People, m.UID)
		}
	})
	t.Run("WithoutAlbumsScope", func(t *testing.T) {
		client := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "mcp photos"}
		require.True(t, client.SeesFullDetail(acl.ResourcePhotos))

		items := newMediaList(rows)
		addMediaDetails(items, rows, client)

		var labels, collections int
		for _, m := range items {
			labels += len(m.Labels)
			collections += len(m.Collections)
		}

		assert.Positive(t, labels)
		assert.Zero(t, collections)
	})
	t.Run("CollectionTypes", func(t *testing.T) {
		// Each collection type is checked by its own resource.
		var momentUID string
		require.NoError(t, entity.Db().Table("albums").Select("album_uid").Where("album_type = ? AND deleted_at IS NULL", entity.AlbumMoment).
			Order("id").Limit(1).Row().Scan(&momentUID))
		require.NoError(t, entity.Db().Create(&entity.PhotoAlbum{PhotoUID: rows[0].PhotoUID, AlbumUID: momentUID}).Error)
		t.Cleanup(func() {
			_ = entity.UnscopedDb().Delete(&entity.PhotoAlbum{}, "photo_uid = ? AND album_uid = ?", rows[0].PhotoUID, momentUID).Error
		})

		hasMoment := func(s *entity.Session) bool {
			items := newMediaList(rows[:1])
			addMediaDetails(items, rows[:1], s)
			for _, c := range items[0].Collections {
				if c.UID == momentUID {
					assert.Equal(t, entity.AlbumMoment, c.Type)
					return true
				}
			}
			return false
		}

		client := &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "*"}
		assert.True(t, hasMoment(admin))
		assert.True(t, hasMoment(client))

		// The token scope of the type applies too.
		assert.False(t, hasMoment(&entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "mcp photos albums"}))
		assert.True(t, hasMoment(&entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "mcp photos albums moments"}))

		acl.RulesMutex.Lock()
		grant := acl.Rules[acl.ResourceMoments][acl.RoleClient]
		delete(acl.Rules[acl.ResourceMoments], acl.RoleClient)
		acl.RulesMutex.Unlock()
		t.Cleanup(func() {
			acl.RulesMutex.Lock()
			acl.Rules[acl.ResourceMoments][acl.RoleClient] = grant
			acl.RulesMutex.Unlock()
		})

		assert.False(t, hasMoment(client))
	})
	t.Run("LabelsWithoutCount", func(t *testing.T) {
		// Labels of private or archived pictures may have no public count, and are still listed.
		ids := make([]uint, len(rows))
		for i, row := range rows {
			ids[i] = row.ID
		}

		var label struct {
			ID       uint
			LabelUID string
			PhotoID  uint
			Count    int
		}
		require.NoError(t, entity.Db().Table("photos_labels pl").Select("l.id, l.label_uid, pl.photo_id, l.photo_count AS count").
			Joins("JOIN labels l ON l.id = pl.label_id AND l.deleted_at IS NULL").Where("pl.uncertainty < 100 AND pl.photo_id IN (?)", ids).
			Order("pl.photo_id").Limit(1).Scan(&label).Error)
		require.NotEmpty(t, label.LabelUID)
		require.NoError(t, entity.Db().Table("labels").Where("id = ?", label.ID).UpdateColumn("photo_count", 0).Error)
		t.Cleanup(func() {
			_ = entity.Db().Table("labels").Where("id = ?", label.ID).UpdateColumn("photo_count", label.Count).Error
		})

		items := newMediaList(rows)
		addMediaDetails(items, rows, admin)

		var found bool
		for i, row := range rows {
			if row.ID == label.PhotoID {
				for _, l := range items[i].Labels {
					found = found || l.UID == label.LabelUID
				}
			}
		}
		assert.True(t, found)
	})
	t.Run("Mismatch", func(t *testing.T) {
		items := newMediaList(rows[:1])
		addMediaDetails(items, rows, admin)
		assert.Empty(t, items[0].Camera)
	})
}

func TestDeviceNames(t *testing.T) {
	var lens entity.Lens
	require.NoError(t, entity.Db().Where("lens_slug <> ?", entity.UnknownID).First(&lens).Error)

	names := deviceNames("lenses", "lens_name", "lens_slug", []uint{lens.ID, entity.UnknownLens.ID, 0})
	assert.Equal(t, map[uint]string{lens.ID: lens.LensName}, names)
	assert.Empty(t, deviceNames("cameras", "camera_name", "camera_slug", nil))
}

func TestRefUIDs(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, refUIDs([]photoRef{{1, "a"}, {2, "b"}, {3, "a"}}))
	assert.Empty(t, refUIDs(nil))
}

func TestAppendUnique(t *testing.T) {
	assert.Equal(t, []string{"a", "b"}, appendUnique(appendUnique([]string{"a"}, "b"), "a"))
}

func TestSearchMediaError(t *testing.T) {
	for _, err := range []error{search.ErrForbidden, search.ErrInvalidId, search.ErrNotFound} {
		assert.Equal(t, "the search found no accessible media for this collection or filter", searchMediaError(err).Error())
	}

	assert.Contains(t, searchMediaError(search.ErrBadSortOrder).Error(), "order must be one of")
	assert.Equal(t, "the search could not be completed; check the query with find_search_filters", searchMediaError(errors.New("db")).Error())
}

func TestOptionsFeatures(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		assert.Equal(t, customize.NewDefaultSettings().Features, Options{}.features())
		assert.Equal(t, customize.NewDefaultSettings().Features, Options{Settings: func() *customize.Settings { return nil }}.features())
	})
	t.Run("Settings", func(t *testing.T) {
		settings := customize.NewDefaultSettings()
		settings.Features.Review = !settings.Features.Review
		assert.Equal(t, settings.Features, Options{Settings: func() *customize.Settings { return settings }}.features())
	})
}

func TestSearchMedia(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")

	t.Run("Page", func(t *testing.T) {
		_, out, err := searchMedia(context.Background(), sessionRequest(admin), SearchMediaInput{Limit: 2}, Options{})
		require.NoError(t, err)
		assert.Len(t, out.Items, 2)
		assert.Equal(t, 2, out.Limit)
		assert.Positive(t, out.NextOffset)
		assert.True(t, out.Truncated)
		require.Len(t, out.Warnings, 1)
		assert.Contains(t, out.Warnings[0], "call again with offset")
	})
	t.Run("NoMatch", func(t *testing.T) {
		_, out, err := searchMedia(context.Background(), sessionRequest(admin), SearchMediaInput{Query: "label:\"\x00\""}, Options{})
		require.NoError(t, err)
		assert.Empty(t, out.Items)
		assert.False(t, out.Truncated)
	})
	t.Run("NoSession", func(t *testing.T) {
		_, _, err := searchMedia(context.Background(), &sdkmcp.CallToolRequest{}, SearchMediaInput{}, Options{})
		assert.Error(t, err)
	})
	t.Run("InvalidInput", func(t *testing.T) {
		_, _, err := searchMedia(context.Background(), sessionRequest(admin), SearchMediaInput{Order: "size"}, Options{})
		assert.Error(t, err)
	})
	t.Run("UnknownCollection", func(t *testing.T) {
		_, _, err := searchMedia(context.Background(), sessionRequest(admin), SearchMediaInput{Collection: "as6sg6bxpogaab99"}, Options{})
		require.Error(t, err)
		assert.Equal(t, "the search found no accessible media for this collection or filter", err.Error())
	})
	t.Run("FullDetail", func(t *testing.T) {
		_, out, err := searchMedia(context.Background(), sessionRequest(admin), SearchMediaInput{Limit: 50, Detail: "full"}, Options{})
		require.NoError(t, err)
		require.NotEmpty(t, out.Items)

		var labels, people, collections int
		for _, m := range out.Items {
			labels += len(m.Labels)
			people += len(m.People)
			collections += len(m.Collections)
		}

		assert.Positive(t, labels)
		assert.Positive(t, people)
		assert.Positive(t, collections)
	})
	t.Run("Canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _, err := searchMedia(ctx, sessionRequest(admin), SearchMediaInput{}, Options{})
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestRegisterSearchMedia(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test"}, nil)
	registerSearchMedia(server, Options{})

	access, ok := ToolDeclaration("search_media")
	assert.True(t, ok)
	assert.Equal(t, searchMediaAccess, access)
	assert.Equal(t, []string{"search_media"}, toolNames(t, connectServer(t, server)))
}

func TestAddMediaCollections(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")
	rows, _, err := search.UserPhotos(form.SearchPhotos{Order: "newest", Count: search.MaxResults}, admin)
	require.NoError(t, err)

	ids := make([]uint, len(rows))
	index := make(map[uint]int, len(rows))

	for i, row := range rows {
		ids[i] = row.ID
		index[row.ID] = i
	}

	count := func(s *entity.Session) (n int) {
		items := newMediaList(rows)
		addMediaCollections(items, index, ids, s)
		for _, m := range items {
			n += len(m.Collections)
			assert.Empty(t, m.Labels)
			assert.Empty(t, m.People)
		}
		return n
	}

	assert.Positive(t, count(admin))
	assert.Zero(t, count(&entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: "mcp photos"}))
}
