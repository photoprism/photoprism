package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/get"
	"github.com/photoprism/photoprism/pkg/rnd"
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

// WebSocket registers the /ws endpoint for establishing websocket connections.
func WebSocket(router *gin.RouterGroup) {
	if router == nil {
		return
	}

	conf := get.Config()

	if conf == nil {
		return
	}

	router.GET("/ws", func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		ws, err := wsConnection.Upgrade(w, r, nil)

		if err != nil {
			return
		}

		var writeMutex sync.Mutex

		defer ws.Close()

		connId := rnd.UUID()

		// Init connection.
		wsAuth.mutex.Lock()

		if conf.Public() {
			wsAuth.user[connId] = entity.Admin
		} else {
			wsAuth.user[connId] = entity.UnknownUser
		}

		wsAuth.mutex.Unlock()

		// Init writer.
		go wsWriter(ws, &writeMutex, connId)

		// Init reader.
		wsReader(ws, &writeMutex, connId, conf)
	})
}

// wsReader initializes a WebSocket reader for receiving messages.
func wsReader(ws *websocket.Conn, writeMutex *sync.Mutex, connId string, conf *config.Config) {
	defer ws.Close()

	ws.SetReadLimit(4096)

	if err := ws.SetReadDeadline(time.Now().Add(wsTimeout)); err != nil {
		return
	}

	ws.SetPongHandler(func(string) error { _ = ws.SetReadDeadline(time.Now().Add(wsTimeout)); return nil })

	for {
		_, m, readErr := ws.ReadMessage()

		if readErr != nil {
			break
		}

		var info wsClient

		if jsonErr := json.Unmarshal(m, &info); jsonErr != nil {
			// Do nothing.
		} else {
			if s := Session(info.AuthToken); s != nil {
				wsAuth.mutex.Lock()
				wsAuth.sid[connId] = s.ID
				wsAuth.rid[connId] = s.RefID
				wsAuth.user[connId] = *s.User()
				wsAuth.mutex.Unlock()

				wsSendMessage("config.updated", event.Data{"config": conf.ClientSession(s)}, ws, writeMutex)
			}
		}
	}
}

// wsWriter initializes a WebSocket writer for sending messages.
func wsWriter(ws *websocket.Conn, writeMutex *sync.Mutex, connId string) {
	pingTicker := time.NewTicker(15 * time.Second)

	// Subscribe to events.
	e := event.Subscribe(
		"user.*.*.*",
		"session.*.*.*",
		"log.fatal",
		"log.error",
		"log.warning",
		"log.warn",
		"log.info",
		"notify.*",
		"index.*",
		"upload.*",
		"import.*",
		"config.*",
		"count.*",
		"photos.*",
		"cameras.*",
		"lenses.*",
		"countries.*",
		"albums.*",
		"labels.*",
		"subjects.*",
		"people.*",
		"sync.*",
	)

	defer func() {
		pingTicker.Stop()
		event.Unsubscribe(e)
		_ = ws.Close()

		wsAuth.mutex.Lock()
		delete(wsAuth.sid, connId)
		delete(wsAuth.rid, connId)
		delete(wsAuth.user, connId)
		wsAuth.mutex.Unlock()
	}()

	for {
		select {
		case <-pingTicker.C:
			writeMutex.Lock()

			if err := ws.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
				writeMutex.Unlock()
				return
			} else if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				writeMutex.Unlock()
				return
			}

			writeMutex.Unlock()
		case msg := <-e.Receiver:
			wsAuth.mutex.RLock()

			sid := wsAuth.sid[connId] // Session ID.
			// rid := wsAuth.rid[connId]  // Session RefID.
			user := entity.UnknownUser // User.

			if hit, ok := wsAuth.user[connId]; ok {
				user = hit
			}

			wsAuth.mutex.RUnlock()

			// Split topic into sub-channels.
			ev := msg.Topic()
			ch := strings.Split(ev, ".")

			// Send the message only to authorized recipients.
			switch len(ch) {
			case 2:
				// Send to everyone who is allowed to subscribe.
				if res := acl.Resource(ch[0]); acl.Events.AllowAll(res, user.AclRole(), wsSubscribePerms) {
					wsSendMessage(ev, msg.Fields, ws, writeMutex)
				}
			case 4:
				ev = strings.Join(ch[2:4], ".")
				if acl.ChannelUser.Equal(ch[0]) && ch[1] == user.UID() || acl.Events.AllowAll(acl.Resource(ch[2]), user.AclRole(), wsSubscribePerms) {
					// Send to matching user uid.
					wsSendMessage(ev, msg.Fields, ws, writeMutex)
				} else if acl.ChannelSession.Equal(ch[0]) && ch[1] == sid {
					// Send to matching session id.
					wsSendMessage(ev, msg.Fields, ws, writeMutex)
				}
			}
		}
	}
}

// wsSendMessage sends a message to the WebSocket client.
func wsSendMessage(topic string, data interface{}, ws *websocket.Conn, writeMutex *sync.Mutex) {
	if topic == "" || ws == nil || writeMutex == nil {
		return
	}

	writeMutex.Lock()
	defer writeMutex.Unlock()

	if err := ws.SetWriteDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return
	} else if err := ws.WriteJSON(gin.H{"event": topic, "data": data}); err != nil {
		return
	}
}
