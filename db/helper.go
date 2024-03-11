package db

import (
	"go_chat/internal/chat"
)

type DBHelper struct {
}

func (helper *DBHelper) LogMessage(message string, user chat.User) {

}

// return last *amount* messages from this chat
func (helper *DBHelper) GetLogs(chat_name string, amount int) {

}
