package models

import (
	"time"
)

type Message struct {
	MessageId uint64
	PostText  string
	PostDate  time.Time
	OwnerId   uint64
	RoomId    uint64
}
