package chat

import "net"

type User struct {
	Username   string
	Connection net.Conn
}

// writes message to user's connection
func (u *User) Write(message string) {
	u.Connection.Write([]byte(message))
}
