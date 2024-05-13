package main

import (
	"context"
	"fmt"
	"go_chat/internal/api"
	"go_chat/internal/chat"
	"go_chat/internal/lib/db"
	"go_chat/internal/repository"
	"go_chat/telegram"
	"time"
)

var server = chat.NewServer()

func main() {
	telegram.Run()
}

func proceed() {
	timeout := time.Second * 10

	ctx := context.Background()

	withTimeout, _ := context.WithTimeout(ctx, timeout)

	database := db.New(withTimeout)

	manager := repository.NewRepositoryManager(database)

	fmt.Println(manager)
	fmt.Println(manager.MsgRepository)
	server.Wg.Add(2)
	go server.Start_Lobby(ctx, ":8000", &manager)
	go api.Start_api(ctx, &server, &manager)
	server.Wg.Wait()
}
