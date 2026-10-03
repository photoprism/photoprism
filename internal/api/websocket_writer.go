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
	"github.com/photoprism/photoprism/pkg/clean"
)

// WebsocketTopics lists the event topics that are forwarded to connected websocket clients.
// Extensions may append additional topics during package initialization so they are subscribed
// as soon as the server starts accepting websocket connections.
var WebsocketTopics = []string{
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
}

// AppendWebsocketTopics adds the provided topics to the global websocket topic list.
func AppendWebsocketTopics(topics ...string) {
	if len(topics) == 0 {
		return
	}

	WebsocketTopics = append(WebsocketTopics, topics...)
}

// wsSendMessage sends a message to the WebSocket client.
func wsSendMessage(topic string, data any, ws *websocket.Conn, writeMutex *sync.Mutex) {
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

// wsRelease drops the principals and session identifiers a connection was holding.
func wsRelease(connId string) {
	wsAuth.mutex.Lock()
	defer wsAuth.mutex.Unlock()

	delete(wsAuth.sid, connId)
	delete(wsAuth.rid, connId)
	delete(wsAuth.user, connId)
	delete(wsAuth.client, connId)
}

// wsSessionClient returns the client a session authenticates, using the same predicate the REST
// handlers apply. It is read before the role, because resolving the role loads the client entity onto
// the session and rewrites the provider the predicate reads.
func wsSessionClient(s *entity.Session) wsClientRole {
	if s == nil || !s.IsClient() {
		return wsClientRole{}
	}

	return wsClientRole{Role: s.GetClientRole(), Present: true}
}

// wsDelivery returns the event name a connection receives for the topic, and reports whether the
// connection is authorized to receive it at all. A client session receives nothing, whatever its
// account may subscribe to: clients and agents read through the REST API.
func wsDelivery(topic, sid string, user entity.User, client wsClientRole) (ev string, ok bool) {
	ch := strings.Split(topic, ".")
	ev = topic

	if len(ch) == 4 {
		ev = strings.Join(ch[2:4], ".")
	}

	if client.Present {
		return ev, false
	}

	switch len(ch) {
	case 2, 3:
		// Two-part topics and three-part ones such as audit.log.info.
		res := acl.Resource(ch[0])

		return ev, acl.Events.AllowAll(res, user.AclRole(), wsSubscribePerms)
	case 4:
		switch {
		case acl.Events.AllowAll(acl.Resource(ch[2]), user.AclRole(), wsSubscribePerms):
			return ev, true
		case acl.ChannelUser.Equal(ch[0]) && ch[1] == user.GetUID():
			// Addressed to a matching user uid.
			return ev, true
		case acl.ChannelSession.Equal(ch[0]) && ch[1] == sid:
			// Addressed to a matching session id.
			return ev, true
		}
	}

	return ev, false
}

// wsWriter initializes a WebSocket writer for sending messages.
func wsWriter(ws *websocket.Conn, writeMutex *sync.Mutex, connId string) {
	pingTicker := time.NewTicker(15 * time.Second)

	// Subscribe to events.
	topics := append([]string(nil), WebsocketTopics...)
	e := event.Subscribe(topics...)

	// Set once the connection is reported as narrowed by its client role.
	reported := false

	defer func() {
		pingTicker.Stop()
		event.Unsubscribe(e)
		_ = ws.Close()

		wsRelease(connId)
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
			user := entity.UnknownUser      // User.
			client := wsAuth.client[connId] // Client role, absent without a client.

			if hit, ok := wsAuth.user[connId]; ok {
				user = hit
			}

			wsAuth.mutex.RUnlock()

			// Send the message only to authorized recipients.
			if ev, ok := wsDelivery(msg.Topic(), sid, user, client); ok {
				wsSendMessage(ev, msg.Fields, ws, writeMutex)
			} else if !reported && client.Present {
				// Report once per connection when the client role is what narrows it, so an
				// integration that goes quiet has a reason to read.
				if _, byAccount := wsDelivery(msg.Topic(), sid, user, wsClientRole{}); byAccount {
					reported = true
					log.Debugf("websocket: client role %s receives no %s", clean.LogQuote(client.Role.String()), clean.Log(msg.Topic()))
				}
			}
		}
	}
}
