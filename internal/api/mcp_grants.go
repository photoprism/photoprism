package api

import (
	"sync"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/photoprism/photoprism/internal/entity"
)

// sdkSessionIdHeader is the response header in which the MCP SDK returns the ID of a new session.
const sdkSessionIdHeader = "Mcp-Session-Id"

// mcpGrantSessions maps the IDs of MCP sessions opened with app access tokens to their grants, since an SDK
// session does not expose the user it is bound to.
type mcpGrantSessions struct {
	server *sdkmcp.Server
	mu     sync.Mutex
	grants map[string]string
}

// mcpGrantExists reports whether the grant with the specified UID still exists.
func mcpGrantExists(grantUID string) bool {
	return entity.FindGrant(grantUID) != nil
}

// newMcpGrantSessions returns an empty map for the MCP sessions of the specified server.
func newMcpGrantSessions(server *sdkmcp.Server) *mcpGrantSessions {
	return &mcpGrantSessions{server: server, grants: make(map[string]string)}
}

// Add remembers the grant of an MCP session and forgets sessions that no longer exist. The grant is checked
// after the session is recorded, so a revocation that deletes it meanwhile either finds the session or is seen here.
func (m *mcpGrantSessions) Add(sessionID, grantUID string, grantExists func(string) bool) {
	if m == nil || sessionID == "" || grantUID == "" {
		return
	}

	m.mu.Lock()
	_, known := m.grants[sessionID]

	if !known {
		active := make(map[string]bool)

		for ss := range m.server.Sessions() {
			active[ss.ID()] = true
		}

		for id := range m.grants {
			if !active[id] {
				delete(m.grants, id)
			}
		}

		m.grants[sessionID] = grantUID
	}

	m.mu.Unlock()

	if !known && grantExists != nil && !grantExists(grantUID) {
		m.Close(grantUID)
	}
}

// Close ends the MCP sessions of the specified grant.
func (m *mcpGrantSessions) Close(grantUID string) {
	if m == nil || grantUID == "" {
		return
	}

	m.mu.Lock()
	ids := make(map[string]bool)

	for id, g := range m.grants {
		if g == grantUID {
			ids[id] = true
			delete(m.grants, id)
		}
	}

	m.mu.Unlock()

	if len(ids) == 0 {
		return
	}

	for ss := range m.server.Sessions() {
		if ids[ss.ID()] {
			_ = ss.Close()
		}
	}
}

// Len returns the number of MCP sessions with a known grant.
func (m *mcpGrantSessions) Len() int {
	if m == nil {
		return 0
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.grants)
}
