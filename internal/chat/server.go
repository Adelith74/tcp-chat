package chat

import "errors"

type Server struct {
	Rooms map[*Room][]*User
	Users []*User
}

func NewServer() *Server {
	return &Server{
		Rooms: make(map[*Room][]*User), 
		Users: []*User{},
	}
}

func (s *Server) AddRoom(room *Room){
	s.Rooms[room] = []*User{}
}

func (s *Server) KickUser(username string) (err error){
	for _, u := range s.Users{
		if u.Username == username {
			u.Connection.Write([]byte("You was kicked from this server"))
			u.Connection.Close()
			break
		}
	}
	return errors.New("User not found")
}