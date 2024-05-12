package chat

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"go_chat/internal/model"
	"go_chat/internal/repository"
	"go_chat/internal/repository/dbModel"
	"log"
	"net"
	"strings"
	"sync"
	"time"
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

func (s *Server) MoveUserToLobby(user *User) {

}

func (s *Server) RunChat(ctx context.Context, c dbModel.Chat, rManager *repository.RepositoryManager) {
	s.Wg.Add(1)
	defer s.Wg.Done()
	rw := sync.RWMutex{}
	rw.Lock()
	chat := &Chat{Chat_name: c.Chat_name, Chat_id: c.Chat_id, Creator: &User{Username: c.Creator, Connection: nil}, IsOpen: c.IsOpen}
	chat.Connections = make(chan *User)
	chat.Alive = make(map[*User]string)
	chat.Dead_connections = make(chan *User)
	chat.Messages = make(chan string)
	s.Chats[chat] = []*User{}
	fmt.Printf("%s was launched during app start \n", chat.Chat_name)
	rw.Unlock()
	for chat.IsOpen {
		select {
		case user := <-chat.Connections:
			for u := range chat.Alive {
				u.Write(fmt.Sprintf("%s has connected\n\r", user.Username))
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
						_, err := rManager.MsgRepository.CreateMsg(ctx, model.Message{Username: user.Username, Chat_id: chat.Chat_id, Message: m, Time: time.Now()})
						if err != nil {
							log.Println("Error during logging message")
						}
						chat.Messages <- fmt.Sprintf("%s: %s \n\r", user.Username, m)
					}
				}
				chat.Dead_connections <- user
				s.Lobby.Connections <- user
			}()
		case msg := <-chat.Messages:
			for u := range chat.Alive {
				u.Write(msg)
			}
		case dconn := <-chat.Dead_connections:
			log.Printf("%v has disconnected\n", chat.Alive[dconn])
			for u := range chat.Alive {
				u.Write(fmt.Sprintf("%s has disconnected\n\r", chat.Alive[dconn]))
			}
			delete(chat.Alive, dconn)
		}
	}
}

func (s *Server) CloseChat(ctx context.Context, chatname string, rManager *repository.RepositoryManager) (err error) {
	if chatname == "Lobby" {
		return errors.New("can't close lobby")
	}
	for c := range s.Chats {
		if c.Chat_name == chatname {
			for _, u := range s.Chats[c] {
				u.Write("This Chat was closed by admin\n\r")
				c.Dead_connections <- u
				s.Lobby.Connections <- u
			}
			c.IsOpen = false
			delete(s.Chats, c)
			rManager.DeleteChat(ctx, c.Chat_id)
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
func (s *Server) CheckChatName(chat_name string) bool {
	flag := true
	for v := range s.Chats {
		if v.Chat_name == chat_name {
			flag = false
		}
	}
	return flag
}

func (s *Server) RunChats(ctx context.Context, rManager *repository.RepositoryManager) {
	chats, err := rManager.GetChats(ctx)
	if err != nil {
		fmt.Println("Failed to run chats from DB")
	}
	for _, chat := range chats {
		go s.RunChat(ctx, dbModel.Chat(chat), rManager)
	}
}

// creates a new chat only when chatname is available (checked before calling this function)
func (s *Server) NewChat(ctx context.Context, chat_name string, rManager *repository.RepositoryManager) error {
	defer s.Wg.Done()
	rw := sync.RWMutex{}
	rw.Lock()
	id, err := rManager.ChatRepository.GetId(ctx)
	chat := NewChat(chat_name, id)
	available_id, err := rManager.ChatRepository.CreateChat(ctx, model.Chat{Chat_name: chat.Chat_name, Chat_id: chat.Chat_id, Creator: chat.Creator.Username, IsOpen: chat.IsOpen})

	if err != nil {
		fmt.Println(err.Error())
	}
	chat.Id = available_id
	s.Available += 1
	rw.Unlock()
	fmt.Println(chat.Chat_name, chat.Chat_id, chat.Creator.Username, chat.IsOpen)

	s.Chats[chat] = []*User{}
	fmt.Printf("%s chat is running \n", chat_name)
	for chat.IsOpen {
		select {
		case user := <-chat.Connections:
			for u := range chat.Alive {
				u.Write(fmt.Sprintf("%s have connected\n\r", user.Username))
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
						_, err := rManager.MsgRepository.CreateMsg(ctx, model.Message{Username: user.Username, Chat_id: chat.Chat_id, Message: m, Time: time.Now()})
						if err != nil {
							log.Println("Error during logging message")
						}
						chat.Messages <- fmt.Sprintf("%s: %s \n\r", user.Username, m)
					}
				}
				chat.Dead_connections <- user
				s.Lobby.Connections <- user
			}()
		case msg := <-chat.Messages:
			for u := range chat.Alive {
				u.Write(msg)
			}
		case dconn := <-chat.Dead_connections:
			log.Printf("%v has disconnected\n", chat.Alive[dconn])
			for u := range chat.Alive {
				u.Write(fmt.Sprintf("%s has disconnected\n\r", chat.Alive[dconn]))
			}
			delete(chat.Alive, dconn)
		}
	}
	return nil
}

// lobby for all users, who just connetcted to chat
func (s *Server) Start_Lobby(ctx context.Context, address string, rManager *repository.RepositoryManager) {
	defer s.Wg.Done()
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err.Error())
		return
	}
	s.RunChats(ctx, rManager)
	//creating lobby
	lobby := NewChat("Lobby", 0)
	lobby.Id = 0
	s.Lobby = lobby
	s.Chats[lobby] = []*User{}

	//listening for new connections.
	s.Wg.Add(1)
	defer s.Wg.Done()
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
						user.Write("Enter your username: \n\r")
						username, err = rd.ReadString('\n')
						if err != nil {
							fmt.Println(err)
						}
						username = strings.TrimSpace(username)
						if !s.Check_username(username) {
							user.Write("This username is already taken\n\r")
						} else {
							break
						}
					}
					user.Username = username
					user.Write(fmt.Sprintf("Welcome to Lobby, %s \n\r", user.Username))
					s.Users = append(s.Users, user)
				}
				s.Chats[lobby] = append(s.Chats[lobby], user)
				lobby.Alive[user] = user.Username
				for {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					if m[0] == '/' {
						if s.parse_command(ctx, m, user, rManager) {
							break
						}
					} else {
						user.Write("No such command, available commands: /help \n\r")
					}
				}
				lobby.Dead_connections <- user
			}()
		case dconn := <-lobby.Dead_connections:
			delete(lobby.Alive, dconn)
		}
	}
}

func (s *Server) parse_command(ctx context.Context, command string, user *User, rManager *repository.RepositoryManager) bool {
	if strings.Contains(command, "/create") {
		chat_name := strings.Split(command, " ")[1]
		chat_name = strings.TrimSpace(chat_name)
		flag := false
		rw := sync.RWMutex{}
		rw.Lock()
		flag = s.CheckChatName(chat_name)
		if flag {
			s.Wg.Add(1)
			go s.NewChat(ctx, strings.TrimSpace(chat_name), rManager)
			user.Write("Room was successfully created \n\r")
		} else {
			user.Write("Room with this name is already exists \n\r")
		}
		rw.Unlock()
	} else if strings.Contains(command, "/chats") {
		for i := range s.Chats {
			if i.Chat_name != "Lobby" {
				user.Write(i.Chat_name + "\n" + "\r")
			}
		}
	} else if strings.Contains(command, "/connect") {
		room_name := strings.TrimSpace(strings.Split(command, " ")[1])
		if room_name == "Lobby" {
			return false
		}
		var room *Chat
		flag := false
		for i := range s.Chats {
			if i.Chat_name == room_name {
				flag = true
				room = i
			}
		}
		if !flag {
			user.Write("There is no such room \n\r")
			return false
		} else {
			s.Chats[room] = append(s.Chats[room], user)
			room.Connections <- user
			delete(s.Lobby.Alive, user)
			user.Write(fmt.Sprintf("You are connected to %s", room.Chat_name) + "\n\r")
			return true
		}
	} else if strings.Contains(command, "/leave") {

	} else if strings.Contains(command, "/help") {
		user.Write("Available commands are:\n\r" + "/chats\n\r" + "/create\n\r" + "/connect\n\r")
	} else {
		user.Write("No such command, available commands: /help \n\r")
		user.Write("Available commands are:\n\r" + "/chats\n\r" + "/create\n\r" + "/connect\n\r")
	}
	return false
}
