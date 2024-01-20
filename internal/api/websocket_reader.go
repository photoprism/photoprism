package api

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
)

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
			if s := Session(ws.RemoteAddr().String(), info.AuthToken); s != nil {
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
