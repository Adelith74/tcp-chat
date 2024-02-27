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

type Server struct {
	rooms map[*chat.Room][]*chat.User
}

var server = Server{rooms: make(map[*chat.Room][]*chat.User)}

var wg sync.WaitGroup

func main() {
	wg.Add(3)
	go Start_Lobby(":8000")
	go api.Start_api(&wg)
	wg.Wait()
}

// if this name is already taken user cant connect
func check_username(users map[net.Conn]string) bool {
	return false
}

// possibly will be used later
func parse_command(command string, user chat.User) bool{
	if strings.Contains(command, "/create") {
		room_name := strings.Split(command, " ")[1]
		flag := false
		for i, _ := range server.rooms {
			if i.Room_name == room_name {
				flag = true
			}
		}
		if !flag {
			wg.Add(1)
			go Runner(room_name)
		}
	} else if strings.Contains(command, "/chats") {
		for i, _ := range server.rooms {
			user.Connection.Write([]byte(i.Room_name))
		}
	} else if strings.Contains(command, "/connect") {
		room_name := strings.Split(command, " ")[1]
		var room *chat.Room
		flag := false
		for i, _ := range server.rooms {
			if i.Room_name == room_name {
				flag = true
				room = i
			}
		}
		if !flag{
			user.Connection.Write([]byte("There is no such room"))
		} else {
			server.rooms[room] = append(server.rooms[room], &user)
			room.Connections <- user.Connection
			for i, _ := range server.rooms {
				if i.Room_name == "Lobby"{
					i.Dead_connections <- user.Connection
				}
			}
			user.Connection.Write([]byte(fmt.Sprintf("You are connected to %s", room.Room_name)))
		}
		return true
	} else if strings.Contains(command, "/leave") {

	} else if strings.Contains(command, "/help"){
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
	lobby.Connections = make(chan net.Conn)
	lobby.Dead_connections = make(chan net.Conn)
	lobby.Messages = make(chan string)
	server.rooms[&lobby] = []*chat.User{}

	//listening for new connections
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println()
			}
			lobby.Connections <- conn
		}
	}()

	for {
		select {
		case conn := <-lobby.Connections:
			conn.Write([]byte("Enter your username:"))
			r := bufio.NewScanner(conn)
			username := ""
			for r.Scan() {
				username = r.Text()
				break
			}
			user := chat.User{Username: username, Connection: conn}
			server.rooms[&lobby] = append(server.rooms[&lobby], &user)
			lobby.Alive[conn] = username
			//listening for user's messages
			go func(user chat.User, username string) {
				rd := bufio.NewReader(user.Connection)
				for {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					if m[0] == '/' {
						if parse_command(m, user){
							break
						}
					} else {
						user.Connection.Write([]byte("No such command, available commands: /help \n"))
					}
				}
				lobby.Dead_connections <- conn
			}(user, username)
		case dconn := <-lobby.Dead_connections:
			log.Printf("%v has disconnected from lobby\n", lobby.Alive[dconn])
			delete(lobby.Alive, dconn)
		}
	}
}

// Creates a room
func Runner(room_name string) {
	defer wg.Done()
	fmt.Print("Server is running on:")
	room := chat.Room{}
	room.Room_name = room_name
	room.Alive = make(map[net.Conn]string)
	room.Connections = make(chan net.Conn)
	room.Dead_connections = make(chan net.Conn)
	room.Messages = make(chan string)
	server.rooms[&room] = []*chat.User{}

	for {
		select {
		case conn := <-room.Connections:
			conn.Write([]byte("Enter your username:"))
			r := bufio.NewScanner(conn)
			username := ""
			for r.Scan() {
				username = r.Text()
				break
			}
			room.Alive[conn] = username
			go func(conn net.Conn, username string) {
				rd := bufio.NewReader(conn)
				for {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					room.Messages <- fmt.Sprintf("%s: %v", username, m)
				}
				room.Dead_connections <- conn
			}(conn, username)
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
