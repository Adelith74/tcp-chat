package chat

import (
	"net"
)

type Room struct {
	room_name        string
	room_id          int
	creator          User
	connections      chan net.Conn
	messages         chan string
	dead_connections chan net.Conn
	alive            map[net.Conn]string
}
