package models

import (
	"time"
)

type User struct {
	UserId       uint64
	Login        string
	Password     string
	RegisterDate time.Time
}
