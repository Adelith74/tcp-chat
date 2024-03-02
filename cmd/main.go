package main

import (
	"bufio"
	"fmt"
	"go_chat/internal/api"
	"go_chat/internal/chat"
	"log"
	"net"
	"strings"
	"sync"
)

var server = chat.NewServer()

var wg sync.WaitGroup

func main() {
	wg.Add(3)
	go Start_Lobby(":8000")
	go api.Start_api(&wg, &server)
	wg.Wait()
}

// if this name is already taken user cant connect
func check_username(users map[net.Conn]string) bool {
	return false
}

// possibly will be used later
func parse_command(command string, user chat.User) bool {
	if strings.Contains(command, "/create") {
		room_name := strings.Split(command, " ")[1]
		flag := false
		for i := range server.Rooms {
			if i.Room_name == room_name {
				flag = true
			}
		}
		if !flag {
			wg.Add(1)
			go Runner(room_name)
		} else {
			user.Connection.Write([]byte("Room with this name is already exists \n"))
		}
	} else if strings.Contains(command, "/chats") {
		for i := range server.Rooms {
			if i.Room_name != "Lobby" {
				user.Connection.Write([]byte(i.Room_name))
			}
		}
	} else if strings.Contains(command, "/connect") {
		room_name := strings.Split(command, " ")[1]
		if room_name == "Lobby" {
			return false
		}
		var room *chat.Room
		flag := false
		for i := range server.Rooms {
			if i.Room_name == room_name {
				flag = true
				room = i
			}
		}
		if !flag {
			user.Connection.Write([]byte("There is no such room"))
		} else {
			server.Rooms[room] = append(server.Rooms[room], &user)
			room.Connections <- user
			for i := range server.Rooms {
				if i.Room_name == "Lobby" {
					i.Dead_connections <- user.Connection
				}
			}
			user.Connection.Write([]byte(fmt.Sprintf("You are connected to %s", room.Room_name)))
		}
		return true
	} else if strings.Contains(command, "/leave") {

	} else if strings.Contains(command, "/help") {
		user.Connection.Write([]byte("Available commands are: \n" + "/chats \n" + "/create \n" + "/connect \n"))
	}

	return false
}

type Body struct {
	chat_id  int    `json:"chat_id"`
	username string `json:"username"`
}

// lobby for all users, who just connetcted to chat
func Start_Lobby(address string) {
	defer wg.Done()

	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//creating lobby
	lobby := chat.Room{}
	lobby.Room_name = "Lobby"
	lobby.Alive = make(map[net.Conn]string)
	lobby.Connections = make(chan chat.User)
	lobby.Dead_connections = make(chan net.Conn)
	lobby.Messages = make(chan string)
	server.Rooms[&lobby] = []*chat.User{}

	//listening for new connections
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println()
			}
			lobby.Connections <- chat.User{Connection: conn}
		}
	}()

	for {
		select {
		case user := <-lobby.Connections:
			//listening for user's messages
			go func() {
				user.Connection.Write([]byte("Enter your username:"))
				var username string
				rd := bufio.NewReader(user.Connection)
				username, err = rd.ReadString('\n')
				if err != nil {
					fmt.Println(err)
				}
				username = strings.TrimSpace(username)
				user.Username = username
				server.Rooms[&lobby] = append(server.Rooms[&lobby], &user)
				lobby.Alive[user.Connection] = username
				for {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					if m[0] == '/' {
						if parse_command(m, user) {
							break
						}
					} else {
						user.Connection.Write([]byte("No such command, available commands: /help \n"))
					}
				}
				lobby.Dead_connections <- user.Connection
			}()
		case dconn := <-lobby.Dead_connections:
			delete(lobby.Alive, dconn)
		}
	}
}

// Creates a room
func Runner(room_name string) {
	defer wg.Done()
	fmt.Printf("%s is running", room_name)
	room := chat.NewRoom(room_name)
	server.Rooms[&room] = []*chat.User{}

	for {
		select {
		case user := <-room.Connections:
			for conn := range room.Alive {
				conn.Write([]byte(fmt.Sprintf("%s have connected\n", user.Username)))
			}
			go func() {
				room.Alive[user.Connection] = user.Username
				rd := bufio.NewReader(user.Connection)
				for {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					if strings.Contains(m, "/leave") {
						break
					} else {
						room.Messages <- fmt.Sprintf("%s: %s", user.Username, m)
					}
				}
				room.Dead_connections <- user.Connection
				for i := range server.Rooms {
					if i.Room_name == "Lobby" {
						i.Connections <- user
					}
				}
			}()
		case msg := <-room.Messages:
			for conn := range room.Alive {
				conn.Write([]byte(msg))
			}
		case dconn := <-room.Dead_connections:
			log.Printf("%v has disconnected\n", room.Alive[dconn])
			for conn := range room.Alive {
				conn.Write([]byte(fmt.Sprintf("%s has disconnected\n", room.Alive[dconn])))
			}
			delete(room.Alive, dconn)
		}
	}
}
