package main

import (
	"go_chat/internal/api"
	"go_chat/internal/chat"
)

var server = chat.NewServer()

func main() {
	server.Wg.Add(2)
	go server.Start_Lobby(":8000")
	go api.Start_api(&server)
	server.Wg.Wait()
}
