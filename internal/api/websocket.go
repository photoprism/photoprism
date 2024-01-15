package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/entity"
)

// wsTimeout specifies the timeout duration for WebSocket connections.
var wsTimeout = 90 * time.Second

// wsSubPerm specifies the permissions required to subscribe to a channel.
var wsSubscribePerms = acl.Permissions{acl.ActionSubscribe}

// wsAuth maps connection IDs to specific users and session IDs.
var wsAuth = struct {
	sid   map[string]string
	rid   map[string]string
	user  map[string]entity.User
	mutex sync.RWMutex
}{
	sid:  make(map[string]string),
	rid:  make(map[string]string),
	user: make(map[string]entity.User),
}

// wsConnection upgrades the HTTP server connection to the WebSocket protocol.
var wsConnection = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// wsClient represents information about the WebSocket client.
type wsClient struct {
	AuthToken string `json:"session"`
	CssUri    string `json:"css"`
	JsUri     string `json:"js"`
	Version   string `json:"version"`
}
