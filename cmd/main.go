package main

import (
	"bufio"
	"fmt"
	"go_chat/internal/chat"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	rooms map[int]chat.Room
}

func main() {
	go Runner(":8080")
	go Start_api()
}

// if this name is already taken user cant connect
func check_username(users map[net.Conn]string) bool {

	return false
}

// possibly will be used later
func parse_command(command string) {

}

func Start_api() {
	router := gin.Default()

	router.DELETE("/disconnect/:chat_id/:username", func(c *gin.Context) {
		chat_id := c.Param("chat_id")
		username := c.Param("username")

		c.JSON(http.StatusAccepted, gin.H{"message": "created"})
	})

	router.GET("/logs/:chat_id")

	router.POST("/create_chat")

	router.Run(":8080")
}

// Creates a room
func Runner(address string) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Println(err.Error())
		return
	}
	room := chat.Room{}
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
