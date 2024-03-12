package model

import (
	"time"
)

type Message struct {
	Chat_id  int
	Message  string
	Username string
	Time     time.Time
}
