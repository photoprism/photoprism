package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/service"
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
	user  map[string]entity.Person
	mutex sync.RWMutex
}{user: make(map[string]entity.Person)}

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

		var info clientInfo

		if err := json.Unmarshal(m, &info); err != nil {
			// Do nothing.
		} else {
			if sess := Session(info.SessionToken); sess.Valid() {
				wsAuth.mutex.Lock()
				wsAuth.user[connId] = sess.User
				wsAuth.mutex.Unlock()

				var clientConfig config.ClientConfig

				if sess.User.Guest() {
					clientConfig = conf.GuestConfig()
				} else if sess.User.Registered() {
					clientConfig = conf.UserConfig()
				} else {
					clientConfig = conf.PublicConfig()
				}

				writeMutex.Lock()
				ws.SetWriteDeadline(time.Now().Add(30 * time.Second))

				if err := ws.WriteJSON(gin.H{"event": "config.updated", "data": event.Data{"config": clientConfig}}); err != nil {
					// Do nothing.
				}

				writeMutex.Unlock()
			}
		}
	}
}

func wsWriter(ws *websocket.Conn, writeMutex *sync.Mutex, connId string) {
	pingTicker := time.NewTicker(15 * time.Second)
	s := event.Subscribe(
		"log.*",
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
		"sync.*",
	)

	defer func() {
		pingTicker.Stop()
		event.Unsubscribe(s)
		ws.Close()

		wsAuth.mutex.Lock()
		wsAuth.user[connId] = entity.UnknownPerson
		wsAuth.mutex.Unlock()
	}()

	for {
		select {
		case <-pingTicker.C:
			writeMutex.Lock()
			ws.SetWriteDeadline(time.Now().Add(30 * time.Second))

			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				writeMutex.Unlock()
				return
			}

			writeMutex.Unlock()
		case msg := <-s.Receiver:
			wsAuth.mutex.RLock()

			user := entity.UnknownPerson

			if hit, ok := wsAuth.user[connId]; ok {
				user = hit
			}

			wsAuth.mutex.RUnlock()

			if user.Registered() {
				writeMutex.Lock()
				ws.SetWriteDeadline(time.Now().Add(30 * time.Second))

				if err := ws.WriteJSON(gin.H{"event": msg.Name, "data": msg.Fields}); err != nil {
					writeMutex.Unlock()
					return
				}

				writeMutex.Unlock()
			}
		}
	}
}

// GET /api/v1/ws
func Websocket(router *gin.RouterGroup) {
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
			wsAuth.user[connId] = entity.UnknownPerson
		}
		wsAuth.mutex.Unlock()

		go wsWriter(ws, &writeMutex, connId)

		wsReader(ws, &writeMutex, connId, conf)
	})
}
