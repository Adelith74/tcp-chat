package chat

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	Lobby     *Chat
	Chats     map[*Chat][]*User
	Users     []*User
	Wg        sync.WaitGroup
	Available int
}

// if this name is available returns true
func (s *Server) Check_username(username string) bool {
	flag := true
	for _, u := range s.Users {
		if u.Username == username {
			flag = false
		}
	}
	return flag
}

func (s *Server) KickUser(username string) (err error) {
	rw := sync.RWMutex{}
	rw.Lock()
	for _, r := range s.Chats {
		var users []*User
		for _, u := range r {
			if u.Username == username {
				u.Write("You was kicked from this server")
				u.Connection.Close()
			} else {
				users = append(users, u)
			}
		}
		s.Users = users
		return
	}
	rw.Unlock()
	return errors.New("user not found")
}

func (s *Server) CloseChat(chatname string) (err error) {
	if chatname == "Lobby" {
		return errors.New("can't close lobby")
	}
	for c := range s.Chats {
		if c.Chat_name == chatname {
			for _, u := range s.Chats[c] {
				u.Write("This Chat was closed by admin\n")
				c.Dead_connections <- u
				s.Lobby.Connections <- u
			}
			c.IsOpen = false
			delete(s.Chats, c)
			return
		}
	}
	return errors.New("there is no such chat")
}

func NewServer() Server {
	return Server{
		Chats:     make(map[*Chat][]*User),
		Users:     []*User{},
		Wg:        sync.WaitGroup{},
		Available: 1,
	}
}

// returns true if chatname is available
func (server *Server) CheckChatName(chat_name string) bool {
	flag := true
	for v := range server.Chats {
		if v.Chat_name == chat_name {
			flag = false
		}
	}
	return flag
}

// creates a new chat
func (server *Server) NewChat(chat_name string) error {
	defer server.Wg.Done()
	rw := sync.RWMutex{}
	rw.Lock()
	chat := NewChat(chat_name, server.Available)
	server.Available += 1
	rw.Unlock()
	server.Chats[chat] = []*User{}
	fmt.Printf("%s chat is running \n", chat_name)
	for chat.IsOpen {
		select {
		case user := <-chat.Connections:
			for u := range chat.Alive {
				u.Write(fmt.Sprintf("%s have connected\n", user.Username))
			}
			go func() {
				chat.Alive[user] = user.Username
				rd := bufio.NewReader(user.Connection)
				for chat.IsOpen {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					if strings.Contains(m, "/leave") {
						break
					} else {
						chat.Messages <- fmt.Sprintf("%s: %s \n", user.Username, m)
					}
				}
				chat.Dead_connections <- user
				server.Lobby.Connections <- user
			}()
		case msg := <-chat.Messages:
			for u := range chat.Alive {
				u.Write(msg)
			}
		case dconn := <-chat.Dead_connections:
			log.Printf("%v has disconnected\n", chat.Alive[dconn])
			for u := range chat.Alive {
				u.Write(fmt.Sprintf("%s has disconnected\n", chat.Alive[dconn]))
			}
			delete(chat.Alive, dconn)
		}
	}
	return nil
}

// lobby for all users, who just connetcted to chat
func (server *Server) Start_Lobby(address string) {
	defer server.Wg.Done()

	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//creating lobby
	lobby := NewChat("Lobby", 0)
	server.Lobby = lobby
	server.Chats[lobby] = []*User{}

	//listening for new connections
	server.Wg.Add(1)
	defer server.Wg.Done()
	go func() {
		for lobby.IsOpen {
			conn, err := ln.Accept()
			if err != nil {
				log.Println()
			}
			lobby.Connections <- &User{Connection: conn}
		}
	}()
	for lobby.IsOpen {
		select {
		case user := <-lobby.Connections:
			//listening for user's messages
			go func() {
				rd := bufio.NewReader(user.Connection)
				if user.Username == "" {
					var username string
					for {
						user.Write("Enter your username: \n")
						username, err = rd.ReadString('\n')
						if err != nil {
							fmt.Println(err)
						}
						username = strings.TrimSpace(username)
						if !server.Check_username(username) {
							user.Write("This username is already taken\n")
						} else {
							break
						}
					}
					user.Username = username
					user.Write(fmt.Sprintf("Welcome to Lobby, %s \n", user.Username))
					server.Users = append(server.Users, user)
				}
				server.Chats[lobby] = append(server.Chats[lobby], user)
				lobby.Alive[user] = user.Username
				for {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					if m[0] == '/' {
						if server.parse_command(m, user) {
							break
						}
					} else {
						user.Write("No such command, available commands: /help \n")
					}
				}
				lobby.Dead_connections <- user
			}()
		case dconn := <-lobby.Dead_connections:
			delete(lobby.Alive, dconn)
		}
	}
}

func (server *Server) parse_command(command string, user *User) bool {
	if strings.Contains(command, "/create") {
		chat_name := strings.Split(command, " ")[1]
		chat_name = strings.TrimSpace(chat_name)
		flag := false
		rw := sync.RWMutex{}
		rw.Lock()
		flag = server.CheckChatName(chat_name)
		if flag {
			server.Wg.Add(1)
			go server.NewChat(strings.TrimSpace(chat_name))
			user.Write("Room was successfully created \n")
		} else {
			user.Write("Room with this name is already exists \n")
		}
		rw.Unlock()
	} else if strings.Contains(command, "/chats") {
		for i := range server.Chats {
			if i.Chat_name != "Lobby" {
				user.Write(i.Chat_name)
			}
		}
	} else if strings.Contains(command, "/connect") {
		room_name := strings.TrimSpace(strings.Split(command, " ")[1])
		if room_name == "Lobby" {
			return false
		}
		var room *Chat
		flag := false
		for i := range server.Chats {
			if i.Chat_name == room_name {
				flag = true
				room = i
			}
		}
		if !flag {
			user.Write("There is no such room \n")
			return false
		} else {
			server.Chats[room] = append(server.Chats[room], user)
			room.Connections <- user
			delete(server.Lobby.Alive, user)
			user.Write(fmt.Sprintf("You are connected to %s", room.Chat_name) + "\n")
			return true
		}
	} else if strings.Contains(command, "/leave") {

	} else if strings.Contains(command, "/help") {
		user.Write("Available commands are:\n" + "/chats\n" + "/create\n" + "/connect\n")
	} else {
		user.Write("No such command, available commands: /help \n")
		user.Write("Available commands are:\n" + "/chats\n" + "/create\n" + "/connect\n")
	}
	return false
}
