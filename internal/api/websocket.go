package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/session"
	"github.com/photoprism/photoprism/pkg/rnd"
)

var wsConnection = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var wsTimeout = 90 * time.Second

type clientInfo struct {
	SessionToken string `json:"session"`
	JsHash       string `json:"js"`
	CssHash      string `json:"css"`
	Version      string `json:"version"`
}

var wsAuth = struct {
	authenticated map[string]bool
	mutex         sync.RWMutex
}{authenticated: make(map[string]bool)}

func wsReader(ws *websocket.Conn, writeMutex *sync.Mutex, connId string, conf *config.Config) {
	defer ws.Close()

	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(wsTimeout))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(wsTimeout)); return nil })

	for {
		_, m, err := ws.ReadMessage()

		if err != nil {
			break
		}

		log.Debugf("websocket: received %d bytes", len(m))

		var info clientInfo

		if err := json.Unmarshal(m, &info); err != nil {
			log.Error(err)
		} else {
			if session.Exists(info.SessionToken) {
				log.Debug("websocket: authenticated")

				wsAuth.mutex.Lock()
				wsAuth.authenticated[connId] = true
				wsAuth.mutex.Unlock()

				writeMutex.Lock()
				ws.SetWriteDeadline(time.Now().Add(30 * time.Second))

				if err := ws.WriteJSON(gin.H{"event": "config.updated", "data": event.Data(conf.ClientConfig())}); err != nil {
					log.Error(err)
				}
				writeMutex.Unlock()
			}
		}
	}
}

func wsWriter(ws *websocket.Conn, writeMutex *sync.Mutex, connId string) {
	pingTicker := time.NewTicker(15 * time.Second)
	s := event.Subscribe("log.*", "notify.*", "index.*", "upload.*", "import.*", "config.*", "count.*", "photos.*", "albums.*", "labels.*")

	defer func() {
		pingTicker.Stop()
		event.Unsubscribe(s)
		ws.Close()

		wsAuth.mutex.Lock()
		wsAuth.authenticated[connId] = false
		wsAuth.mutex.Unlock()
	}()

	for {
		select {
		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(30 * time.Second))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case msg := <-s.Receiver:
			wsAuth.mutex.RLock()
			auth := wsAuth.authenticated[connId]
			wsAuth.mutex.RUnlock()

			if auth {
				writeMutex.Lock()
				ws.SetWriteDeadline(time.Now().Add(30 * time.Second))

				if err := ws.WriteJSON(gin.H{"event": msg.Name, "data": msg.Fields}); err != nil {
					writeMutex.Unlock()
					log.Debug(err)
					return
				}
				writeMutex.Unlock()
			}
		}
	}
}

// GET /api/v1/ws
func Websocket(router *gin.RouterGroup, conf *config.Config) {
	if router == nil {
		log.Error("websocket: router is nil")
		return
	}

	if conf == nil {
		log.Error("websocket: conf is nil")
		return
	}

	router.GET("/ws", func(c *gin.Context) {
		w := c.Writer
		r := c.Request

		ws, err := wsConnection.Upgrade(w, r, nil)
		if err != nil {
			log.Error(err)
			return
		}

		var writeMutex sync.Mutex

		defer ws.Close()

		connId := rnd.UUID()

		if conf.Public() {
			wsAuth.mutex.Lock()
			wsAuth.authenticated[connId] = true
			wsAuth.mutex.Unlock()
		}

		log.Debug("websocket: connected")

		go wsWriter(ws, &writeMutex, connId)

		wsReader(ws, &writeMutex, connId, conf)
	})
}
