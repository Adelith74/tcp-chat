package chat

import (
	"errors"
	"fmt"
)

type Server struct {
	Rooms map[*Room][]*User
	Users []*User
}

func (s *Server) KickUser(username string) (err error) {
	for _, r := range s.Rooms {
		for _, u := range r {
			fmt.Println(u.Username)
			if u.Username == username {
				u.Connection.Write([]byte("You was kicked from this server"))
				u.Connection.Close()
				return
			}
		}
	}
	return errors.New("User not found")
}

func NewServer() Server {
	return Server{
		Rooms: make(map[*Room][]*User),
		Users: []*User{},
	}
}
