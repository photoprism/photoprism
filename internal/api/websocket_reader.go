package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/pkg/authn"
)

// wsRateLimitedEvent is sent to a client over the authentication failure limit before the connection closes.
const wsRateLimitedEvent = "websocket.rate-limited"

// wsClose sends a close message with the specified code and reason.
func wsClose(ws *websocket.Conn, writeMutex *sync.Mutex, code int, reason string) {
	writeMutex.Lock()
	defer writeMutex.Unlock()

	_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, reason), time.Now().Add(time.Second))
}

// wsReader initializes a WebSocket reader for receiving messages.
func wsReader(ws *websocket.Conn, writeMutex *sync.Mutex, connId string, conf *config.Config, clientIp string) {
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
			s, err := LookupSession(clientIp, info.AuthToken)

			// Tell a client over the authentication failure limit to try again later, and close.
			if errors.Is(err, authn.ErrRateLimitExceeded) {
				wsSendMessage(wsRateLimitedEvent, event.Data{"code": http.StatusTooManyRequests}, ws, writeMutex)
				wsClose(ws, writeMutex, websocket.CloseTryAgainLater, authn.ErrRateLimitExceeded.Error())
				return
			}

			if s != nil {
				// Resolve both principals before taking the lock, since either may query the database.
				user := *s.GetUser()
				client := wsSessionClient(s)

				wsAuth.mutex.Lock()
				wsAuth.sid[connId] = s.ID
				wsAuth.rid[connId] = s.RefID
				wsAuth.user[connId] = user
				wsAuth.client[connId] = client
				wsAuth.mutex.Unlock()

				// Send the session config only if the session may view it, as GET /api/v1/config does.
				if authorizeSession(clientIp, s, acl.ResourceConfig, acl.Permissions{acl.ActionView}).Valid() {
					wsSendMessage("config.updated", event.Data{"config": conf.ClientSession(s)}, ws, writeMutex)
				} else {
					wsSendMessage("config.updated", event.Data{"config": conf.ClientPublic()}, ws, writeMutex)
				}
			}
		}
	}
}
