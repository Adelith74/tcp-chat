package main

import (
	"context"
	"go_chat/internal/api"
	"go_chat/internal/chat"
	"go_chat/internal/lib/db"
	"go_chat/internal/repository"
	"go_chat/telegram"
	"sync"
	"time"
)

var server = chat.NewServer()

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	timeout := time.Second * 10

	ctx := context.Background()

	withTimeout, _ := context.WithTimeout(ctx, timeout)

	database := db.New(withTimeout)

	manager := repository.NewRepositoryManager(database)
	go func() {
		defer wg.Done()
		br := telegram.BotRunner{RM: &manager, CTX: withTimeout}
		br.Run()
	}()
	go func() {
		defer wg.Done()

		server.Wg.Add(2)

		go server.Start_Lobby(ctx, ":8000", &manager)

		go api.Start_api(ctx, &server, &manager)

		server.Wg.Wait()
	}()
	wg.Wait()
}
