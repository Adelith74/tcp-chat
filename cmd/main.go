package main

import (
	"context"
	"go_chat/internal/api"
	"go_chat/internal/chat"
	"go_chat/internal/lib/db"
	"go_chat/internal/repository"
	"time"
)

var server = chat.NewServer()

func main() {

	timeout := time.Second * 10

	ctx := context.Background()

	withTimeout, _ := context.WithTimeout(ctx, timeout)

	database := db.New(withTimeout)

	manager := repository.NewRepositoryManager(database)

	server.Wg.Add(2)
	go server.Start_Lobby(":8000", manager)
	go api.Start_api(&server)
	server.Wg.Wait()
}
