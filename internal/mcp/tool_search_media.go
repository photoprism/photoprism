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
	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/media"
	"github.com/photoprism/photoprism/pkg/rnd"
)

// SearchMediaOrders lists the sort orders search_media accepts, each of which keeps the result rows of a media
// adjacent, which paging relies on.
var SearchMediaOrders = []string{sortby.Newest, sortby.Oldest, sortby.Added, sortby.Edited, sortby.Relevance}

// searchMediaTypes lists the media types search_media accepts.
var searchMediaTypes = []string{
	media.Image.String(), media.Raw.String(), media.Live.String(), media.Video.String(),
	media.Animated.String(), media.Audio.String(), media.Vector.String(), media.Document.String(),
}

// Result rows read per query: the first window holds at most searchMediaMaxWindow rows, and further windows of
// one call double in size up to searchMediaMaxReadOn, so a media with many files takes few queries.
const (
	searchMediaMaxWindow = 150
	searchMediaMaxReadOn = 5000
)

// personUID is the prefix of person UIDs.
const personUID = byte('j')

// searchMediaAccess is the declaration of search_media.
var searchMediaAccess = ToolAccess{Resource: acl.ResourcePhotos, Action: acl.ActionSearch}

// SearchMediaInput defines the inputs of the search_media tool.
type SearchMediaInput struct {
	Query       string `json:"query,omitempty" jsonschema:"optional search filters, e.g. label:beach year:2019; call find_search_filters for the syntax"`
	Type        string `json:"type,omitempty" jsonschema:"optional media type: image, raw, live, video, animated, audio, vector, or document; combine with |"`
	Collection  string `json:"collection,omitempty" jsonschema:"optional collection UID from list_collections; finds the media in that collection"`
	Person      string `json:"person,omitempty" jsonschema:"optional person UID from list_people"`
	TakenAfter  string `json:"taken_after,omitempty" jsonschema:"optional date (YYYY-MM-DD); finds media taken on or after it"`
	TakenBefore string `json:"taken_before,omitempty" jsonschema:"optional date (YYYY-MM-DD); finds media taken before it"`
	Order       string `json:"order,omitempty" jsonschema:"optional sort order: newest (default), oldest, added, edited, or relevance"`
	Limit       int    `json:"limit,omitempty" jsonschema:"optional maximum number of media to return (default 20, max 50)"`
	Offset      int    `json:"offset,omitempty" jsonschema:"optional position to continue from; pass next_offset of the previous result"`
	Detail      string `json:"detail,omitempty" jsonschema:"optional detail level: brief (default) or full, which adds caption, camera, lens, labels, people, and collections"`
}

// SearchMediaOutput defines the structured output of the search_media tool.
type SearchMediaOutput struct {
	Items      []Media  `json:"items"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	NextOffset int      `json:"next_offset,omitempty"`
	Truncated  bool     `json:"truncated"`
	Warnings   []string `json:"warnings,omitempty"`
}

// searchMediaRequest contains the validated inputs of a search_media call.
type searchMediaRequest struct {
	Form   form.SearchPhotos
	Limit  int
	Offset int
	Full   bool
}

// searchMedia returns a page of the media the session finds with the web app's search.
func searchMedia(ctx context.Context, req *sdkmcp.CallToolRequest, input SearchMediaInput, opt Options) (*sdkmcp.CallToolResult, SearchMediaOutput, error) {
	s := RequestSession(req)

	if s == nil {
		return nil, SearchMediaOutput{}, errors.New("search_media requires an authenticated session")
	}

	r, noMatch, err := newSearchMediaRequest(input, s, opt.features())

	if err != nil {
		return nil, SearchMediaOutput{}, err
	}

	result := SearchMediaOutput{Items: []Media{}, Limit: r.Limit, Offset: r.Offset}

	if noMatch {
		return nil, result, nil
	}

	rows, next, err := pageMedia(ctx, func(offset, count int) (search.PhotoResults, error) {
		frm := r.Form
		frm.Offset = offset
		frm.Count = count
		results, _, searchErr := search.UserPhotos(frm, s)
		return results, searchErr
	}, r.Limit, r.Offset)

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil, SearchMediaOutput{}, err
	} else if err != nil {
		return nil, SearchMediaOutput{}, searchMediaError(err)
	}

	result.Items = newMediaList(rows)

	if r.Full {
		addMediaDetails(result.Items, rows, s)
	}

	if next > 0 {
		result.NextOffset = next
		result.Truncated = true
		result.Warnings = []string{fmt.Sprintf("returned %d media starting at offset %d; call again with offset %d to retrieve more", len(result.Items), r.Offset, next)}
	}

	return nil, result, nil
}

// newSearchMediaRequest validates the inputs and returns the search form, or reports that the query can match
// nothing. It applies the feature settings like the web app's search; the other structured inputs take precedence
// over the same filters in the query.
func newSearchMediaRequest(input SearchMediaInput, s *entity.Session, features customize.FeatureSettings) (r searchMediaRequest, noMatch bool, err error) {
	if r.Limit, err = validateLimit(input.Limit); err != nil {
		return r, false, err
	} else if r.Offset, err = validateOffset(input.Offset); err != nil {
		return r, false, err
	}

	query, err := validateString("query", input.Query)

	if err != nil {
		return r, false, err
	}

	// Keep the case of the query, which validateString lowers for matching.
	if query != "" {
		query = strings.TrimSpace(input.Query)
	}

	r.Form = form.SearchPhotos{Query: query}

	if input.Collection != "" {
		if !rnd.IsUID(strings.ToLower(input.Collection), entity.AlbumUID) {
			return r, false, fmt.Errorf("collection must be a collection UID; call list_collections to find collections")
		}

		r.Form.Scope = strings.ToLower(input.Collection)
	}

	search.ApplyPhotoFeatures(&r.Form, s, features.Private, features.Review)

	if parseErr := r.Form.ParseQueryString(); errors.Is(parseErr, form.ErrNoMatch) {
		noMatch = true
	} else if parseErr != nil {
		return r, false, fmt.Errorf("query could not be parsed; call find_search_filters for the syntax")
	}

	// The search parses the remaining text again, so text with a colon, which was quoted, is quoted again.
	if strings.Contains(r.Form.Query, ":") && !strings.Contains(r.Form.Query, "\"") {
		r.Form.Query = "\"" + r.Form.Query + "\""
	}

	if mediaType, typeErr := validateMediaType(input.Type); typeErr != nil {
		return r, false, typeErr
	} else if mediaType != "" {
		r.Form.Type = mediaType
	}

	if input.Person != "" {
		if !rnd.IsUID(strings.ToLower(input.Person), personUID) {
			return r, false, fmt.Errorf("person must be a person UID; call list_people to find people")
		}

		r.Form.Subject = strings.ToLower(input.Person)
		r.Form.Person = ""
	}

	if after, dateErr := validateDate("taken_after", input.TakenAfter); dateErr != nil {
		return r, false, dateErr
	} else if !after.IsZero() {
		r.Form.After = after
	}

	if before, dateErr := validateDate("taken_before", input.TakenBefore); dateErr != nil {
		return r, false, dateErr
	} else if !before.IsZero() {
		r.Form.Before = before
	}

	if r.Form.Order, err = validateOrder(input.Order); err != nil {
		return r, false, err
	}

	switch strings.ToLower(strings.TrimSpace(input.Detail)) {
	case "", "brief":
	case "full":
		r.Full = true
	default:
		return r, false, fmt.Errorf("detail must be one of brief or full")
	}

	// The tool reads the result rows itself to page by media.
	r.Form.Merged = false

	return r, noMatch, nil
}

// validateMediaType normalizes one or more media types separated by |.
func validateMediaType(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	if value == "" {
		return "", nil
	}

	var types []string

	for _, t := range strings.Split(value, "|") {
		if t = strings.TrimSpace(t); !slices.Contains(searchMediaTypes, t) {
			return "", fmt.Errorf("type must be one of %s", strings.Join(searchMediaTypes, ", "))
		} else if !slices.Contains(types, t) {
			types = append(types, t)
		}
	}

	return strings.Join(types, "|"), nil
}

// validateDate parses a date in the form YYYY-MM-DD, which the search compares without a time of day.
func validateDate(field, value string) (time.Time, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return time.Time{}, nil
	} else if t, err := time.Parse(time.DateOnly, value); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("%s must be a date in the form YYYY-MM-DD", field)
}

// validateOrder returns the sort order, newest by default.
func validateOrder(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))

	if value == "" {
		return sortby.Newest, nil
	} else if slices.Contains(SearchMediaOrders, value) {
		return value, nil
	}

	return "", fmt.Errorf("order must be one of %s", strings.Join(SearchMediaOrders, ", "))
}

// searchMediaError returns the error message of a failed search, which names no internal details.
func searchMediaError(err error) error {
	switch {
	case errors.Is(err, search.ErrForbidden), errors.Is(err, search.ErrInvalidId), errors.Is(err, search.ErrNotFound):
		return errors.New("the search found no accessible media for this collection or filter")
	case errors.Is(err, search.ErrBadSortOrder):
		return fmt.Errorf("order must be one of %s", strings.Join(SearchMediaOrders, ", "))
	default:
		return errors.New("the search could not be completed; check the query with find_search_filters")
	}
}

// fetchRows returns the search result rows starting at the specified row position.
type fetchRows func(offset, count int) (search.PhotoResults, error)

// pageMedia returns up to limit media, one row each, starting at the row position offset, and the row position of
// the first row of the next media, or 0 if the results end. It reads one row before the position, skips the
// leading rows of a media that began before it, and reads further windows until it has limit media and has seen
// where the next one begins. The rows of a media must be adjacent, which the orders in SearchMediaOrders ensure.
func pageMedia(ctx context.Context, fetch fetchRows, limit, offset int) (items search.PhotoResults, next int, err error) {
	window := min(3*limit, searchMediaMaxWindow)
	items = make(search.PhotoResults, 0, limit)
	seen := make(map[uint]bool, limit)

	var last uint
	var skip uint

	readFrom := offset
	count := window + 1

	// Read one row before the position to recognize a media that began before it.
	if offset > 0 {
		readFrom = offset - 1
		count++
	}

	for {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return items, 0, ctxErr
		}

		rows, fetchErr := fetch(readFrom, count)

		if fetchErr != nil {
			return items, 0, fetchErr
		}

		for i, row := range rows {
			pos := readFrom + i

			switch {
			case pos < offset:
				skip = row.ID
				continue
			case skip != 0 && row.ID == skip:
				continue
			case row.ID == last:
				continue
			}

			skip = 0

			if len(items) == limit {
				return items, pos, nil
			}

			last = row.ID

			if seen[row.ID] {
				continue
			}

			seen[row.ID] = true
			items = append(items, row)
		}

		if len(rows) < count {
			return items, 0, nil
		}

		readFrom += len(rows)
		window = min(2*window, searchMediaMaxReadOn)
		count = window + 1
	}
}

// registerSearchMedia registers the search_media tool.
func registerSearchMedia(server *sdkmcp.Server, opt Options) {
	AddTool(server, searchMediaAccess, &sdkmcp.Tool{
		Name: "search_media",
		Description: "Searches the library with the web app's search filters, e.g. label:beach year:2019 (call find_search_filters " +
			"for the syntax), and returns the media the account may see, one item per picture or video. Narrow the search " +
			"with type, collection (a UID from list_collections), person (a UID from list_people), and taken_after or " +
			"taken_before. Returns at most limit items (default 20, max 50); to continue, pass next_offset as offset.",
		Annotations: readOnly(),
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, input SearchMediaInput) (*sdkmcp.CallToolResult, SearchMediaOutput, error) {
		return searchMedia(ctx, req, input, opt)
	})
}
