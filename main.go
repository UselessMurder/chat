package main

import (
	"./db/pgwrapper"
	"./logical"
	"./routes"
	"./sessions"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

var userSessions sessions.SessionList

func Middle(c *gin.Context) {

	log.Println("Connected", c.ClientIP())

	id := userSessions.GetCookie(c.Request, c.Writer)

	err, currentSession := userSessions.GetSession(id)

	if err != nil {
		currentSession = &sessions.Session{id, 0, 0, "Guest", 0, time.Now().Add(24 * time.Hour)}
		userSessions.SetSession(currentSession)
	}

	c.Set("currentSession", currentSession)
	c.Set("database", &pgwrapper.Wrapper)
	c.Set("wsrouter", &logical.Router)

	c.Next()

	userSessions.SetSession(currentSession)

	log.Println("Disconnected", c.ClientIP())
}

func main() {

	fmt.Println("Start listening 3000")
	err := pgwrapper.Wrapper.ReplaceRequestList("requests.sqls")
	if err != nil {
		log.Panicln("Sql error:", err)
	}
	userSessions.OpenSessionManager()
	defer userSessions.CloseSessionManager()

	logical.Router.InitRouter()
	defer logical.Router.Stop()

	r := gin.Default()
	r.Use(Middle)
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")

	r.GET("/", routes.IndexHandler)
	r.GET("/login", routes.GetLoginHandler)
	r.GET("/register", routes.GetRegisterHandler)
	r.GET("/leave", routes.LogoutHandler)

	r.GET("/enterRoom", routes.EnterRoomHandler)
	r.GET("/createRoom", routes.GetCreateRoomHandler)
	r.GET("/enterRoomPassword", routes.GetEnterRoomPasswordHandler)
	r.GET("/leaveRoom", routes.LeaveRoomHandler)

	r.GET("/ws", routes.WebsocketWrapper)

	r.POST("/login", routes.PostLoginHandler)
	r.POST("/register", routes.PostRegisterHandler)

	r.POST("/createRoom", routes.PostCreateRoomHandler)
	r.POST("/enterRoomPassword", routes.PostEnterRoomPasswordHandler)
	r.Run(":3000")
}
