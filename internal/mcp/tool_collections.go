package mcp

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// Declarations of the collection tools.
var (
	listCollectionsAccess = ToolAccess{Resource: acl.ResourceAlbums, Action: acl.ActionSearch}
	getCollectionAccess   = ToolAccess{Resource: acl.ResourceAlbums, Action: acl.ActionView}
)

// collectionTypes lists the collection types in the order list_collections returns them with type all.
var collectionTypes = []string{entity.AlbumManual, entity.AlbumFolder, entity.AlbumMoment, entity.AlbumMonth, entity.AlbumState}

// Collection is a group of media, as returned by the collection tools.
type Collection struct {
	UID         string `json:"uid"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	MediaCount  int    `json:"media_count"`
	Cover       string `json:"cover,omitempty"`
	Year        int    `json:"year,omitempty"`
	Month       int    `json:"month,omitempty"`
	Day         int    `json:"day,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ListCollectionsInput defines the inputs of the list_collections tool.
type ListCollectionsInput struct {
	Type   string `json:"type,omitempty" jsonschema:"optional collection type: album (default), folder, moment, month, state, or all"`
	Query  string `json:"query,omitempty" jsonschema:"optional search in titles and locations"`
	Limit  int    `json:"limit,omitempty" jsonschema:"optional maximum number of collections to return (default 20, max 50)"`
	Offset int    `json:"offset,omitempty" jsonschema:"optional number of collections to skip"`
}

// ListCollectionsOutput defines the structured output of the list_collections tool.
type ListCollectionsOutput struct {
	Items     []Collection `json:"items"`
	Total     int          `json:"total"`
	Limit     int          `json:"limit"`
	Offset    int          `json:"offset"`
	Truncated bool         `json:"truncated"`
	Warnings  []string     `json:"warnings,omitempty"`
}

// GetCollectionInput defines the inputs of the get_collection tool.
type GetCollectionInput struct {
	UID string `json:"uid" jsonschema:"collection UID from list_collections"`
}

// newCollection returns a collection from an album search result.
func newCollection(a search.Album) Collection {
	c := Collection{
		UID:         a.AlbumUID,
		Type:        a.AlbumType,
		Title:       a.AlbumTitle,
		Description: a.AlbumDescription,
		MediaCount:  a.PhotoCount,
		CreatedAt:   a.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   a.UpdatedAt.UTC().Format(time.RFC3339),
	}

	// Months and moments are defined by their dates.
	if a.AlbumType == entity.AlbumMonth || a.AlbumType == entity.AlbumMoment {
		c.Year = max(a.AlbumYear, 0)
		c.Month = max(a.AlbumMonth, 0)
		c.Day = max(a.AlbumDay, 0)
	}

	return c
}

// addCollectionCovers sets the cover of each collection to the UID of its cover picture, if the session may see it.
func addCollectionCovers(items []Collection, albums search.AlbumResults, s *entity.Session) {
	hashes := make([]string, len(albums))

	for i, a := range albums {
		hashes[i] = a.Thumb
	}

	covers := coverUIDs(hashes, s)

	for i, a := range albums {
		items[i].Cover = covers[a.Thumb]
	}
}

// validateCollectionType returns the collection types to list, album by default.
func validateCollectionType(value string) ([]string, error) {
	switch value = strings.ToLower(strings.TrimSpace(value)); value {
	case "":
		return []string{entity.AlbumManual}, nil
	case "all":
		return collectionTypes, nil
	default:
		if slices.Contains(collectionTypes, value) {
			return []string{value}, nil
		}
	}

	return nil, fmt.Errorf("type must be one of %s, or all", strings.Join(collectionTypes, ", "))
}

// collectionTypeScope returns the token scope of a collection type, and whether the session's scope permits it;
// sessions without a scope are not restricted by one.
func collectionTypeScope(s *entity.Session, albumType string) (string, bool) {
	resource := search.AlbumTypeResource(albumType)

	if !s.IsClient() && !s.HasScope() {
		return resource.String(), true
	}

	return resource.String(), s.ValidateScope(resource, acl.Permissions{acl.ActionSearch})
}

// plainQuery quotes a free-text query, so the search reads it as text and not as filters.
func plainQuery(query string) string {
	if query = strings.TrimSpace(strings.ReplaceAll(query, "\"", "")); query == "" {
		return ""
	}

	return "\"" + query + "\""
}

// listCollections returns the collections the session may see, like GET /api/v1/albums for each type.
func listCollections(_ context.Context, req *sdkmcp.CallToolRequest, input ListCollectionsInput) (*sdkmcp.CallToolResult, ListCollectionsOutput, error) {
	s := RequestSession(req)

	if s == nil {
		return nil, ListCollectionsOutput{}, errors.New("list_collections requires an authenticated session")
	}

	limit, err := validateLimit(input.Limit)

	if err != nil {
		return nil, ListCollectionsOutput{}, err
	}

	offset, err := validateOffset(input.Offset)

	if err != nil {
		return nil, ListCollectionsOutput{}, err
	}

	if _, err = validateString("query", input.Query); err != nil {
		return nil, ListCollectionsOutput{}, err
	}

	types, err := validateCollectionType(input.Type)

	if err != nil {
		return nil, ListCollectionsOutput{}, err
	}

	albums := search.AlbumResults{}

	var warnings []string

	for _, t := range types {
		// Each type needs the token scope of its own resource; with type all, types without it are left out.
		if scope, ok := collectionTypeScope(s, t); !ok && len(types) > 1 {
			warnings = append(warnings, fmt.Sprintf("collections of type %s need the %s scope", t, scope))
			continue
		} else if !ok {
			return nil, ListCollectionsOutput{}, fmt.Errorf("collections of type %s need the %s scope", t, scope)
		}

		results, searchErr := search.UserAlbums(form.SearchAlbums{Type: t, Query: plainQuery(input.Query), Count: search.MaxResults}, s)

		// With type all, types the role may not see are left out.
		if errors.Is(searchErr, search.ErrForbidden) && len(types) > 1 {
			continue
		} else if searchErr != nil {
			return nil, ListCollectionsOutput{}, fmt.Errorf("collections of type %s could not be listed", t)
		}

		albums = append(albums, results...)
	}

	page, truncated := pageOf(albums, limit, offset)
	items := make([]Collection, len(page))

	for i, a := range page {
		items[i] = newCollection(a)
	}

	addCollectionCovers(items, page, s)

	return nil, ListCollectionsOutput{
		Items:     items,
		Total:     len(albums),
		Limit:     limit,
		Offset:    offset,
		Truncated: truncated,
		Warnings:  append(warnings, pageWarnings(len(items), len(albums), offset, truncated)...),
	}, nil
}

// errCollectionNotFound is returned for a UID that does not exist or that the session cannot see.
func errCollectionNotFound(uid string) error {
	return fmt.Errorf("collection %s not found; call list_collections to find collections", uid)
}

// getCollection returns one collection, checked by the rules of its own type like GET /api/v1/albums.
func getCollection(_ context.Context, req *sdkmcp.CallToolRequest, input GetCollectionInput) (*sdkmcp.CallToolResult, Collection, error) {
	s := RequestSession(req)

	if s == nil {
		return nil, Collection{}, errors.New("get_collection requires an authenticated session")
	}

	uid := strings.ToLower(strings.TrimSpace(input.UID))

	if !rnd.IsUID(uid, entity.AlbumUID) {
		return nil, Collection{}, fmt.Errorf("uid must be a collection UID; call list_collections to find collections")
	}

	album, err := entity.CachedAlbumByUID(uid)

	if err != nil || album.AlbumUID == "" {
		return nil, Collection{}, errCollectionNotFound(uid)
	} else if _, ok := collectionTypeScope(s, album.AlbumType); !ok {
		return nil, Collection{}, errCollectionNotFound(uid)
	}

	results, err := search.UserAlbums(form.SearchAlbums{UID: uid, Type: album.AlbumType, Count: 1}, s)

	if err != nil || len(results) != 1 {
		return nil, Collection{}, errCollectionNotFound(uid)
	}

	items := []Collection{newCollection(results[0])}
	addCollectionCovers(items, results, s)

	return nil, items[0], nil
}

// registerCollectionTools registers the list_collections and get_collection tools.
func registerCollectionTools(server *sdkmcp.Server) {
	AddTool(server, listCollectionsAccess, &sdkmcp.Tool{
		Name: "list_collections",
		Description: "Lists the collections the account may see: albums by default, or folders, moments, months, " +
			"states, or all types. Only albums can be changed; the other types are generated from the library. " +
			"Pass a collection UID to search_media as collection to find its media. query searches titles as text. " +
			"Returns at most limit items (default 20, max 50) starting at offset.",
		Annotations: readOnly(),
	}, listCollections)

	AddTool(server, getCollectionAccess, &sdkmcp.Tool{
		Name: "get_collection",
		Description: "Returns one collection with its title, description, media count, cover media UID, and dates. " +
			"Takes a collection UID from list_collections; search_media with collection returns its media.",
		Annotations: readOnly(),
	}, getCollection)
}
