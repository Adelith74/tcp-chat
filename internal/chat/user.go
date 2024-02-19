package chat

import "net"

type User struct {
	username   string
	connection net.Conn
}