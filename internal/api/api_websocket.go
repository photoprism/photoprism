package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/photoprism/photoprism/internal/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service"
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

// clientInfo represents information provided by the WebSocket client.
type clientInfo struct {
	SessionID string `json:"session"`
	CssUri    string `json:"css"`
	JsUri     string `json:"js"`
	Version   string `json:"version"`
}

// WebSocket registers the /ws endpoint for establishing websocket connections.
func WebSocket(router *gin.RouterGroup) {
	if router == nil {
		return
	}

	conf := service.Config()

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
		_, m, err := ws.ReadMessage()

		if err != nil {
			break
		}

		var info clientInfo

		if err := json.Unmarshal(m, &info); err != nil {
			// Do nothing.
		} else {
			if s := Session(info.SessionID); s != nil {
				wsAuth.mutex.Lock()
				wsAuth.sid[connId] = s.ID
				wsAuth.rid[connId] = s.RefID
				wsAuth.user[connId] = *s.User()
				wsAuth.mutex.Unlock()

				var clientConfig config.ClientConfig

				if s.User().IsVisitor() {
					clientConfig = conf.ClientShare()
				} else if s.User().IsRegistered() {
					clientConfig = conf.ClientSession(s)
				} else {
					clientConfig = conf.ClientPublic()
				}

				wsSendMessage("config.updated", event.Data{"config": clientConfig}, ws, writeMutex)
			}
		}
	}
}

// wsWriter initializes a WebSocket writer for sending messages.
func wsWriter(ws *websocket.Conn, writeMutex *sync.Mutex, connId string) {
	pingTicker := time.NewTicker(15 * time.Second)

	// Subscribe to events.
	e := event.Subscribe(
		"session.*",
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

			// Split topic into channel and event name.
			ch, ev := event.Topic(msg.Topic())

			// Message intended for a specific session only?
			if acl.ChannelSession.Equal(ch) {
				if s, topic := event.Topic(ev); s == sid && topic != "" {
					// Send to client with the matching session ID.
					wsSendMessage(topic, msg.Fields, ws, writeMutex)
				}
			} else if chRes := acl.Resource(ch); acl.Events.AllowAll(chRes, user.AclRole(), wsSubscribePerms) {
				// Send the message to authorized recipient.
				// event.AuditDebug([]string{"websocket", "session %s", "%s %s as %s", "granted"}, rid, wsSubscribePerms.String(), chRes.String(), user.AclRole().String())
				wsSendMessage(msg.Topic(), msg.Fields, ws, writeMutex)
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
