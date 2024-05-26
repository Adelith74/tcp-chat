package chat

import (
	"log"
	"net"
)

type User struct {
	Username   string
	Connection net.Conn
}

// writes message to user's connection
func (u *User) Write(message string) {
	_, err := u.Connection.Write([]byte(message))
	if err != nil {
		log.Fatal(err)
	}
}
