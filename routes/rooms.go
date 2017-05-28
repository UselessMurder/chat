package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/murder/chat/db/pgwrapper"
	"github.com/murder/chat/hashgenerator"
	"github.com/murder/chat/logical"
	"github.com/murder/chat/models"
	"github.com/murder/chat/sessions"
	"net/http"
	"strconv"
	"time"
)

func LeaveRoomHandler(c *gin.Context) {

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 2 {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	currentSession.RoomId = 0
	currentSession.Status = 1

	c.Redirect(http.StatusTemporaryRedirect, "/")
	return
}

func EnterRoomHandler(c *gin.Context) {

	roomNumber := c.DefaultQuery("num", "0")
	token := c.DefaultQuery("token", "default")

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 1 {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	num, err := strconv.Atoi(roomNumber)

	if err != nil {
		num = 0
	}

	if num == 0 {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	intr, _ = c.Get("database")
	db := intr.(*pgwrapper.DataBaseWrapper)

	row, err := db.QueryRow("GetRoomById", num)

	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	var currentRoom models.Room

	err = row.Scan(&currentRoom.RoomId, &currentRoom.RoomName, &currentRoom.PrivatePassword, &currentRoom.CreateDate, &currentRoom.OwnerId, &currentRoom.Private)

	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	if token == currentRoom.PrivatePassword {

		currentSession.RoomId = uint64(num)
		currentSession.Status = 2

		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	} else {
		c.Redirect(http.StatusTemporaryRedirect, "/enterRoomPassword?num="+strconv.Itoa(num))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, "/")
	return
}

func GetEnterRoomPasswordHandler(c *gin.Context) {

	roomNumber := c.DefaultQuery("num", "0")
	roomError := c.DefaultQuery("error", "0")

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 1 {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	num, err := strconv.Atoi(roomNumber)

	if err != nil {
		num = 0
	}

	inputError, err := strconv.Atoi(roomError)

	if err != nil {
		inputError = 0
	}

	intr, _ = c.Get("database")
	db := intr.(*pgwrapper.DataBaseWrapper)

	row, err := db.QueryRow("GetRoomById", num)

	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	var currentRoom models.Room

	err = row.Scan(&currentRoom.RoomId, &currentRoom.RoomName, &currentRoom.PrivatePassword, &currentRoom.CreateDate, &currentRoom.OwnerId, &currentRoom.Private)

	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	if currentRoom.PrivatePassword == "default" {
		c.Redirect(http.StatusTemporaryRedirect, "/enterRoom?num="+roomNumber)
	} else {

		type rps struct {
			Stage uint8
			Irp   int
			Num   string
		}

		nrps := rps{currentSession.Status, inputError, roomNumber}

		c.HTML(http.StatusOK, "roomPswd", nrps)
	}

	return
}

func PostEnterRoomPasswordHandler(c *gin.Context) {

	password := c.PostForm("password")

	roomNumber := c.DefaultQuery("num", "0")

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 1 {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	num, err := strconv.Atoi(roomNumber)

	if err != nil {
		num = 0
	}

	intr, _ = c.Get("database")
	db := intr.(*pgwrapper.DataBaseWrapper)

	row, err := db.QueryRow("GetRoomById", num)

	if err != nil {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	var currentRoom models.Room

	err = row.Scan(&currentRoom.RoomId, &currentRoom.RoomName, &currentRoom.PrivatePassword, &currentRoom.CreateDate, &currentRoom.OwnerId, &currentRoom.Private)

	if err != nil {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	hash, _ := hashgenerator.GetHashSum28(password, "password")

	if hash == currentRoom.PrivatePassword {
		c.Redirect(http.StatusMovedPermanently, "/enterRoom?num="+roomNumber+"&token="+hash)
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/enterRoomPassword?num="+roomNumber+"&error=1")
	return
}

func GetCreateRoomHandler(c *gin.Context) {

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 1 {
		c.Redirect(http.StatusTemporaryRedirect, "/")
		return
	}

	inputError := c.DefaultQuery("error", "0")

	type crs struct {
		Icr   int
		Stage uint8
	}

	errValue, err := strconv.Atoi(inputError)

	if err != nil {
		errValue = 0
	}

	cr := crs{errValue, currentSession.Status}

	c.HTML(http.StatusOK, "createRoom", cr)
}

func PostCreateRoomHandler(c *gin.Context) {

	roomName := c.PostForm("roomname")
	private := c.PostForm("private")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirm-password")

	intr, _ := c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	if currentSession.Status != 1 {
		c.Redirect(http.StatusMovedPermanently, "/")
		return
	}

	var currentRoom models.Room

	if len(roomName) < 3 {
		c.Redirect(http.StatusMovedPermanently, "/createRoom?error=1")
		return
	}

	currentRoom.RoomName = roomName

	if private == "done" {

		currentRoom.Private = true

		if len(password) < 5 {
			c.Redirect(http.StatusMovedPermanently, "/createRoom?error=1")
			return
		}

		if password != confirmPassword {
			c.Redirect(http.StatusMovedPermanently, "/createRoom?error=1")
			return
		}

		currentRoom.PrivatePassword, _ = hashgenerator.GetHashSum28(password, "password")

	} else {

		currentRoom.Private = false
		currentRoom.PrivatePassword = "default"
	}

	currentRoom.OwnerId = currentSession.UserId

	currentRoom.CreateDate = time.Now()

	intr, _ = c.Get("database")
	db := intr.(*pgwrapper.DataBaseWrapper)

	err := db.ExecTransact("CreateRoom", currentRoom.RoomName, currentRoom.PrivatePassword, currentRoom.CreateDate, currentRoom.OwnerId, currentRoom.Private)

	if err != nil {
		c.Redirect(http.StatusMovedPermanently, "/createRoom?error=1")
		return
	}

	intr, _ = c.Get("wsrouter")
	router := intr.(*logical.ChatRouter)

	row, err := db.QueryRow("GetRoomByName", currentRoom.RoomName)

	if err == nil {
		var val uint64
		err := row.Scan(&val)
		if err == nil {
			router.AddRoom(val)
		}
	}

	c.Redirect(http.StatusMovedPermanently, "/")
}
