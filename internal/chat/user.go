package chat

import "net"

type User struct {
	Username   string
	Connection net.Conn
}