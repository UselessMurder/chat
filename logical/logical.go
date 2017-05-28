package logical

import (
	"encoding/json"
	"github.com/murder/chat/db/pgwrapper"
	"github.com/murder/chat/models"
	"golang.org/x/net/websocket"
	"log"
	"time"
)

var Router ChatRouter

type Wmsg struct {
	Username    string    `json:"username"`
	MessageText string    `json:"messagetext"`
	MessageType string    `json:"messagetype"`
	MessageTime time.Time `json:"messagetime"`
}

type Hmsg struct {
	ClientKey string
	RoomId    uint64
	msg       *Wmsg
}

type client struct {
	RoomId    uint64
	ClientKey string
	MsgChan   chan *Wmsg
}

type roomObj struct {
	RoomId  uint64
	Clients map[string]*client
}

func newRoom(RoomId uint64) *roomObj {

	var room roomObj

	room.Clients = make(map[string]*client)
	room.RoomId = RoomId

	return &room
}

func newClient(ClientKey string, RoomId uint64) *client {
	var cl client

	cl.ClientKey = ClientKey
	cl.RoomId = RoomId
	cl.MsgChan = make(chan *Wmsg, 1000)

	return &cl
}

type ChatRouter struct {
	rooms        map[uint64]*roomObj
	addRoom      chan *roomObj
	addClient    chan *client
	removeClient chan *client
	addMsg       chan *Hmsg
	done         chan struct{}
}

func (cr *ChatRouter) AddRoom(RoomId uint64) {
	cr.addRoom <- newRoom(RoomId)
}

func (cr *ChatRouter) AddClient(ClientKey string, RoomId uint64, ws *websocket.Conn) (error, chan *Wmsg) {

	rows, err := pgwrapper.Wrapper.Query("GetMessagesByRoomId", RoomId)

	if err != nil {
		return err, nil
	}

	defer rows.Close()

	for rows.Next() {
		var msg models.Message
		var OwnerName string
		err := rows.Scan(&OwnerName, &msg.PostText, &msg.PostDate)
		if err != nil {
			return err, nil
		}

		nmsg := Wmsg{OwnerName, msg.PostText, "User", msg.PostDate}
		bytes, err := json.Marshal(nmsg)
		if err != nil {
			return err, nil
		}

		_, err = ws.Write(bytes)
		if err != nil {
			return err, nil
		}
	}

	nclient := newClient(ClientKey, RoomId)

	cr.addClient <- nclient

	return nil, nclient.MsgChan
}

func (cr *ChatRouter) RemoveClient(ClientKey string, RoomId uint64) {
	cr.removeClient <- &client{RoomId, ClientKey, nil}
}

func (cr *ChatRouter) AddMessage(ClientKey string, RoomId uint64, UserId uint64, message Wmsg, store bool) error {

	if store {
		err := pgwrapper.Wrapper.ExecTransact("AddMessage", message.MessageText, message.MessageTime, UserId, RoomId)

		if err != nil {
			return err
		}
	}

	cr.addMsg <- &Hmsg{ClientKey, RoomId, &message}

	return nil
}

func (cr *ChatRouter) InitRouter() {
	cr.rooms = make(map[uint64]*roomObj)
	cr.addRoom = make(chan *roomObj, 1000)
	cr.addClient = make(chan *client, 1000)
	cr.removeClient = make(chan *client, 1000)
	cr.addMsg = make(chan *Hmsg, 1000)
	cr.done = make(chan struct{})

	rows, err := pgwrapper.Wrapper.Query("GetAllRoom")

	if err != nil {
		log.Panicln(err)
	}

	for rows.Next() {
		var currentRoom models.Room

		err = rows.Scan(&currentRoom.RoomId, &currentRoom.RoomName, &currentRoom.PrivatePassword, &currentRoom.CreateDate, &currentRoom.OwnerId, &currentRoom.Private)
		if err != nil {
			rows.Close()
			log.Panicln(err)
		}

		cr.rooms[currentRoom.RoomId] = newRoom(currentRoom.RoomId)
	}

	rows.Close()

	go func() {
		for {
			select {
			case currentRoom := <-cr.addRoom:
				cr.rooms[currentRoom.RoomId] = currentRoom
			case currentClient := <-cr.addClient:
				cr.rooms[currentClient.RoomId].Clients[currentClient.ClientKey] = currentClient
			case currentClient := <-cr.removeClient:
				close(cr.rooms[currentClient.RoomId].Clients[currentClient.ClientKey].MsgChan)
				delete(cr.rooms[currentClient.RoomId].Clients, currentClient.ClientKey)
			case currentMessage := <-cr.addMsg:
				for _, currentClient := range cr.rooms[currentMessage.RoomId].Clients {
					if len(currentClient.MsgChan) < cap(currentClient.MsgChan) {
						currentClient.MsgChan <- currentMessage.msg
					}
				}
			case <-cr.done:
				return
			}
		}
	}()
}

func (cr *ChatRouter) Stop() {
	cr.done <- struct{}{}
}
