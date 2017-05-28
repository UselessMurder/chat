package routes

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/murder/chat/hashgenerator"
	"github.com/murder/chat/logical"
	"github.com/murder/chat/sessions"
	"golang.org/x/net/websocket"
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var insider *sessions.Session
var insiderGuard = &sync.Mutex{}

func WebsocketWrapper(c *gin.Context) {

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 2 {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	insiderGuard.Lock()
	insider = currentSession

	handler := websocket.Handler(WebsocketServer)
	handler.ServeHTTP(c.Writer, c.Request)
}

func WebsocketServer(ws *websocket.Conn) {

	currentSession := insider
	insiderGuard.Unlock()

	hash, _ := hashgenerator.GenerateHash28(time.Now().String(), "websocket")

	rid := currentSession.RoomId

	uid := currentSession.UserId

	err, channel := logical.Router.AddClient(hash, rid, ws)

	if err != nil {
		log.Println(err)
		return
	}

	go func() {
		for msg := range channel {
			bytes, err := json.Marshal(msg)
			if err == nil {
				_, err := ws.Write(bytes)
				if err != nil {
					return
				}
			}
		}
	}()

	enterMessage := logical.Wmsg{currentSession.UserName, "Connected!", "Admin", time.Now()}

	err = logical.Router.AddMessage(hash, rid, uid, enterMessage, false)

	if err != nil {
		log.Println(err)
		return
	}

	defer func() {

		logical.Router.RemoveClient(hash, rid)

		leaveMessage := logical.Wmsg{currentSession.UserName, "Disconnected!", "Admin", time.Now()}

		err := logical.Router.AddMessage(hash, rid, uid, leaveMessage, false)

		if err != nil {
			log.Println(err)
		}

	}()

	for {

		lenBuf := make([]byte, 8)

		_, err := ws.Read(lenBuf[:])
		if err != nil {
			log.Println(err)
			return
		}

		length, _ := strconv.Atoi(strings.TrimSpace(string(lenBuf)))

		if length > 65536 {
			log.Println(errors.New("Incorrect size!"))
			return
		}

		if length < 0 {
			log.Println(errors.New("Incorrect size!"))
			return
		}

		buf := make([]byte, length)

		_, err = ws.Read(buf)

		if err != nil {
			log.Println(err)
			return
		}

		str := html.EscapeString(string(buf))

		currentMessage := logical.Wmsg{currentSession.UserName, str, "User", time.Now()}

		if err != nil {
			log.Println(err)
			return
		}

		err = logical.Router.AddMessage(hash, rid, uid, currentMessage, true)

		if err != nil {
			log.Println(err)
			return
		}
	}
}
