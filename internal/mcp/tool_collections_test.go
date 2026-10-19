package mcp

import (
	"context"
	"testing"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/authn"
)

func TestNewCollection(t *testing.T) {
	created := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)

	t.Run("Album", func(t *testing.T) {
		c := newCollection(search.Album{AlbumUID: "as6sg6bxpogaaba7", AlbumType: entity.AlbumManual, AlbumTitle: "Christmas",
			AlbumDescription: "Family", PhotoCount: 3, AlbumYear: 2020, CreatedAt: created, UpdatedAt: created})
		assert.Equal(t, Collection{UID: "as6sg6bxpogaaba7", Type: "album", Title: "Christmas", Description: "Family", MediaCount: 3,
			CreatedAt: "2020-01-02T03:04:05Z", UpdatedAt: "2020-01-02T03:04:05Z"}, c)
	})
	t.Run("Month", func(t *testing.T) {
		c := newCollection(search.Album{AlbumUID: "as6sg6bxpogaaba8", AlbumType: entity.AlbumMonth, AlbumYear: 2019, AlbumMonth: 7, AlbumDay: -1})
		assert.Equal(t, 2019, c.Year)
		assert.Equal(t, 7, c.Month)
		assert.Equal(t, 0, c.Day)
	})
}

func TestValidateCollectionType(t *testing.T) {
	types, err := validateCollectionType("")
	require.NoError(t, err)
	assert.Equal(t, []string{"album"}, types)

	types, err = validateCollectionType(" All ")
	require.NoError(t, err)
	assert.Equal(t, collectionTypes, types)

	types, err = validateCollectionType("folder")
	require.NoError(t, err)
	assert.Equal(t, []string{"folder"}, types)

	_, err = validateCollectionType("albums")
	require.Error(t, err)
	assert.Equal(t, "type must be one of album, folder, moment, month, state, or all", err.Error())
}

func TestCoverUIDs(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")

	var private struct {
		Hash     string
		PhotoUID string
	}
	require.NoError(t, entity.Db().Table("files").Select("files.file_hash AS hash, photos.photo_uid").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.photo_private = ? AND photos.deleted_at IS NULL", true).
		Where("files.file_hash <> '' AND files.deleted_at IS NULL").Limit(1).Scan(&private).Error)
	require.NotEmpty(t, private.Hash)

	assert.Equal(t, map[string]string{private.Hash: private.PhotoUID}, coverUIDs([]string{private.Hash, "", private.Hash}, admin))
	assert.Equal(t, map[string]string{private.Hash + "-0910162fd2fd": private.PhotoUID}, coverUIDs([]string{private.Hash + "-0910162fd2fd"}, admin))
	assert.Empty(t, coverUIDs(nil, admin))

	// Covers of pictures the session may not see are left out.
	guest := &entity.Session{}
	guest.SetUser(entity.UserFixtures.Pointer("guest"))
	assert.Empty(t, coverUIDs([]string{private.Hash}, guest))
}

func TestAddCollectionCovers(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")

	var cover struct {
		Hash     string
		PhotoUID string
	}
	require.NoError(t, entity.Db().Table("files").Select("files.file_hash AS hash, photos.photo_uid").
		Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL").
		Where("files.file_hash <> '' AND files.deleted_at IS NULL").Order("files.id").Limit(1).Scan(&cover).Error)

	albums := search.AlbumResults{{AlbumUID: "as6sg6bxpogaaba7", Thumb: cover.Hash}, {AlbumUID: "as6sg6bxpogaaba8"}}
	items := make([]Collection, len(albums))
	addCollectionCovers(items, albums, admin)
	assert.Equal(t, cover.PhotoUID, items[0].Cover)
	assert.Empty(t, items[1].Cover)
}

// collectionSession returns a session of an API client with the scope, which acts with the client role.
func collectionSession(scope string) *entity.Session {
	return &entity.Session{AuthProvider: authn.ProviderClient.String(), AuthScope: scope}
}

func TestListCollections(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")

	t.Run("Albums", func(t *testing.T) {
		expected, err := search.UserAlbums(form.SearchAlbums{Type: entity.AlbumManual, Count: search.MaxResults}, admin)
		require.NoError(t, err)
		require.NotEmpty(t, expected)

		_, out, err := listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Limit: 50})
		require.NoError(t, err)
		assert.Equal(t, len(expected), out.Total)

		for i, c := range out.Items {
			assert.Equal(t, expected[i].AlbumUID, c.UID)
			assert.Equal(t, "album", c.Type)
		}
	})
	t.Run("Paging", func(t *testing.T) {
		_, all, err := listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Limit: 50})
		require.NoError(t, err)
		require.Greater(t, all.Total, 2)

		_, page, err := listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Limit: 2, Offset: 1})
		require.NoError(t, err)
		assert.Equal(t, all.Items[1:3], page.Items)
		assert.Equal(t, all.Total, page.Total)
		assert.True(t, page.Truncated)
		require.Len(t, page.Warnings, 1)
	})
	t.Run("AllTypes", func(t *testing.T) {
		_, out, err := listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Type: "all", Limit: 50})
		require.NoError(t, err)

		order := map[string]int{}
		for i, c := range collectionTypes {
			order[c] = i
		}

		types := map[string]bool{}
		last := 0
		for _, c := range out.Items {
			types[c.Type] = true
			assert.GreaterOrEqual(t, order[c.Type], last, "types are listed in order")
			last = order[c.Type]
		}
		assert.Greater(t, len(types), 1)
	})
	t.Run("Query", func(t *testing.T) {
		_, out, err := listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Query: "christmas"})
		require.NoError(t, err)
		require.NotEmpty(t, out.Items)

		for _, c := range out.Items {
			assert.Contains(t, c.Title+c.Description, "Christmas")
		}
	})
	t.Run("TypeNotAvailable", func(t *testing.T) {
		// A role without access to a type gets an error for that type, and the other types with type all.
		acl.RulesMutex.Lock()
		grant := acl.Rules[acl.ResourceMoments][acl.RoleClient]
		delete(acl.Rules[acl.ResourceMoments], acl.RoleClient)
		acl.RulesMutex.Unlock()
		t.Cleanup(func() {
			acl.RulesMutex.Lock()
			acl.Rules[acl.ResourceMoments][acl.RoleClient] = grant
			acl.RulesMutex.Unlock()
		})

		client := collectionSession("*")
		_, _, err := listCollections(context.Background(), sessionRequest(client), ListCollectionsInput{Type: "moment"})
		assert.Error(t, err)

		_, out, err := listCollections(context.Background(), sessionRequest(client), ListCollectionsInput{Type: "all", Limit: 50})
		require.NoError(t, err)
		require.NotEmpty(t, out.Items)

		for _, c := range out.Items {
			assert.NotEqual(t, "moment", c.Type)
		}
	})
	t.Run("AllTypesOnce", func(t *testing.T) {
		_, all, err := listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Type: "all", Limit: 50})
		require.NoError(t, err)

		sum := 0
		for _, albumType := range collectionTypes {
			_, out, typeErr := listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Type: albumType})
			require.NoError(t, typeErr, albumType)
			sum += out.Total
		}
		assert.Equal(t, sum, all.Total)

		_, page, err := listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Type: "all", Limit: 50, Offset: 0})
		require.NoError(t, err)
		seen := map[string]bool{}
		for _, c := range page.Items {
			assert.False(t, seen[c.UID], c.UID)
			seen[c.UID] = true
		}
	})
	t.Run("QueryIsText", func(t *testing.T) {
		// Filter syntax in the query is searched as text, so it cannot change the type.
		_, out, err := listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Type: "album", Query: "type:folder"})
		require.NoError(t, err)
		for _, c := range out.Items {
			assert.Equal(t, "album", c.Type)
		}

		_, _, err = listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Query: "Trip: Paris"})
		assert.NoError(t, err)
	})
	t.Run("TypeScope", func(t *testing.T) {
		// Each type needs the token scope of its own resource.
		albumsOnly := collectionSession("mcp albums")

		_, out, err := listCollections(context.Background(), sessionRequest(albumsOnly), ListCollectionsInput{Type: "all", Limit: 50})
		require.NoError(t, err)
		require.NotEmpty(t, out.Items)
		for _, c := range out.Items {
			assert.Equal(t, "album", c.Type)
		}
		assert.Contains(t, out.Warnings, "collections of type folder need the folders scope")
		assert.Contains(t, out.Warnings, "collections of type month need the calendar scope")

		_, _, err = listCollections(context.Background(), sessionRequest(albumsOnly), ListCollectionsInput{Type: "folder"})
		require.Error(t, err)
		assert.Equal(t, "collections of type folder need the folders scope", err.Error())

		_, out, err = listCollections(context.Background(), sessionRequest(collectionSession("mcp albums folders")), ListCollectionsInput{Type: "folder"})
		require.NoError(t, err)
		assert.NotEmpty(t, out.Items)
	})
	t.Run("Covers", func(t *testing.T) {
		var cover struct {
			Hash     string
			PhotoUID string
		}
		require.NoError(t, entity.Db().Table("files").Select("files.file_hash AS hash, photos.photo_uid").
			Joins("JOIN photos ON photos.id = files.photo_id AND photos.deleted_at IS NULL").
			Where("files.file_hash <> '' AND files.deleted_at IS NULL").Order("files.id").Limit(1).Scan(&cover).Error)
		var thumb string
		require.NoError(t, entity.Db().Table("albums").Select("thumb").Where("album_uid = ?", "as6sg6bxpogaaba7").Row().Scan(&thumb))
		require.NoError(t, entity.Db().Model(&entity.Album{}).Where("album_uid = ?", "as6sg6bxpogaaba7").UpdateColumn("thumb", cover.Hash).Error)
		t.Cleanup(func() {
			_ = entity.Db().Model(&entity.Album{}).Where("album_uid = ?", "as6sg6bxpogaaba7").UpdateColumn("thumb", thumb).Error
		})

		_, out, err := listCollections(context.Background(), sessionRequest(admin), ListCollectionsInput{Limit: 50})
		require.NoError(t, err)

		found := false
		for _, c := range out.Items {
			if c.UID == "as6sg6bxpogaaba7" {
				found = true
				assert.Equal(t, cover.PhotoUID, c.Cover)
			}
		}
		assert.True(t, found)

		_, one, err := getCollection(context.Background(), sessionRequest(admin), GetCollectionInput{UID: "as6sg6bxpogaaba7"})
		require.NoError(t, err)
		assert.Equal(t, cover.PhotoUID, one.Cover)
	})
	t.Run("Invalid", func(t *testing.T) {
		for name, input := range map[string]ListCollectionsInput{
			"type must be one of":          {Type: "x"},
			"limit must be greater than 0": {Limit: -1},
			"offset must be 0 or greater":  {Offset: -1},
			"query must not exceed 200":    {Query: string(make([]byte, 201))},
		} {
			_, _, err := listCollections(context.Background(), sessionRequest(admin), input)
			require.Error(t, err, name)
			assert.Contains(t, err.Error(), name)
		}

		_, _, err := listCollections(context.Background(), &sdkmcp.CallToolRequest{}, ListCollectionsInput{})
		assert.Error(t, err)
	})
}

func TestGetCollection(t *testing.T) {
	admin := entity.SessionFixtures.Pointer("alice")

	t.Run("Types", func(t *testing.T) {
		for _, albumType := range collectionTypes {
			var uid string
			require.NoError(t, entity.Db().Table("albums").Select("album_uid").Where("album_type = ? AND deleted_at IS NULL", albumType).
				Order("id").Limit(1).Row().Scan(&uid), albumType)

			_, out, err := getCollection(context.Background(), sessionRequest(admin), GetCollectionInput{UID: uid})
			require.NoError(t, err, albumType)
			assert.Equal(t, uid, out.UID)
			assert.Equal(t, albumType, out.Type)
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		_, _, err := getCollection(context.Background(), sessionRequest(admin), GetCollectionInput{UID: "as6sg6bxpogaab99"})
		require.Error(t, err)
		assert.Equal(t, "collection as6sg6bxpogaab99 not found; call list_collections to find collections", err.Error())
	})
	t.Run("NotVisible", func(t *testing.T) {
		// A collection of a type the role may not see reads like one that does not exist.
		var uid string
		require.NoError(t, entity.Db().Table("albums").Select("album_uid").Where("album_type = ? AND deleted_at IS NULL", entity.AlbumMoment).
			Order("id").Limit(1).Row().Scan(&uid))

		acl.RulesMutex.Lock()
		grant := acl.Rules[acl.ResourceMoments][acl.RoleClient]
		delete(acl.Rules[acl.ResourceMoments], acl.RoleClient)
		acl.RulesMutex.Unlock()
		t.Cleanup(func() {
			acl.RulesMutex.Lock()
			acl.Rules[acl.ResourceMoments][acl.RoleClient] = grant
			acl.RulesMutex.Unlock()
		})

		_, _, err := getCollection(context.Background(), sessionRequest(collectionSession("*")), GetCollectionInput{UID: uid})
		require.Error(t, err)
		assert.Equal(t, "collection "+uid+" not found; call list_collections to find collections", err.Error())
	})
	t.Run("TypeScope", func(t *testing.T) {
		var uid string
		require.NoError(t, entity.Db().Table("albums").Select("album_uid").Where("album_type = ? AND deleted_at IS NULL", entity.AlbumFolder).
			Order("id").Limit(1).Row().Scan(&uid))

		_, _, err := getCollection(context.Background(), sessionRequest(collectionSession("mcp albums")), GetCollectionInput{UID: uid})
		require.Error(t, err)
		assert.Equal(t, "collection "+uid+" not found; call list_collections to find collections", err.Error())

		_, out, err := getCollection(context.Background(), sessionRequest(collectionSession("mcp albums folders")), GetCollectionInput{UID: uid})
		require.NoError(t, err)
		assert.Equal(t, uid, out.UID)
	})
	t.Run("Invalid", func(t *testing.T) {
		_, _, err := getCollection(context.Background(), sessionRequest(admin), GetCollectionInput{UID: "ps6sg6be2lvl0y11"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "uid must be a collection UID")

		_, _, err = getCollection(context.Background(), &sdkmcp.CallToolRequest{}, GetCollectionInput{UID: "as6sg6bxpogaaba7"})
		assert.Error(t, err)
	})
}

func TestErrCollectionNotFound(t *testing.T) {
	assert.Equal(t, "collection as6sg6bxpogaab99 not found; call list_collections to find collections", errCollectionNotFound("as6sg6bxpogaab99").Error())
}

func TestRegisterCollectionTools(t *testing.T) {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test"}, nil)
	registerCollectionTools(server)

	access, ok := ToolDeclaration("list_collections")
	assert.True(t, ok)
	assert.Equal(t, ToolAccess{Resource: acl.ResourceAlbums, Action: acl.ActionSearch}, access)

	access, ok = ToolDeclaration("get_collection")
	assert.True(t, ok)
	assert.Equal(t, ToolAccess{Resource: acl.ResourceAlbums, Action: acl.ActionView}, access)

	assert.ElementsMatch(t, []string{"list_collections", "get_collection"}, toolNames(t, connectServer(t, server)))
}

func TestCollectionTypeScope(t *testing.T) {
	scope, ok := collectionTypeScope(entity.SessionFixtures.Pointer("alice"), entity.AlbumFolder)
	assert.Equal(t, "folders", scope)
	assert.True(t, ok)

	scope, ok = collectionTypeScope(collectionSession("mcp albums"), entity.AlbumMonth)
	assert.Equal(t, "calendar", scope)
	assert.False(t, ok)

	_, ok = collectionTypeScope(collectionSession("mcp albums"), entity.AlbumManual)
	assert.True(t, ok)
	_, ok = collectionTypeScope(collectionSession("read"), entity.AlbumState)
	assert.True(t, ok)
}

func TestPlainQuery(t *testing.T) {
	assert.Equal(t, "", plainQuery(" "))
	assert.Equal(t, `"type:folder"`, plainQuery("type:folder"))
	assert.Equal(t, `"Trip: Paris"`, plainQuery(` "Trip: Paris" `))

	frm := form.SearchAlbums{Type: entity.AlbumManual, Query: plainQuery("type:folder uid:as6sg6bxpogaaba7")}
	require.NoError(t, frm.ParseQueryString())
	assert.Equal(t, entity.AlbumManual, frm.Type)
	assert.Empty(t, frm.UID)
}
