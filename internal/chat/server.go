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
	Lobby *Chat
	Chats map[*Chat][]*User
	Users []*User
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
	for _, r := range s.Chats {
		for _, u := range r {
			fmt.Println(u.Username)
			if u.Username == username {
				u.Connection.Write([]byte("You was kicked from this server"))
				u.Connection.Close()
				return
			}
		}
	}
	return errors.New("user not found")
}

func (s *Server) MoveUserToLobby(user *User) {

}

func (s *Server) CloseChat(chatname string) (err error) {
	if chatname == "Lobby" {
		return errors.New("can't close lobby")
	}
	for c := range s.Chats {
		fmt.Println(chatname)
		if c.Chat_name == chatname {
			fmt.Println("Chat with this name is found")
			for _, u := range s.Chats[c] {
				u.Connection.Write([]byte("This Chat was closed by admin"))
				c.IsOpen = false
				return
			}
		}
	}
	return errors.New("there is no such chat")
}

func NewServer() Server {
	return Server{
		Chats: make(map[*Chat][]*User),
		Users: []*User{},
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
func (server *Server) NewChat(chat_name string, wg *sync.WaitGroup) error {
	defer wg.Done()
	rw := sync.RWMutex{}
	rw.Lock()
	if !server.CheckChatName(chat_name) {
		return errors.New("chat with this name is already exists")
	}
	chat := NewChat(chat_name)
	server.Chats[&chat] = []*User{}
	fmt.Printf("%s chat is running \n", chat_name)
	rw.Unlock()

	for chat.IsOpen {
		select {
		case user := <-chat.Connections:
			for conn := range chat.Alive {
				conn.Write([]byte(fmt.Sprintf("%s have connected\n", user.Username)))
			}
			go func() {
				chat.Alive[user.Connection] = user.Username
				rd := bufio.NewReader(user.Connection)
				for chat.IsOpen {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					if strings.Contains(m, "/leave") {
						break
					} else {
						chat.Messages <- fmt.Sprintf("%s: %s", user.Username, m)
					}
				}
				chat.Dead_connections <- user.Connection
				for i := range server.Chats {
					if i.Chat_name == "Lobby" {
						i.Connections <- user
					}
				}
			}()
		case msg := <-chat.Messages:
			for conn := range chat.Alive {
				conn.Write([]byte(msg))
			}
		case dconn := <-chat.Dead_connections:
			log.Printf("%v has disconnected\n", chat.Alive[dconn])
			for conn := range chat.Alive {
				conn.Write([]byte(fmt.Sprintf("%s has disconnected\n", chat.Alive[dconn])))
			}
			delete(chat.Alive, dconn)
		}
	}
	return nil
}

// lobby for all users, who just connetcted to chat
func (server Server) Start_Lobby(address string, wg *sync.WaitGroup) {
	defer wg.Done()

	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//creating lobby
	lobby := NewChat("Lobby")
	server.Chats[&lobby] = []*User{}

	//listening for new connections
	wg.Add(1)
	go func() {
		for lobby.IsOpen {
			conn, err := ln.Accept()
			if err != nil {
				log.Println()
			}
			lobby.Connections <- User{Connection: conn}
		}
	}()

	for lobby.IsOpen {
		select {
		case user := <-lobby.Connections:
			//listening for user's messages
			go func() {
				var username string
				rd := bufio.NewReader(user.Connection)
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
				server.Chats[&lobby] = append(server.Chats[&lobby], &user)
				server.Users = append(server.Users, &user)
				lobby.Alive[user.Connection] = username
				for {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					if m[0] == '/' {
						if server.parse_command(m, user, wg) {
							break
						}
					} else {
						user.Write("No such command, available commands: /help \n")
					}
				}
				lobby.Dead_connections <- user.Connection
			}()
		case dconn := <-lobby.Dead_connections:
			delete(lobby.Alive, dconn)
		}
	}
}

func (server Server) parse_command(command string, user User, wg *sync.WaitGroup) bool {
	if strings.Contains(command, "/create") {
		room_name := strings.Split(command, " ")[1]
		flag := false
		for i := range server.Chats {
			if i.Chat_name == room_name {
				flag = true
			}
		}
		if !flag {
			wg.Add(1)
			server.NewChat(strings.TrimSpace(room_name), wg)
		} else {
			user.Write("Room with this name is already exists \n")
		}
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
			server.Chats[room] = append(server.Chats[room], &user)
			room.Connections <- user
			for i := range server.Chats {
				if i.Chat_name == "Lobby" {
					i.Dead_connections <- user.Connection
				}
			}
			user.Write(fmt.Sprintf("You are connected to %s\n", room.Chat_name))
		}
		return true
	} else if strings.Contains(command, "/leave") {

	} else if strings.Contains(command, "/help") {
		user.Write("Available commands are:\n" + "/chats\n" + "/create\n" + "/connect\n")
	}

	return false
}
