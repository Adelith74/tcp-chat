package chat

import (
	"net"
)

type Room struct {
	Room_name        string
	Room_id          int
	Creator          User
	Connections      chan net.Conn
	Messages         chan string
	Dead_connections chan net.Conn
	Alive            map[net.Conn]string
}
