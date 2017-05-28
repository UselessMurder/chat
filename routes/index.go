package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/murder/chat/db/pgwrapper"
	"github.com/murder/chat/models"
	"github.com/murder/chat/sessions"
	"log"
	"net/http"
)

func IndexHandler(c *gin.Context) {

	intr, _ := c.Get("database")
	db := intr.(*pgwrapper.DataBaseWrapper)

	intr, _ = c.Get("currentSession")
	currentSession := intr.(*sessions.Session)

	switch currentSession.Status {

	case 0:
		row, err := db.QueryRow("GetCount")
		if err != nil {
			log.Println("Request error:", err)
			return
		}

		type cs struct {
			Uc    uint64
			Rc    uint64
			Mc    uint64
			Stage uint8
		}
		var counts cs
		row.Scan(&counts.Uc, &counts.Rc, &counts.Mc)
		counts.Stage = currentSession.Status
		c.HTML(http.StatusOK, "index", counts)
		return
	case 1:
		row, err := db.QueryRow("GetCount")
		if err != nil {
			log.Println("Request error:", err)
			return
		}

		type coos struct {
			Uc    uint64
			Rc    uint64
			Mc    uint64
			Yr    map[string]*models.Room
			Ar    map[string]*models.Room
			Stage uint8
		}
		var counts coos
		row.Scan(&counts.Uc, &counts.Rc, &counts.Mc)
		counts.Stage = currentSession.Status

		counts.Yr = make(map[string]*models.Room)
		counts.Ar = make(map[string]*models.Room)

		rows, _ := db.Query("GetRoomByOwner", currentSession.UserId)

		for rows.Next() {
			var currentRoom models.Room
			rows.Scan(&currentRoom.RoomId, &currentRoom.RoomName, &currentRoom.PrivatePassword, &currentRoom.CreateDate, &currentRoom.OwnerId, &currentRoom.Private)
			counts.Yr[currentRoom.RoomName] = &currentRoom
		}

		rows.Close()

		rows, _ = db.Query("GetAllRoom")

		for rows.Next() {
			var currentRoom models.Room
			rows.Scan(&currentRoom.RoomId, &currentRoom.RoomName, &currentRoom.PrivatePassword, &currentRoom.CreateDate, &currentRoom.OwnerId, &currentRoom.Private)
			counts.Ar[currentRoom.RoomName] = &currentRoom
		}

		rows.Close()

		c.HTML(http.StatusOK, "choose", counts)
		return
	case 2:
		type cus struct {
			Stage uint8
		}
		ncus := cus{currentSession.Status}
		c.HTML(http.StatusOK, "chat", ncus)
	}
}
