package dbModel

import (
	"time"
)

type Message struct {
	Chat_id  int       `db:"chat_id"`
	Message  string    `db:"message"`
	Username string    `db:"username"`
	Time     time.Time `db:"time"`
}
