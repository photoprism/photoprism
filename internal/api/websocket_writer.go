package api

import (
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
)

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
				if acl.ChannelUser.Equal(ch[0]) && ch[1] == user.GetUID() || acl.Events.AllowAll(acl.Resource(ch[2]), user.AclRole(), wsSubscribePerms) {
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
