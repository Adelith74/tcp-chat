package main

import (
	"go_chat/internal/api"
	"go_chat/internal/chat"
	"sync"
)

var server = chat.NewServer()

var wg sync.WaitGroup

func main() {
	wg.Add(2)
	go server.Start_Lobby(":8000", &wg)
	go api.Start_api(&wg, &server)
	wg.Wait()
}
