package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
)

// wsTimeout specifies the timeout duration for WebSocket connections.
var wsTimeout = 90 * time.Second

// wsSubPerm specifies the permissions required to subscribe to a channel.
var wsSubscribePerms = acl.Permissions{acl.ActionSubscribe}

// wsClientRole is the role of the API client a session authenticates. Present distinguishes a session
// that authenticates none from one whose role is acl.RoleNone, which Client.AclRole also returns for a
// role the running edition does not list.
type wsClientRole struct {
	Role    acl.Role
	Present bool
}

// wsAuth maps connection IDs to the principals of a session and its session IDs. Roles are stored as
// resolved values, so the delivery loop holds no entity another request may be mutating.
var wsAuth = struct {
	sid    map[string]string
	rid    map[string]string
	user   map[string]entity.User
	client map[string]wsClientRole
	mutex  sync.RWMutex
}{
	sid:    make(map[string]string),
	rid:    make(map[string]string),
	user:   make(map[string]entity.User),
	client: make(map[string]wsClientRole),
}

// wsConnection upgrades the HTTP server connection to the WebSocket protocol.
var wsConnection = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// wsClient represents information about the WebSocket client.
type wsClient struct {
	AuthToken string `json:"session"` //nolint:gosec // API payload field used for websocket auth handshake.
	CssUri    string `json:"css"`
	JsUri     string `json:"js"`
	Version   string `json:"version"`
}
