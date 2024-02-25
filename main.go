package main

import (
	"bufio"
	"fmt"
	"go_chat/internal/chat"
	"log"
	"net"
	"go_chat/internal/api"
	"sync"

)

type Server struct {
	rooms map[int]chat.Room
}

var server = Server{make(map[int]chat.Room)}

var wg sync.WaitGroup

func main() {
	wg.Add(2)
	go Runner(":8000")
	go api.Start_api(&wg)
	wg.Wait()
}

// if this name is already taken user cant connect
func check_username(users map[net.Conn]string) bool {
	return false
}

// possibly will be used later
func parse_command(command string) {

}

type Body struct {
	chat_id  int    `json:"chat_id"`
	username string `json:"username"`
}


// Creates a room
func Runner(address string) {
	defer wg.Done()
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err.Error())
		return
	}
	alive := make(map[net.Conn]string)
	conns := make(chan net.Conn)
	dconns := make(chan net.Conn)
	msgs := make(chan string)

	//accepting new users
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println()
			}
			conns <- conn
		}
	}()

	for {
		select {
		case conn := <-conns:
			conn.Write([]byte("Enter your username:"))
			r := bufio.NewScanner(conn)
			username := ""
			for r.Scan() {
				username = r.Text()
				break
			}
			alive[conn] = username
			go func(conn net.Conn, username string) {
				rd := bufio.NewReader(conn)
				for {
					m, err := rd.ReadString('\n')
					if err != nil {
						break
					}
					msgs <- fmt.Sprintf("%s: %v", username, m)
				}
				dconns <- conn
			}(conn, username)
		case msg := <-msgs:
			for conn := range alive {
				conn.Write([]byte(msg))
			}
		case dconn := <-dconns:
			log.Printf("%v has disconnected\n", alive[dconn])
			for conn := range alive {
				conn.Write([]byte(fmt.Sprintf("%s has disconnected\n", alive[dconn])))
			}
			delete(alive, dconn)
		}
	}
}
