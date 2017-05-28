package models

import (
	"time"
)

type Room struct {
	RoomId          uint64
	RoomName        string
	PrivatePassword string
	CreateDate      time.Time
	OwnerId         uint64
	Private         bool
}
