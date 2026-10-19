package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/config/customize"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/entity/sortby"
)

// testRows returns search result rows for media with the specified numbers of rows, where media i has ID i+1.
func testRows(rowsPerMedia ...int) search.PhotoResults {
	var rows search.PhotoResults

	for i, n := range rowsPerMedia {
		for j := 0; j < n; j++ {
			rows = append(rows, search.Photo{ID: uint(i + 1), PhotoUID: "p" + string(rune('a'+i))})
		}
	}

	return rows
}

// testFetch returns a fetchRows function that reads from the rows and counts the queries.
func testFetch(rows search.PhotoResults, queries *int) fetchRows {
	return func(offset, count int) (search.PhotoResults, error) {
		*queries++

		if offset >= len(rows) {
			return search.PhotoResults{}, nil
		}

		return rows[offset:min(offset+count, len(rows))], nil
	}
}

// mediaIDs returns the media IDs of the rows.
func mediaIDs(rows search.PhotoResults) (ids []uint) {
	for _, row := range rows {
		ids = append(ids, row.ID)
	}

	return ids
}

// allPages pages through the rows and returns the media IDs of every page.
func allPages(t *testing.T, rows search.PhotoResults, limit int) (ids []uint, pages int) {
	t.Helper()

	offset := 0

	for {
		var queries int
		items, next, err := pageMedia(context.Background(), testFetch(rows, &queries), limit, offset)
		require.NoError(t, err)
		pages++

		if next > 0 {
			require.Len(t, items, limit, "a page with more results holds limit media")
			require.Greater(t, next, offset)
		}

		ids = append(ids, mediaIDs(items)...)

		if next == 0 {
			return ids, pages
		}

		offset = next
		require.Less(t, pages, 1000)
	}
}

func TestPageMedia(t *testing.T) {
	t.Run("OneRowEach", func(t *testing.T) {
		var queries int
		rows := testRows(1, 1, 1, 1, 1)
		items, next, err := pageMedia(context.Background(), testFetch(rows, &queries), 2, 0)
		require.NoError(t, err)
		assert.Equal(t, []uint{1, 2}, mediaIDs(items))
		assert.Equal(t, 2, next)
		assert.Equal(t, 1, queries)

		items, next, err = pageMedia(context.Background(), testFetch(rows, &queries), 2, next)
		require.NoError(t, err)
		assert.Equal(t, []uint{3, 4}, mediaIDs(items))
		assert.Equal(t, 4, next)

		items, next, err = pageMedia(context.Background(), testFetch(rows, &queries), 2, next)
		require.NoError(t, err)
		assert.Equal(t, []uint{5}, mediaIDs(items))
		assert.Equal(t, 0, next)
	})
	t.Run("MultiRowMedia", func(t *testing.T) {
		rows := testRows(2, 1, 3, 1, 2, 2, 1)
		items, next, err := pageMedia(context.Background(), testFetch(rows, new(int)), 3, 0)
		require.NoError(t, err)
		assert.Equal(t, []uint{1, 2, 3}, mediaIDs(items))
		assert.Equal(t, 6, next)
	})
	t.Run("ConcatenatedPages", func(t *testing.T) {
		rows := testRows(2, 1, 3, 1, 2, 2, 1, 5, 1, 1, 4, 1)
		expected := []uint{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

		for limit := 1; limit <= 13; limit++ {
			ids, _ := allPages(t, rows, limit)
			assert.Equal(t, expected, ids, "limit %d", limit)
		}
	})
	t.Run("MediaLargerThanWindow", func(t *testing.T) {
		// The window of limit 2 is 6 rows, so media 2 spans several windows.
		rows := testRows(1, 20, 1, 1)
		var queries int
		items, next, err := pageMedia(context.Background(), testFetch(rows, &queries), 2, 0)
		require.NoError(t, err)
		assert.Equal(t, []uint{1, 2}, mediaIDs(items))
		assert.Equal(t, 21, next)
		assert.Greater(t, queries, 1)

		ids, _ := allPages(t, rows, 2)
		assert.Equal(t, []uint{1, 2, 3, 4}, ids)
	})
	t.Run("OffsetInsideMedia", func(t *testing.T) {
		// An offset not taken from next_offset skips the rest of the media it falls into.
		rows := testRows(1, 20, 1, 1)
		items, next, err := pageMedia(context.Background(), testFetch(rows, new(int)), 1, 5)
		require.NoError(t, err)
		assert.Equal(t, []uint{3}, mediaIDs(items))
		assert.Equal(t, 22, next)

		items, next, err = pageMedia(context.Background(), testFetch(rows, new(int)), 5, 20)
		require.NoError(t, err)
		assert.Equal(t, []uint{3, 4}, mediaIDs(items))
		assert.Equal(t, 0, next)
	})
	t.Run("OffsetInsideLastMedia", func(t *testing.T) {
		rows := testRows(1, 20)
		items, next, err := pageMedia(context.Background(), testFetch(rows, new(int)), 2, 10)
		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, 0, next)
	})
	t.Run("LastPageBoundary", func(t *testing.T) {
		// The window of limit 2 is 6 rows plus one, so the results end exactly at the end of the read.
		rows := testRows(3, 4)
		var queries int
		items, next, err := pageMedia(context.Background(), testFetch(rows, &queries), 2, 0)
		require.NoError(t, err)
		assert.Equal(t, []uint{1, 2}, mediaIDs(items))
		assert.Equal(t, 0, next)

		// One more media follows.
		rows = testRows(3, 4, 1)
		items, next, err = pageMedia(context.Background(), testFetch(rows, &queries), 2, 0)
		require.NoError(t, err)
		assert.Equal(t, []uint{1, 2}, mediaIDs(items))
		assert.Equal(t, 7, next)

		// The same with an offset, which reads one row before it.
		rows = testRows(1, 3, 3)
		items, next, err = pageMedia(context.Background(), testFetch(rows, &queries), 2, 1)
		require.NoError(t, err)
		assert.Equal(t, []uint{2, 3}, mediaIDs(items))
		assert.Equal(t, 0, next)

		rows = testRows(1, 3, 3, 2)
		items, next, err = pageMedia(context.Background(), testFetch(rows, &queries), 2, 1)
		require.NoError(t, err)
		assert.Equal(t, []uint{2, 3}, mediaIDs(items))
		assert.Equal(t, 7, next)
	})
	t.Run("OffsetPastEnd", func(t *testing.T) {
		items, next, err := pageMedia(context.Background(), testFetch(testRows(1, 1), new(int)), 5, 10)
		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, 0, next)
	})
	t.Run("NoResults", func(t *testing.T) {
		items, next, err := pageMedia(context.Background(), testFetch(nil, new(int)), 5, 0)
		require.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, 0, next)
	})
	t.Run("RepeatedMedia", func(t *testing.T) {
		// A media whose rows are not adjacent is returned once per page.
		rows := search.PhotoResults{{ID: 1}, {ID: 2}, {ID: 1}, {ID: 3}}
		items, next, err := pageMedia(context.Background(), testFetch(rows, new(int)), 5, 0)
		require.NoError(t, err)
		assert.Equal(t, []uint{1, 2, 3}, mediaIDs(items))
		assert.Equal(t, 0, next)
	})
	t.Run("GrowingWindows", func(t *testing.T) {
		// Further windows of one call double in size, so a media with many rows takes few queries.
		rows := testRows(1, 10000, 1)
		var queries int
		items, next, err := pageMedia(context.Background(), testFetch(rows, &queries), 1, 0)
		require.NoError(t, err)
		assert.Equal(t, []uint{1}, mediaIDs(items))
		assert.Equal(t, 1, next)

		queries = 0
		items, next, err = pageMedia(context.Background(), testFetch(rows, &queries), 1, next)
		require.NoError(t, err)
		assert.Equal(t, []uint{2}, mediaIDs(items))
		assert.Equal(t, 10001, next)
		assert.LessOrEqual(t, queries, 12)

		var counts []int
		fetch := func(offset, count int) (search.PhotoResults, error) {
			counts = append(counts, count)
			return testFetch(rows, new(int))(offset, count)
		}
		_, _, _ = pageMedia(context.Background(), fetch, 1, 1)
		assert.Equal(t, []int{5, 7, 13, 25, 49, 97, 193, 385, 769, 1537, 3073, 5001}, counts[:12])
	})
	t.Run("Canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, _, err := pageMedia(ctx, testFetch(testRows(1, 1), new(int)), 5, 0)
		assert.ErrorIs(t, err, context.Canceled)
	})
	t.Run("Error", func(t *testing.T) {
		_, _, err := pageMedia(context.Background(), func(int, int) (search.PhotoResults, error) { return nil, errors.New("failed") }, 5, 0)
		assert.Error(t, err)
	})
	t.Run("Window", func(t *testing.T) {
		var counts []int
		fetch := func(offset, count int) (search.PhotoResults, error) {
			counts = append(counts, count)
			return search.PhotoResults{}, nil
		}
		_, _, _ = pageMedia(context.Background(), fetch, 20, 0)
		_, _, _ = pageMedia(context.Background(), fetch, 50, 10)
		assert.Equal(t, []int{61, 152}, counts)
	})
}

func TestValidateOrder(t *testing.T) {
	t.Run("Default", func(t *testing.T) {
		order, err := validateOrder("")
		require.NoError(t, err)
		assert.Equal(t, sortby.Newest, order)
	})
	t.Run("Accepted", func(t *testing.T) {
		for _, o := range []string{"newest", "oldest", "added", "edited", "relevance", "Oldest "} {
			_, err := validateOrder(o)
			assert.NoError(t, err, o)
		}
	})
	t.Run("Rejected", func(t *testing.T) {
		for _, o := range []string{"size", "similar", "random", "name", "title", "duration"} {
			_, err := validateOrder(o)
			require.Error(t, err, o)
			assert.Equal(t, "order must be one of newest, oldest, added, edited, relevance", err.Error())
		}
	})
}

func TestValidateMediaType(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		v, err := validateMediaType(" ")
		require.NoError(t, err)
		assert.Empty(t, v)
	})
	t.Run("Combined", func(t *testing.T) {
		v, err := validateMediaType("Video | live|video")
		require.NoError(t, err)
		assert.Equal(t, "video|live", v)
	})
	t.Run("Unknown", func(t *testing.T) {
		_, err := validateMediaType("sidecar")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "type must be one of image, raw, live, video")
	})
}

func TestValidateDate(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		d, err := validateDate("taken_after", "")
		require.NoError(t, err)
		assert.True(t, d.IsZero())
	})
	t.Run("Date", func(t *testing.T) {
		d, err := validateDate("taken_after", "2019-07-01")
		require.NoError(t, err)
		assert.Equal(t, "2019-07-01T00:00:00Z", d.Format("2006-01-02T15:04:05Z07:00"))
	})
	t.Run("Invalid", func(t *testing.T) {
		for _, value := range []string{"yesterday", "2019-07-01T12:00:00+02:00", "2019"} {
			_, err := validateDate("taken_before", value)
			require.Error(t, err, value)
			assert.Equal(t, "taken_before must be a date in the form YYYY-MM-DD", err.Error())
		}
	})
}

func TestNewSearchMediaRequest(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		r, noMatch, err := newSearchMediaRequest(SearchMediaInput{}, nil, customize.FeatureSettings{})
		require.NoError(t, err)
		assert.False(t, noMatch)
		assert.Equal(t, 20, r.Limit)
		assert.Equal(t, 0, r.Offset)
		assert.Equal(t, sortby.Newest, r.Form.Order)
		assert.False(t, r.Form.Merged)
		assert.False(t, r.Form.Primary)
		assert.False(t, r.Full)
	})
	t.Run("StructuredInputsTakePrecedence", func(t *testing.T) {
		r, _, err := newSearchMediaRequest(SearchMediaInput{
			Query:       "Beach subject:js6sg6b1qekk9jx8 type:raw after:2001-01-01",
			Type:        "image",
			Person:      "JS6SG6B1H1NJAAAA",
			Collection:  "as6sg6bxpogaaba7",
			TakenAfter:  "2019-01-01",
			TakenBefore: "2020-01-01",
			Order:       "oldest",
			Limit:       60,
			Offset:      7,
			Detail:      "full",
		}, nil, customize.FeatureSettings{})
		require.NoError(t, err)
		assert.Equal(t, "beach", r.Form.Query)
		assert.Equal(t, "image", r.Form.Type)
		assert.Equal(t, "js6sg6b1h1njaaaa", r.Form.Subject)
		assert.Equal(t, "as6sg6bxpogaaba7", r.Form.Scope)
		assert.Equal(t, "2019-01-01", r.Form.After.Format("2006-01-02"))
		assert.Equal(t, "2020-01-01", r.Form.Before.Format("2006-01-02"))
		assert.Equal(t, sortby.Oldest, r.Form.Order)
		assert.Equal(t, 50, r.Limit)
		assert.Equal(t, 7, r.Offset)
		assert.True(t, r.Full)
	})
	t.Run("QueryFilters", func(t *testing.T) {
		r, _, err := newSearchMediaRequest(SearchMediaInput{Query: "codec:avc1 primary:true"}, nil, customize.FeatureSettings{})
		require.NoError(t, err)
		assert.Equal(t, "avc1", r.Form.Codec)
		assert.True(t, r.Form.Primary)
	})
	t.Run("QuotedColon", func(t *testing.T) {
		// The search parses the remaining text again, so it must stay quoted.
		r, _, err := newSearchMediaRequest(SearchMediaInput{Query: `"12:30"`}, nil, customize.FeatureSettings{})
		require.NoError(t, err)
		assert.Equal(t, `"12:30"`, r.Form.Query)
		require.NoError(t, r.Form.ParseQueryString())
		assert.Equal(t, "12:30", r.Form.Query)

		r, _, err = newSearchMediaRequest(SearchMediaInput{Query: `beach "foo:bar"`}, nil, customize.FeatureSettings{})
		require.NoError(t, err)
		require.NoError(t, r.Form.ParseQueryString())
		assert.Equal(t, "beach foo:bar", r.Form.Query)
	})
	t.Run("NoMatch", func(t *testing.T) {
		_, noMatch, err := newSearchMediaRequest(SearchMediaInput{Query: "label:\"\x00\""}, nil, customize.FeatureSettings{})
		require.NoError(t, err)
		assert.True(t, noMatch)
	})
	t.Run("Features", func(t *testing.T) {
		guest := &entity.Session{}
		guest.SetUser(&entity.User{ID: 1234567, UserUID: "urqdrfb72479n047", UserName: "mcp-test", UserRole: "guest"})
		features := customize.FeatureSettings{Private: false, Review: true}

		r, _, err := newSearchMediaRequest(SearchMediaInput{}, guest, features)
		require.NoError(t, err)
		assert.Equal(t, 3, r.Form.Quality)
		r, _, err = newSearchMediaRequest(SearchMediaInput{Collection: "as6sg6bxpogaaba7"}, guest, features)
		require.NoError(t, err)
		assert.Equal(t, 0, r.Form.Quality)
	})
	t.Run("Invalid", func(t *testing.T) {
		for name, input := range map[string]SearchMediaInput{
			"limit must be greater than 0":    {Limit: -1},
			"offset must be 0 or greater":     {Offset: -1},
			"collection must be a collection": {Collection: "pt9jtdre2lvl0yh7"},
			"person must be a person UID":     {Person: "jane"},
			"type must be one of":             {Type: "photo"},
			"taken_after must be a date":      {TakenAfter: "2019"},
			"order must be one of":            {Order: "size"},
			"detail must be one of brief":     {Detail: "all"},
			"query must not exceed 200":       {Query: string(make([]byte, 201))},
		} {
			_, _, err := newSearchMediaRequest(input, nil, customize.FeatureSettings{})
			require.Error(t, err, name)
			assert.Contains(t, err.Error(), name)
		}
	})
}
