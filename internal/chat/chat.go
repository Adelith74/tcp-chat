package chat

import (
	"net"
)

type Chat struct {
	Chat_name        string
	Chat_id          int
	Creator          User
	Connections      chan *User
	Messages         chan string
	Dead_connections chan net.Conn
	Alive            map[net.Conn]string
	IsOpen           bool
}

func NewChat(name string) Chat {
	return Chat{
		Chat_name:        name,
		Alive:            make(map[net.Conn]string),
		Connections:      make(chan *User),
		Dead_connections: make(chan net.Conn),
		Messages:         make(chan string),
		IsOpen:           true,
	}
}
