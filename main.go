package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {	
	Runner()
}

func Runner(){
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Println(err.Error())
		return
	}
	alive := make(map[net.Conn]string)
	conns := make(chan net.Conn)
	dconns := make(chan net.Conn)
	msgs := make(chan string)
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
			for r.Scan(){
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

