package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/entity/sortby"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/internal/mcp"
	"github.com/photoprism/photoprism/pkg/authn"
)

// mcpResult returns the JSON-RPC response of a request, which the SDK sends as JSON or as a server-sent event.
func mcpResult(t *testing.T, body string) gjson.Result {
	t.Helper()

	if trimmed := strings.TrimSpace(body); strings.HasPrefix(trimmed, "{") {
		return gjson.Parse(trimmed)
	}

	for _, line := range strings.Split(body, "\n") {
		if data, ok := strings.CutPrefix(line, "data: "); ok && gjson.Valid(data) {
			return gjson.Parse(data)
		}
	}

	t.Fatalf("no JSON-RPC response in %q", body)

	return gjson.Result{}
}

// mcpCallTool calls a tool and returns its structured content, or fails the test if the call failed.
func mcpCallTool(t *testing.T, app http.Handler, token, sessionID, name string, args map[string]any) gjson.Result {
	t.Helper()

	params, err := json.Marshal(map[string]any{"name": name, "arguments": args})
	require.NoError(t, err)

	w := mcpPost(app, `{"jsonrpc":"2.0","id":100,"method":"tools/call","params":`+string(params)+`}`, token, sessionID)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	result := mcpResult(t, w.Body.String())
	require.False(t, result.Get("result.isError").Bool(), result.Raw)

	return result.Get("result.structuredContent")
}

// searchMediaUIDs pages through search_media and returns the UIDs of all media, checking the paging fields.
func searchMediaUIDs(t *testing.T, app http.Handler, token, sessionID string, args map[string]any, limit int) (uids []string) {
	t.Helper()

	offset := 0

	for page := 0; page < 1000; page++ {
		args["limit"] = limit
		args["offset"] = offset
		out := mcpCallTool(t, app, token, sessionID, "search_media", args)
		items := out.Get("items").Array()
		next := out.Get("next_offset")

		for _, item := range items {
			uids = append(uids, item.Get("uid").String())
		}

		if !next.Exists() {
			assert.False(t, out.Get("truncated").Bool())
			return uids
		}

		require.Len(t, items, limit, "a page with more results holds limit media")
		require.Greater(t, int(next.Int()), offset)
		assert.True(t, out.Get("truncated").Bool())
		offset = int(next.Int())
	}

	t.Fatal("too many pages")

	return uids
}

// restMergedUIDs returns the distinct media UIDs, in order, of the merged photo search for the same session.
func restMergedUIDs(t *testing.T, app http.Handler, token, order, query string) (uids []string) {
	t.Helper()

	url := "/api/v1/photos?count=100000&merged=true&order=" + order

	if query != "" {
		url += "&q=" + query
	}

	r := AuthenticatedRequest(app, http.MethodGet, url, token)
	require.Equal(t, http.StatusOK, r.Code, r.Body.String())

	seen := make(map[string]bool)

	for _, photo := range gjson.Parse(r.Body.String()).Array() {
		if uid := photo.Get("UID").String(); !seen[uid] {
			seen[uid] = true
			uids = append(uids, uid)
		}
	}

	return uids
}

func TestServeMCP_SearchMedia(t *testing.T) {
	app, router, _ := NewApiTest()
	prepareMCPTest(t)
	ServeMCP(router)
	SearchPhotos(router)

	token := AuthenticateAdmin(app, router)
	sessionID := mcpInitSession(t, app, token)

	t.Run("SameAsWebApp", func(t *testing.T) {
		for _, order := range mcp.SearchMediaOrders {
			expected := restMergedUIDs(t, app, token, order, "")

			if order == sortby.Edited {
				require.NotEmpty(t, expected, order)
			} else {
				require.Greater(t, len(expected), 20, order)
			}

			for _, limit := range []int{1, 7, 50} {
				assert.Equal(t, expected, searchMediaUIDs(t, app, token, sessionID, map[string]any{"order": order}, limit), "%s %d", order, limit)
			}
		}
	})
	t.Run("FileFilter", func(t *testing.T) {
		expected := restMergedUIDs(t, app, token, "newest", "codec:avc1")
		require.NotEmpty(t, expected)
		assert.Equal(t, expected, searchMediaUIDs(t, app, token, sessionID, map[string]any{"query": "codec:avc1"}, 2))
	})
	t.Run("Output", func(t *testing.T) {
		out := mcpCallTool(t, app, token, sessionID, "search_media", map[string]any{"limit": 3})
		require.Len(t, out.Get("items").Array(), 3)
		item := out.Get("items.0")
		assert.NotEmpty(t, item.Get("uid").String())
		assert.NotEmpty(t, item.Get("type").String())
		assert.True(t, item.Get("taken_at").Exists())
		assert.True(t, item.Get("taken_at_local").Exists())
		assert.True(t, item.Get("favorite").Exists())
		assert.False(t, item.Get("caption").Exists())
		assert.Equal(t, int64(3), out.Get("limit").Int())
		assert.Equal(t, int64(0), out.Get("offset").Int())
		assert.False(t, out.Get("total").Exists())

		// Media carry only the documented fields, so no coordinates, paths, or internal IDs.
		allowed := []string{"uid", "type", "title", "taken_at", "taken_at_local", "place", "favorite",
			"caption", "camera", "lens", "labels", "people", "collections"}
		for _, item := range mcpCallTool(t, app, token, sessionID, "search_media", map[string]any{"limit": 50, "detail": "full"}).Get("items").Array() {
			for key := range item.Map() {
				assert.Contains(t, allowed, key)
			}
			for key := range item.Get("place").Map() {
				assert.Contains(t, []string{"name", "country"}, key)
			}
		}
	})
	t.Run("QuotedColon", func(t *testing.T) {
		expected := restMergedUIDs(t, app, token, "newest", url.QueryEscape(`"12:30"`))
		assert.Equal(t, expected, searchMediaUIDs(t, app, token, sessionID, map[string]any{"query": `"12:30"`}, 5))
	})
	t.Run("FullDetail", func(t *testing.T) {
		// The admin sees the labels, people, and collections of the media.
		var found [3]bool

		for _, item := range mcpCallTool(t, app, token, sessionID, "search_media", map[string]any{"limit": 50, "detail": "full"}).Get("items").Array() {
			found[0] = found[0] || len(item.Get("labels").Array()) > 0
			found[1] = found[1] || len(item.Get("people").Array()) > 0
			found[2] = found[2] || len(item.Get("collections").Array()) > 0
		}

		assert.Equal(t, [3]bool{true, true, true}, found)
	})
	t.Run("Validation", func(t *testing.T) {
		w := mcpPost(app, `{"jsonrpc":"2.0","id":101,"method":"tools/call","params":{"name":"search_media","arguments":{"order":"size"}}}`, token, sessionID)
		require.Equal(t, http.StatusOK, w.Code)
		result := mcpResult(t, w.Body.String())
		assert.True(t, result.Get("result.isError").Bool())
		assert.Contains(t, result.Raw, "order must be one of newest, oldest, added, edited, relevance")
	})
}

// TestSearchMediaOrders checks that each order search_media offers keeps the rows of a media adjacent, which its
// paging relies on.
func TestSearchMediaOrders(t *testing.T) {
	multiRow := false

	for _, order := range mcp.SearchMediaOrders {
		rows, _, err := search.UserPhotos(form.SearchPhotos{Order: order, Count: search.MaxResults}, nil)
		require.NoError(t, err, order)
		require.NotEmpty(t, rows, order)

		done := make(map[uint]bool)

		for i, row := range rows {
			if i > 0 && rows[i-1].ID == row.ID {
				multiRow = true
				continue
			}

			assert.False(t, done[row.ID], "rows of photo %s are not adjacent with order %s", row.PhotoUID, order)
			done[row.ID] = true
		}
	}

	assert.True(t, multiRow, "the fixtures have media with more than one row")
}

// TestServeMCP_SearchMediaScopes checks the scope step-up and the scope verbs with search_media.
func TestServeMCP_SearchMediaScopes(t *testing.T) {
	app, router, conf := NewApiTest()
	prepareMCPTest(t)
	ServeMCP(router)

	newClient := func(t *testing.T, scope string) string {
		sess, err := entity.AddClientSession("mcp-test-search-"+strings.ReplaceAll(scope, " ", "-"), 3600, scope, authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })
		return sess.AuthToken()
	}

	call := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search_media","arguments":{"limit":2}}}`

	t.Run("MCPOnly", func(t *testing.T) {
		token := newClient(t, "mcp")
		sessionID := mcpInitSession(t, app, token)
		assert.Contains(t, mcpPost(app, `{"jsonrpc":"2.0","id":3,"method":"tools/list"}`, token, sessionID).Body.String(), `"name":"search_media"`)
		w := mcpPost(app, call, token, sessionID)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos"), w.Header().Get("WWW-Authenticate"))
	})
	t.Run("ReadVerb", func(t *testing.T) {
		token := newClient(t, "read")
		out := mcpCallTool(t, app, token, mcpInitSession(t, app, token), "search_media", map[string]any{"limit": 2})
		assert.Len(t, out.Get("items").Array(), 2)
	})
	t.Run("MCPReadVerb", func(t *testing.T) {
		token := newClient(t, "mcp read")
		w := mcpPost(app, call, token, mcpInitSession(t, app, token))
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos"), w.Header().Get("WWW-Authenticate"))
	})
	t.Run("MCPAndPhotos", func(t *testing.T) {
		token := newClient(t, "mcp photos")
		out := mcpCallTool(t, app, token, mcpInitSession(t, app, token), "search_media", map[string]any{"limit": 2})
		assert.Len(t, out.Get("items").Array(), 2)
	})
	t.Run("PublicMode", func(t *testing.T) {
		app, router, _ := NewApiTest()
		conf := prepareMCPTest(t)
		conf.Options().Public = true
		ServeMCP(router)

		sessionID := mcpInitSession(t, app, "")
		w := mcpPost(app, `{"jsonrpc":"2.0","id":3,"method":"tools/list"}`, "", sessionID)
		require.Equal(t, http.StatusOK, w.Code)
		assert.NotContains(t, w.Body.String(), "search_media")
		assert.Equal(t, len(mcp.ToolNames), strings.Count(w.Body.String(), `"inputSchema"`))

		w = mcpPost(app, call, "", sessionID)
		assert.Contains(t, w.Body.String(), `unknown tool`)
		assert.NotContains(t, w.Body.String(), `"items"`)
	})
}

// TestServeMCP_SearchMediaStepUp checks that an app connected with the mcp scope only is asked for the photos
// scope when it calls search_media, and continues in a new MCP session after it is authorized again.
func TestServeMCP_SearchMediaStepUp(t *testing.T) {
	app, router, conf, client := prepareAppTokenTest(t)
	require.NoError(t, client.SetScope("mcp photos").Save())
	adminToken := AuthenticateAdmin(app, router)

	token := func(scopes ...string) string {
		q := testAuthorizeQuery(conf, client.ClientUID)
		q.Set("scope", strings.Join(scopes, " "))
		id := authorizeRequestID(t, app, q)
		scopeList, _ := json.Marshal(scopes)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/oauth/consent/approve", `{"request":"`+id+`","scopes":`+string(scopeList)+`}`, adminToken)
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		loc, err := url.Parse(gjson.Get(r.Body.String(), "redirectUri").String())
		require.NoError(t, err)
		r = postTokenForm(app, "/api/v1/oauth/token", codeTokenForm(client.ClientUID, loc.Query().Get("code")), "", "")
		require.Equal(t, http.StatusOK, r.Code, r.Body.String())
		return gjson.Get(r.Body.String(), "access_token").String()
	}

	call := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search_media","arguments":{"limit":2}}}`

	first := token("mcp")
	firstSession := mcpInitSession(t, app, first)
	assert.Contains(t, mcpPost(app, `{"jsonrpc":"2.0","id":3,"method":"tools/list"}`, first, firstSession).Body.String(), `"name":"search_media"`)

	w := mcpPost(app, call, first, firstSession)
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos"), w.Header().Get("WWW-Authenticate"))

	// The new authorization replaces the grant, which closes the MCP session of the old one.
	second := token("mcp", "photos")
	assert.Equal(t, http.StatusNotFound, mcpPost(app, call, second, firstSession).Code)

	out := mcpCallTool(t, app, second, mcpInitSession(t, app, second), "search_media", map[string]any{"limit": 2})
	assert.Len(t, out.Get("items").Array(), 2)
}

// TestServeMCP_GetMedia checks that get_media returns the picture GET /api/v1/photos/:uid returns for the session.
func TestServeMCP_GetMedia(t *testing.T) {
	app, router, conf := NewApiTest()
	prepareMCPTest(t)
	ServeMCP(router)
	GetPhoto(router)

	token := AuthenticateAdmin(app, router)
	sessionID := mcpInitSession(t, app, token)

	t.Run("SameAsWebApp", func(t *testing.T) {
		for _, item := range mcpCallTool(t, app, token, sessionID, "search_media", map[string]any{"limit": 10}).Get("items").Array() {
			uid := item.Get("uid").String()
			r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/photos/"+uid, token)
			require.Equal(t, http.StatusOK, r.Code)
			photo := gjson.Parse(r.Body.String())

			out := mcpCallTool(t, app, token, sessionID, "get_media", map[string]any{"uid": uid})
			assert.Equal(t, uid, out.Get("uid").String())
			assert.Equal(t, photo.Get("Type").String(), out.Get("type").String())
			assert.Equal(t, photo.Get("Title").String(), out.Get("title").String())
			assert.Equal(t, photo.Get("Quality").Int(), out.Get("quality").Int())
			assert.Equal(t, photo.Get("Private").Bool(), out.Get("private").Bool())

			var names []string
			for _, f := range photo.Get("Files").Array() {
				if !f.Get("Missing").Bool() {
					name := f.Get("Name").String()
					names = append(names, name[strings.LastIndex(name, "/")+1:])
				}
			}

			var got []string
			for _, f := range out.Get("files").Array() {
				got = append(got, f.Get("name").String())
			}

			assert.ElementsMatch(t, names, got, uid)

			// Only the documented fields, so no paths, hashes, owners, or internal IDs.
			allowed := []string{"uid", "type", "title", "taken_at", "taken_at_local", "place", "favorite", "caption", "camera",
				"lens", "labels", "people", "collections", "width", "height", "duration", "location", "keywords", "files",
				"quality", "private", "archived", "review"}
			for key := range out.Map() {
				assert.Contains(t, allowed, key)
			}
			for _, f := range out.Get("files").Array() {
				for key := range f.Map() {
					assert.Contains(t, []string{"name", "type", "size", "primary"}, key)
				}
			}
		}
	})
	t.Run("NotFound", func(t *testing.T) {
		w := mcpPost(app, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_media","arguments":{"uid":"ps6sg6be2lvl0yzz"}}}`, token, sessionID)
		require.Equal(t, http.StatusOK, w.Code)
		result := mcpResult(t, w.Body.String())
		assert.True(t, result.Get("result.isError").Bool())
		assert.Contains(t, result.Raw, "media ps6sg6be2lvl0yzz not found")
	})
	t.Run("StepUp", func(t *testing.T) {
		sess, err := entity.AddClientSession("mcp-test-get-media", 3600, "mcp", authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })

		w := mcpPost(app, `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_media","arguments":{"uid":"ps6sg6be2lvl0yzz"}}}`, sess.AuthToken(), mcpInitSession(t, app, sess.AuthToken()))
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos"), w.Header().Get("WWW-Authenticate"))
	})
}

// TestServeMCP_Collections checks that list_collections and get_collection return the albums GET /api/v1/albums
// returns for the same session.
func TestServeMCP_Collections(t *testing.T) {
	app, router, conf := NewApiTest()
	prepareMCPTest(t)
	ServeMCP(router)
	SearchAlbums(router)

	token := AuthenticateAdmin(app, router)
	sessionID := mcpInitSession(t, app, token)

	t.Run("SameAsWebApp", func(t *testing.T) {
		for _, albumType := range []string{"album", "folder", "moment", "month", "state"} {
			r := AuthenticatedRequest(app, http.MethodGet, "/api/v1/albums?count=100000&type="+albumType, token)
			require.Equal(t, http.StatusOK, r.Code)

			var expected []string
			for _, a := range gjson.Parse(r.Body.String()).Array() {
				expected = append(expected, a.Get("UID").String())
			}

			var got []string
			for offset := 0; ; offset += 7 {
				out := mcpCallTool(t, app, token, sessionID, "list_collections", map[string]any{"type": albumType, "limit": 7, "offset": offset})
				assert.Equal(t, int64(len(expected)), out.Get("total").Int(), albumType)
				for _, c := range out.Get("items").Array() {
					got = append(got, c.Get("uid").String())
				}
				if !out.Get("truncated").Bool() {
					break
				}
			}

			assert.Equal(t, expected, got, albumType)

			for _, uid := range expected {
				out := mcpCallTool(t, app, token, sessionID, "get_collection", map[string]any{"uid": uid})
				assert.Equal(t, uid, out.Get("uid").String())
				assert.Equal(t, albumType, out.Get("type").String())
			}
		}
	})
	t.Run("StepUp", func(t *testing.T) {
		sess, err := entity.AddClientSession("mcp-test-collections", 3600, "mcp photos", authn.GrantClientCredentials, nil)
		require.NoError(t, err)
		t.Cleanup(func() { _ = sess.Delete() })

		w := mcpPost(app, `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"list_collections","arguments":{}}}`, sess.AuthToken(), mcpInitSession(t, app, sess.AuthToken()))
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Equal(t, McpChallenge(conf, McpErrInsufficientScope, "mcp", "photos", "albums"), w.Header().Get("WWW-Authenticate"))
	})
}
