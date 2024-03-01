package chat

import (
	"net"
)

type Room struct {
	Room_name        string
	Room_id          int
	Creator          User
	Connections      chan User
	Messages         chan string
	Dead_connections chan net.Conn
	Alive            map[net.Conn]string
}

func NewRoom(name string) *Room{
	return &Room{
		Alive: make(map[net.Conn]string),
		Connections: make(chan User),
		Dead_connections: make(chan net.Conn),
		Messages: make(chan string),
	}
}
